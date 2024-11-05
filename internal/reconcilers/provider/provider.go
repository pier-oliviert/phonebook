package provider

import (
	"context"
	"errors"
	"fmt"
	"slices"

	core "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/providers"
)

var ErrProviderDidNotSetCondition = errors.New("PB-#0100: Provider didn't set a condition status upon returning from function")

// ProviderReconciler handles all incoming reconciliation requests
// for DNSRecord that matches the Integration as defined. It ignores
// any DNSRecord that doesn't fit the requirement set for the
// Provider (ie. zone defined)
type ProviderReconciler struct {
	Store       *providers.ProviderStore
	Integration string

	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

// Reconciliation runs for the Provider.
func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	record, err := r.GetRecord(ctx, req)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		return result, err
	}

	su := &stageUpdater{
		record: record,
	}

	conditionType := konditions.ConditionType(r.Integration)
	condition := record.Status.Conditions.FindType(conditionType)
	if condition == nil || condition.Status == konditions.ConditionError || condition.Status == konditions.ConditionCompleted {
		return result, nil
	}

	if !slices.Contains(r.Store.Provider().Zones(), record.Spec.Zone) {
		// This Provider doesn't have authority over the zone specified by the
		// record.
		return result, nil
	}

	lock := konditions.NewLock(record, r.Client, conditionType)
	if lock.Condition().Status == konditions.ConditionError {
		return result, nil
	}

	switch {
	case !record.DeletionTimestamp.IsZero():
		err = lock.Execute(ctx, func(c konditions.Condition) (konditions.Condition, error) {
			if err = r.Store.Provider().Delete(ctx, *record.DeepCopy(), su); err != nil {
				return c, err
			}

			if su.status == nil {
				return c, ErrProviderDidNotSetCondition
			}

			c.Status = *su.status

			if su.reason != nil {
				c.Reason = *su.reason
			}

			if su.info != nil {
				record.Status.RemoteInfo[r.Integration] = su.info
			}

			return c, nil
		})

	case lock.Condition().Status == konditions.ConditionInitialized:
		// Execute will update the DNSRecord's Status subresource before
		// it returns. Unless there is an error while updating, any field set on the status
		// will be persisted by the end of this method.
		err = lock.Execute(ctx, func(c konditions.Condition) (konditions.Condition, error) {
			if err = r.Store.Provider().Create(ctx, *record.DeepCopy(), su); err != nil {
				return c, err
			}
			if su.status == nil {
				return c, ErrProviderDidNotSetCondition
			}
			c.Status = *su.status

			if su.reason != nil {
				c.Reason = *su.reason
			}

			if su.info != nil {
				if record.Status.RemoteInfo == nil {
					record.Status.RemoteInfo = make(map[string]phonebook.IntegrationInfo)
				}

				record.Status.RemoteInfo[r.Integration] = su.info
			}

			return c, nil
		})
	}

	if k8sErrors.IsConflict(err) {
		log.FromContext(ctx).Info("Conflict error while updating the DNSRecord, retrying.", "Error", err)
		result.Requeue = true
		return result, nil
	}

	if err != nil {
		r.Event(
			record,
			core.EventTypeWarning,
			string(lock.Condition().Status),
			err.Error())
	}

	return result, err
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&phonebook.DNSRecord{}).
		Complete(r)
}

func (r *ProviderReconciler) GetRecord(ctx context.Context, req ctrl.Request) (*phonebook.DNSRecord, error) {
	var record phonebook.DNSRecord
	if err := r.Get(ctx, req.NamespacedName, &record); err != nil {
		return nil, fmt.Errorf("PB#0002: Couldn't retrieve the resource (%s) -- %w", req.NamespacedName, err)
	}

	return &record, nil
}

type stageUpdater struct {
	record *phonebook.DNSRecord
	status *konditions.ConditionStatus
	reason *string
	info   phonebook.IntegrationInfo
}

func (su *stageUpdater) StageCondition(status konditions.ConditionStatus, reason string) {
	su.status = &status
	su.reason = &reason
}

func (su *stageUpdater) StageRemoteInfo(info phonebook.IntegrationInfo) {
	su.info = info
}
