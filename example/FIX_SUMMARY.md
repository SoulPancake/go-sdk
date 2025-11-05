# Fix Summary: Credentials + HTTPClient Bug

## Problem Statement
When both `HTTPClient` and `Credentials` are provided to the SDK client configuration, the credentials are completely ignored. This occurs because credential processing only happens when `cfg.HTTPClient == nil`.

## Root Cause
In `api_client.go`, the `NewAPIClient` function had the following logic:

```go
if cfg.HTTPClient == nil {
    if cfg.Credentials == nil {
        cfg.HTTPClient = http.DefaultClient
    } else {
        // Process credentials and create HTTPClient
    }
}
```

This meant that when a custom HTTPClient was provided, the entire credentials processing block was skipped.

## Solution
The fix ensures credentials are always processed when provided, regardless of whether a custom HTTPClient is present:

1. **Modified `NewAPIClient` in `api_client.go`**:
   - Moved credential processing outside the `if cfg.HTTPClient == nil` check
   - Credentials are now processed unconditionally when provided

2. **Added `GetHttpClientAndHeaderOverridesWithBaseClient` in `credentials/credentials.go`**:
   - New function that accepts a base HTTP client
   - For **ClientCredentials**: Creates an OAuth2 transport that wraps the base client's transport
   - For **ApiToken**: Uses the base client directly and adds authorization headers
   - Preserves custom client settings (Timeout, CheckRedirect, Jar)

## Changes Made

### Files Modified
1. `api_client.go` - Updated `NewAPIClient` function
2. `credentials/credentials.go` - Added new function to support base client
3. `api_client_credentials_test.go` - New comprehensive test file
4. `Makefile` - Added demo verification targets
5. `example/demo_before/main.go` - Demo showing the bug
6. `example/demo_after/main.go` - Demo showing the fix

### Test Coverage
Added 5 new test cases covering all scenarios:
- ✅ Custom HTTPClient + ClientCredentials
- ✅ Custom HTTPClient + ApiToken
- ✅ Credentials only
- ✅ Custom HTTPClient only
- ✅ Neither credentials nor HTTPClient

## Verification

### Run the Demo
```bash
# Show the bug before the fix
make demo-before

# Show the fix working
make demo-after

# Or verify the fix directly
make demo-verify
```

### Run the Tests
```bash
# Run all tests
go test ./...

# Run only the new credential tests
go test -v -run="TestAPIClient"
```

### Expected Behavior

#### Before Fix
- ❌ Credentials ignored when custom HTTPClient provided
- ✅ Credentials work without custom HTTPClient
- ✅ Custom HTTPClient works without credentials

#### After Fix
- ✅ Credentials processed when custom HTTPClient provided
- ✅ Custom HTTPClient settings preserved (timeout, etc.)
- ✅ Both ClientCredentials and ApiToken methods work correctly
- ✅ Backward compatible - all existing tests pass

## Technical Details

### How It Works

#### For ClientCredentials
1. Base client is passed via context to OAuth2
2. OAuth2 creates a new client with a transport that:
   - Wraps the base client's transport
   - Adds OAuth2 token handling
3. Custom settings (Timeout, CheckRedirect, Jar) are copied to the new client

#### For ApiToken
1. Base client is used directly (no wrapping needed)
2. Authorization header is added to DefaultHeaders
3. All custom client settings are preserved

### Code Example
```go
// Before: This would ignore credentials
fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
    ApiUrl: "http://localhost:8080",
    Credentials: &credentials.Credentials{
        Method: credentials.CredentialsMethodClientCredentials,
        Config: &credentials.Config{
            ClientCredentialsClientId:       "client-id",
            ClientCredentialsClientSecret:   "client-secret",
            ClientCredentialsApiAudience:    "https://api.example/",
            ClientCredentialsApiTokenIssuer: "issuer.example",
        },
    },
    HTTPClient: &http.Client{Timeout: 30 * time.Second},
})

// After: Both credentials and custom HTTPClient are honored
// - Credentials create OAuth2 transport
// - Custom timeout of 30s is preserved
```

## Security Analysis
- ✅ No security vulnerabilities found (CodeQL analysis)
- ✅ All existing tests pass
- ✅ No breaking changes to API

## Backward Compatibility
✅ **Fully backward compatible**
- Existing code without custom HTTPClient works exactly as before
- Existing code with custom HTTPClient but without credentials works as before
- Only the previously broken case (custom HTTPClient + credentials) is fixed

## Impact
This fix enables users to:
- Combine custom transport settings (timeouts, proxies, etc.) with authentication
- Use enterprise HTTP clients with custom configurations while maintaining security
- Properly test authenticated scenarios with custom test clients
