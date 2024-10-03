package rfc2136

import (
	"context"
	"fmt"
	"time"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	utils "github.com/pier-oliviert/phonebook/pkg/utils"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kRFC2136Server    = "RFC2136_SERVER"
	kRFC2136Port      = "RFC2136_PORT"
	kRFC2136ZoneName  = "RFC2136_ZONE_NAME"
	kRFC2136Insecure  = "RFC2136_INSECURE"
	kRFC2136Keyname   = "RFC2136_KEYNAME"
	kRFC2136Secret    = "RFC2136_SECRET"
	kRFC2136SecretAlg = "RFC2136_SECRET_ALG"
	defaultTTL        = int64(60) // Default TTL for DNS records in seconds if not specified
	retryInterval     = 1 * time.Second
	retryCount        = 3
)

type rfc2136DNS struct {
	server     string
	port       int32
	zoneName   string
	keyname    string
	secret     string
	secretAlg  string
	insecure   bool
	defaultTTL int64
	client     client.Client
}

// NewClient initializes an RFC2136 DNS client
func NewClient(ctx context.Context) (*rfc2136DNS, error) {
	logger := log.FromContext(ctx)

	// First lets get our environment variables and make sure they are set properly
	server, err := utils.RetrieveValueFromEnvOrFile(kRFC2136Server)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0001: RFC2136 Server not found -- %w", err)
	}

	portEnv, err := utils.RetrieveValueFromEnvOrFile(kRFC2136Port)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0002: RFC2136 Port not found -- %w", err)
	}

	// Convert port to int32 because the env-var is a string
	port := utils.ToInt32Ptr(portEnv)

	zoneName, err := utils.RetrieveValueFromEnvOrFile(kRFC2136ZoneName)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0003: RFC2136 Zone Name not found -- %w", err)
	}

	insecure, err := utils.RetrieveValueFromEnvOrFile(kRFC2136Insecure)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0004: RFC2136 Insecure not found -- %w", err)
	}

	// if insecure is set, we can ignore the keyname and secret values
	// Using insecure mode is not recommended so we will log a warning just in case it's an accident
	if insecure == "true" {
		logger.Info(("WARNING: RFC2136 INSECURE MODE ENABLED"))
		logger.Info("[Provider] RFC2136 Configured", "Zone Name", zoneName, "Server", server, "Port", port)
		return &rfc2136DNS{
			server:     server,
			port:       *port,
			zoneName:   zoneName,
			insecure:   true,
			defaultTTL: defaultTTL,
		}, nil
	}

	keyname, err := utils.RetrieveValueFromEnvOrFile(kRFC2136Keyname)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0005: RFC2136 Keyname not found -- %w", err)
	}

	secret, err := utils.RetrieveValueFromEnvOrFile(kRFC2136Secret)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0006: RFC2136 Secret not found -- %w", err)
	}

	secretAlg, err := utils.RetrieveValueFromEnvOrFile(kRFC2136SecretAlg)
	if err != nil {
		return nil, fmt.Errorf("PB-RFC2136-#0007: RFC2136 Secret Alg not found -- %w", err)
	}

	logger.Info("[Provider] RFC2136 Configured", "Zone Name", zoneName, "Server", server, "Port", port)

	return &rfc2136DNS{
		server:     server,
		port:       *port,
		zoneName:   zoneName,
		keyname:    keyname,
		secret:     secret,
		secretAlg:  secretAlg,
		insecure:   false,
		defaultTTL: defaultTTL,
	}, nil
}

// Create DNS record in RFC2136
func (c *rfc2136DNS) Create(ctx context.Context, record *phonebook.DNSRecord) error {
	logger := log.FromContext(ctx)

	zoneName := c.zoneName
	if record.Spec.Zone != "" {
		zoneName = record.Spec.Zone
	}

	// Ensure that zoneName ends with a dot for proper DNS formatting
	if zoneName[len(zoneName)-1] != '.' {
		zoneName = zoneName + "."
	}

	// Retry logic for handling conflicts
	for i := 0; i < retryCount; i++ {
		err := c.createDNSRecord(ctx, record, zoneName)
		if err == nil {
			return nil
		}

		// Handle conflict errors (409)
		if kerrors.IsConflict(err) {
			logger.Info("Conflict encountered while updating DNS record, retrying...", "attempt", i+1)
			time.Sleep(retryInterval)

			// Re-fetch the latest version of the record
			err = c.reloadDNSRecord(ctx, record)
			if err != nil {
				return fmt.Errorf("PB-RFC2136-#0008: Failed to reload DNS record after conflict: %w", err)
			}

			continue // Retry the operation
		}

		// If it's not a conflict error, return it
		return err
	}

	// If retries exhausted, return error
	return fmt.Errorf("PB-RFC2136-#0009: Failed to create DNS record after %d retries due to conflict", retryCount)
}


// createDNSRecord is the function that performs the actual DNS record creation
func (c *rfc2136DNS) createDNSRecord(ctx context.Context, record *phonebook.DNSRecord, zoneName string) error {
	logger := log.FromContext(ctx)

	// Check if insecure mode is enabled
	if c.insecure {
		logger.Info("Performing insecure DNS update", "Record", record.Spec.Name)
		return c.performInsecureUpdate(record, zoneName)
	}

	// Secure update
	logger.Info("Performing secure DNS update", "Record", record.Spec.Name)
	return c.performSecureUpdate(record, zoneName)
}

// Placeholder for delete function
func (c *rfc2136DNS) Delete(ctx context.Context, record *phonebook.DNSRecord) error {
	return nil
}
