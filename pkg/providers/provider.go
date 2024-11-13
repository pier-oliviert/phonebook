package providers

import (
	"context"
	"fmt"
	"sync"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

// This constant needs to be configured through a build flag when Phonebook is released
var ProviderVersion = "0.0.0"

type ProviderStore struct {
	mu       sync.Mutex
	provider Provider
}

func (ps *ProviderStore) Store(p Provider) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.provider = p
}

func (ps *ProviderStore) Provider() Provider {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	return ps.provider
}

type Provider interface {
	Configure(ctx context.Context, integration string, zones []string) error

	// Create a DNS Record
	Create(context.Context, phonebook.DNSRecord, phonebook.StagingUpdater) error

	// Delete a DNS Record
	Delete(context.Context, phonebook.DNSRecord, phonebook.StagingUpdater) error

	// Zones the Provider has authority over
	Zones() []string
}

var ProviderImages = map[string]string{
	"aws":        fmt.Sprintf("ghcr.io/pier-oliviert/providers-aws:v%s", ProviderVersion),
	"azure":      fmt.Sprintf("ghcr.io/pier-oliviert/providers-azure:v%s", ProviderVersion),
	"cloudflare": fmt.Sprintf("ghcr.io/pier-oliviert/providers-cloudflare:v%s", ProviderVersion),
	"desec":      fmt.Sprintf("ghcr.io/pier-oliviert/providers-desec:v%s", ProviderVersion),
	"gcore":      fmt.Sprintf("ghcr.io/pier-oliviert/providers-gcore:v%s", ProviderVersion),
}
