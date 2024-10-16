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

