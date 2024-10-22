package gcore

import (
	"context"
	"fmt"
	"net/url"
	"os"

	gdns "github.com/G-Core/gcore-dns-sdk-go"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

const (
	EnvAPIURL   = "GCORE_API_URL"
	EnvAPIToken = "GCORE_API_TOKEN"
	DefaultTTL  = int64(120) // gcore doesn't support shorter TTL for the free plan, so 120 is the basis to avoid confusions
)

type gcore struct {
	integration string
	zoneID      string
	zones       []string
	api         *gdns.Client
}

func NewClient(ctx context.Context) (*gcore, error) {
	var err error

	apiURL := os.Getenv(EnvAPIURL)

	token := os.Getenv(EnvAPIToken)
	api := gdns.NewClient(gdns.PermanentAPIKeyAuth(token))

	if apiURL != "" {
		api.BaseURL, err = url.Parse(apiURL)
		if err != nil {
			return nil, err
		}
	}

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

func (c *gcore) Create(ctx context.Context, record *phonebook.DNSRecord) error {
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

	return c.api.AddZoneRRSet(ctx, record.Spec.Zone, fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone), record.Spec.RecordType, values, int(*ttl))
}

func (c *gcore) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	return c.api.DeleteRRSet(ctx, record.Spec.Zone, fmt.Sprintf("%s.%s", record.Spec.Name, record.Spec.Zone), record.Spec.RecordType)
}
