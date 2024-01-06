package shared

import (
	"fmt"
	"os"
	"px/api"
	"px/authentication"
	"px/configmap"
	"px/etc"
	"px/proxmox/clusters"
	"px/queries"
	"sort"
	"time"
)

// CreateLoginConfigsFromNodes transforms cluster node information into login configurations.
func CreateLoginConfigsFromNodes(clusterNodes []map[string]interface{}) ([]*authentication.LoginConfig, error) {
	var loginConfigs []*authentication.LoginConfig

	for _, node := range clusterNodes {
		if enabled := configmap.GetBoolWithDefault(node, "enabled", true); !enabled {
			continue
		}

		url, urlExists := configmap.GetString(node, "url")
		username, usernameExists := configmap.GetString(node, "username")
		domain, _ := configmap.GetString(node, "domain")

		if !urlExists {
			return nil, fmt.Errorf("node configuration is missing url")
		}
		if !usernameExists {
			return nil, fmt.Errorf("node configuration is missing username")
		}
		/*
			if !domainExists {
				return nil, fmt.Errorf("node configuration is missing domain")
			}
		*/

		insecureskipverify := configmap.GetBoolWithDefault(node, "insecureskipverify", false)
		loginConfig := authentication.NewLoginConfig(url, domain, username, insecureskipverify)
		loginConfigs = append(loginConfigs, loginConfig)
	}

	return loginConfigs, nil
}

// GeneratePxClientSlice creates a slice of PxClient objects from login configurations.
func GeneratePxClientSlice(loginConfigs []*authentication.LoginConfig) ([]etc.PxClient, error) {
	var pxClients []etc.PxClient

	for i, loginConfig := range loginConfigs {
		if !loginConfig.GetSuccess() {
			continue
		}

		pxClient := etc.PxClient{
			Context:   loginConfig.GetContext(),
			ApiClient: loginConfig.GetApiClient(),
			OrigIndex: i,
		}
		pxClients = append(pxClients, pxClient)
	}

	return pxClients, nil
}

// AuthenticateClusterNodes handles authentication for all nodes in a cluster database.
func AuthenticateClusterNodes(clusterDatabase *etc.ClusterDatabase, passwordManager *authentication.PasswordManager) ([]*authentication.LoginConfig, error) {
	timeout := time.Millisecond * time.Duration(clusterDatabase.GetTimeout())
	clusterNodes := clusterDatabase.GetNodes()

	loginConfigs, err := CreateLoginConfigsFromNodes(clusterNodes)
	if err != nil {
		return nil, err
	}

	if err := authentication.AuthenticateClusterNodes(loginConfigs, passwordManager, timeout); err != nil {
		return nil, err
	}

	return loginConfigs, nil
}

// InitializePxClientSettings configures PxClient objects with settings derived from a cluster database.
func InitializePxClientSettings(clusterDB *etc.ClusterDatabase, pxClients []etc.PxClient) ([]etc.PxClient, error) {
	var initializedClients []etc.PxClient

	clusterVars := clusterDB.GetVars()
	clusterIgnition := clusterDB.GetIgnition()
	clusterAliases := clusterDB.GetAliases()
	clusterNodes := clusterDB.GetNodes()

	for _, client := range pxClients {
		nodeConfig := clusterNodes[client.OrigIndex]

		client.Vars = mergeConfigMaps(clusterVars, nodeConfig, "vars")
		client.Ignition = mergeConfigMaps(clusterIgnition, nodeConfig, "ignition")
		client.Aliases = mergeConfigMaps(clusterAliases, nodeConfig, "aliases")

		storageResponse, err := queries.GetStorage(client.ApiClient, client.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage: %v", err)
		}
		if storageResponse == nil {
			continue
		}

		storageSlice, err := convertToStorageSlice(storageResponse)
		if err != nil {
			return nil, err
		}
		client.Storage = storageSlice
		client.StorageAliases = createStorageAliases(client, storageSlice)
		client.Parent = clusterDB

		initializedClients = append(initializedClients, client)
	}

	return queries.AssignStorage(initializedClients), nil
}

// mergeConfigMaps combines specific configuration maps from the node and the cluster.
func mergeConfigMaps(clusterConfig, nodeConfig map[string]interface{}, key string) map[string]interface{} {
	nodeSpecificConfig := configmap.GetMapEntryWithDefault(nodeConfig, key, map[string]interface{}{})
	return configmap.MergeMapRecursive(clusterConfig, nodeSpecificConfig)
}

// convertToStorageSlice converts the storage response to a slice of map[string]interface{}.
func convertToStorageSlice(storageResponse interface{}) ([]map[string]interface{}, error) {
	responseMap, ok := storageResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid storage response format")
	}

	data, ok := responseMap["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("data field is missing or not a slice")
	}

	var storageSlice []map[string]interface{}
	for _, item := range data {
		storageItem, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid storage item format")
		}
		storageSlice = append(storageSlice, storageItem)
	}

	return storageSlice, nil
}

// createStorageAliases generates storage aliases for a PxClient based on its storage configuration.
func createStorageAliases(client etc.PxClient, storageSlice []map[string]interface{}) map[string]interface{} {
	aliases := make(map[string]interface{})
	for _, node := range client.Nodes {
		localStorage := FilterStorageByNodeName(storageSlice, node)
		sort.Strings(localStorage)

		nodeAliases := buildNodeStorageAliases(client.Aliases, localStorage)
		aliases[node] = nodeAliases
	}
	return aliases
}

// buildNodeStorageAliases creates storage aliases for a specific node.
func buildNodeStorageAliases(aliases map[string]interface{}, localStorage []string) map[string]string {
	nodeAliases := make(map[string]string)
	for key := range aliases {
		aliasValues, ok := aliases[key].([]string)
		if !ok {
			continue
		}
		if result := GetPriorityMatch(aliasValues, localStorage); result != "" {
			nodeAliases[key] = result
		}
	}
	return nodeAliases
}

/*
// sortedKeys returns sorted keys of a map.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
*/

func InitPxCluster(clusterName string) *etc.PxCluster {

	clusterConfigData := clusters.GetSystemConfiguration()
	clustersDatabase := etc.NewClustersDatabase(&clusterConfigData)
	clusterDatabase, _ := clustersDatabase.GetClusterDatabaseByName(clusterName)

	clusterNodes := clusterDatabase.GetNodes()
	var simplePasswordManager authentication.PasswordManager = authentication.NewSimplePasswordManager(clusterNodes)

	logins, err := AuthenticateClusterNodes(clusterDatabase, &simplePasswordManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)

	}

	pxClients, err := GeneratePxClientSlice(logins)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)

	}
	pxClients = queries.AddClusterResources(pxClients)

	pxClients, _ = InitializePxClientSettings(clusterDatabase, pxClients)

	pxCluster := etc.PxCluster{}
	pxCluster.SetGlobalConfigData(&clusterConfigData)
	etc.InitCluster(&pxCluster, pxClients)

	return &pxCluster
}
func BuildMappingTable(pxClients []etc.PxClient) map[string]*api.Connection {
	nodeIndexMap := make(map[string]*api.Connection)

	for index, client := range pxClients {
		for _, nodeName := range client.Nodes {
			if existingIndex, exists := nodeIndexMap[nodeName]; exists {
				fmt.Fprintf(os.Stderr, "WARN: Duplicate node '%v' found in clusters %v and %v\n", nodeName, index, existingIndex)
				continue
			}
			c := api.NewConnection(nodeName, client.ApiClient, client.Context)
			nodeIndexMap[nodeName] = c
		}
	}
	return nodeIndexMap
}
