package gcore

import (
	"context"
	"fmt"

	gdns "github.com/G-Core/gcore-dns-sdk-go"
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/utils"
)

const (
	EnvAPIToken = "GCORE_API_TOKEN"
	DefaultTTL  = int64(120) // gcore doesn't support shorter TTL for the free plan, so 120 is the basis to avoid confusions
)

// This interface is created so mocks can be done on testing. Since gcore doesn't have any interface to work with,
// this needs to exists here in order for phonebook to have proper testing
type api interface {
	AddZoneRRSet(context.Context, string, string, string, []gdns.ResourceRecord, int, ...gdns.AddZoneOpt) error
	DeleteRRSet(context.Context, string, string, string) error
}

type gcore struct {
	integration string
	zoneID      string
	zones       []string
	api         api
}

func NewClient(ctx context.Context) (*gcore, error) {
	var err error

	token, err := utils.RetrieveValueFromEnvOrFile(EnvAPIToken)
	if err != nil {
		return nil, err
	}
	api := gdns.NewClient(gdns.PermanentAPIKeyAuth(token))

	return &gcore{
		api: api,
	}, err
}

func (c *gcore) Configure(ctx context.Context, integration string, zones []string) error {
	c.zones = zones
	c.integration = integration

	return nil
}

func (c *gcore) Zones() []string {
	return c.zones
}

func (c *gcore) Create(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	values := []gdns.ResourceRecord{{
		Enabled: true,
	}}

	// Need to copy each to satisfy the []any types
	for _, t := range record.Spec.Targets {
		values[0].Content = append(values[0].Content, t)
	}

	// TTL is an optional field, so check if it is set before storing the default value
	ttl := record.Spec.TTL
	if ttl == nil {
		ttl = new(int64)
		*ttl = DefaultTTL
	}

	err := c.api.AddZoneRRSet(ctx, record.Spec.Zone, fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone), record.Spec.RecordType, values, int(*ttl))
	if err != nil {
		return err
	}

	su.StageCondition(konditions.ConditionCreated, "G-Core record created")
	return nil
}

func (c *gcore) Delete(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	err := c.api.DeleteRRSet(ctx, record.Spec.Zone, fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone), record.Spec.RecordType)
	if err != nil {
		return err
	}

	su.StageCondition(konditions.ConditionTerminated, "G-Core record deleted")
	return nil
}
