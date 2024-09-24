package aws

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(context.TODO())
	if !strings.HasPrefix(err.Error(), "PB#0100: Zone ID not found --") {
		t.Error("Client should require a Zone ID")
	}

	os.Setenv("AWS_ZONE_ID", "Some Value")
}
