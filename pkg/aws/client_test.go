package aws

import (
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	if !strings.HasPrefix(err.Error(), "PB#0100: Zone ID not found --") {
		t.Error("Client should require a Zone ID")
	}

	os.Setenv("AWS_ZONE_ID", "Some Value")

	_, err = NewClient()
	if !strings.HasPrefix(err.Error(), "PB#0100: Hosted Zone ID not found --") {
		t.Error("Client should require a Hosted Zone ID")
	}
	os.Setenv("AWS_HOSTED_ZONE_ID", "Some Value")

	_, err = NewClient()
	if !strings.HasPrefix(err.Error(), "PB#0100: Load balancer host not found --") {
		t.Error("Client should require a Load Balancer host")
	}
}
