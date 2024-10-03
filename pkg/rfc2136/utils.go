package rfc2136

import (
	"context"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *rfc2136DNS) reloadDNSRecord(ctx context.Context, record *phonebook.DNSRecord) error {
	logger := log.FromContext(ctx)

	// Re-fetch the latest version of the DNS record using the Kubernetes client
	if err := c.client.Get(ctx, client.ObjectKey{Name: record.Name, Namespace: record.Namespace}, record); err != nil {
		logger.Error(err, "Failed to reload DNS record")
		return err
	}

	return nil
}