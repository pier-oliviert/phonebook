package rfc2136

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"
	phonebook "github.com/pier-oliviert/phonebook/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDNSClient is a mock for the DNS client
type MockDNSClient struct {
	mock.Mock
}

func (m *MockDNSClient) Exchange(msg *dns.Msg, addr string) (*dns.Msg, time.Duration, error) {
    args := m.Called(msg, addr)
    return args.Get(0).(*dns.Msg), args.Get(1).(time.Duration), args.Error(2)
}

func (m *MockDNSClient) SetTsigSecret(secret map[string]string) {
    m.Called(secret)
}

func TestNewClient(t *testing.T) {
    // Set up test environment variables
    os.Setenv(kRFC2136Server, "dns.example.com")
    os.Setenv(kRFC2136Port, "53")
    os.Setenv(kRFC2136ZoneName, "example.com")
    os.Setenv(kRFC2136Insecure, "false")
    os.Setenv(kRFC2136Keyname, "keyname")
    os.Setenv(kRFC2136Secret, "secret")
    os.Setenv(kRFC2136SecretAlg, "hmac-sha256")

    // Clean up environment variables after the test
    defer func() {
        os.Unsetenv(kRFC2136Server)
        os.Unsetenv(kRFC2136Port)
        os.Unsetenv(kRFC2136ZoneName)
        os.Unsetenv(kRFC2136Insecure)
        os.Unsetenv(kRFC2136Keyname)
        os.Unsetenv(kRFC2136Secret)
        os.Unsetenv(kRFC2136SecretAlg)
    }()

    client, err := NewClient(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, client)
    assert.Equal(t, "dns.example.com", client.server)
    assert.Equal(t, int32(53), client.port)
    assert.Equal(t, "example.com", client.zoneName)
    assert.False(t, client.insecure)
    assert.Equal(t, "keyname", client.keyname)
    assert.Equal(t, "secret", client.secret)
    assert.Equal(t, "hmac-sha256", client.secretAlg)
}

func TestNewClientMissingServer(t *testing.T) {
	// Set up test environment variables
	os.Setenv(kRFC2136Port, "53")
	os.Setenv(kRFC2136ZoneName, "example.com")
	os.Setenv(kRFC2136Insecure, "false")
	os.Setenv(kRFC2136Keyname, "keyname")
	os.Setenv(kRFC2136Secret , "secret")
	os.Setenv(kRFC2136SecretAlg, "hmac")

	// Check we get an error when the server is missing
	client, err := NewClient(context.Background())
	assert.Error(t, err)
	assert.Nil(t, client)
}

// Check that the client uses port 53 by default if the port is not set
func TestNewClientMissingPort(t *testing.T) {
	// Set up test environment variables
	os.Setenv(kRFC2136Server, "dns.example.com")
	os.Setenv(kRFC2136ZoneName, "example.com")
	os.Setenv(kRFC2136Insecure, "false")
	os.Setenv(kRFC2136Keyname, "keyname")
	os.Setenv(kRFC2136Secret, "secret")
	os.Setenv(kRFC2136SecretAlg, "hmac")

	client, err := NewClient(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, int32(53), client.port)

	// Clean up environment variables after the test
	defer func() {
		os.Unsetenv(kRFC2136Server)
		os.Unsetenv(kRFC2136ZoneName)
		os.Unsetenv(kRFC2136Insecure)
		os.Unsetenv(kRFC2136Keyname)
		os.Unsetenv(kRFC2136Secret)
		os.Unsetenv(kRFC2136SecretAlg)
	}()
}

func TestNewClientMissingZoneName(t *testing.T) {
	// Set up test environment variables
	os.Setenv(kRFC2136Server, "dns.example.com")
	os.Setenv(kRFC2136Port, "53")
	os.Setenv(kRFC2136Insecure, "false")
	os.Setenv(kRFC2136Keyname, "keyname")
	os.Setenv(kRFC2136Secret, "secret")
	os.Setenv(kRFC2136SecretAlg, "hmac")

	client, err := NewClient(context.Background())
	assert.Error(t, err)
	assert.Nil(t, client)

	// Clean up environment variables after the test
	defer func() {
		os.Unsetenv(kRFC2136Server)
		os.Unsetenv(kRFC2136Port)
		os.Unsetenv(kRFC2136Insecure)
		os.Unsetenv(kRFC2136Keyname)
		os.Unsetenv(kRFC2136Secret)
		os.Unsetenv(kRFC2136SecretAlg)
	}()
}
	
// Check that we get a warning when the insecure flag is set to true
func TestNewClientInsecure(t *testing.T) {
	// Set up test environment variables
	os.Setenv(kRFC2136Server, "dns.example.com")
	os.Setenv(kRFC2136Port, "53")
	os.Setenv(kRFC2136ZoneName, "example.com")
	os.Setenv(kRFC2136Insecure, "true")

	client, err := NewClient(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.insecure)

	// Clean up environment variables after the test
	defer func() {
		os.Unsetenv(kRFC2136Server)
		os.Unsetenv(kRFC2136Port)
		os.Unsetenv(kRFC2136ZoneName)
		os.Unsetenv(kRFC2136Insecure)
	}()
}

