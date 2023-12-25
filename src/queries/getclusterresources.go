package queries

import(
	"github.com/DirkTheDaring/px-api-client-go"
	"context"
	"fmt"
	"os"
	"px/shared"
)

func GetClusterResources(apiClient *pxapiflat.APIClient, context context.Context) map[string]interface{} {
	_, r, err := apiClient.ClusterAPI.GetClusterResources(context).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetClusterResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil
	}
	//resources := clusterResourcesResponse.GetData()
	restResponse := shared.ConvertJsonHttpResponseToMap(r)
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)
	return restResponse
}
