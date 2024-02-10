package etc

import (
	"fmt"
	"os"
	"px/configmap"
	"px/proxmox/query"
	"strconv"
	"strings"
)

type PxCluster struct {
	pxClients2       []*PxClient
	pxClientLookup   map[string]int
	nodes            []string
	uniqueMachines   map[string]map[string]interface{}
	machines         []map[string]interface{}
	globalConfigData *map[string]interface{}
}

func (pxCluster *PxCluster) GetMachines() []map[string]interface{} {
	return pxCluster.machines
}

func (pxCluster *PxCluster) Exists(node string, vmid int64) bool {

	key := node + "/" + strconv.FormatInt(vmid, 10)

	//fmt.Fprintf(os.Stderr, "key=%v\n", key)
	//fmt.Fprintf(os.Stderr, "hash=%+v\n", pxCluster.uniqueMachines)

	_, ok := pxCluster.uniqueMachines[key]
	return ok
}

func (pxCluster *PxCluster) IsVirtualCluster() bool {
	// if we have more than one entry in pxClients, this is a virtual cluster, which
	// has different Storage per pxClient and also non unique vmids
	return len(pxCluster.pxClients2) > 1
}
func (pxCluster *PxCluster) GetPxClient(node string) (PxClient, error) {
	pos, ok := pxCluster.pxClientLookup[node]
	if !ok {
		return PxClient{}, fmt.Errorf("node not found: %v", node)

	}
	return *pxCluster.pxClients2[pos], nil
}

func (pxCluster *PxCluster) GetPxClients() []*PxClient {
	return pxCluster.pxClients2
}

// SELECT * FROM  pxCluster.PxClients as result
// SELECT * FROM result.GetStorage()

func (pxCluster *PxCluster) GetStorage(types []string) []map[string]interface{} {

	if !pxCluster.IsVirtualCluster() {
		//fmt.Printf("GetStorage() not virtual\n")
		pxClient := pxCluster.pxClients2[0]
		return pxClient.GetStorage(types)
	}

	account := map[string]int{}
	// Count storage (names) , used to determine, if a storage is on every Client
	for _, pxClient := range pxCluster.pxClients2 {
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
	total := len(pxCluster.pxClients2)
	commonStorage := []string{}
	for key, value := range account {
		if value == total {
			items := strings.Split(key, " ")
			commonStorage = append(commonStorage, items[0])
		}
	}

	//	fmt.Fprintf(os.Stderr, "commonStorage: %v\n", commonStorage)

	list := []map[string]interface{}{}
	pxClient := pxCluster.pxClients2[0]
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

	//	jsonBinary, _ := json.Marshal(list)
	//	fmt.Fprintf(os.Stdout, "list: %v\n", string(jsonBinary))

	// Now add the local
	for _, pxClient := range pxCluster.pxClients2 {
		storageList := pxClient.GetStorage(types)

		//		jsonBinary, _ := json.Marshal(storageList)
		//		fmt.Fprintf(os.Stdout, "storageList = %v\n", string(jsonBinary))

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

func (pxCluster *PxCluster) GetStorageContent() []map[string]interface{} {
	types := []string{"dir", "nfs"}

	storageList := pxCluster.GetStorage(types)

	//jsonBinary, _ := json.Marshal(storageList)
	//fmt.Fprintf(os.Stdout, "storageList = %v\n", string(jsonBinary))

	list := []map[string]interface{}{}

	for _, item := range storageList {

		var nodeList []string

		_type, _ := item["type"].(string)
		node, _ := item["node"].(string)

		// Global * or "" are handle differently
		if node == "*" || node == "" {
			if _type == "nfs" {
				nodeList = []string{pxCluster.nodes[0]}
			} else {
				nodeList = pxCluster.nodes
			}
		} else {
			nodeList = []string{node}
		}

		storage := item["storage"].(string)

		//fmt.Fprintf(os.Stderr, "nodeList: %v\n", nodeList)
		for _, nodeItem := range nodeList {

			pxClient, _ := pxCluster.GetPxClient(nodeItem)
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

func (pxCluster *PxCluster) GetStorageNames() []string {
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
func (pxCluster *PxCluster) GetStorageNamesOnNode(node string) []string {
	if !query.In(pxCluster.nodes, node) {
		fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): node not found: %s (%v)\n", node, pxCluster.nodes)
		return []string{}
	}
	pxClient, _ := pxCluster.GetPxClient(node)
	return pxClient.GetStorageNames()
}
func (pxCluster *PxCluster) GetAliasOnNode(node string) map[string]string {

	if !query.In(pxCluster.nodes, node) {
		fmt.Fprintf(os.Stderr, "GetStorageNameOnNode(): node not found: %s (%v)\n", node, pxCluster.nodes)
		return map[string]string{}
	}
	pxClient, _ := pxCluster.GetPxClient(node)

	aliasesValue, ok := pxClient.StorageAliases[node]

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

func (pxCluster *PxCluster) GetNodeNames() []string {
	return pxCluster.nodes
}

func (pxCluster *PxCluster) HasNode(node string) bool {
	//fmt.Fprintf(os.Stderr, "HasNode() %v %v\n", node, pxCluster.Nodes)
	return query.In(pxCluster.nodes, node)
}

func (pxCluster *PxCluster) GetNodeCount() int {
	return len(pxCluster.nodes)
}

func (pxCluster *PxCluster) GetNodeLookup() map[string]int {
	nodeLookup := map[string]int{}
	for i, name := range pxCluster.nodes {
		nodeLookup[name] = i
	}
	return nodeLookup
}

func (pxCluster *PxCluster) GetPxClientLookup() map[string]int {
	return pxCluster.pxClientLookup
}

func (pxCluster *PxCluster) SetGlobalConfigData(globalConfigData *map[string]interface{}) {
	pxCluster.globalConfigData = globalConfigData
}
func (pxCluster *PxCluster) PickCluster(name string) (*ClusterDatabase, error) {

	clustersDatabase := ClustersDatabase{table: pxCluster.globalConfigData}

	return clustersDatabase.GetClusterDatabaseByName(name)

}

func (pxCluster *PxCluster) GetPxClientMap() map[string]*PxClient {
	newMap := make(map[string]*PxClient)
	for key, index := range pxCluster.pxClientLookup {
		newMap[key] = pxCluster.pxClients2[index]
	}
	return newMap
}

// GetMachinesByKey filters a slice of maps within a PxCluster object,
// returning only those that contain a specific key-value pair.
func (pxCluster *PxCluster) GetMachinesByKey(key, value string) ([]map[string]interface{}, error) {
	var list []map[string]interface{}

	for _, mapItem := range pxCluster.machines {
		val, ok := configmap.GetString(mapItem, key)
		if !ok {
			// Skip items that do not contain the key
			continue
		}
		if val == value {
			list = append(list, mapItem)
		}
	}

	return list, nil
}
