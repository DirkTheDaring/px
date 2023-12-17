package shared

import (
	"fmt"
	"os"
)

func JSONGetVMConfig(node string, vmid int64) (map[string]interface{}, error) {
	_, apiClient, context, err := GetPxClient(node)
	if err != nil {
		return nil, err
	}
	_, r, err := apiClient.NodesAPI.GetVMConfig(context, node, vmid).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	restResponse := ConvertJsonHttpResponseToMap(r)
	return restResponse, nil
}
