package main

import (
	"context"
	"fmt"
	"os"
	"time"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	creds := credentials.Credentials{}
	if os.Getenv("FGA_CLIENT_ID") != "" {
		creds = credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       os.Getenv("FGA_CLIENT_ID"),
				ClientCredentialsClientSecret:   os.Getenv("FGA_CLIENT_SECRET"),
				ClientCredentialsApiAudience:    os.Getenv("FGA_API_AUDIENCE"),
				ClientCredentialsApiTokenIssuer: os.Getenv("FGA_API_TOKEN_ISSUER"),
			},
		}
	}

	apiUrl := os.Getenv("FGA_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:8080"
	}

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:      apiUrl,
		StoreId:     os.Getenv("FGA_STORE_ID"),
		Credentials: &creds,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	storeId, err := fgaClient.GetStoreId()
	if err != nil {
		return fmt.Errorf("failed to get store id: %w", err)
	}

	if storeId == "" {
		store, err := fgaClient.CreateStore(ctx).Body(client.ClientCreateStoreRequest{
			Name: "StreamedListObjects Demo Store",
		}).Execute()
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}
		fmt.Printf("Created store: %s\n", store.Id)
		fgaClient.SetStoreId(store.Id)
		defer func() {
			fgaClient.DeleteStore(ctx).Execute()
			fmt.Printf("Deleted store: %s\n", store.Id)
		}()
	}

	fmt.Println("Writing Authorization Model...")
	authModel := client.ClientWriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{
				Type:      "user",
				Relations: &map[string]openfga.Userset{},
			},
			{
				Type: "document",
				Relations: &map[string]openfga.Userset{
					"reader": {This: &map[string]interface{}{}},
					"writer": {This: &map[string]interface{}{}},
				},
				Metadata: &openfga.Metadata{
					Relations: &map[string]openfga.RelationMetadata{
						"reader": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
						"writer": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
					},
				},
			},
		},
	}

	writeModelResp, err := fgaClient.WriteAuthorizationModel(ctx).Body(authModel).Execute()
	if err != nil {
		return fmt.Errorf("failed to write authorization model: %w", err)
	}
	fmt.Printf("Authorization Model ID: %s\n", writeModelResp.AuthorizationModelId)

	fgaClient.SetAuthorizationModelId(writeModelResp.AuthorizationModelId)

	fmt.Println("\nWriting tuples...")
	tuples := []client.ClientTupleKey{
		{User: "user:alice", Relation: "reader", Object: "document:roadmap"},
		{User: "user:alice", Relation: "reader", Object: "document:budget"},
		{User: "user:alice", Relation: "reader", Object: "document:plan"},
		{User: "user:alice", Relation: "writer", Object: "document:strategy"},
		{User: "user:bob", Relation: "reader", Object: "document:roadmap"},
		{User: "user:bob", Relation: "reader", Object: "document:plan"},
	}

	_, err = fgaClient.Write(ctx).Body(client.ClientWriteRequest{Writes: tuples}).Execute()
	if err != nil {
		return fmt.Errorf("failed to write tuples: %w", err)
	}
	fmt.Printf("Wrote %d tuples\n", len(tuples))

	fmt.Println("\n=== Demo 1: List all documents alice can read ===")
	objectChan, errorChan := fgaClient.StreamedListObjects(ctx).
		Body(client.ClientStreamedListObjectsRequest{
			User:     "user:alice",
			Relation: "reader",
			Type:     "document",
		}).
		Execute()

	objects := []string{}
	done := false
	for !done {
		select {
		case obj, ok := <-objectChan:
			if !ok {
				done = true
				break
			}
			fmt.Printf("  -> %s\n", obj.Object)
			objects = append(objects, obj.Object)
		case err := <-errorChan:
			if err != nil {
				return fmt.Errorf("error during streaming: %w", err)
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timeout waiting for objects")
		}
	}
	fmt.Printf("Total objects found: %d\n", len(objects))

	fmt.Println("\n=== Demo 2: List all documents bob can read ===")
	objectChan2, errorChan2 := fgaClient.StreamedListObjects(ctx).
		Body(client.ClientStreamedListObjectsRequest{
			User:     "user:bob",
			Relation: "reader",
			Type:     "document",
		}).
		Execute()

	objects2 := []string{}
	done2 := false
	for !done2 {
		select {
		case obj, ok := <-objectChan2:
			if !ok {
				done2 = true
				break
			}
			fmt.Printf("  -> %s\n", obj.Object)
			objects2 = append(objects2, obj.Object)
		case err := <-errorChan2:
			if err != nil {
				return fmt.Errorf("error during streaming: %w", err)
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timeout waiting for objects")
		}
	}
	fmt.Printf("Total objects found: %d\n", len(objects2))

	fmt.Println("\n=== Demo 3: List with contextual tuples ===")
	objectChan3, errorChan3 := fgaClient.StreamedListObjects(ctx).
		Body(client.ClientStreamedListObjectsRequest{
			User:     "user:charlie",
			Relation: "reader",
			Type:     "document",
			ContextualTuples: []client.ClientContextualTupleKey{
				{
					User:     "user:charlie",
					Relation: "reader",
					Object:   "document:temp-doc",
				},
			},
		}).
		Execute()

	objects3 := []string{}
	done3 := false
	for !done3 {
		select {
		case obj, ok := <-objectChan3:
			if !ok {
				done3 = true
				break
			}
			fmt.Printf("  -> %s\n", obj.Object)
			objects3 = append(objects3, obj.Object)
		case err := <-errorChan3:
			if err != nil {
				return fmt.Errorf("error during streaming: %w", err)
			}
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timeout waiting for objects")
		}
	}
	fmt.Printf("Total objects found (with contextual tuple): %d\n", len(objects3))

	fmt.Println("\n=== Demo 4: Stream with cancellation ===")
	cancelCtx, cancel := context.WithCancel(ctx)

	objectChan4, errorChan4 := fgaClient.StreamedListObjects(cancelCtx).
		Body(client.ClientStreamedListObjectsRequest{
			User:     "user:alice",
			Relation: "reader",
			Type:     "document",
		}).
		Execute()

	receivedCount := 0
	cancelled := false
	done4 := false
	for !done4 {
		select {
		case obj, ok := <-objectChan4:
			if !ok {
				done4 = true
				break
			}
			receivedCount++
			fmt.Printf("  -> Received object #%d: %s\n", receivedCount, obj.Object)
			if receivedCount >= 2 {
				fmt.Println("  -> Cancelling stream after 2 objects...")
				cancel()
				cancelled = true
			}
		case err := <-errorChan4:
			if err != nil && !cancelled {
				return fmt.Errorf("error during streaming: %w", err)
			}
		case <-time.After(5 * time.Second):
			done4 = true
		}
	}
	fmt.Printf("Received %d objects before cancellation\n", receivedCount)

	fmt.Println("\n=== Demo Complete ===")
	return nil
}
