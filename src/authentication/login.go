package authentication

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"px/configmap"
	"px/etc"
	"time"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

func CreateHttpClient(insecureSkipVerify bool) http.Client {
	tlsconf := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	tr := &http.Transport{
		TLSClientConfig:    tlsconf,
		DisableCompression: true,
		Proxy:              nil,
	}
	return http.Client{Transport: tr}
}

func LoginNode(url string, username string, password string, insecureskipverify bool, timeout time.Duration) (*pxapiflat.APIClient, string, string, error) {

	//fmt.Fprintf(os.Stderr, "%v %v %v %v\n", url, username, password, insecureskipverify)
	/// ---
	context_timeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpClient := CreateHttpClient(insecureskipverify)
	configuration := pxapiflat.NewConfiguration()
	configuration.HTTPClient = &httpClient
	configuration.Servers[0] = pxapiflat.ServerConfiguration{
		URL:         url,
		Description: "local",
	}
	apiClient := pxapiflat.NewAPIClient(configuration)
	createAccessTicketRequest := *pxapiflat.NewCreateAccessTicketRequest(password, username)

	resp, _, err := apiClient.AccessAPI.CreateAccessTicket(context_timeout).CreateAccessTicketRequest(createAccessTicketRequest).Execute()

	if err != nil {
		// cluster is not available, skip
		// No error message therefore
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil, "", "", err
	}

	data := resp.GetData()
	ticket := data.GetTicket()

	csrfpreventiontoken := data.GetCSRFPreventionToken()
	return apiClient, ticket, csrfpreventiontoken, nil
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

func LoginClusterNodes(nodes []map[string]interface{}, timeout time.Duration) []etc.PxClient {
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
		password, _ := configmap.GetString(node, "password")
		insecureskipverify := configmap.GetBoolWithDefault(node, "insecureskipverify", false)

		apiClient, ticket, csrfpreventiontoken, err := LoginNode(url, username, password, insecureskipverify, timeout)

		if err != nil {
			// should depend on policy if we stay silent here
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
