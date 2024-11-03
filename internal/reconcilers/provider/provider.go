package provider

import (
	"context"
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

// ProviderReconciler handles all incoming reconciliation requests
// for DNSRecord that matches the Integration as defined. It ignores
// any DNSRecord that doesn't fit the requirement set for the
// Provider (ie. zone defined)
type ProviderReconciler struct {
	Store *providers.ProviderStore

	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	record, err := r.GetRecord(ctx, req)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		return result, err
	}

	condition := record.Status.Conditions.FindType(phonebook.IntegrationCondition)
	if condition == nil || condition.Status != konditions.ConditionCompleted {
		return result, nil
	}

	if !slices.Contains(r.Store.Provider().Zones(), record.Spec.Zone) {
		// This Provider doesn't have authority over the zone specified by the
		// record.
		return result, nil
	}

	lock := konditions.NewLock(record, r.Client, phonebook.ProviderCondition)
	if lock.Condition().Status == konditions.ConditionError {
		return result, nil
	}

	if !record.DeletionTimestamp.IsZero() {
		err = lock.Execute(ctx, func(c konditions.Condition) (konditions.Condition, error) {
			if err = r.Store.Provider().Delete(ctx, record); err != nil {
				return c, err
			}

			c.Status = konditions.ConditionTerminated
			c.Reason = "DNS Record Deleted"
			return c, nil
		})
	}

	if lock.Condition().Status == konditions.ConditionInitialized {
		err = lock.Execute(ctx, func(c konditions.Condition) (konditions.Condition, error) {
			if err = r.Store.Provider().Create(ctx, record); err != nil {
				return c, err
			}
			c.Status = konditions.ConditionCreated
			c.Reason = "DNS Record Created"
			return c, nil
		})
	}

	if k8sErrors.IsConflict(err) {
		log.FromContext(ctx).Info("Conflict error while updating the DNSRecord, retrying.", "Error", err)
		result.Requeue = true
		return result, nil
	}

	if err != nil {
		r.EventRecorder.Event(
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
