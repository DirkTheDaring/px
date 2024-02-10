package queries

import (
	"fmt"
	"os"
	"px/api"
	"px/configmap"
)

func JSONGetVMConfig(node string, vmid int64) (map[string]interface{}, error) {
	apiClient, context, err := api.GetPxClient(node)
	if err != nil {
		return nil, err
	}

	_, r, err := apiClient.NodesAPI.GetVMConfig(context, node, vmid).Execute()
	if err != nil {
		// HACK: This fixes it for proxmox 7
		if fmt.Sprintf("%v", err) == "json: cannot unmarshal number into Go struct field GetVMConfig200ResponseData.data.memory of type string" {

			//fmt.Fprintf(os.Stderr, "Body:\n%s\n", r.Body)

			restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)

			data, ok := configmap.GetMapEntry(restResponse, "data")

			if !ok {
				return nil, err
			}

			memory, ok := data["memory"]

			if !ok {
				fmt.Fprintf(os.Stderr, "not found: memory\n")
			}

			memoryStr := fmt.Sprintf("%f", memory)

			data["memory"] = memoryStr
			//fmt.Fprintf(os.Stderr, "memory = %v (%T)\n", memory, memory)

			return restResponse, nil

		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetVMConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}

	restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)
	return restResponse, nil
}
