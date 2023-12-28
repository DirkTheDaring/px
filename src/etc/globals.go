package etc

import (
	"context"
	"fmt"
	"os"
	"px/configmap"
	"px/proxmox/query"
	"strings"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
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
}

type PxCluster struct {
	PxClients      []PxClient
	PxClientLookup map[string]int
	Nodes          []string
	UniqueMachines map[int]map[string]interface{}
	Machines       []map[string]interface{}
}

// --------------------------------------------------------------

var GlobalPxCluster PxCluster
var GlobalConfigData map[string]interface{}

// --------------------------------------------------------------

func InOrSkipIfEmpty(haystack []string, needle string) bool {
	// we have the needle if the haystack is empty ...
	if len(haystack) == 0 {
		return true
	}
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func (pxCluster PxCluster) IsVirtualCluster() bool {
	// if we have more than one entry in pxClients, this is a virtual cluster, which
	// has different Storage per pxClient and also non unique vmids
	if len(pxCluster.PxClients) > 1 {
		return true
	}
	return false
}
func (pxCluster PxCluster) GetPxClient(node string) PxClient {
	pos := pxCluster.PxClientLookup[node]
	return pxCluster.PxClients[pos]
}

// SELECT * FROM  pxCluster.PxClients as result
// SELECT * FROM result.GetStorage()

func (pxCluster PxCluster) GetStorage(types []string) []map[string]interface{} {
	if !pxCluster.IsVirtualCluster() {
		return pxCluster.PxClients[0].GetStorage(types)
	}
	account := map[string]int{}
	// Count storage (names) , used to determine, if a storage is on every Client
	for _, pxClient := range pxCluster.PxClients {
		storageList := pxClient.GetStorage(types)
		var uniqueKey string
		for _, item := range storageList {
			_type, _ := item["type"].(string)
			storage, _ := item["storage"].(string)
			path, _ := item["path"].(string)
			if _type == "nfs" {
				server, _ := item["server"].(string)
				export, _ := item["export"].(string)
				uniqueKey = storage + " " + _type + " " + path + " " + server + " " + export
			} else {
				uniqueKey = storage + " " + _type + " " + path
			}

			count, ok := account[uniqueKey]
			if ok {
				account[uniqueKey] = count + 1
			} else {
				account[uniqueKey] = 1
			}
		}
	}
	total := len(pxCluster.PxClients)
	commonStorage := []string{}
	for key, value := range account {
		if value == total {
			items := strings.Split(key, " ")
			commonStorage = append(commonStorage, items[0])
		}
	}

	//fmt.Fprintf(os.Stderr, "commonStorage: %v\n", commonStorage)

	list := []map[string]interface{}{}
	pxClient := pxCluster.PxClients[0]
	storageList := pxClient.GetStorage(types)
	for _, needle := range commonStorage {
		for _, item := range storageList {
			storage, _ := item["storage"].(string)
			if storage != needle {
				continue
			}
			item["node"] = "*"
			list = append(list, item)
		}
	}
	// Now add the local
	for _, pxClient := range pxCluster.PxClients {
		storageList := pxClient.GetStorage(types)
		for _, node := range pxClient.Nodes {
			for _, item := range storageList {
				storage, _ := item["storage"].(string)
				if InOrSkipIfEmpty(commonStorage, storage) {
					continue
				}
				item["node"] = node
				list = append(list, item)
			}
		}
	}

	return list
}

func (pxCluster PxCluster) GetStorageContent() []map[string]interface{} {
	types := []string{"dir", "nfs"}

	storageList := pxCluster.GetStorage(types)

	list := []map[string]interface{}{}

	for _, item := range storageList {

		var nodeList []string

		_type, _ := item["type"].(string)
		node, _ := item["node"].(string)

		// Global * or "" are handle differently
		if node == "*" || node == "" {
			if _type == "nfs" {
				nodeList = []string{pxCluster.Nodes[0]}
			} else {
				nodeList = pxCluster.Nodes
			}
		} else {
			nodeList = []string{node}
		}

		storage := item["storage"].(string)

		//fmt.Fprintf(os.Stderr, "nodeList: %v\n", nodeList)
		for _, nodeItem := range nodeList {

			pxClient := pxCluster.GetPxClient(nodeItem)
			storageContent, _ := configmap.GetMapEntry(pxClient.StorageContent, nodeItem)
			if len(storageContent) == 0 {
				continue
			}
			storageContentList := storageContent[storage].([]map[string]interface{})

			for _, storageContent := range storageContentList {
				if (node == "*" || node == "") && _type == "nfs" {
					storageContent["node"] = node
				} else {
					storageContent["node"] = nodeItem
				}
				storageContent["storage"] = storage
				storageContent["type"] = _type
				list = append(list, storageContent)
			}
		}
	}
	return list
}

func (pxCluster PxCluster) GetStorageNames() []string {
	types := []string{"dir", "nfs"}
	storage := pxCluster.GetStorage(types)
	storageNames := []string{}
	for _, storageItem := range storage {
		storage, _ := storageItem["storage"].(string)
		//fmt.Fprintf(os.Stderr, "storage:  %v\n", storage)
		if !query.In(storageNames, storage) {
			storageNames = append(storageNames, storage)
		}
	}
	return storageNames
}
func (pxCluster PxCluster) GetStorageNamesOnNode(node string) []string {
	if !query.In(pxCluster.Nodes, node) {
		fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): node not found: %s (%v)\n", node, pxCluster.Nodes)
		return []string{}
	}
	pxClient := pxCluster.GetPxClient(node)
	return pxClient.GetStorageNames()
}
func (pxCluster PxCluster) GetAliasOnNode(node string) map[string]string {

	if !query.In(pxCluster.Nodes, node) {
		fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): node not found: %s (%v)\n", node, pxCluster.Nodes)
		return map[string]string{}
	}
	pxClient := pxCluster.GetPxClient(node)
	aliasesValue, ok := pxClient.StorageAliases[node]
	//fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): alias not found: %T\n", aliasesValue)
	if !ok {
		fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): alias not found: %s (%v) (%v)\n", node, aliasesValue, pxClient.StorageAliases)
		return map[string]string{}
	}
	aliases, ok := aliasesValue.(map[string]string)
	if ok {
		return aliases
	}

	fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): alias not found: %T\n", aliasesValue)

	return map[string]string{}
}

func (pxCluster PxCluster) HasNode(node string) bool {
	//fmt.Fprintf(os.Stderr, "HasNode() %v %v\n", node, pxCluster.Nodes)
	return query.In(pxCluster.Nodes, node)
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
