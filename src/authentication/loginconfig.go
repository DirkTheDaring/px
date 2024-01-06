package authentication

import (
	"context"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

type LoginConfig struct {
	url                string
	username           string
	domain             string
	insecureskipverify bool

	ticket              string
	csrfPreventionToken string
	apiClient           *pxapiflat.APIClient
	context             context.Context

	success bool
	error   error
}

func NewLoginConfig(url string, domain string, username string, insecureskipverify bool) *LoginConfig {
	return &LoginConfig{url: url, domain: domain, username: username, insecureskipverify: insecureskipverify}
}

func (loginConfig *LoginConfig) GetContext() context.Context {
	return loginConfig.context
}
func (loginConfig *LoginConfig) GetApiClient() *pxapiflat.APIClient {
	return loginConfig.apiClient
}
func (loginConfig *LoginConfig) GetSuccess() bool {
	return loginConfig.success
}
