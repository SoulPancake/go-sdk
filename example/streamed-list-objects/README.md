# StreamedListObjects Example

This example demonstrates how to use the StreamedListObjects API in the OpenFGA Go SDK.

## What is StreamedListObjects?

StreamedListObjects is a streaming version of the ListObjects API that returns objects as they are computed, rather than waiting for all results to be gathered. This is useful for:

- Large result sets that would be expensive to compute all at once
- Real-time feedback as objects are found
- Better resource utilization with incremental processing

## Features Demonstrated

1. **Basic Streaming**: Retrieve objects a user has access to with a specific relation
2. **Multiple Users**: Show how different users have access to different objects
3. **Contextual Tuples**: Use contextual tuples to evaluate access without persisting data
4. **Cancellation**: Cancel streaming requests mid-flight using context cancellation

## Prerequisites

You need either:
- A running OpenFGA server (default: http://localhost:8080)
- OR OpenFGA cloud credentials

## Environment Variables

- `FGA_API_URL`: OpenFGA API URL (default: http://localhost:8080)
- `FGA_STORE_ID`: Store ID (optional, will create one if not provided)
- `FGA_CLIENT_ID`: Client ID for authentication (optional)
- `FGA_CLIENT_SECRET`: Client secret for authentication (optional)
- `FGA_API_AUDIENCE`: API audience for authentication (optional)
- `FGA_API_TOKEN_ISSUER`: Token issuer URL for authentication (optional)

## Running the Example

### With a local OpenFGA server:

```bash
# Start OpenFGA server first
docker run -p 8080:8080 openfga/openfga run

# Run the example
go run main.go
```

### With OpenFGA Cloud:

```bash
export FGA_API_URL=https://api.us1.fga.dev
export FGA_STORE_ID=your-store-id
export FGA_CLIENT_ID=your-client-id
export FGA_CLIENT_SECRET=your-client-secret
export FGA_API_AUDIENCE=your-audience
export FGA_API_TOKEN_ISSUER=your-issuer

go run main.go
```

## Example Output

```
Created store: 01GXSB9YR785C4FYS3C0RTG7B2
Writing Authorization Model...
Authorization Model ID: 01GXSA8YR785C4FYS3C0RTG7B1

Writing tuples...
Wrote 6 tuples

=== Demo 1: List all documents alice can read ===
  -> document:roadmap
  -> document:budget
  -> document:plan
Total objects found: 3

=== Demo 2: List all documents bob can read ===
  -> document:roadmap
  -> document:plan
Total objects found: 2

=== Demo 3: List with contextual tuples ===
  -> document:temp-doc
Total objects found (with contextual tuple): 1

=== Demo 4: Stream with cancellation ===
  -> Received object #1: document:roadmap
  -> Received object #2: document:budget
  -> Cancelling stream after 2 objects...
Received 2 objects before cancellation

=== Demo Complete ===
Deleted store: 01GXSB9YR785C4FYS3C0RTG7B2
```

## Code Walkthrough

### Basic Usage

```go
objectChan, errorChan := fgaClient.StreamedListObjects(ctx).
    Body(client.ClientStreamedListObjectsRequest{
        User:     "user:alice",
        Relation: "reader",
        Type:     "document",
    }).
    Execute()

// Process objects as they arrive
for {
    select {
    case obj, ok := <-objectChan:
        if !ok {
            // Channel closed, all objects received
            break
        }
        fmt.Printf("Found: %s\n", obj.Object)
    case err := <-errorChan:
        if err != nil {
            // Handle error
            return err
        }
    }
}
```

### With Options

```go
objectChan, errorChan := fgaClient.StreamedListObjects(ctx).
    Body(client.ClientStreamedListObjectsRequest{
        User:     "user:alice",
        Relation: "reader",
        Type:     "document",
        ContextualTuples: []client.ClientContextualTupleKey{
            {
                User:     "user:alice",
                Relation: "writer",
                Object:   "document:temp",
            },
        },
        Context: &map[string]interface{}{
            "ViewCount": 100,
        },
    }).
    Options(client.ClientStreamedListObjectsOptions{
        Consistency: &openfga.CONSISTENCYPREFERENCE_HIGHER_CONSISTENCY,
    }).
    Execute()
```

## Key Differences from ListObjects

| Feature | ListObjects | StreamedListObjects |
|---------|-------------|---------------------|
| Return type | Slice of all objects | Channel streaming objects |
| Memory usage | All objects in memory | Objects processed incrementally |
| Response time | Wait for all results | First result arrives quickly |
| Cancellation | No mid-request cancellation | Can cancel via context |
| Error handling | Single error return | Error channel for async errors |

## See Also

- [OpenFGA Documentation](https://openfga.dev/docs)
- [StreamedListObjects API Reference](https://openfga.dev/api/service#/Relationship%20Queries/StreamedListObjects)
- [OpenFGA Go SDK](https://github.com/openfga/go-sdk)
