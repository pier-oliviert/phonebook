package azure

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/mocks"
)

// MockRecordSetsClient is a mock for the Azure RecordSetsClient
type MockRecordSetsClient struct {
	mock.Mock
}

func (m *MockRecordSetsClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, zoneName string, relativeRecordSetName string, recordType armdns.RecordType, parameters armdns.RecordSet, options *armdns.RecordSetsClientCreateOrUpdateOptions) (armdns.RecordSetsClientCreateOrUpdateResponse, error) {
	args := m.Called(ctx, resourceGroupName, zoneName, relativeRecordSetName, recordType, parameters, options)
	return args.Get(0).(armdns.RecordSetsClientCreateOrUpdateResponse), args.Error(1)
}

func (m *MockRecordSetsClient) Delete(ctx context.Context, resourceGroupName string, zoneName string, relativeRecordSetName string, recordType armdns.RecordType, options *armdns.RecordSetsClientDeleteOptions) (armdns.RecordSetsClientDeleteResponse, error) {
	args := m.Called(ctx, resourceGroupName, zoneName, relativeRecordSetName, recordType, options)
	return args.Get(0).(armdns.RecordSetsClientDeleteResponse), args.Error(1)
}

func TestNewClient(t *testing.T) {
	// Mock environment variables
	os.Setenv(kAzureSubscriptionID, "SomeSubscriptionID")
	os.Setenv(kAzureClientID, "SomeClientID")
	os.Setenv(kAzureClientSecret, "SomeClientSecret")
	os.Setenv(kAzureTenantID, "SomeTenantID")
	os.Setenv(kAzureZoneName, "SomeZoneName")
	os.Setenv(kAzureResourceGroup, "SomeResourceGroup")

	_, err := NewClient(context.TODO())
	if err != nil {
		t.Errorf("Client initialization failed: %v", err)
	}

	// Clean up the environment variables after the test
	os.Unsetenv(kAzureSubscriptionID)
	os.Unsetenv(kAzureClientID)
	os.Unsetenv(kAzureClientSecret)
	os.Unsetenv(kAzureTenantID)
	os.Unsetenv(kAzureZoneName)
	os.Unsetenv(kAzureResourceGroup)
}

func TestDNSNameConcatenation(t *testing.T) {
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "mydomain.com",
			Name:       "subdomain",
			RecordType: "A",
			Targets:    []string{"127.0.0.1"},
		},
	}

	c := &azureDNS{
		zoneName:      "example.com",
		resourceGroup: "SomeResourceGroup",
	}

	params, err := c.resourceRecordSet(context.TODO(), &record)
	if err != nil {
		t.Fatalf("resourceRecordSet failed: %v", err)
	}

	// Validate that the ARecord type is set for an A record
	if len(params.Properties.ARecords) == 0 || *params.Properties.ARecords[0].IPv4Address != "127.0.0.1" {
		t.Errorf("Expected an ARecord with IP 127.0.0.1, got %v", params.Properties.ARecords)
	}
}

func TestAliasTargetProperty(t *testing.T) {
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "example.com",
			Name:       "subdomain",
			Targets:    []string{"alias.example.com"},
			RecordType: "CNAME",
		},
	}

	c := &azureDNS{
		zoneName:      "example.com",
		resourceGroup: "SomeResourceGroup",
	}

	params, err := c.resourceRecordSet(context.TODO(), &record)
	if err != nil {
		t.Fatalf("resourceRecordSet failed: %v", err)
	}

	// Validate that the CNAME record is properly set
	if params.Properties.CnameRecord == nil || *params.Properties.CnameRecord.Cname != "alias.example.com" {
		t.Errorf("Expected a CNAME record with alias.example.com, got %v", params.Properties.CnameRecord)
	}
}

func TestResourceRecordSetWithTTL(t *testing.T) {
	tests := []struct {
		name     string
		record   *phonebook.DNSRecord
		expected int64
	}{
		{
			name: "With custom TTL",
			record: &phonebook.DNSRecord{
				Spec: phonebook.DNSRecordSpec{
					Zone:       "example.com",
					Name:       "custom-ttl",
					Targets:    []string{"1.2.3.4"},
					RecordType: "A",
					TTL:        to.Ptr(int64(7200)),
				},
			},
			expected: 7200,
		},
		{
			name: "Without TTL (default)",
			record: &phonebook.DNSRecord{
				Spec: phonebook.DNSRecordSpec{
					Zone:       "example.com",
					Name:       "default-ttl",
					Targets:    []string{"5.6.7.8"},
					RecordType: "A",
				},
			},
			expected: defaultTTL,
		},
	}

	c := &azureDNS{
		zoneName:      "example.com",
		resourceGroup: "SomeResourceGroup",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := c.resourceRecordSet(context.TODO(), tt.record)
			if err != nil {
				t.Fatalf("resourceRecordSet failed: %v", err)
			}
			assert.Equal(t, tt.expected, *params.Properties.TTL)
		})
	}
}

