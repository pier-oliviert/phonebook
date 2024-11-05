package cloudflare

import (
	"context"
	"fmt"
	"strings"

	client "github.com/cloudflare/cloudflare-go"
	"github.com/pier-oliviert/konditionner/pkg/konditions"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/utils"
)

const (
	kCloudflareAPIKeyName        = "CF_API_TOKEN"
	kCloudflareZoneID            = "CF_ZONE_ID"
	defaultTTL                   = int64(60) // Default TTL for DNS records in seconds if not specified
	kCloudflarePropertiesProxied = "proxied"
)

type cf struct {
	integration string
	zoneID      string
	zones       []string

	client.API
}

// Generate a new Cloudflare Provider that can be used to create DNS records. The
// provider requires values to be defined by the user in order to be configured properly.
//
// The CF_API_TOKEN value can either be sourced from an environment variable, or from a file.
// The file needs to be located at `${kProviderConfigPath}/CF_API_TOKEN`
// The file path is preferred as that's easier to work with different providers and Kubernetes secret system.
func NewClient(ctx context.Context) (*cf, error) {
	token, err := utils.RetrieveValueFromEnvOrFile(kCloudflareAPIKeyName)
	if err != nil {
		return nil, fmt.Errorf("PB-CF-#0001: API Key not found -- %w", err)
	}

	zoneID, err := utils.RetrieveValueFromEnvOrFile(kCloudflareZoneID)
	if err != nil {
		return nil, fmt.Errorf("PB-CF-#0002: Zone ID not found -- %w", err)
	}

	// Trimming space in case the user included a space when copying the token over. This small
	// quality of life fix might just make it easier to work with token (debugging white spaces when trying new tools can be frustrating)
	api, err := client.NewWithAPIToken(strings.TrimSpace(token))
	if err != nil {
		return nil, fmt.Errorf("PB-CF-#0003: Could not create new Cloudflare Client -- %w", err)
	}

	return &cf{
		zoneID: zoneID,
		API:    *api,
	}, nil
}

func (c *cf) Configure(ctx context.Context, integration string, zones []string) error {
	c.integration = integration
	c.zones = zones

	return nil
}

func (c *cf) Zones() []string {
	return c.zones
}

func (c *cf) Create(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	dnsParams := client.CreateDNSRecordParams{
		Type:    record.Spec.RecordType,
		Name:    record.Spec.Name,
		Content: record.Spec.Targets[0],
	}

	// It doesn't seem the cloudflare api library has a way of supporting multiple targets
	// I tried to create multiple entries for the same hostname in the CF dashboard and it provides an error, so I'm assuming it's not supported. Shame.
	if len(record.Spec.Targets) > 1 {
		// Throw an error if the user tries to create multiple targets for the same hostname
		return fmt.Errorf("PB-CF-#0004: Cloudflare does not support multiple targets for the same hostname")
	}

	// Set TTL
	// The cloudflare library only accepts int, so we need to convert the int64 to int
	// Shame because it means we have to type convert the default value as well, only for this provider.
	if record.Spec.TTL != nil {
		dnsParams.TTL = int(*record.Spec.TTL)
	} else {
		dnsParams.TTL = int(defaultTTL)
	}

	if proxied, ok := record.Spec.Properties[kCloudflarePropertiesProxied]; ok {
		dnsParams.Proxied = new(bool)
		*dnsParams.Proxied = strings.EqualFold(proxied, "true")
	}

	response, err := c.CreateDNSRecord(ctx, client.ZoneIdentifier(c.zoneID), dnsParams)
	if err != nil {
		return fmt.Errorf("PB-CF-#0005: Failed to create DNS record -- %w", err)
	}

	su.StageRemoteInfo(phonebook.IntegrationInfo{
		"recordID": response.ID,
	})
	su.StageCondition(konditions.ConditionCreated, "Cloudflare record created")

	return nil
}

func (c *cf) Delete(ctx context.Context, record phonebook.DNSRecord, su phonebook.StagingUpdater) error {
	if record.Status.RemoteInfo[c.integration] == nil {
		// Nothing to delete if the RemoteID was never added to this resource. It could
		// cause an orphan record in Cloudflare, but it might be the better option as the system would
		// never be able to recover from a lack of remoteID.
		return nil
	}

	err := c.DeleteDNSRecord(ctx, client.ZoneIdentifier(c.zoneID), record.Status.RemoteInfo[c.integration]["recordID"])
	if err != nil {
		return fmt.Errorf("PB-CF-#0006: Failed to delete DNS record -- %w", err)
	}

	su.StageCondition(konditions.ConditionTerminated, "Cloudflare record deleted")

	return nil
}
