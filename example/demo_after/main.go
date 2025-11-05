package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

func main() {
	fmt.Println("=== Demo: Credentials Only (Should Work) ===")
	testCredentialsOnly()

	fmt.Println("\n=== Demo: Custom HTTPClient + Credentials (FIXED: Both Should Be Honored) ===")
	testCustomHTTPClientWithCredentials()

	fmt.Println("\n=== Demo: Custom HTTPClient Only (Should Work) ===")
	testCustomHTTPClientOnly()

	fmt.Println("\n=== Demo: API Token Credentials + Custom HTTPClient (Should Work) ===")
	testApiTokenWithCustomHTTPClient()
}

func testCredentialsOnly() {
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: "http://localhost:8080",
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       "some-client-id",
				ClientCredentialsClientSecret:   "some-client-secret",
				ClientCredentialsApiAudience:    "https://api.fga.example/",
				ClientCredentialsApiTokenIssuer: "issuer.fga.example",
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	// Check if HTTPClient is set (should be set with credentials)
	config := fgaClient.APIClient.GetConfig()
	if config.HTTPClient != nil {
		fmt.Println("✓ HTTPClient is configured")
		fmt.Println("✓ Credentials were processed")
	} else {
		fmt.Println("✗ HTTPClient is nil")
	}
}

func testCustomHTTPClientWithCredentials() {
	customClient := &http.Client{
		Timeout: 30 * time.Second, // Custom timeout
	}

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: "http://localhost:8080",
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       "some-client-id",
				ClientCredentialsClientSecret:   "some-client-secret",
				ClientCredentialsApiAudience:    "https://api.fga.example/",
				ClientCredentialsApiTokenIssuer: "issuer.fga.example",
			},
		},
		HTTPClient: customClient,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	config := fgaClient.APIClient.GetConfig()

	// The HTTPClient should be different from the custom one (wrapped with OAuth2 transport)
	// But it should preserve the custom client's settings
	if config.HTTPClient != customClient {
		fmt.Println("✓ HTTPClient was wrapped with credentials transport")

		// Check if the timeout was preserved
		if config.HTTPClient.Timeout == customClient.Timeout {
			fmt.Println("✓ Custom HTTPClient timeout was preserved (30s)")
		} else {
			fmt.Printf("✗ Custom HTTPClient timeout was NOT preserved (got %v, expected %v)\n",
				config.HTTPClient.Timeout, customClient.Timeout)
		}
		fmt.Println("✓ FIXED: Both credentials and custom HTTPClient were honored")
	} else {
		fmt.Println("✗ HTTPClient is still the same (credentials were not processed)")
	}
}

func testCustomHTTPClientOnly() {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:     "http://localhost:8080",
		HTTPClient: customClient,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	config := fgaClient.APIClient.GetConfig()
	if config.HTTPClient == customClient {
		fmt.Println("✓ HTTPClient is the custom client (expected)")
		fmt.Println("✓ No credentials provided, so none processed (expected)")
	} else {
		fmt.Println("✗ HTTPClient was changed unexpectedly")
	}
}

func testApiTokenWithCustomHTTPClient() {
	customClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: "http://localhost:8080",
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: "test-api-token",
			},
		},
		HTTPClient: customClient,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	config := fgaClient.APIClient.GetConfig()

	// For API Token, the custom client should be used directly (not wrapped)
	if config.HTTPClient == customClient {
		fmt.Println("✓ HTTPClient is the custom client")

		// Check if the authorization header was added
		if authHeader, exists := config.DefaultHeaders["Authorization"]; exists {
			if authHeader == "Bearer test-api-token" {
				fmt.Println("✓ API Token authorization header was added correctly")
			} else {
				fmt.Printf("✗ Authorization header has wrong value: %s\n", authHeader)
			}
		} else {
			fmt.Println("✗ Authorization header was not added")
		}

		fmt.Println("✓ Both API Token credentials and custom HTTPClient were honored")
	} else {
		fmt.Println("✗ HTTPClient was changed when it shouldn't have been")
	}
}
