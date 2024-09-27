package solver

import (
	"context"
	"reflect"
	"testing"

	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestPresent(t *testing.T) {
	var result *phonebook.DNSRecord
	client := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {

		var ok bool
		result, ok = obj.(*phonebook.DNSRecord)
		if !ok {
			t.Error("Expected client.Object to be a DNSRecord")
		}

		return nil
	},
	}).Build()

	solver := Solver{
		Client: client,
	}

	err := solver.Present(&whapi.ChallengeRequest{
		UID:               types.UID("testerino"),
		Key:               "test-1234",
		ResolvedFQDN:      "my.domain.test",
		ResourceNamespace: "phonebook-test",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(result.Spec.Targets) != 1 || result.Spec.Targets[0] != "test-1234" {
		t.Error("Target doesn't match the challenge request", "Target", result.Spec.Targets)
	}

	if result.Namespace != "phonebook-test" {
		t.Error("Namespace doesn't match the challenge request", "Namespace", result.Namespace)
	}

	if result.Spec.Name != "my.domain.test" {
		t.Error("ResolvedFQDN doesn't match the challenge request", "ResolvedFQDN", result.Spec.Name)
	}

	if result.Spec.RecordType != "TXT" {
		t.Error("RecordType isn't TXT", "RecordType", result.Spec.RecordType)
	}

	if val, ok := result.Labels[kChallengeLabel]; !ok || val != kChallengeKey {
		t.Error("Label for DNS-01 not properly set", "Labels", result.Labels)
	}
}

func TestCleanUpRecordWithMultipleTargets(t *testing.T) {
	records := &phonebook.DNSRecordList{
		Items: []phonebook.DNSRecord{{
			Spec: phonebook.DNSRecordSpec{
				Targets: []string{"one", "two"},
			},
		}},
	}

	client := fake.NewClientBuilder().
		WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				// Store the value of `records` in `list`
				val := reflect.ValueOf(list).Elem()
				val.Set(reflect.ValueOf(records).Elem())

				return nil
			},
			Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				return nil
			},
		}).Build()

	solver := Solver{
		Client: client,
	}

	err := solver.CleanUp(&whapi.ChallengeRequest{
		UID:               types.UID("testerino"),
		Key:               "test-1234",
		ResolvedFQDN:      "my.domain.test",
		ResourceNamespace: "phonebook-test",
	})

	if err == nil {
		t.Error("Expected an error from the CleanUp method due to multiple targets", "List", records)
	}
}

func TestCleanUpRecordValidRecord(t *testing.T) {
	deletedCount := 0
	var lastDeletedRecord *phonebook.DNSRecord

	records := &phonebook.DNSRecordList{
		Items: []phonebook.DNSRecord{{
			Spec: phonebook.DNSRecordSpec{
				Targets: []string{"test-1234"},
			},
		}},
	}

	client := fake.NewClientBuilder().
		WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, client client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				// Store the value of `records` in `list`
				val := reflect.ValueOf(list).Elem()
				val.Set(reflect.ValueOf(records).Elem())

				return nil
			},
			Delete: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				var ok bool
				lastDeletedRecord, ok = obj.(*phonebook.DNSRecord)
				if !ok {
					t.Error("Expected client.Object to be a DNSRecord")
				}

				deletedCount = deletedCount + 1
				return nil
			},
		}).Build()

	solver := Solver{
		Client: client,
	}

	err := solver.CleanUp(&whapi.ChallengeRequest{
		UID:               types.UID("testerino"),
		Key:               "test-1234",
		ResolvedFQDN:      "my.domain.test",
		ResourceNamespace: "phonebook-test",
	})

	if deletedCount != 1 {
		t.Error("Expected only 1 record to be deleted")
	}

	if err != nil {
		t.Fatal(err)
	}

	// Only comparing specs as other fields are dynamic and irrelevant to this test
	if !reflect.DeepEqual(lastDeletedRecord.Spec, records.Items[0].Spec) {
		t.Errorf("Unexpected deleted record: %#v", lastDeletedRecord)
	}
}
