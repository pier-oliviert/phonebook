package rfc2136

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

func (c *rfc2136DNS) performInsecureUpdate(record *phonebook.DNSRecord, zoneName string) error {
	// Ensure that the zoneName ends with a dot
	if zoneName[len(zoneName)-1] != '.' {
		zoneName = zoneName + "."
	}

	// Prepare the DNS message for the insecure update
	msg := new(dns.Msg)
	msg.SetUpdate(zoneName)

	// Construct the fully qualified domain name (FQDN) for the record
	// If the record name doesn't already end with a dot, append the zone to it
	recordName := record.Spec.Name
	if !strings.HasSuffix(recordName, ".") {
		recordName = fmt.Sprintf("%s.%s", recordName, zoneName)
	}

	// Create the DNS RR (resource record) for the A record
	rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", recordName, c.defaultTTL, record.Spec.Targets[0]))
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0010: Failed to create DNS RR: %w", err)
	}

	// Add the new record to the message
	msg.Insert([]dns.RR{rr})

	// Send the update to the RFC2136 server
	client := new(dns.Client)
	serverAddr := fmt.Sprintf("%s:%d", c.server, c.port)
	_, _, err = client.Exchange(msg, serverAddr)
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0011: Insecure DNS update failed: %w", err)
	}

	return nil
}
