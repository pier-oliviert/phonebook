package rfc2136

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

func (c *rfc2136DNS) performSecureUpdate(record *phonebook.DNSRecord, zoneName string) error {
	// Prepare the DNS message for the secure update
	msg := new(dns.Msg)
	msg.SetUpdate(zoneName)

	// Create the DNS RR (resource record) for the A record
	rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", record.Spec.Name, c.defaultTTL, record.Spec.Targets[0]))
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0008: Failed to create DNS RR: %w", err)
	}

	// Add the new record
	msg.Insert([]dns.RR{rr})

	// Add TSIG for secure updates
	msg.SetTsig(c.keyname+".", c.secretAlg, 300, time.Now().Unix())

	// Set TSIG secret
    c.dnsClient.SetTsigSecret(map[string]string{c.keyname + ".": c.secret})

	serverAddr := fmt.Sprintf("%s:%d", c.server, c.port)
	_, _, err = c.dnsClient.Exchange(msg, serverAddr)
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0009: Secure DNS update failed: %w", err)
	}

	return nil
}


// implement a delete function
func (c *rfc2136DNS) performSecureDelete(record *phonebook.DNSRecord, zoneName string) error {
	// Prepare the DNS message for the delete operation
	msg := new(dns.Msg)
	msg.SetUpdate(zoneName)

	// Create the DNS RR (resource record) for the A record with TTL 0
	rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", record.Spec.Name, 0, record.Spec.Targets[0]))
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0010: Failed to create DNS RR: %w", err)
	}

	// Add the new record
	msg.Insert([]dns.RR{rr})

	// Add TSIG for secure updates
	msg.SetTsig(c.keyname+".", c.secretAlg, 300, time.Now().Unix())

	// Set TSIG secret
    c.dnsClient.SetTsigSecret(map[string]string{c.keyname + ".": c.secret})
	
	// Send the update to the RFC2136 server
	serverAddr := fmt.Sprintf("%s:%d", c.server, c.port)
	_, _, err = c.dnsClient.Exchange(msg, serverAddr)
	if err != nil {
		return fmt.Errorf("PB-RFC2136-#0011: Secure DNS delete failed: %w", err)
	}

	return nil
}