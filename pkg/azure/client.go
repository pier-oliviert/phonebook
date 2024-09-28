package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kAzureZoneName       = "AZURE_ZONE_NAME"
	kAzureResourceGroup  = "AZURE_RESOURCE_GROUP"
	kAzureSubscriptionID = "AZURE_SUBSCRIPTION_ID"
	kAzureClientID       = "AZURE_CLIENT_ID"
	kAzureClientSecret   = "AZURE_CLIENT_SECRET"
	kAzureTenantID       = "AZURE_TENANT_ID"
	defaultTTL           = 60 // Default TTL for DNS records in seconds if not specified
)

type azureDNS struct {
	zoneName         string
	resourceGroup    string
	recordSetsClient interface {
		CreateOrUpdate(ctx context.Context, resourceGroupName string, zoneName string, relativeRecordSetName string, recordType armdns.RecordType, parameters armdns.RecordSet, options *armdns.RecordSetsClientCreateOrUpdateOptions) (armdns.RecordSetsClientCreateOrUpdateResponse, error)
		Delete(ctx context.Context, resourceGroupName string, zoneName string, relativeRecordSetName string, recordType armdns.RecordType, options *armdns.RecordSetsClientDeleteOptions) (armdns.RecordSetsClientDeleteResponse, error)
	}
}

// NewClient initializes an Azure DNS client
func NewClient(ctx context.Context) (*azureDNS, error) {
	logger := log.FromContext(ctx)

	clientID, err := utils.RetrieveValueFromEnvOrFile(kAzureClientID)
	if err != nil {
		return nil, fmt.Errorf("PB#0104: Azure Client ID not found -- %w", err)
	}

	clientSecret, err := utils.RetrieveValueFromEnvOrFile(kAzureClientSecret)
	if err != nil {
		return nil, fmt.Errorf("PB#0105: Azure Client Secret not found -- %w", err)
	}

	tenantID, err := utils.RetrieveValueFromEnvOrFile(kAzureTenantID)
	if err != nil {
		return nil, fmt.Errorf("PB#0106: Azure Tenant ID not found -- %w", err)
	}

	subscriptionID, err := utils.RetrieveValueFromEnvOrFile(kAzureSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("PB#0101: Azure Subscription ID not found -- %w", err)
	}

	zoneName, err := utils.RetrieveValueFromEnvOrFile(kAzureZoneName)
	if err != nil {
		return nil, fmt.Errorf("PB#0102: Azure Zone Name not found -- %w", err)
	}

	resourceGroup, err := utils.RetrieveValueFromEnvOrFile(kAzureResourceGroup)
	if err != nil {
		return nil, fmt.Errorf("PB#0103: Azure Resource Group not found -- %w", err)
	}

	// Create the credential
	credential, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create Azure credential: %w", err)
	}

	// Initialize the DNS client
	dnsClient, err := armdns.NewRecordSetsClient(subscriptionID, credential, &arm.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create Azure DNS client: %w", err)
	}

	logger.Info("[Provider] Azure Configured", "Zone Name", zoneName, "Resource Group", resourceGroup)

	return &azureDNS{
		zoneName:         zoneName,
		resourceGroup:    resourceGroup,
		recordSetsClient: dnsClient,
	}, nil
}

// Create DNS record in Azure
func (c *azureDNS) Create(ctx context.Context, record *phonebook.DNSRecord) error {
	params := c.resourceRecordSet(record)
	response, err := c.recordSetsClient.CreateOrUpdate(ctx, c.resourceGroup, c.zoneName, record.Spec.Name, armdns.RecordType(record.Spec.RecordType), params, nil)
	if err != nil {
		return fmt.Errorf("failed to create Azure DNS record: %w", err)
	}

	record.Status.Provider = "Azure"
	record.Status.RemoteID = response.ID

	// Log the record creation to the console
	log.FromContext(ctx).Info("[Provider] Azure DNS Record Created", "Name", record.Spec.Name, "Type", record.Spec.RecordType, "Targets", record.Spec.Targets, "TTL", *params.Properties.TTL)
	return nil
}

// Delete DNS record from Azure
func (c *azureDNS) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	_, err := c.recordSetsClient.Delete(ctx, c.resourceGroup, c.zoneName, record.Spec.Name, armdns.RecordType(record.Spec.RecordType), nil)
	if err != nil {
		return fmt.Errorf("failed to delete Azure DNS record: %w", err)
	}

	record.Status.Provider = "Azure"
	record.Status.RemoteID = nil

	// Log the record deletion to the console
	log.FromContext(ctx).Info("[Provider] Azure DNS Record Deleted", "Name", record.Spec.Name, "Type", record.Spec.RecordType, "Targets", record.Spec.Targets)

	return nil
}

// Convert a DNSRecord to an Azure DNS record set
func (c *azureDNS) resourceRecordSet(record *phonebook.DNSRecord) armdns.RecordSet {
	ttl := defaultTTL
	if record.Spec.TTL != nil {
		ttl = *record.Spec.TTL
	}

	params := armdns.RecordSet{
		Properties: &armdns.RecordSetProperties{
			TTL: to.Ptr(int64(ttl)),
		},
	}

	// Create specific record types based on the DNS type (e.g., A, CNAME)
	switch armdns.RecordType(record.Spec.RecordType) {
	case armdns.RecordTypeA:
		aRecords := make([]*armdns.ARecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			aRecords[i] = &armdns.ARecord{IPv4Address: to.Ptr(target)}
		}
		params.Properties.ARecords = aRecords
	case armdns.RecordTypeCNAME:
		params.Properties.CnameRecord = &armdns.CnameRecord{
			Cname: to.Ptr(record.Spec.Targets[0]),
		}
	}

	return params
}
