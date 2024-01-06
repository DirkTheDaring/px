package etc

import (
	"context"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

type PxClient struct {
	Parent    *ClusterDatabase
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
}

func (pxClient *PxClient) GetStorage(types []string) []map[string]interface{} {
	list := []map[string]interface{}{}
	for _, value := range pxClient.Storage {
		_type, _ := value["type"].(string)
		if !InOrSkipIfEmpty(types, _type) {
			continue
		}
		// Side Effect! we change the map and add a value YOU MUST NOT DO THIS!!!
		// It currently screws up the latest algorithm
		//value["node"] = pxClient.Nodes[0]

		list = append(list, value)
	}
	return list
}

func (pxClient *PxClient) GetStorageByName(name string) map[string]interface{} {
	for _, value := range pxClient.Storage {
		if value["storage"].(string) == name {
			return value
		}

	}
	return nil
}

func (pxClient *PxClient) GetStorageNames() []string {
	list := []string{}
	for _, value := range pxClient.Storage {
		storage, _ := value["storage"].(string)
		list = append(list, storage)
	}
	return list
}
