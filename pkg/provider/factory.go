package provider

import (
	"context"
	"fmt"
	"log"

	aws "github.com/pier-oliviert/phonebook/pkg/aws"
	"github.com/pier-oliviert/phonebook/pkg/cloudflare"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
)

const kPhonebookProvider = "PHONEBOOK_PROVIDER"

func NewProvider(ctx context.Context, name string) (Provider, error) {
	log.Default().Print("Initializing provider: ", name)

	switch name {
	case "aws":
		return aws.NewClient(ctx)
	case "cloudflare":
		return cloudflare.NewClient(ctx)
	case "":
		return nil, fmt.Errorf("PB#0001: The environment variable %s need to be set with a valid provider name", kPhonebookProvider)
	}

	return nil, fmt.Errorf("PB#0001: The environment variable %s need to be set with a valid provider name, got %s", kPhonebookProvider, name)
}

// Same as NewProvider but throw a fatal exception if
// the configuration settings can't initialize a provider.
func DefaultProvider() Provider {
	value, err := utils.RetrieveValueFromEnvOrFile(kPhonebookProvider)
	if err != nil {
		log.Fatal(err)
	}

	p, err := NewProvider(context.Background(), value)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