func TestPerformSecureUpdate(t *testing.T) {
    // Set up test environment variables
    os.Setenv(kRFC2136Server, "dns.example.com")
    os.Setenv(kRFC2136Port, "53")
    os.Setenv(kRFC2136ZoneName, "example.com")
    os.Setenv(kRFC2136Insecure, "false")
    os.Setenv(kRFC2136Keyname, "keyname")
    os.Setenv(kRFC2136Secret, "secret")
    os.Setenv(kRFC2136SecretAlg, "hmac-sha256")

    // Clean up environment variables after the test
    defer func() {
        os.Unsetenv(kRFC2136Server)
        os.Unsetenv(kRFC2136Port)
        os.Unsetenv(kRFC2136ZoneName)
        os.Unsetenv(kRFC2136Insecure)
        os.Unsetenv(kRFC2136Keyname)
        os.Unsetenv(kRFC2136Secret)
        os.Unsetenv(kRFC2136SecretAlg)
    }()

    // Create a mock DNS client
    mockClient := new(MockDNSClient)
    
    // Set up expectations with the exact map we expect
    expectedSecret := map[string]string{"keyname.": "secret"}
    mockClient.On("SetTsigSecret", expectedSecret).Return()
    
    mockClient.On("Exchange", mock.Anything, mock.Anything).Return(new(dns.Msg), time.Second, nil)

    // Create a new RFC2136 client
    client, err := NewClient(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, client)

    // Replace the real DNS client with our mock
    client.dnsClient = mockClient

    // Create a new DNS record 
    record := &phonebook.DNSRecord{
        Spec: phonebook.DNSRecordSpec{
            Name:    "test",
            Targets: []string{"127.0.0.1"},
        },
    }

    // Perform the secure update
    err = client.performSecureUpdate(record, "example.com")
    assert.NoError(t, err)

    // Assert that the SetTsigSecret and Exchange methods were called
    mockClient.AssertCalled(t, "SetTsigSecret", expectedSecret)
    mockClient.AssertCalled(t, "Exchange", mock.Anything, mock.Anything)
}

func TestPerformSecureDelete(t *testing.T) {
    // Set up test environment variables
    os.Setenv(kRFC2136Server, "dns.example.com")
    os.Setenv(kRFC2136Port, "53")
    os.Setenv(kRFC2136ZoneName, "example.com")
    os.Setenv(kRFC2136Insecure, "false")
    os.Setenv(kRFC2136Keyname, "keyname")
    os.Setenv(kRFC2136Secret, "secret")
    os.Setenv(kRFC2136SecretAlg, "hmac-sha256")

    // Clean up environment variables after the test
    defer func() {
        os.Unsetenv(kRFC2136Server)
        os.Unsetenv(kRFC2136Port)
        os.Unsetenv(kRFC2136ZoneName)
        os.Unsetenv(kRFC2136Insecure)
        os.Unsetenv(kRFC2136Keyname)
        os.Unsetenv(kRFC2136Secret)
        os.Unsetenv(kRFC2136SecretAlg)
    }()

    // Create a mock DNS client
    mockClient := new(MockDNSClient)
    
    // Set up expectations with the exact map we expect
    expectedSecret := map[string]string{"keyname.": "secret"}
    mockClient.On("SetTsigSecret", expectedSecret).Return()
    
    // Set up expectation for the Exchange method
    mockClient.On("Exchange", mock.Anything, mock.Anything).Return(new(dns.Msg), time.Second, nil)

    // Create a new RFC2136 client
    client, err := NewClient(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, client)

    // Replace the real DNS client with our mock
    client.dnsClient = mockClient

    // Create a new DNS record to delete
    record := &phonebook.DNSRecord{
        Spec: phonebook.DNSRecordSpec{
            Name:    "test.example.com",
            Targets: []string{"127.0.0.1"},
        },
    }

    // Perform the secure delete
    err = client.performSecureDelete(record, "example.com")
    assert.NoError(t, err)

    // Assert that the SetTsigSecret method was called with the expected secret
    mockClient.AssertCalled(t, "SetTsigSecret", expectedSecret)

    // Assert that the Exchange method was called
    mockClient.AssertCalled(t, "Exchange", mock.Anything, mock.Anything)

	// Assert that the DNS record was deleted
	mockClient.AssertExpectations(t)

	// Clean up environment variables after the test
	defer func() {
		os.Unsetenv(kRFC2136Server)
		os.Unsetenv(kRFC2136Port)
		os.Unsetenv(kRFC2136ZoneName)
		os.Unsetenv(kRFC2136Insecure)
		os.Unsetenv(kRFC2136Keyname)
		os.Unsetenv(kRFC2136Secret)
		os.Unsetenv(kRFC2136SecretAlg)
	}()


    
}