package queries

import (
	"fmt"
	"os"
	"px/api"
	//pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

func JSONGetCTConfig(node string, vmid int64) (map[string]interface{}, error) {
	apiClient, context, err := api.GetPxClient(node)
	if err != nil {
		return nil, err
	}
	_, r, err := apiClient.NodesApi.GetContainerConfig(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)
	return restResponse, nil
}
