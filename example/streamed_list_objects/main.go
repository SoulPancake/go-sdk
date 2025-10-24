package main

import (
	"context"
	"fmt"
	"log"
	"os"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

func main() {
	apiUrl := os.Getenv("FGA_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:8080"
	}

	storeId := os.Getenv("FGA_STORE_ID")
	if storeId == "" {
		log.Fatal("FGA_STORE_ID environment variable is required")
	}

	config := client.ClientConfiguration{
		ApiUrl:  apiUrl,
		StoreId: storeId,
	}

	fgaClient, err := client.NewSdkClient(&config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("OpenFGA StreamedListObjects Example")
	fmt.Println("====================================")
	fmt.Printf("API URL: %s\n", apiUrl)
	fmt.Printf("Store ID: %s\n", storeId)
	fmt.Println()

	authModelId := os.Getenv("FGA_MODEL_ID")
	if authModelId != "" {
		fmt.Printf("Authorization Model ID: %s\n\n", authModelId)
	}

	ctx := context.Background()

	if err := createTestData(ctx, fgaClient, authModelId); err != nil {
		log.Printf("Warning: Failed to create test data: %v", err)
		log.Println("Continuing with example...")
	}

	fmt.Println("Calling StreamedListObjects...")
	fmt.Println()

	request := client.ClientStreamedListObjectsRequest{
		Type:     "document",
		Relation: "viewer",
		User:     "user:anne",
	}

	options := client.ClientStreamedListObjectsOptions{}
	if authModelId != "" {
		options.AuthorizationModelId = &authModelId
	}

	response, err := fgaClient.StreamedListObjects(ctx).
		Body(request).
		Options(options).
		Execute()

	if err != nil {
		log.Fatalf("StreamedListObjects failed: %v", err)
	}

	defer response.Close()

	fmt.Println("Streaming objects:")
	count := 0
	for obj := range response.Objects {
		count++
		fmt.Printf("  %d. %s\n", count, obj.Object)
	}

	if err := <-response.Errors; err != nil {
		log.Fatalf("Error during streaming: %v", err)
	}

	fmt.Printf("\nTotal objects received: %d\n", count)

	if count == 0 {
		fmt.Println("\nNote: No objects found. This might be because:")
		fmt.Println("  1. The store has no data")
		fmt.Println("  2. The authorization model is not set up")
		fmt.Println("  3. user:anne has no viewer relationship with any documents")
		fmt.Println("\nYou can add test data using the Write API first.")
	}
}

func createTestData(ctx context.Context, fgaClient *client.OpenFgaClient, authModelId string) error {
	if authModelId == "" {
		fmt.Println("Creating authorization model...")
		
		viewerRelations := map[string]openfga.Userset{
			"viewer": {
				This: &map[string]interface{}{},
			},
		}
		
		viewerMetadata := map[string]openfga.RelationMetadata{
			"viewer": {
				DirectlyRelatedUserTypes: &[]openfga.RelationReference{
					{
						Type: "user",
					},
				},
			},
		}
		
		model := openfga.AuthorizationModel{
			SchemaVersion: "1.1",
			TypeDefinitions: []openfga.TypeDefinition{
				{
					Type: "user",
				},
				{
					Type:      "document",
					Relations: &viewerRelations,
					Metadata: &openfga.Metadata{
						Relations: &viewerMetadata,
					},
				},
			},
		}

		writeModelResp, err := fgaClient.WriteAuthorizationModel(ctx).
			Body(client.ClientWriteAuthorizationModelRequest{
				SchemaVersion:   model.SchemaVersion,
				TypeDefinitions: model.TypeDefinitions,
			}).
			Execute()

		if err != nil {
			return fmt.Errorf("failed to create authorization model: %w", err)
		}

		authModelId = writeModelResp.AuthorizationModelId
		fmt.Printf("Created authorization model: %s\n\n", authModelId)
	}

	fmt.Println("Writing test tuples...")
	
	tuples := []client.ClientTupleKey{
		{
			User:     "user:anne",
			Relation: "viewer",
			Object:   "document:1",
		},
		{
			User:     "user:anne",
			Relation: "viewer",
			Object:   "document:2",
		},
		{
			User:     "user:anne",
			Relation: "viewer",
			Object:   "document:3",
		},
	}

	_, err := fgaClient.WriteTuples(ctx).
		Body(tuples).
		Execute()

	if err != nil {
		return fmt.Errorf("failed to write tuples: %w", err)
	}

	fmt.Printf("Wrote %d test tuples\n\n", len(tuples))
	return nil
}
