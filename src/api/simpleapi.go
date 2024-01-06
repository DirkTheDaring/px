package api

import (
	"context"
	"fmt"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

type SimpleAPI struct {
	lookup map[string]*Connection
}

func NewSimpleAPI(lookup map[string]*Connection) *SimpleAPI {
	simpleAPI := SimpleAPI{lookup: lookup}
	return &simpleAPI
}

func (simpleAPI *SimpleAPI) GetPxClient(node string) (*pxapiflat.APIClient, context.Context, error) {

	var apiClient *pxapiflat.APIClient
	var context context.Context

	pxClientPtr, ok := simpleAPI.lookup[node]
	if !ok {
		return nil, context, fmt.Errorf("node not found: %v", node)
	}

	apiClient = pxClientPtr.apiClient
	context = pxClientPtr.context
	return apiClient, context, nil
}

/*
func NewSimpleAPI(lookup map[string]*etc.PxClient) *SimpleAPI {
	simpleAPI := SimpleAPI{lookup: lookup}
	return &simpleAPI
}

func (simpleAPI *SimpleAPI) GetPxClient(node string) (*etc.PxClient, *pxapiflat.APIClient, context.Context, error) {

	var apiClient *pxapiflat.APIClient
	var context context.Context

	pxClientPtr, ok := simpleAPI.lookup[node]
	if !ok {
		return nil, nil, context, fmt.Errorf("node not found: %v", node)
	}

	apiClient = pxClientPtr.ApiClient
	context = pxClientPtr.Context
	return pxClientPtr, apiClient, context, nil
}
*/
