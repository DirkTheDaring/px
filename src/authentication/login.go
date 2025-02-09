package authentication

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

// createTLSConfig generates a TLS configuration with the option to skip verification.
func createTLSConfig(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{InsecureSkipVerify: insecureSkipVerify}
}

// createTransport creates an HTTP transport configuration.
// It allows for TLS configuration, compression settings, and an optional proxy.
func createTransport(insecureSkipVerify, disableCompression bool, proxyURL string) (*http.Transport, error) {
	tlsConfig := createTLSConfig(insecureSkipVerify)

	var proxyFunc func(*http.Request) (*url.URL, error)
	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		proxyFunc = http.ProxyURL(parsedURL)
	}

	return &http.Transport{
		TLSClientConfig:    tlsConfig,
		DisableCompression: disableCompression,
		Proxy:              proxyFunc,
	}, nil
}

// createHTTPClient generates an HTTP client with specific transport settings.
func createHTTPClient(insecureSkipVerify, disableCompression bool, proxyURL string) (http.Client, error) {
	transport, err := createTransport(insecureSkipVerify, disableCompression, proxyURL)
	if err != nil {
		return http.Client{}, err
	}
	return http.Client{Transport: transport}, nil
}

// LoginNode authenticates with the specified node using provided credentials and returns API client and tokens.
func LoginNode(login *LoginConfig, passwordManager *PasswordManager, timeout time.Duration) error {
	// Set up the HTTP client with specific configurations.
	disableCompression := true
	proxyURL := "" // Set your proxy URL here if needed
	httpClient, err := createHTTPClient(login.insecureskipverify, disableCompression, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Setting up API client configuration.
	configuration := pxapiflat.NewConfiguration()
	configuration.HTTPClient = &httpClient
	configuration.Servers[0] = pxapiflat.ServerConfiguration{URL: login.url, Description: "local"}
	apiClient := pxapiflat.NewAPIClient(configuration)

	// Create a context with timeout for the login request.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute the login request.

	password, err := (*passwordManager).GetCredentials(login.url, login.domain, login.username)
	if err != nil {
		return fmt.Errorf("failed to get credentials for %s in domain '%s' at url %s", login.username, login.domain, login.url)
	}

	// FIXME here you forgot domain!!!
	createAccessTicketRequest := *pxapiflat.NewCreateAccessTicketRequest(password, login.username)
	resp, _, err := apiClient.AccessApi.CreateAccessTicket(ctx).CreateAccessTicketRequest(createAccessTicketRequest).Execute()
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	// Extracting the data from the response.
	data := resp.GetData()

	login.ticket = data.GetTicket()
	login.csrfPreventionToken = data.GetCSRFPreventionToken()
	login.apiClient = apiClient
	login.context = CreateContext(login.ticket, login.csrfPreventionToken)

	login.success = true
	login.error = nil

	return nil
}

func CreateContext(ticket string, csrfpreventiontoken string) context.Context {

	cookie := "PVEAuthCookie=" + ticket
	newContext := context.WithValue(
		context.Background(),
		pxapiflat.ContextAPIKeys,
		map[string]pxapiflat.APIKey{
			"cookie": {
				Key: cookie,
			},
			"token": {
				Key: csrfpreventiontoken,
			},
		},
	)
	return newContext
}

func AuthenticateClusterNodes(logins []*LoginConfig, passwordManager *PasswordManager, timeout time.Duration) error {
	for _, login := range logins {
		err := LoginNode(login, passwordManager, timeout)
		if err != nil {
			login.success = false
			login.error = err
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
	return nil
}
