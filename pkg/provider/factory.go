package provider

import (
	"fmt"
	"log"

	aws "github.com/pier-oliviert/phonebook/pkg/aws"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
)

const kPhoneBookProvider = "PHONEBOOK_PROVIDER"

func NewProvider(name string) (Provider, error) {
	switch name {
	case "aws":
		return aws.NewClient()
	case "":
		return nil, fmt.Errorf("E#4001: The environment variable %s need to be set with a valid provider name", kPhoneBookProvider)
	}

	return nil, fmt.Errorf("E#4001: The environment variable %s need to be set with a valid provider name, got %s", kPhoneBookProvider, name)
}

// Same as NewProvider but throw a fatal exception if
// the configuration settings can't initialize a provider.
func DefaultProvider() Provider {
	value, err := utils.RetrieveValueFromEnvOrFile(kPhoneBookProvider)
	if err != nil {
		log.Fatal(err)
	}

	p, err := NewProvider(value)
	if err != nil {
		log.Fatal(err)
	}
	return p
}
