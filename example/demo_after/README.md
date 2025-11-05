# Demo: After Fix

This demo demonstrates that the fix allows both custom HTTPClient and credentials to be honored simultaneously.

## Running the Demo

```bash
cd /home/runner/work/go-sdk/go-sdk
make demo-after
```

Or directly:

```bash
cd example/demo_after
go run main.go
```

## Expected Output

The demo shows four scenarios:

1. **Credentials Only**: Works correctly - credentials are processed
2. **Custom HTTPClient + ClientCredentials**: **FIXED** - Both are honored, custom timeout preserved
3. **Custom HTTPClient Only**: Works correctly - no credentials to process
4. **API Token + Custom HTTPClient**: **FIXED** - Both are honored

All scenarios now work correctly with the fix applied.

## What Changed?

The fix ensures that:
- Credentials are always processed when provided, regardless of whether a custom HTTPClient is provided
- For ClientCredentials: The OAuth2 transport wraps the custom client's transport, preserving settings like Timeout
- For ApiToken: The custom client is used directly with authorization headers added
- Custom HTTPClient settings (Timeout, CheckRedirect, Jar) are preserved
