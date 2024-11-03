package cloudflare

import (
	"context"
	"os"
	"strings"
	"testing"

	client "github.com/cloudflare/cloudflare-go"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

type CloudflareAPI interface {
	CreateDNSRecord(ctx context.Context, zoneID string, params client.CreateDNSRecordParams) (client.DNSRecordResponse, error)
	DeleteDNSRecord(ctx context.Context, zoneID string, recordID string) error
}

type mockAPI struct {
	CreateDNSRecordFunc func(ctx context.Context, zoneID string, params client.CreateDNSRecordParams) (client.DNSRecordResponse, error)
	DeleteDNSRecordFunc func(ctx context.Context, zoneID string, recordID string) error
}

type cft struct {
	integration string
	zoneID      string
	API         CloudflareAPI
}

func (m *mockAPI) CreateDNSRecord(ctx context.Context, zoneID string, params client.CreateDNSRecordParams) (client.DNSRecordResponse, error) {
	if m.CreateDNSRecordFunc != nil {
		return m.CreateDNSRecordFunc(ctx, zoneID, params)
	}
	return client.DNSRecordResponse{}, nil
}

func (m *mockAPI) DeleteDNSRecord(ctx context.Context, zoneID string, recordID string) error {
	if m.DeleteDNSRecordFunc != nil {
		return m.DeleteDNSRecordFunc(ctx, zoneID, recordID)
	}
	return nil
}

func (c *cft) CreateDNSRecord(ctx context.Context, record *phonebook.DNSRecord) error {
	params := client.CreateDNSRecordParams{
		Type:    record.Spec.RecordType,
		Name:    record.Spec.Name,
		Content: record.Spec.Targets[0], // Assuming the first target is the content
		TTL:     1,                      // You might want to make this configurable
	}

	response, err := c.API.CreateDNSRecord(ctx, c.zoneID, params)
	if err != nil {
		return err
	}

	// Set the RemoteID in the record's status
	remoteID := response.Result.ID
	record.Status.RemoteInfo = map[string]phonebook.IntegrationInfo{
		c.integration: {
			"recordID": remoteID,
		},
	}

	return nil
}

// Add Delete method to cft struct (you'll need to implement this)
func (c *cft) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	if record.Status.RemoteInfo[c.integration] == nil {
		return nil // Nothing to delete if RemoteID is not set
	}
	return c.API.DeleteDNSRecord(ctx, c.zoneID, record.Status.RemoteInfo[c.integration]["recordID"])
}

// Test for the NewClient function
func TestNewClient(t *testing.T) {
	// Test missing API Key
	_, err := NewClient(context.TODO())
	if err == nil || !strings.HasPrefix(err.Error(), "PB-CF-#0001: API Key not found --") {
		t.Error("Expected error for missing API Key")
	}

	// Set API token environment variable
	os.Setenv("CF_API_TOKEN", "Some Value")

	// Test missing Zone ID
	_, err = NewClient(context.TODO())
	if err == nil || !strings.HasPrefix(err.Error(), "PB-CF-#0002: Zone ID not found --") {
		t.Error("Expected error for missing Zone ID")
	}

	// Set Zone ID environment variable
	os.Setenv("CF_ZONE_ID", "Some Zone ID")

	// Test successful client creation
	client, err := NewClient(context.TODO())
	if err != nil {
		t.Errorf("Expected successful client creation, got error: %v", err)
	}

	if client == nil {
		t.Error("Expected a valid client, got nil")
	}
}

// Test for the DNS record creation function
func TestDNSCreation(t *testing.T) {
	// Mock environment variables
	os.Setenv("CF_API_TOKEN", "Some Value")
	os.Setenv("CF_ZONE_ID", "Some Zone ID")

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
		CreateDNSRecordFunc: func(ctx context.Context, zoneID string, params client.CreateDNSRecordParams) (client.DNSRecordResponse, error) {
			// Return a mock response with a fake record ID
			return client.DNSRecordResponse{
				Result: client.DNSRecord{ID: "fake-record-id"},
			}, nil
		},
	}

	// Create cf struct with mock API
	c := cft{
		integration: "cloudflare-test",
		zoneID:      "Some Zone ID",
		API:         mockAPI, // Use the mock API here
	}

	// Test DNS record creation
	err := c.CreateDNSRecord(context.TODO(), &record)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// Check if RemoteID was set correctly
	if record.Status.RemoteInfo[c.integration]["recordID"] == "" || record.Status.RemoteInfo[c.integration]["recordID"] != "fake-record-id" {
		t.Errorf("Expected RemoteID to be 'fake-record-id', but got: %v", record.Status.RemoteInfo[c.integration]["recordID"])
	}
}

// Test for the DNS record deletion function
func TestDNSDeletion(t *testing.T) {
	// Mock environment variables
	os.Setenv("CF_API_TOKEN", "Some Value")
	os.Setenv("CF_ZONE_ID", "Some Zone ID")

	// Mock the DeleteDNSRecord function
	mockAPI := &mockAPI{
		DeleteDNSRecordFunc: func(ctx context.Context, zoneID string, recordID string) error {
			// Check if the correct record ID is being deleted
			if recordID != "fake-record-id" {
				t.Errorf("Expected record ID 'fake-record-id', got '%s'", recordID)
			}
			return nil
		},
	}

	// Create cft struct with mock API
	c := cft{
		integration: "my-test",
		zoneID:      "Some Zone ID",
		API:         mockAPI,
	}
	//
	// Prepare test DNS record
	recordID := "fake-record-id"
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
					"recordID": recordID,
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
