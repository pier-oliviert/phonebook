package desec

import (
	"context"
	"fmt"

	"github.com/nrdcg/desec"
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kDesecToken = "DESEC_TOKEN"
	defaultTTL  = int64(3600) // Default TTL for DNS records in seconds if not specified
)

type deSEC struct {
	integration string
	token       string
	client      *desec.Client
	zones       []string
}

// NewClient initializes a deSEC DNS client
func NewClient(ctx context.Context) (*deSEC, error) {
	logger := log.FromContext(ctx)

	token, err := utils.RetrieveValueFromEnvOrFile(kDesecToken)
	if err != nil {
		return nil, fmt.Errorf("PB-DESEC-#0001: deSEC Token not found -- %w", err)
	}

	// Create a new deSEC client with the default options and set the token
	options := desec.NewDefaultClientOptions()
	client := desec.New(token, options)

	logger.Info("[Provider] deSEC Configured")

	return &deSEC{
		integration: "deSEC",
		token:       token,
		client:      client,
	}, nil
}

func (d *deSEC) Configure(ctx context.Context, integration string, zones []string) error {
	d.integration = integration
	d.zones = zones

	return nil
}

func (d *deSEC) Zones() []string {
	return d.zones
}

// Create DNS record in deSEC
func (d *deSEC) Create(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	logger := log.FromContext(ctx)

	ttl := defaultTTL
	if record.Spec.TTL != nil {
		ttl = *record.Spec.TTL
	}

	// Create a new RRSet
	rrset := desec.RRSet{
		Domain:  record.Spec.Zone,
		Name:    record.Spec.Name,
		SubName: record.Spec.Name,
		Type:    record.Spec.RecordType,
		TTL:     int(ttl),
		Records: record.Spec.Targets,
	}

	// Create the RRSet
	_, err := d.client.Records.Create(ctx, rrset)
	if err != nil {
		return fmt.Errorf("PB-DESEC-#0002: Unable to create record -- %w", err)
	}

	logger.Info("[Provider] deSEC Record Created")
	su.StageCondition(konditions.ConditionCreated, "deSEC record created")

	return nil
}

// Delete DNS record in deSEC
func (d *deSEC) Delete(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	logger := log.FromContext(ctx)

	// Delete the RRSet
	err := d.client.Records.Delete(ctx, record.Spec.Zone, record.Spec.Name, record.Spec.RecordType)
	if err != nil {
		return fmt.Errorf("PB-DESEC-#0003: Unable to delete record -- %w", err)
	}

	logger.Info("[Provider] deSEC Record Deleted")

	su.StageCondition(konditions.ConditionTerminated, "deSEC record deleted")

	return nil
}
