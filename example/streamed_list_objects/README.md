# StreamedListObjects Example

This example demonstrates how to use the `StreamedListObjects` API in the OpenFGA Go SDK.

## Prerequisites

1. Have an OpenFGA server running (default: `http://localhost:8080`)
2. Have a store created in OpenFGA

## Running the Example

### Option 1: With an existing store and model

```bash
export FGA_API_URL=http://localhost:8080
export FGA_STORE_ID=your-store-id
export FGA_MODEL_ID=your-model-id  # Optional
go run main.go
```

### Option 2: Let the example create test data

If you don't provide `FGA_MODEL_ID`, the example will:
1. Create an authorization model with `user` and `document` types
2. Write test tuples for `user:anne` viewing `document:1`, `document:2`, and `document:3`
3. Stream the objects

```bash
export FGA_API_URL=http://localhost:8080
export FGA_STORE_ID=your-store-id
go run main.go
```

## What This Example Shows

- How to create a client with the SDK
- How to call the `StreamedListObjects` API
- How to consume objects from the streaming channel
- How to handle errors from the streaming API
- The difference between streaming and regular ListObjects (streaming returns results as they're computed)

## Expected Output

```
OpenFGA StreamedListObjects Example
====================================
API URL: http://localhost:8080
Store ID: 01ARZ3NDEKTSV4RRFFQ69G5FAV

Creating authorization model...
Created authorization model: 01ARZ3NDEKTSV4RRFFQ69G5FAX

Writing test tuples...
Wrote 3 test tuples

Calling StreamedListObjects...

Streaming objects:
  1. document:1
  2. document:2
  3. document:3

Total objects received: 3
```