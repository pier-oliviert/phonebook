package providers

import (
	"context"
	"sync"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

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
	Create(context.Context, *phonebook.DNSRecord) error

	// Delete a DNS Record
	Delete(context.Context, *phonebook.DNSRecord) error

	// Zones the Provider has authority over
	Zones() []string
}
