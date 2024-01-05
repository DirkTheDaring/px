package authentication

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"px/configmap"
	"px/etc"
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
func LoginNode(url string, username string, getCredentials GetCredentialsCallback, insecureSkipVerify bool, timeout time.Duration) (*pxapiflat.APIClient, string, string, error) {
	// Set up the HTTP client with specific configurations.
	disableCompression := true
	proxyURL := "" // Set your proxy URL here if needed
	httpClient, err := createHTTPClient(insecureSkipVerify, disableCompression, proxyURL)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Setting up API client configuration.
	configuration := pxapiflat.NewConfiguration()
	configuration.HTTPClient = &httpClient
	configuration.Servers[0] = pxapiflat.ServerConfiguration{URL: url, Description: "local"}
	apiClient := pxapiflat.NewAPIClient(configuration)

	// Create a context with timeout for the login request.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute the login request.
	password, err := getCredentials(url, "", username)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get credentials for %s at url %s", username, url)
	}

	createAccessTicketRequest := *pxapiflat.NewCreateAccessTicketRequest(password, username)
	resp, _, err := apiClient.AccessAPI.CreateAccessTicket(ctx).CreateAccessTicketRequest(createAccessTicketRequest).Execute()
	if err != nil {
		return nil, "", "", fmt.Errorf("login request failed: %w", err)
	}

	// Extracting the data from the response.
	data := resp.GetData()
	ticket := data.GetTicket()
	csrfPreventionToken := data.GetCSRFPreventionToken()

	return apiClient, ticket, csrfPreventionToken, nil
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

func LoginClusterNodes(nodes []map[string]interface{}, getCredentials GetCredentialsCallback, timeout time.Duration) []etc.PxClient {
	// Usually only one node per cluster is needed
	// However especially in testlabs, the nodes might come and go and might
	// not be even joined. We then handle this by joining the nodes, if there
	// is no conflict of nodenames (which MUST BE UNIQUE) AND Vmids which
	// MUST BE UNIQUE (same requirements as for a proxmox cluster which is joined)
	pxClients := []etc.PxClient{}

	for i, node := range nodes {
		enabled := configmap.GetBoolWithDefault(node, "enabled", true)
		if !enabled {
			continue
		}

		url, _ := configmap.GetString(node, "url")
		username, _ := configmap.GetString(node, "username")
		insecureskipverify := configmap.GetBoolWithDefault(node, "insecureskipverify", false)

		apiClient, ticket, csrfpreventiontoken, err := LoginNode(url, username, getCredentials, insecureskipverify, timeout)
		if err != nil {
			// FIXME should depend on policy if we stay silent here
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		//fmt.Fprintf(os.Stderr, "%v %v\n", ticket, csrfpreventiontoken)
		context := CreateContext(ticket, csrfpreventiontoken)

		pxClient := etc.PxClient{}
		pxClient.Context = context
		pxClient.ApiClient = apiClient
		pxClient.OrigIndex = i
		pxClients = append(pxClients, pxClient)
	}
	return pxClients
}
