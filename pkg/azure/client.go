package azure

import (
	"context"
	"fmt"
	"strings"

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
	defaultTTL           = int64(60) // Default TTL for DNS records in seconds if not specified
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
	params, err := c.resourceRecordSet(ctx, record)
	if err != nil {
		return fmt.Errorf("failed to create resource record set: %w", err)
	}
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
func (c *azureDNS) resourceRecordSet(ctx context.Context, record *phonebook.DNSRecord) (armdns.RecordSet, error) {
	ttl := defaultTTL
	if record.Spec.TTL != nil {
		ttl = *record.Spec.TTL
	}

	params := armdns.RecordSet{
		Properties: &armdns.RecordSetProperties{
			TTL: to.Ptr(int64(ttl)),
		},
	}

	// Create specific record types based on the DNS type
	switch armdns.RecordType(record.Spec.RecordType) {
	case armdns.RecordTypeA:
		aRecords := make([]*armdns.ARecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			aRecords[i] = &armdns.ARecord{IPv4Address: to.Ptr(target)}
		}
		params.Properties.ARecords = aRecords

	case armdns.RecordTypeAAAA:
		aaaaRecords := make([]*armdns.AaaaRecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			aaaaRecords[i] = &armdns.AaaaRecord{IPv6Address: to.Ptr(target)}
		}
		params.Properties.AaaaRecords = aaaaRecords

	case armdns.RecordTypeCNAME:
		// CNAME can only have one target
		// If Targets is more than one, throw an error
		if len(record.Spec.Targets) > 1 {
			err := fmt.Errorf("CNAME record can only have one target")
			log.FromContext(ctx).Error(err, "CNAME record can only have one target")
			return armdns.RecordSet{}, err
		}

		params.Properties.CnameRecord = &armdns.CnameRecord{
			Cname: to.Ptr(record.Spec.Targets[0]),
		}

	case armdns.RecordTypeMX:
		mxRecords := make([]*armdns.MxRecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			parts := strings.SplitN(target, " ", 2)
			if len(parts) != 2 {
				err := fmt.Errorf("invalid MX record: %s", target)
				log.FromContext(ctx).Error(err, "Invalid MX record")
				return armdns.RecordSet{}, err
			}
			mxRecords[i] = &armdns.MxRecord{
				Preference: toInt32Ptr(parts[0]),
				Exchange:   to.Ptr(parts[1]),
			}
		}
		params.Properties.MxRecords = mxRecords

	case armdns.RecordTypeTXT:
		txtRecords := make([]*armdns.TxtRecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			txtRecords[i] = &armdns.TxtRecord{Value: []*string{to.Ptr(target)}}
		}
		params.Properties.TxtRecords = txtRecords

	case armdns.RecordTypeSRV:
		srvRecords := make([]*armdns.SrvRecord, len(record.Spec.Targets))
		for i, target := range record.Spec.Targets {
			parts := strings.Split(target, " ")
			if len(parts) != 4 {
				err := fmt.Errorf("invalid SRV record: %s", target)
				log.FromContext(ctx).Error(err, "Invalid SRV record")
				return armdns.RecordSet{}, err
			}
			srvRecords[i] = &armdns.SrvRecord{
				Priority: toInt32Ptr(parts[0]),
				Weight:   toInt32Ptr(parts[1]),
				Port:     toInt32Ptr(parts[2]),
				Target:   to.Ptr(parts[3]),
			}
		}
		params.Properties.SrvRecords = srvRecords

	default:
		// Unsupported record type and return an error
		err := fmt.Errorf("unsupported record type: %s", record.Spec.RecordType)
		log.FromContext(ctx).Error(err, "Unsupported record type")
	}

	return params, nil
}
