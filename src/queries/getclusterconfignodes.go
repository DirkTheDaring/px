package queries

import(
)
/*
func GetClusterConfigNodes(apiClient *pxapiflat.APIClient, context context.Context) map[string]interface{} {
	_, r, err := apiClient.ClusterAPI.GetClusterConfigNodes(context).Execute()
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
	restResponse := shared.ConvertJsonHttpResponseToMap(r)
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)
	return restResponse
}

/*
func GetClusterNodes(pxClients []shared.PxClient) []shared.PxClient {
	list := []shared.PxClient{}
	for _, pxClient := range pxClients {
		nodeList := []string{}
		result := GetClusterConfigNodes(pxClient.ApiClient, pxClient.Context)
		nodes, err := configmap.GetInterfaceSliceValue(result, "data")
		if err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)

			pxClient.Nodes = nodeList
			continue
		}
		if len(nodes) > 0 {
			for _, item := range nodes {
				nodeName := item["node"].(string)
				nodeList = append(nodeList, nodeName)
			}
			sort.Strings(nodeList)
			pxClient.Nodes = nodeList
		}
		list = append(list, pxClient)
	}
	return list
}
*/

