package openfga

import (
	"net/http"
	"testing"
	"time"

	"github.com/openfga/go-sdk/credentials"
)

// TestAPIClientWithCustomHTTPClientAndClientCredentials verifies that when both
// a custom HTTPClient and ClientCredentials are provided, both are honored.
func TestAPIClientWithCustomHTTPClientAndClientCredentials(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	cfg := Configuration{
		ApiUrl: "http://localhost:8080/",
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       "test-client-id",
				ClientCredentialsClientSecret:   "test-client-secret",
				ClientCredentialsApiAudience:    "https://api.test.example/",
				ClientCredentialsApiTokenIssuer: "issuer.test.example",
			},
		},
		HTTPClient: customClient,
	}

	apiClient := NewAPIClient(&cfg)

	// Verify that credentials were processed (HTTPClient should be different)
	if apiClient.cfg.HTTPClient == customClient {
		t.Fatal("Expected HTTPClient to be wrapped with OAuth2 transport, but it's the same as the custom client")
	}

	// Verify that custom client settings were preserved
	if apiClient.cfg.HTTPClient.Timeout != customClient.Timeout {
		t.Fatalf("Expected HTTPClient timeout to be %v, got %v",
			customClient.Timeout, apiClient.cfg.HTTPClient.Timeout)
	}
}

// TestAPIClientWithCustomHTTPClientAndApiToken verifies that when both
// a custom HTTPClient and ApiToken credentials are provided, both are honored.
func TestAPIClientWithCustomHTTPClientAndApiToken(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	cfg := Configuration{
		ApiUrl:         "http://localhost:8080/",
		DefaultHeaders: make(map[string]string),
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: "test-api-token",
			},
		},
		HTTPClient: customClient,
	}

	apiClient := NewAPIClient(&cfg)

	// For ApiToken, the HTTPClient should be the same custom client
	if apiClient.cfg.HTTPClient != customClient {
		t.Fatal("Expected HTTPClient to be the same as the custom client for ApiToken method")
	}

	// Verify that the authorization header was added
	authHeader, exists := apiClient.cfg.DefaultHeaders["Authorization"]
	if !exists {
		t.Fatal("Expected Authorization header to be added, but it's missing")
	}

	expectedAuthHeader := "Bearer test-api-token"
	if authHeader != expectedAuthHeader {
		t.Fatalf("Expected Authorization header to be %q, got %q",
			expectedAuthHeader, authHeader)
	}
}

// TestAPIClientWithCredentialsOnly verifies that credentials work without a custom HTTPClient
func TestAPIClientWithCredentialsOnly(t *testing.T) {
	cfg := Configuration{
		ApiUrl: "http://localhost:8080/",
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       "test-client-id",
				ClientCredentialsClientSecret:   "test-client-secret",
				ClientCredentialsApiAudience:    "https://api.test.example/",
				ClientCredentialsApiTokenIssuer: "issuer.test.example",
			},
		},
	}

	apiClient := NewAPIClient(&cfg)

	// Verify that HTTPClient was created
	if apiClient.cfg.HTTPClient == nil {
		t.Fatal("Expected HTTPClient to be created, but it's nil")
	}
}

// TestAPIClientWithCustomHTTPClientOnly verifies that a custom HTTPClient works without credentials
func TestAPIClientWithCustomHTTPClientOnly(t *testing.T) {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	cfg := Configuration{
		ApiUrl:     "http://localhost:8080/",
		HTTPClient: customClient,
	}

	apiClient := NewAPIClient(&cfg)

	// Verify that the custom client is used as-is
	if apiClient.cfg.HTTPClient != customClient {
		t.Fatal("Expected HTTPClient to be the same as the custom client")
	}
}

// TestAPIClientWithoutCredentialsAndHTTPClient verifies that default HTTPClient is used
func TestAPIClientWithoutCredentialsAndHTTPClient(t *testing.T) {
	cfg := Configuration{
		ApiUrl: "http://localhost:8080/",
	}

	apiClient := NewAPIClient(&cfg)

	// Verify that HTTPClient was set to default
	if apiClient.cfg.HTTPClient == nil {
		t.Fatal("Expected HTTPClient to be set to default, but it's nil")
	}
}
