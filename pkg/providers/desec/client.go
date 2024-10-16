package desec

import (
	"context"
	"fmt"

	"github.com/nrdcg/desec"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kDesecToken = "DESEC_TOKEN"
)

type deSEC struct {
	integration string
	token       string
	client      *desec.Client
	zones	   []string
	zoneName   string
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

// Create DNS record in deSEC
func (d *deSEC) CreateDNSRecord(ctx context.Context, zoneName string, recordName string, recordType string, recordValue string, ttl int) error {
	logger := log.FromContext(ctx)

	// Create a new RRSet
	rrset := desec.RRSet{
		Name:  recordName,
		Type:  recordType,
		TTL:   ttl,
		Records: []string{recordValue},
	}

	// Create the RRSet
	_, err := d.client.Records.Create(ctx, rrset)
	if err != nil {
		return fmt.Errorf("PB-DESEC-#0002: Unable to create record -- %w", err)
	}

	logger.Info("[Provider] deSEC Record Created")

	return nil
}

// Delete DNS record in deSEC
func (d *deSEC) DeleteDNSRecord(ctx context.Context, zoneName string, recordName string, recordType string) error {
	logger := log.FromContext(ctx)

	// Delete the RRSet
	err := d.client.Records.Delete(ctx, zoneName, recordName, recordType)
	if err != nil {
		return fmt.Errorf("PB-DESEC-#0003: Unable to delete record -- %w", err)
	}

	logger.Info("[Provider] deSEC Record Deleted")

	return nil
}
