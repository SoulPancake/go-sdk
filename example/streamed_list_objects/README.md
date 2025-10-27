# StreamedListObjects Example

This example demonstrates how to use the `StreamedListObjects` API in the OpenFGA Go SDK in both:
- Synchronous mode (range over the channel)
- Asynchronous mode (consume in a goroutine)

It creates (if not provided) a temporary store and authorization model, writes mock tuples, streams objects, and optionally cleans up.

## Prerequisites

1. An OpenFGA server running (default: `http://localhost:8080`)
2. (Optional) Existing store and authorization model IDs

## Environment Variables

- `FGA_API_URL` (default: `http://localhost:8080`)
- `FGA_STORE_ID` (optional; if absent a new store is created and later deleted)
- `FGA_MODEL_ID` (optional; if absent a simple model is created)
- `FGA_TUPLE_COUNT` (optional; number of tuples to write; overridden by CLI arg if passed)
- `FGA_RELATION` (optional; relation used when writing and listing; must be `viewer` or `owner`; overridden by CLI arg if passed)

## CLI Arguments

```
go run . [mode] [tupleCount] [relation]
```

- `mode`: `sync` (default) or `async`
- `tupleCount`: positive integer (default: 3 if omitted and env var not set)
- `relation`: `viewer` or `owner` (default: `viewer`)

Examples:

```bash
# Basic sync (defaults to 3 tuples, relation viewer)
go run .

# Explicit sync, 10 tuples, relation owner
go run . sync 10 owner

# Async mode with 50 tuples and viewer relation
go run . async 50 viewer

# Using environment variables (relation owner, 25 tuples)
export FGA_TUPLE_COUNT=25
export FGA_RELATION=owner
go run . async
```

## What Happens Internally

1. Store creation (if `FGA_STORE_ID` not provided)
2. Authorization model creation (if `FGA_MODEL_ID` not provided) with `viewer` and `owner` relations
3. Tuple writes: `user:anne` assigned chosen relation for `document:0 .. document:N-1`
4. Streaming request (`StreamedListObjects`) for the chosen relation
5. Consumption pattern:
   - Sync: range directly over `response.Objects`
   - Async: consume in a goroutine while main goroutine reports progress
6. Final error check via `response.Errors`
7. Temporary store deletion (only if the example created it)

## Expected Output (Sync Mode Example)

```
OpenFGA StreamedListObjects Example
====================================
API URL: http://localhost:8080
Store ID: 01ARZ3NDEKTSV4RRFFQ69G5FAV

Creating authorization model...
Created authorization model: 01ARZ3NDEKTSV4RRFFQ69G5FAX

Writing 3 test tuples for relation 'viewer'...
Wrote 3 test tuples

Selected mode: sync | relation: viewer | tuple count: 3 (pass 'sync|async [count] [relation]')
Mode: sync streaming (range over channel)
Streaming objects (sync):
  1. document:0
  2. document:1
  3. document:2

Total objects received (sync): 3 (expected up to 3)

Deleting temporary store...
Deleted temporary store (01ARZ3NDEKTSV4RRFFQ69G5FAV)

Done.
```

## Expected Output (Async Mode Example)

```
OpenFGA StreamedListObjects Example
====================================
API URL: http://localhost:8080
Store ID: 01ARZ3NDEKTSV4RRFFQ69G5FAV

Creating authorization model...
Created authorization model: 01ARZ3NDEKTSV4RRFFQ69G5FAX

Writing 5 test tuples for relation 'owner'...
Wrote 5 test tuples

Selected mode: async | relation: owner | tuple count: 5 (pass 'sync|async [count] [relation]')
Mode: async streaming (consume in goroutine)
Performing other work while streaming...
  (main goroutine still free to do work)
  async -> 1. document:0
  (main goroutine still free to do work)
  async -> 2. document:1
  async -> 3. document:2
  (main goroutine still free to do work)
  async -> 4. document:3
  async -> 5. document:4

Total objects received (async): 5 (expected up to 5)

Deleting temporary store...
Deleted temporary store (01ARZ3NDEKTSV4RRFFQ69G5FAV)

Done.
```

