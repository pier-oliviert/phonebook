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

	core "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/provider"
)

const kDNSRecordFinalizer string = "phonebook.se.quencer.io/finalizer"

// DNSRecordReconciler reconciles a DNSRecord object
type DNSRecordReconciler struct {
	Provider provider.Provider

	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=se.quencer.io,resources=dnsrecords/finalizers,verbs=update
func (r *DNSRecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var record phonebook.DNSRecord
	if err := r.Get(ctx, req.NamespacedName, &record); err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("PB#0002: Couldn't retrieve the DNSRecord (%s) -- %w", req.NamespacedName, err)
	}

	condition := record.Status.Conditions.FindOrInitializeFor(phonebook.ProviderCondition)
	record.Status.Conditions.SetCondition(condition)

	if condition.Status == konditions.ConditionInitialized {
		err := r.createRecord(ctx, &record)
		if err != nil {
		}

		return ctrl.Result{}, err
	}

	if !record.DeletionTimestamp.IsZero() || condition.Status == konditions.ConditionTerminating {
		err := r.deleteRecord(ctx, &record)
		if err != nil {
			r.Event(&record, core.EventTypeWarning, string(phonebook.ProviderCondition), err.Error())
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Status().Update(ctx, &record)
}

func (r *DNSRecordReconciler) createRecord(ctx context.Context, record *phonebook.DNSRecord) error {
	logger := log.FromContext(ctx)

	lock := konditions.NewLock(record, r.Client, phonebook.ProviderCondition)
	return lock.Execute(ctx, func(condition konditions.Condition) error {
		if controllerutil.AddFinalizer(record, kDNSRecordFinalizer) {
			if err := r.Update(ctx, record); err != nil {
				return err
			}
		}

		if err := r.Provider.Create(ctx, record); err != nil {
			// TODO:Eventually, this should move up the stack to the  main Reconciler method for this struct
			// as Konditionner should return errors from the Task as expected.
			logger.Error(err, "DNS Record could not be created", "Zone", record.Spec.Zone, "Subdomain", record.Spec.Name)
			r.Event(record, core.EventTypeWarning, string(phonebook.ProviderCondition), err.Error())

			return err
		}

		condition.Status = konditions.ConditionCreated
		condition.Reason = "DNS Record created according to the spec"
		record.Conditions().SetCondition(condition)

		return nil
	})
}

func (r *DNSRecordReconciler) deleteRecord(ctx context.Context, record *phonebook.DNSRecord) error {
	logger := log.FromContext(ctx)

	lock := konditions.NewLock(record, r.Client, phonebook.ProviderCondition)

	if lock.Condition().Status == konditions.ConditionTerminated || lock.Condition().Status == konditions.ConditionError {
		if controllerutil.ContainsFinalizer(record, kDNSRecordFinalizer) {
			controllerutil.RemoveFinalizer(record, kDNSRecordFinalizer)
			return r.Update(ctx, record)
		}
	}

	return lock.Execute(ctx, func(condition konditions.Condition) error {

		if err := r.Provider.Delete(ctx, record); err != nil {
			logger.Error(err, "PB#0003: Could not delete the record upstream")

			// The lock is in charge of updating the condition for an error
			return err
		}

		condition.Status = konditions.ConditionTerminated
		condition.Reason = "Record deleted"
		record.Conditions().SetCondition(condition)

		return r.Status().Update(ctx, record)
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSRecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&phonebook.DNSRecord{}).
		Complete(r)
}
