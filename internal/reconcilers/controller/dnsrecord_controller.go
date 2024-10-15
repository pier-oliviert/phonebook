/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"slices"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

const kDNSRecordFinalizer string = "phonebook.se.quencer.io/finalizer"

// DNSRecordReconciler reconciles a DNSRecord object
type DNSRecordReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

// DNSRecordReconciler's job is to validate the DNSRecord as well as making sure that
// the finalizer for the record is in its proper state (present or removed)
//
// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords/finalizers,verbs=update
func (r *DNSRecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	record, err := r.GetRecord(ctx, req)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	log.FromContext(ctx).Info("Reconciling", "Record", record)
	if !record.DeletionTimestamp.IsZero() {
		condition := record.Status.Conditions.FindType(phonebook.ProviderCondition)
		if condition == nil || (condition.Status == konditions.ConditionError || condition.Status == konditions.ConditionTerminated) {
			if controllerutil.RemoveFinalizer(record, kDNSRecordFinalizer) {
				return ctrl.Result{Requeue: true}, r.Update(ctx, record)
			}
		}

		return ctrl.Result{}, err
	}

	if controllerutil.AddFinalizer(record, kDNSRecordFinalizer) {
		log.FromContext(ctx).Info("Finalizer didn't exists")
		return ctrl.Result{Requeue: true}, r.Update(ctx, record)
	}

	lock := konditions.NewLock(record, r.Client, phonebook.IntegrationCondition)
	if lock.Condition().Status == konditions.ConditionCompleted {
		return ctrl.Result{}, nil
	}

	if lock.Condition().Status == konditions.ConditionInitialized {

		lock.Execute(ctx, func(c konditions.Condition) (konditions.Condition, error) {
			var integrations phonebook.DNSIntegrationList
			var integration *phonebook.DNSIntegration

			if err := r.List(ctx, &integrations); err != nil {
				return c, err
			}

			for _, i := range integrations.Items {
				if record.Spec.Integration != nil {
					if *record.Spec.Integration != i.Name {
						continue
					}
				}

				if slices.Contains(i.Spec.Zones, record.Spec.Zone) {
					integration = &i
					goto FOUND_INTEGRATION
				}
			}

			goto NO_INTEGRATION

		FOUND_INTEGRATION:
			c.Status = konditions.ConditionCompleted
			c.Reason = fmt.Sprintf("Integration found: %s", integration.Name)
			return c, nil

		NO_INTEGRATION:
			c.Status = konditions.ConditionError
			c.Reason = fmt.Sprintf("No Integration matches the zone for this record: %s", record.Spec.Zone)
			return c, nil
		})
	}

	return ctrl.Result{}, nil
}

func (r *DNSRecordReconciler) GetRecord(ctx context.Context, req ctrl.Request) (*phonebook.DNSRecord, error) {
	var record phonebook.DNSRecord
	if err := r.Get(ctx, req.NamespacedName, &record); err != nil {
		return nil, fmt.Errorf("PB#0002: Couldn't retrieve the resource (%s) -- %w", req.NamespacedName, err)
	}

	return &record, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSRecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&phonebook.DNSRecord{}).
		Complete(r)
}
