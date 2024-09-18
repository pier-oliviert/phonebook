package cloudflare

import (
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient()
	if !strings.HasPrefix(err.Error(), "PB#0100: API Key not found --") {
		t.Error("Client should require an API Key")
	}

	os.Setenv("CF_API_TOKEN", "Some Value")

	_, err = NewClient()
	if !strings.HasPrefix(err.Error(), "PB#0100: Zone ID not found --") {
		t.Error("Client should require a Zone ID")
	}

}
