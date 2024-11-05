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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

type TestProvider struct{}

func (tp TestProvider) Create(context.Context, *phonebook.DNSRecord) error {
	return nil
}

func (tp TestProvider) Delete(context.Context, *phonebook.DNSRecord) error {
	return nil
}

var _ = Describe("DNSRecord Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		dnsrecord := &phonebook.DNSRecord{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind DNSRecord")
			err := k8sClient.Get(ctx, typeNamespacedName, dnsrecord)
			if err != nil && errors.IsNotFound(err) {
				resource := &phonebook.DNSRecord{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: phonebook.DNSRecordSpec{
						Zone:       "example.com",
						RecordType: "A",
						Name:       "subdomain",
						Targets:    []string{"127.0.0.1"},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &phonebook.DNSRecord{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance DNSRecord")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DNSRecordReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})

	Context("AllProvidersMatchesOneOf", func() {
		It("returns false if no conditions are present", func() {
			r := &DNSRecordReconciler{}
			Expect(r.AllProvidersMatchesOneOf(konditions.Conditions{})).To(Equal(true))
		})

		It("returns false if no condition status present", func() {
			r := &DNSRecordReconciler{}
			conditions := konditions.Conditions{{
				Type:   konditions.ConditionType("provider://test"),
				Status: konditions.ConditionError,
			}}

			Expect(r.AllProvidersMatchesOneOf(conditions)).To(Equal(false))
		})

		It("returns true if no condition for providers present", func() {
			r := &DNSRecordReconciler{}

			conditions := konditions.Conditions{{
				Type:   konditions.ConditionType("test"),
				Status: konditions.ConditionError,
			}}

			Expect(r.AllProvidersMatchesOneOf(conditions)).To(Equal(true))
		})

		It("returns true if all condition for providers matches one of the condition status", func() {
			r := &DNSRecordReconciler{}

			conditions := konditions.Conditions{
				{
					Type:   konditions.ConditionType("not-a-provider"),
					Status: konditions.ConditionCreated,
				},
				{
					Type:   konditions.ConditionType("provider://test"),
					Status: konditions.ConditionError,
				},
				{
					Type:   konditions.ConditionType("provider://test-1"),
					Status: konditions.ConditionTerminated,
				},
				{
					Type:   konditions.ConditionType("provider://test-2"),
					Status: konditions.ConditionTerminated,
				},
			}

			Expect(r.AllProvidersMatchesOneOf(
				conditions,
				konditions.ConditionTerminated,
				konditions.ConditionError,
				konditions.ConditionCompleted),
			).To(Equal(true))
		})

		It("returns false if one condition for providers matches one of the condition status", func() {
			r := &DNSRecordReconciler{}

			conditions := konditions.Conditions{
				{
					Type:   konditions.ConditionType("provider://test"),
					Status: konditions.ConditionLocked,
				},
				{
					Type:   konditions.ConditionType("provider://test-1"),
					Status: konditions.ConditionTerminated,
				},
				{
					Type:   konditions.ConditionType("provider://test-2"),
					Status: konditions.ConditionTerminated,
				},
			}

			Expect(r.AllProvidersMatchesOneOf(
				conditions,
				konditions.ConditionTerminated,
				konditions.ConditionError,
				konditions.ConditionCompleted),
			).To(Equal(false))
		})
	})
})
