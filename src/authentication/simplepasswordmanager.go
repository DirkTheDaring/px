package authentication

import (
	"fmt"
	"net/url"
	"px/configmap"

)

type Password struct {
	Url      string
	Domain   string
	Username string
	Password string
}

type SimplePasswordManager struct {
	passwordTable []Password
}

func NewSimplePasswordManager(nodes []map[string]interface{}) *SimplePasswordManager {

	spm := SimplePasswordManager{passwordTable: make([]Password, 0)}
	spm.fillPasswordTable(nodes)
	return &spm
}

func (spm *SimplePasswordManager) fillPasswordTable(nodes []map[string]interface{}) {
	for _, node := range nodes {
		enabled := configmap.GetBoolWithDefault(node, "enabled", true)
		if !enabled {
			continue
		}
		url, _ := configmap.GetString(node, "url")
		username, _ := configmap.GetString(node, "username")
		password, _ := configmap.GetString(node, "password")
		domain, _ := configmap.GetString(node, "domain")

		spm.passwordTable = append(spm.passwordTable, Password{
			Url:      url,
			Domain:   domain,
			Username: username,
			Password: password,
		})
	}
}

func (spm *SimplePasswordManager) lookupPassword(url string, domain string, username string) (string, error) {
	for _, item := range spm.passwordTable {
		if item.Url == url && item.Username == username && item.Domain == domain {
			return item.Password, nil
		}
	}
	return "", fmt.Errorf("no password found for %s at url %s", username, url)
}

func (spm *SimplePasswordManager) GetCredentials(urlString, domain string, username string) (string, error) {
	if _, err := url.Parse(urlString); err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	return spm.lookupPassword(urlString, domain, username)
}
