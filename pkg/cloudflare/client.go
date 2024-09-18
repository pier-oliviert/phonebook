package cloudflare

import (
	"context"
	"fmt"
	"strings"

	client "github.com/cloudflare/cloudflare-go"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/pier-oliviert/phonebook/pkg/utils"
)

const kCloudflareAPIKeyName = "CF_API_TOKEN"
const kCloudflareZoneID = "CF_ZONE_ID"

const kCloudflarePropertiesProxied = "proxied"

type cf struct {
	zoneID string

	client.API
}

// Generate a new Cloudflare Provider that can be used to create DNS records. The
// provider requires values to be defined by the user in order to be configured properly.
//
// The CF_API_TOKEN value can either be sourced from an environment variable, or from a file.
// The file needs to be located at `${kProviderConfigPath}/CF_API_TOKEN`
// The file path is preferred as that's easier to work with different providers and Kubernetes secret system.
func NewClient() (*cf, error) {
	token, err := utils.RetrieveValueFromEnvOrFile(kCloudflareAPIKeyName)
	if err != nil {
		return nil, fmt.Errorf("PB#0100: API Key not found -- %w", err)
	}

	zoneID, err := utils.RetrieveValueFromEnvOrFile(kCloudflareZoneID)
	if err != nil {
		return nil, fmt.Errorf("PB#0100: Zone ID not found -- %w", err)
	}

	// Trimming space in case the user included a space when copying the token over. This small
	// quality of life fix might just make it easier to work with token (debugging white spaces when trying new tools can be frustrating)
	api, err := client.NewWithAPIToken(strings.TrimSpace(token))
	if err != nil {
		return nil, fmt.Errorf("PB#4003: Could not create new Cloudflare Client -- %w", err)
	}

	return &cf{
		zoneID: zoneID,
		API:    *api,
	}, nil
}

func (c *cf) Create(ctx context.Context, record *phonebook.DNSRecord) error {
	dnsParams := client.CreateDNSRecordParams{
		Type:    record.Spec.RecordType,
		Name:    record.Spec.Name,
		Content: record.Spec.Targets[0],
	}

	if proxied, ok := record.Spec.Properties[kCloudflarePropertiesProxied]; ok {
		dnsParams.Proxied = new(bool)
		*dnsParams.Proxied = strings.EqualFold(proxied, "true")
	}

	response, err := c.CreateDNSRecord(ctx, client.ZoneIdentifier(c.zoneID), dnsParams)
	if err != nil {
		return err
	}

	record.Status.RemoteID = new(string)
	*record.Status.RemoteID = response.ID
	return nil
}

func (c *cf) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	if record.Status.RemoteID == nil {
		// Nothing to delete if the RemoteID was never added to this resource. It could
		// cause an orphan record in Cloudflare, but it might be the better option as the system would
		// never be able to recover from a lack of remoteID.
		return nil
	}

	return c.DeleteDNSRecord(ctx, client.ZoneIdentifier(c.zoneID), *record.Status.RemoteID)
}
