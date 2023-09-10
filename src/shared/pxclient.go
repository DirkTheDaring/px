package shared

import (
	"context"
	"px/configmap"
	"github.com/DirkTheDaring/px-api-client-go"
)

type PxClient struct {
	OrigIndex int // internal
	Context   context.Context
	ApiClient *pxapiflat.APIClient
	Nodes     []string
	Storage   []map[string]interface{}
	Machines  []map[string]interface{}

	StorageAliases map[string]interface{} // map[nodeName]map[alias]name
	StorageContent map[string]interface{}

	Vars     map[string]interface{}
	Ignition map[string]interface{}
	Aliases  map[string]interface{}

	//Resources []pxapiflat.GetClusterResources200ResponseDataInner
	//Storage   []pxapiflat.GetStorage200ResponseDataInner
	//Resources map[string]interface{}
	//Storage   map[string]interface{}
}

func (pxClient PxClient) GetStorage(types []string) []map[string]interface{} {
	list := []map[string]interface{}{}
	for _, value := range pxClient.Storage {
		_type, _ := value["type"].(string)
		if !InOrSkipIfEmpty(types, _type) {
			continue
		}
		list = append(list, value)
	}
	return list
}

func (pxClient PxClient) GetStorageByName(name string) map[string]interface{} {
	for _, value := range pxClient.Storage {
		if value["storage"].(string) == name {
			return value
		}

	}
	return nil
}

func (pxClient PxClient) GetStorageNames() []string {
	list := []string{}
	for _, value := range pxClient.Storage {
		storage, _ := value["storage"].(string)
		list = append(list, storage)
	}
	return list
}

func GetStorageContentAll(pxClients []PxClient) []PxClient {
	list := []PxClient{}
	storageContent := map[string]interface{}{}

	for _, pxClient := range pxClients {
		for _, node := range pxClient.Nodes {
			lookup := map[string]interface{}{}
			for _, item := range pxClient.Storage {
				_type := item["type"].(string)
				if !(_type == "dir" || _type == "nfs") {
					continue
				}
				storage := item["storage"].(string)
				result := GetJsonStorageContent(pxClient, node, storage)
				data, _ := configmap.GetInterfaceSliceValue(result, "data")
				lookup[storage] = data
				//json, _ := json.Marshal(data)
				//fmt.Fprintf(os.Stdout, "%s\n", json)
			}
			storageContent[node] = lookup
		}
		pxClient.StorageContent = storageContent
		list = append(list, pxClient)
	}
	return list
}
