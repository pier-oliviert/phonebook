package gcore

import (
	"context"
	"fmt"
	"os"
	"testing"

	gdns "github.com/G-Core/gcore-dns-sdk-go"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/mocks"
	"github.com/stretchr/testify/mock"
)

type MockRecordSetsClient struct {
	mock.Mock
	recordCreated rCreated
	recordDeleted rDeleted
}

type rCreated struct {
	zone       string
	name       string
	recordType string
	ttl        int
}

type rDeleted struct {
	zone       string
	name       string
	recordType string
}

func (m *MockRecordSetsClient) AddZoneRRSet(_ context.Context, zone, name, rType string, values []gdns.ResourceRecord, ttl int, _ ...gdns.AddZoneOpt) error {
	m.recordCreated = rCreated{
		zone:       zone,
		name:       name,
		recordType: rType,
		ttl:        ttl,
	}

	return nil
}

func (m *MockRecordSetsClient) DeleteRRSet(_ context.Context, zone, name, rType string) error {
	m.recordDeleted = rDeleted{
		zone:       zone,
		name:       name,
		recordType: rType,
	}

	return nil
}

func TestNewClient(t *testing.T) {
	os.Setenv("GCORE_API_TOKEN", "Mytoken")
	_, err := NewClient(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestCreation(t *testing.T) {
	os.Setenv("GCORE_API_TOKEN", "Mytoken")
	client, err := NewClient(context.Background())
	if err != nil {
		t.Error(err)
	}

	mock := &MockRecordSetsClient{}
	client.api = mock

	// Prepare test DNS record
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "mydomain.com",
			Name:       "subdomain",
			RecordType: "A",
			Targets:    []string{"127.0.0.1"},
		},
		Status: phonebook.DNSRecordStatus{},
	}

	err = client.Create(context.Background(), record, &mocks.Updater{})
	if err != nil {
		t.Error(err)
	}

	if mock.recordCreated.zone != record.Spec.Zone {
		t.Errorf("Record did not match zone %s: %s", mock.recordCreated.zone, record.Spec.Zone)
	}

	if mock.recordCreated.name != fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone) {
		t.Errorf("Record did not match name %s: %s.%s", mock.recordCreated.name, record.Spec.Zone, record.Spec.Zone)
	}

	if mock.recordCreated.recordType != record.Spec.RecordType {
		t.Errorf("Record did not match record type %s: %s", mock.recordCreated.recordType, record.Spec.RecordType)
	}

	if mock.recordCreated.ttl != int(DefaultTTL) {
		t.Errorf("Record did not match default TTL %d: %d", mock.recordCreated.ttl, DefaultTTL)
	}
}

func TestCreationTTLSet(t *testing.T) {
	os.Setenv("GCORE_API_TOKEN", "Mytoken")
	client, err := NewClient(context.Background())
	if err != nil {
		t.Error(err)
	}

	mock := &MockRecordSetsClient{}
	client.api = mock

	// Prepare test DNS record
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "mydomain.com",
			Name:       "subdomain",
			RecordType: "A",
			Targets:    []string{"127.0.0.1"},
			TTL:        new(int64),
		},
		Status: phonebook.DNSRecordStatus{},
	}
	*record.Spec.TTL = 900

	err = client.Create(context.Background(), record, &mocks.Updater{})
	if err != nil {
		t.Error(err)
	}

	if mock.recordCreated.ttl != 900 {
		t.Errorf("Record did not match default TTL %d: %d", mock.recordCreated.ttl, 900)
	}
}
