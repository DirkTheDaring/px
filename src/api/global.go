package api

import (
	"context"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

type Connection struct {
	nodeName  string
	apiClient *pxapiflat.APIClient
	context   context.Context
}

func NewConnection(nodeName string, apiCient *pxapiflat.APIClient, context context.Context) *Connection {
	connection := Connection{
		nodeName:  nodeName,
		apiClient: apiCient,
		context:   context,
	}
	return &connection
}
