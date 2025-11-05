# Demo: Before Fix

This demo demonstrates the bug where credentials are ignored when a custom HTTPClient is provided.

## Running the Demo

```bash
cd /home/runner/work/go-sdk/go-sdk
make demo-before
```

Or directly:

```bash
cd example/demo_before
go run main.go
```

## Expected Output

The demo shows three scenarios:

1. **Credentials Only**: Works correctly - credentials are processed
2. **Custom HTTPClient + Credentials**: **BUG** - Credentials are ignored
3. **Custom HTTPClient Only**: Works correctly - no credentials to process

The second scenario demonstrates the bug where providing both a custom HTTPClient and credentials causes the credentials to be completely ignored.
