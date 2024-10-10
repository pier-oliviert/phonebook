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

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/api/v1alpha1/integrations"
	tasks "github.com/pier-oliviert/phonebook/internal/reconcilers/controller/tasks/integrations"
)

// DNSProviderReconciler reconciles a DNSProvider object
type DNSIntegrationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder
}

// +kubebuilder:rbac:groups=se.quencer.io.se.quencer.io,resources=dnsintegrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=se.quencer.io.se.quencer.io,resources=dnsintegrations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=se.quencer.io.se.quencer.io,resources=dnsintegrations/finalizers,verbs=update
func (r *DNSIntegrationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	integration, err := r.GetIntegration(ctx, req)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	if !integration.DeletionTimestamp.IsZero() {
		condition := integration.Status.Conditions.FindOrInitializeFor(integrations.DeploymentCondition)
		if condition.Status == konditions.ConditionTerminated {
			if controllerutil.RemoveFinalizer(integration, integrations.DeploymentFinalizer) {
				return ctrl.Result{Requeue: true}, r.Update(ctx, integration)
			}
		}

		condition.Status = konditions.ConditionTerminated
		condition.Reason = "Tearing down the integration"
		integration.Status.Conditions.SetCondition(condition)

		return ctrl.Result{}, r.Status().Update(ctx, integration)
	}

	if controllerutil.AddFinalizer(integration, integrations.DeploymentFinalizer) {
		return ctrl.Result{Requeue: true}, r.Update(ctx, integration)
	}

	lock := konditions.NewLock(integration, r.Client, integrations.DeploymentCondition)
	if lock.Condition().Status == konditions.ConditionInitialized {
		if err := lock.Execute(ctx, tasks.DeploymentTask(ctx, r.Client, integration)); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DNSIntegrationReconciler) GetIntegration(ctx context.Context, req ctrl.Request) (*phonebook.DNSIntegration, error) {
	var integration phonebook.DNSIntegration
	if err := r.Get(ctx, req.NamespacedName, &integration); err != nil {
		return nil, fmt.Errorf("PB#0002: Couldn't retrieve the resource (%s) -- %w", req.NamespacedName, err)
	}

	return &integration, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSIntegrationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&phonebook.DNSIntegration{}).
		Complete(r)
}
