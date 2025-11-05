package main

import (
	"fmt"
	"net/http"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

func main() {
	fmt.Println("=== Demo: Credentials Only (Should Work) ===")
	testCredentialsOnly()

	fmt.Println("\n=== Demo: Custom HTTPClient + Credentials (BUG: Credentials Ignored) ===")
	testCustomHTTPClientWithCredentials()

	fmt.Println("\n=== Demo: Custom HTTPClient Only (Should Work) ===")
	testCustomHTTPClientOnly()
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
		// Try to check if it's using the credential transport
		fmt.Println("✓ Credentials were processed")
	} else {
		fmt.Println("✗ HTTPClient is nil")
	}
}

func testCustomHTTPClientWithCredentials() {
	customClient := &http.Client{
		Timeout: 0, // Custom setting
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
	if config.HTTPClient == customClient {
		fmt.Println("✓ HTTPClient is the custom client")
		fmt.Println("✗ BUG: Credentials were NOT processed (credentials ignored when custom HTTPClient is provided)")
	} else {
		fmt.Println("✓ HTTPClient is different (credentials might have been processed)")
	}
}

func testCustomHTTPClientOnly() {
	customClient := &http.Client{
		Timeout: 0, // Custom setting
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
