package credentials

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/openfga/go-sdk/internal/utils/retryutils"
	"github.com/openfga/go-sdk/oauth2"
	"github.com/openfga/go-sdk/oauth2/clientcredentials"
)

const ApiTokenHeaderKey = "Authorization"
const ApiTokenHeaderValuePrefix = "Bearer"

// Available credential methods
type CredentialsMethod string

const (
	// No credentials (default)
	CredentialsMethodNone CredentialsMethod = "none"
	// API Token credentials (will be sent in "Authorization: Bearer $TOKEN" header)
	CredentialsMethodApiToken CredentialsMethod = "api_token"
	// Client Credentials flow will be performed, resulting token will be sent in "Authorization: Bearer $TOKEN" header
	CredentialsMethodClientCredentials CredentialsMethod = "client_credentials"
)

type Config struct {
	ApiToken                        string `json:"apiToken,omitempty"`
	ClientCredentialsApiTokenIssuer string `json:"apiTokenIssuer,omitempty"`
	ClientCredentialsApiAudience    string `json:"apiAudience,omitempty"`
	ClientCredentialsClientId       string `json:"clientId,omitempty"`
	ClientCredentialsClientSecret   string `json:"clientSecret,omitempty"`
	ClientCredentialsScopes         string `json:"scopes,omitempty"`
}

type Credentials struct {
	Method  CredentialsMethod `json:"method,omitempty"`
	Config  *Config           `json:"config,omitempty"`
	Context context.Context
}

func NewCredentials(config Credentials) (*Credentials, error) {
	creds := &Credentials{
		Method: config.Method,
		Config: config.Config,
	}

	if creds.Method == "" {
		creds.Method = CredentialsMethodNone
	}

	err := creds.ValidateCredentialsConfig()

	if err != nil {
		return nil, err
	}

	return creds, nil
}

func (c *Credentials) ValidateCredentialsConfig() error {
	conf := c.Config
	if c.Method == CredentialsMethodApiToken && (conf == nil || conf.ApiToken == "") {
		return fmt.Errorf("CredentialsConfig.ApiToken is required when CredentialsMethod is CredentialsMethodApiToken (%s)", c.Method)
	} else if c.Method == CredentialsMethodClientCredentials {
		if conf == nil ||
			conf.ClientCredentialsClientId == "" ||
			conf.ClientCredentialsClientSecret == "" ||
			conf.ClientCredentialsApiTokenIssuer == "" {
			return fmt.Errorf("all of CredentialsConfig.ClientId, CredentialsConfig.ClientSecret and CredentialsConfig.ApiTokenIssuer are required when CredentialsMethod is CredentialsMethodClientCredentials (%s)", c.Method)
		}
		tokenURL, err := buildApiTokenURL(conf.ClientCredentialsApiTokenIssuer)
		if err != nil {
			return err
		}
		conf.ClientCredentialsApiTokenIssuer = tokenURL
	}

	return nil
}

type HeaderParams struct {
	Key   string
	Value string
}

func (c *Credentials) GetApiTokenHeader() *HeaderParams {
	if c.Method != CredentialsMethodApiToken {
		return nil
	}

	return &HeaderParams{
		Key:   ApiTokenHeaderKey,
		Value: ApiTokenHeaderValuePrefix + " " + c.Config.ApiToken,
	}
}

// GetHttpClientAndHeaderOverrides
// The main export the client uses to get a configuration with the necessary
// httpClient and header overrides based on the chosen credential method
func (c *Credentials) GetHttpClientAndHeaderOverrides(retryParams retryutils.RetryParams, debug bool) (*http.Client, []*HeaderParams) {
	return c.GetHttpClientAndHeaderOverridesWithBaseClient(retryParams, debug, nil)
}

// GetHttpClientAndHeaderOverridesWithBaseClient
// Similar to GetHttpClientAndHeaderOverrides but allows specifying a base HTTP client.
// If baseClient is provided and credentials use ClientCredentials method, the baseClient's
// transport will be used as the base transport for the OAuth2 transport.
func (c *Credentials) GetHttpClientAndHeaderOverridesWithBaseClient(retryParams retryutils.RetryParams, debug bool, baseClient *http.Client) (*http.Client, []*HeaderParams) {
	var headers []*HeaderParams
	var client *http.Client
	switch c.Method {
	case CredentialsMethodClientCredentials:
		requestConfig := clientcredentials.RequestConfig{
			RetryParams: retryParams,
			Debug:       debug,
		}
		_ = requestConfig.New(requestConfig)
		ccConfig := clientcredentials.Config{
			ClientID:      c.Config.ClientCredentialsClientId,
			ClientSecret:  c.Config.ClientCredentialsClientSecret,
			TokenURL:      c.Config.ClientCredentialsApiTokenIssuer,
			RequestConfig: requestConfig,
		}
		if c.Config.ClientCredentialsApiAudience != "" {
			ccConfig.EndpointParams = map[string][]string{
				"audience": {c.Config.ClientCredentialsApiAudience},
			}
		}
		if c.Config.ClientCredentialsScopes != "" {
			scopes := strings.Split(strings.TrimSpace(c.Config.ClientCredentialsScopes), " ")
			ccConfig.Scopes = append(ccConfig.Scopes, scopes...)
		}
		if c.Context == nil {
			c.Context = context.Background()
		}

		// If a base client is provided, use it to create the OAuth client
		if baseClient != nil {
			// Store the base client in context so oauth2 can use it
			ctx := context.WithValue(c.Context, oauth2.HTTPClient, baseClient)
			client = ccConfig.Client(ctx)

			// Copy client settings from base client to preserve custom configurations
			// Note: oauth2.NewClient creates a new http.Client with only the Transport from base
			// We need to manually copy other settings like Timeout, CheckRedirect, Jar, etc.
			if baseClient.Timeout != 0 {
				client.Timeout = baseClient.Timeout
			}
			if baseClient.CheckRedirect != nil {
				client.CheckRedirect = baseClient.CheckRedirect
			}
			if baseClient.Jar != nil {
				client.Jar = baseClient.Jar
			}
		} else {
			client = ccConfig.Client(c.Context)
		}
	case CredentialsMethodApiToken:
		var header = c.GetApiTokenHeader()
		if header != nil {
			headers = append(headers, header)
		}
		// For ApiToken, we can use the base client directly since we're just adding headers
		client = baseClient
	case CredentialsMethodNone:
		client = baseClient
	default:
		client = baseClient
	}

	return client, headers
}

var defaultTokenEndpointPath = "oauth/token"

func buildApiTokenURL(issuer string) (string, error) {
	u, err := url.Parse(issuer)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u, _ = url.Parse(fmt.Sprintf("https://%s", issuer))
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid issuer scheme '%s' (must be http or https)", u.Scheme)
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = defaultTokenEndpointPath
	}
	return u.String(), nil
}
