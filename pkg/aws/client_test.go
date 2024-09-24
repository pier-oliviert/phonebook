package aws

import (
	"context"
	"os"
	"strings"
	"testing"

	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(context.TODO())
	if !strings.HasPrefix(err.Error(), "PB#0100: Zone ID not found --") {
		t.Error("Client should require a Zone ID")
	}

	os.Setenv("AWS_ZONE_ID", "Some Value")
}

func TestDNSNameConcatenation(t *testing.T) {
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:    "mydomain.com",
			Name:    "subdomain",
			Targets: []string{"127.0.0.1"},
		},
	}

	c := &r53{
		zoneID: "MyZone123",
	}

	set := c.resourceRecordSet(&record)

	if *set.Name != "subdomain.mydomain.com" {
		t.Error("Expected name to include both zone and name", "Name", set.Name)
	}
}

func TestAliastTargetProperty(t *testing.T) {
	record := phonebook.DNSRecord{
		Spec: phonebook.DNSRecordSpec{
			Zone:    "mydomain.com",
			Name:    "subdomain",
			Targets: []string{"127.0.0.1"},
			Properties: map[string]string{
				AliasTarget: "myTargetZoneID",
			},
		},
	}

	c := &r53{
		zoneID: "MyZone123",
	}

	set := c.resourceRecordSet(&record)

	if len(set.ResourceRecords) > 0 {
		t.Error("Expected record set to not have any resource records when using AliasTarget", "ResourceRecords", set.ResourceRecords)
	}

	if set.AliasTarget == nil {
		t.Error("Expected alias target to be set when using AliasTarget")
	}

	if *set.AliasTarget.DNSName != record.Spec.Targets[0] {
		t.Error("Expected alias target DNSNAme to be set to the target", "Targets", record.Spec.Targets)
	}

	if *set.AliasTarget.HostedZoneId != record.Spec.Properties[AliasTarget] {
		t.Error("Expected alias hosted zone id to be set to the alias target property", "Properties", record.Spec.Properties)
	}
}
