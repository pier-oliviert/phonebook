package provider

import (
	"context"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

type Provider interface {
	Create(context.Context, *phonebook.DNSRecord) error
	Delete(context.Context, *phonebook.DNSRecord) error
}
