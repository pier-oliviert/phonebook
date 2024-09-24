package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const kAWSZoneID = "AWS_ZONE_ID"

type r53 struct {
	hostedZoneID string
	zoneID       string
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
		return nil, err
	}
	zoneID, err := utils.RetrieveValueFromEnvOrFile(kAWSZoneID)
	if err != nil {
		return nil, fmt.Errorf("PB#0100: Zone ID not found -- %w", err)
	}

	logger.Info("[Provider] AWS Configured", "Zone ID", zoneID)

	return &r53{
		zoneID: zoneID,
		Client: route53.NewFromConfig(cfg),
	}, nil
}

func (c *r53) Create(ctx context.Context, record *phonebook.DNSRecord) error {
	inputs := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &c.zoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{{
				Action:            types.ChangeActionCreate,
				ResourceRecordSet: c.resourceRecordSet(record),
			}},
		},
	}

	_, err := c.ChangeResourceRecordSets(ctx, &inputs)
	return err
}

func (c *r53) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	inputs := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &c.zoneID,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{{
				Action:            types.ChangeActionDelete,
				ResourceRecordSet: c.resourceRecordSet(record),
			}},
		},
	}

	_, err := c.ChangeResourceRecordSets(ctx, &inputs)
	return err
}

// Convert a DNSRecord to a resourceRecordSet
func (c *r53) resourceRecordSet(record *phonebook.DNSRecord) *types.ResourceRecordSet {
	set := types.ResourceRecordSet{
		Name: &record.Spec.Name,
		Type: types.RRType(record.Spec.RecordType),
	}

	if set.Type == types.RRTypeA || set.Type == types.RRTypeCname {
		set.AliasTarget = &types.AliasTarget{
			DNSName:      &record.Spec.Targets[0],
			HostedZoneId: &c.hostedZoneID,
		}
	} else {
		set.ResourceRecords = append(set.ResourceRecords, types.ResourceRecord{
			Value: &record.Spec.Targets[0],
		})
		set.TTL = new(int64)
		*set.TTL = 60

	}

	return &set
}