func TestCreateDNSRecordWithTTL(t *testing.T) {
	// Create a fake record with custom TTL
	record := &phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "example.com",
			Name:       "testrecord",
			Targets:    []string{"1.2.3.4"},
			RecordType: "A",
			TTL:        to.Ptr(int64(3600)),
		},
	}

	// Create a mock RecordSetsClient
	mockClient := new(MockRecordSetsClient)

	// Set up expectations
	mockClient.On("CreateOrUpdate",
		mock.Anything,                           // context
		"SomeResourceGroup",                     // resourceGroupName
		"example.com",                           // zoneName
		"testrecord",                            // relativeRecordSetName
		armdns.RecordTypeA,                      // recordType
		mock.AnythingOfType("armdns.RecordSet"), // parameters
		mock.Anything,                           // options
	).Run(func(args mock.Arguments) {
		// Check if the TTL is correctly set in the parameters
		params := args.Get(5).(armdns.RecordSet)
		assert.Equal(t, int64(3600), *params.Properties.TTL)
	}).Return(armdns.RecordSetsClientCreateOrUpdateResponse{
		RecordSet: armdns.RecordSet{
			ID: to.Ptr("fake-id"),
		},
	}, nil)

	// Create the azureDNS client with the mock
	c := &azureDNS{
		integration:      "azure-test",
		zoneName:         "example.com",
		resourceGroup:    "SomeResourceGroup",
		recordSetsClient: mockClient,
	}

	// Perform the Create operation
	updater := &mocks.Updater{}
	err := c.Create(context.TODO(), *record, updater)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "fake-id", updater.Info["recordID"])

	// Verify that our expectations were met
	mockClient.AssertExpectations(t)
}

func TestCreateDNSRecordWithDefaultTTL(t *testing.T) {
	// Create a fake record without specifying TTL
	record := &phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "example.com",
			Name:       "testrecord",
			Targets:    []string{"1.2.3.4"},
			RecordType: "A",
		},
	}

	// Create a mock RecordSetsClient
	mockClient := new(MockRecordSetsClient)

	// Set up expectations
	mockClient.On("CreateOrUpdate",
		mock.Anything,                           // context
		"SomeResourceGroup",                     // resourceGroupName
		"example.com",                           // zoneName
		"testrecord",                            // relativeRecordSetName
		armdns.RecordTypeA,                      // recordType
		mock.AnythingOfType("armdns.RecordSet"), // parameters
		mock.Anything,                           // options
	).Run(func(args mock.Arguments) {
		// Check if the default TTL is correctly set in the parameters
		params := args.Get(5).(armdns.RecordSet)
		assert.Equal(t, int64(defaultTTL), *params.Properties.TTL)
	}).Return(armdns.RecordSetsClientCreateOrUpdateResponse{
		RecordSet: armdns.RecordSet{
			ID: to.Ptr("fake-id"),
		},
	}, nil)

	// Create the azureDNS client with the mock
	c := &azureDNS{
		zoneName:         "example.com",
		resourceGroup:    "SomeResourceGroup",
		recordSetsClient: mockClient,
	}

	// Perform the Create operation
	updater := &mocks.Updater{}
	err := c.Create(context.TODO(), *record, updater)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "fake-id", updater.Info["recordID"])

	// Verify that our expectations were met
	mockClient.AssertExpectations(t)
}

func TestDeleteDNSRecord(t *testing.T) {
	// Create a fake record
	record := &phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:       "example.com",
			Name:       "testrecord",
			RecordType: "A",
		},
	}

	// Create a mock RecordSetsClient
	mockClient := new(MockRecordSetsClient)

	// Set up expectations
	mockClient.On("Delete",
		mock.Anything,       // context
		"SomeResourceGroup", // resourceGroupName
		"example.com",       // zoneName
		"testrecord",        // relativeRecordSetName
		armdns.RecordTypeA,  // recordType
		mock.Anything,       // options
	).Return(armdns.RecordSetsClientDeleteResponse{}, nil)

	// Create the azureDNS client with the mock
	c := &azureDNS{
		zoneName:         "example.com",
		resourceGroup:    "SomeResourceGroup",
		recordSetsClient: mockClient,
	}

	// Perform the Delete operation
	err := c.Delete(context.TODO(), *record, &mocks.Updater{})

	// Assert
	assert.NoError(t, err)

	// Verify that our expectations were met
	mockClient.AssertExpectations(t)
}
