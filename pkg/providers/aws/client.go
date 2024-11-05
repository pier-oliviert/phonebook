package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// Since the Azure provider already uses this package, it
	// makes sense to reuse it here too. This is only a convenience
	// package for converting values to pointer. Other package exists with
	// a similar functionality, but I rather reuse packages to keep the
	// depedency tree smaller
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

const (
	kAWSZoneID  = "AWS_ZONE_ID"
	AliasTarget = "AliasHostedZoneID"
	defaultTTL  = int64(60) // Default TTL for DNS records in seconds if not specified
)

type r53 struct {
	integration string
	zones       []string
	zoneID      string
	*route53.Client
}

// NewClient doesn't include arguments as all configuration/secret options should be stored
// as environment variable or as secret file mounted by Kubernetes. Since the name of those variables
// and secret files are unique to the provider, it's better for the Client to inspect the system itself
// by using the tools available and return an error if the client cannot be created.
func NewClient(ctx context.Context) (*r53, error) {
	logger := log.FromContext(ctx)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("PB-AWS-#0001: Failed to load AWS configuration -- %w", err)
	}
	zoneID, err := utils.RetrieveValueFromEnvOrFile(kAWSZoneID)
	if err != nil {
		return nil, fmt.Errorf("PB-AWS-#0002: Zone ID not found -- %w", err)
	}

	logger.Info("[Provider] AWS Configured", "Zone ID", zoneID)

	return &r53{
		zoneID: zoneID,
		Client: route53.NewFromConfig(cfg),
	}, nil
}

func (c *r53) Configure(ctx context.Context, integration string, zones []string) error {
	c.zones = zones
	c.integration = integration

	return nil
}

func (c *r53) Zones() []string {
	return c.zones
}

func (c *r53) Create(ctx context.Context, record phonebook.DNSRecord, updater phonebook.StagingUpdater) error {
	inputs := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &c.zoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{{
				Action:            types.ChangeActionCreate,
				ResourceRecordSet: c.resourceRecordSet(ctx, &record),
			}},
		},
	}

	_, err := c.ChangeResourceRecordSets(ctx, &inputs)
	if err != nil {
		return fmt.Errorf("PB-AWS-#0003: Failed to create DNS record -- %w", err)
	}

	updater.StageCondition(konditions.ConditionCreated, "Route53 created the record")
	return nil
}

func (c *r53) Delete(ctx context.Context, record phonebook.DNSRecord, updater phonebook.StagingUpdater) error {
	inputs := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &c.zoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{{
				Action:            types.ChangeActionDelete,
				ResourceRecordSet: c.resourceRecordSet(ctx, &record),
			}},
		},
	}

	_, err := c.ChangeResourceRecordSets(ctx, &inputs)
	if err != nil {
		return fmt.Errorf("PB-AWS-#0004: Failed to delete DNS record -- %w", err)
	}

	updater.StageCondition(konditions.ConditionTerminated, "Route53 record deleted")

	return nil
}

// Convert a DNSRecord to a resourceRecordSet
func (c *r53) resourceRecordSet(ctx context.Context, record *phonebook.DNSRecord) *types.ResourceRecordSet {
	fullName := fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone)

	set := types.ResourceRecordSet{
		Name: &fullName,
		Type: types.RRType(record.Spec.RecordType),
	}

	// Set TTL
	ttl := defaultTTL
	if record.Spec.TTL != nil {
		ttl = *record.Spec.TTL
	}

	set.TTL = &ttl

	if hostedZoneID, ok := record.Spec.Properties[AliasTarget]; ok {
		// User specified Alias Hosted Zone ID. As such, Phonebook will
		// create a DNS record using AWS' Alias Target function(1).
		//
		// Alias Target is useful when you want to create a DNS record that points to
		// an AWS service. Since AWS can validate that the route doesn't leave their infra, you
		// get some benefits like cost reduction, etc. The official documentation will have more
		// information about this.
		//
		// 1. https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html

		set.AliasTarget = &types.AliasTarget{
			DNSName:      &record.Spec.Targets[0],
			HostedZoneId: &hostedZoneID,
		}
		// Note: For Alias records, TTL is not used and should be omitted
		set.TTL = nil
	} else {
		// Handle different record types
		switch types.RRType(record.Spec.RecordType) {
		case types.RRTypeA, types.RRTypeAaaa, types.RRTypeCname:
			set.ResourceRecords = make([]types.ResourceRecord, len(record.Spec.Targets))
			for i, target := range record.Spec.Targets {
				set.ResourceRecords[i] = types.ResourceRecord{Value: &target}
			}
		case types.RRTypeTxt:
			set.ResourceRecords = make([]types.ResourceRecord, len(record.Spec.Targets))
			for i, target := range record.Spec.Targets {
				// AWS TXT Records requires value to be "quoted". For this reason, the Sprintf() method
				// is called before wrapping the returned value in a pointer for ResourceRecord.
				set.ResourceRecords[i] = types.ResourceRecord{Value: to.Ptr(fmt.Sprintf("\"%s\"", target))}
			}

		case types.RRTypeMx:
			set.ResourceRecords = make([]types.ResourceRecord, len(record.Spec.Targets))
			for i, target := range record.Spec.Targets {
				// Assuming MX records are in the format "priority target"
				set.ResourceRecords[i] = types.ResourceRecord{Value: &target}
			}
		case types.RRTypeSrv:
			set.ResourceRecords = make([]types.ResourceRecord, len(record.Spec.Targets))
			for i, target := range record.Spec.Targets {
				// Assuming SRV records are in the format "priority weight port target"
				set.ResourceRecords[i] = types.ResourceRecord{Value: &target}
			}

		default:
			// For unsupported types, log an error
			log.FromContext(ctx).Error(fmt.Errorf("PB-AWS-#0005: Unsupported record type"), "Record Type", record.Spec.RecordType)
		}
	}

	return &set
}
