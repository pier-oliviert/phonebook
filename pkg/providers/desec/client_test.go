package desec_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

const (
	kDesecToken           = "DESEC_TOKEN"
	kDesecMetaRecordIDKey = "desec.io/record_id"
)

type DesecAPI interface {
	CreateDNSRecord(ctx context.Context, domain, name, recordType, content string, ttl int) error
	DeleteDNSRecord(ctx context.Context, domain, name, recordType, content string) error
}

type mockAPI struct {
	CreateDNSRecordFunc func(ctx context.Context, domain, name, recordType, content string, ttl int) error
	DeleteDNSRecordFunc func(ctx context.Context, domain, name, recordType, content string) error
}

type desecClient struct {
	integration string
	API         DesecAPI
}

func (m *mockAPI) CreateDNSRecord(ctx context.Context, domain, name, recordType, content string, ttl int) error {
	if m.CreateDNSRecordFunc != nil {
		return m.CreateDNSRecordFunc(ctx, domain, name, recordType, content, ttl)
	}
	return nil
}

func (m *mockAPI) DeleteDNSRecord(ctx context.Context, domain, name, recordType, content string) error {
	if m.DeleteDNSRecordFunc != nil {
		return m.DeleteDNSRecordFunc(ctx, domain, name, recordType, content)
	}
	return nil
}

func (c *desecClient) CreateDNSRecord(ctx context.Context, record *phonebook.DNSRecord) error {
	domain := record.Spec.Zone
	name := record.Spec.Name
	recordType := record.Spec.RecordType
	content := record.Spec.Targets[0] // Assuming the first target is the content
	ttl := 3600                       // You might want to make this configurable

	err := c.API.CreateDNSRecord(ctx, domain, name, recordType, content, ttl)
	if err != nil {
		return err
	}

	// Set the RemoteID in the record's status
	// For deSEC, we don't have a specific ID, so we'll use a combination of name and type
	remoteID := fmt.Sprintf("%s-%s", name, recordType)
	record.Status.RemoteInfo = map[string]phonebook.IntegrationInfo{
		c.integration: {
			"recordID": remoteID,
		},
	}

	return nil
}

func (c *desecClient) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	if record.Status.RemoteInfo[c.integration] == nil {
		return nil // Nothing to delete if RemoteID is not set
	}

	domain := record.Spec.Zone
	name := record.Spec.Name
	recordType := record.Spec.RecordType
	content := record.Spec.Targets[0] // Assuming the first target is the content

	return c.API.DeleteDNSRecord(ctx, domain, name, recordType, content)
}

func NewClient(ctx context.Context) (*desecClient, error) {
	token := os.Getenv(kDesecToken)
	if token == "" {
		return nil, fmt.Errorf("PB-DESEC-#0001: API Token not found -- Please set the %s environment variable", kDesecToken)
	}

	// In a real implementation, you'd create an actual API client here
	// For this mock, we'll just return a client with a mock API
	return &desecClient{
		API: &mockAPI{},
	}, nil
}

// Test functions

func TestNewClient(t *testing.T) {
	// Test missing API Token
	_, err := NewClient(context.TODO())
	if err == nil || err.Error() != fmt.Sprintf("PB-DESEC-#0001: API Token not found -- Please set the %s environment variable", kDesecToken) {
		t.Error("Expected error for missing API Token")
	}

	// Set API token environment variable
	os.Setenv(kDesecToken, "SomeValue")
	defer os.Unsetenv(kDesecToken)

	// Test successful client creation
	client, err := NewClient(context.TODO())
	if err != nil {
		t.Errorf("Expected successful client creation, got error: %v", err)
	}

	if client == nil {
		t.Error("Expected a valid client, got nil")
	}
}

func TestDNSCreation(t *testing.T) {
	// Set API token environment variable
	os.Setenv(kDesecToken, "SomeValue")
	defer os.Unsetenv(kDesecToken)

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

	// Mock the CreateDNSRecord function
	mockAPI := &mockAPI{
		CreateDNSRecordFunc: func(ctx context.Context, domain, name, recordType, content string, ttl int) error {
			// Perform some checks
			if domain != "mydomain.com" || name != "subdomain" || recordType != "A" || content != "127.0.0.1" {
				return fmt.Errorf("unexpected values in CreateDNSRecord")
			}
			return nil
		},
	}

	// Create desecClient with mock API
	c := desecClient{
		integration: "mytest",
		API:         mockAPI,
	}

	// Test DNS record creation
	err := c.CreateDNSRecord(context.TODO(), &record)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check if RemoteID was set correctly
	expectedRemoteID := "subdomain-A"
	if record.Status.RemoteInfo[c.integration]["recordID"] == "" || record.Status.RemoteInfo[c.integration]["recordID"] != expectedRemoteID {
		t.Errorf("Expected RemoteID to be '%s', but got: %v", expectedRemoteID, record.Status.RemoteInfo[c.integration]["recordID"])
	}
}

func TestDNSDeletion(t *testing.T) {
	// Set API token environment variable
	os.Setenv(kDesecToken, "SomeValue")
	defer os.Unsetenv(kDesecToken)

	// Mock the DeleteDNSRecord function
	mockAPI := &mockAPI{
		DeleteDNSRecordFunc: func(ctx context.Context, domain, name, recordType, content string) error {
			// Perform some checks
			if domain != "mydomain.com" || name != "subdomain" || recordType != "A" || content != "127.0.0.1" {
				return fmt.Errorf("unexpected values in DeleteDNSRecord")
			}
			return nil
		},
	}

	// Create desecClient with mock API
	c := desecClient{
		integration: "Another-test",
		API:         mockAPI,
	}
	//
	// Prepare test DNS record
	remoteID := "subdomain-A"
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "mydomain.com",
			Name:       "subdomain",
			RecordType: "A",
			Targets:    []string{"127.0.0.1"},
		},
		Status: phonebook.DNSRecordStatus{
			RemoteInfo: map[string]phonebook.IntegrationInfo{
				c.integration: {
					"recordID": remoteID,
				},
			},
		},
	}

	// Test DNS record deletion
	err := c.Delete(context.TODO(), &record)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Test deleting a record with no RemoteID
	recordWithNoRemoteID := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "mydomain.com",
			Name:       "another-subdomain",
			RecordType: "A",
			Targets:    []string{"192.168.0.1"},
		},
		Status: phonebook.DNSRecordStatus{},
	}

	err = c.Delete(context.TODO(), &recordWithNoRemoteID)
	if err != nil {
		t.Errorf("Expected no error when deleting record with no RemoteID, but got: %v", err)
	}
}
