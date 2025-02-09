package queries

import (
	"context"
	"fmt"
	"os"
	"px/etc"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

func GetClusterConfigNodes(apiClient *pxapiflat.APIClient, context context.Context) map[string]interface{} {
	_, r, err := apiClient.ClusterApi.GetClusterConfigNodes(context).Execute()
	// {
	//  "data": [
	//    {
	//      "name": "denue6pr248",
	//      "node": "denue6pr248",
	//      "nodeid": "7",
	//      "quorum_votes": "1",
	//      "ring0_addr": "172.16.0.26"
	//    },
	//    {
	//      "name": "denue6pr095",
	//      "node": "denue6pr095",
	//      "nodeid": "4",
	//      "quorum_votes": "1",
	//      "ring0_addr": "172.16.0.22"
	//    }
	//  ]
	//}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetClusterResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil
	}
	//resources := clusterResourcesResponse.GetData()
	restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)
	return restResponse
}

func DumpClusterNodes(pxClients []*etc.PxClient) {

	for _, pxClient := range pxClients {
		result := GetClusterConfigNodes(pxClient.ApiClient, pxClient.Context)
		fmt.Fprintf(os.Stderr, "%v\n", result["data"])
	}
}
