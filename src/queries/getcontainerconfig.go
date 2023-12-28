package queries

import (
	"fmt"
	"os"
	"px/api"
	//pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

func JSONGetCTConfig(node string, vmid int64) (map[string]interface{}, error) {
	_, apiClient, context, err := api.GetPxClient(node)
	if err != nil {
		return nil, err
	}
	_, r, err := apiClient.NodesAPI.GetContainerConfig(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	restResponse, _ := ConvertJsonHttpResponseToMap2(r)
	return restResponse, nil
}
