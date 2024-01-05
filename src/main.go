package main

import (
	"os"
	"px/authentication"
	"px/cmd"
	"px/configmap"
	"px/etc"
	"px/proxmox/clusters"
	"px/queries"
	"px/shared"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func InitConfig() {
	etc.GlobalPxCluster = InitPxCluster()
}

func InitPxCluster() *etc.PxCluster {

	clusterConfigData := clusters.GetSystemConfiguration()
	clustersDatabase := etc.NewClustersDatabase(&clusterConfigData)
	clusterDatabase, _ := clustersDatabase.GetClusterDatabaseByName(cmd.ClusterName)

	pxClients := InitPxClients(clusterDatabase)

	pxCluster := etc.PxCluster{}
	pxCluster.SetGlobalConfigData(&clusterConfigData)
	etc.InitCluster(&pxCluster, pxClients)

	return &pxCluster
}

func InitPxClients(clusterDatabase *etc.ClusterDatabase) []etc.PxClient {

	clusterNodes := clusterDatabase.GetNodes()

	var pm authentication.PasswordManager = authentication.NewSimplePasswordManager(clusterNodes)
	var timeout time.Duration = time.Millisecond * time.Duration(clusterDatabase.GetTimeout())

	pxClients := authentication.LoginClusterNodes(clusterNodes, pm.GetCredentials, timeout)

	pxClients = queries.AddClusterResources(pxClients)

	list := []etc.PxClient{}

	clusterVars := clusterDatabase.GetVars()
	clusterIgnition := clusterDatabase.GetIgnition()
	clusterAliases := clusterDatabase.GetAliases()

	for _, pxClient := range pxClients {
		nodeConfig := clusterNodes[pxClient.OrigIndex]

		vars := configmap.GetMapEntryWithDefault(nodeConfig, "vars", map[string]interface{}{})
		pxClient.Vars = configmap.MergeMapRecursive(clusterVars, vars)

		ignition := configmap.GetMapEntryWithDefault(nodeConfig, "ignition", map[string]interface{}{})
		pxClient.Ignition = configmap.MergeMapRecursive(clusterIgnition, ignition)

		aliases := configmap.GetMapEntryWithDefault(nodeConfig, "aliases", map[string]interface{}{})
		pxClient.Aliases = configmap.MergeMapRecursive(clusterAliases, aliases)

		storageResponse := queries.GetStorage(pxClient.ApiClient, pxClient.Context)
		if storageResponse == nil {
			// FIXME
			continue
		}
		storageSlice, _ := configmap.GetInterfaceSliceValue(storageResponse, "data")
		pxClient.Storage = storageSlice

		storageAliases := map[string]interface{}{}
		for _, node := range pxClient.Nodes {
			localStorage := shared.FilterStorageByNodeName(storageSlice, node)
			sort.Strings(localStorage)
			keys := []string{}
			for aliasName := range pxClient.Aliases {
				keys = append(keys, aliasName)
			}
			sort.Strings(keys)
			storageAliases2 := map[string]string{}
			for _, key := range keys {
				aliasValues := configmap.GetStringSliceWithDefault(pxClient.Aliases, key, []string{})
				result := shared.GetPriorityMatch(aliasValues, localStorage)
				if result == "" {
					continue
				}
				storageAliases2[key] = result
			}
			storageAliases[node] = storageAliases2
		}
		//proxmox.DumpJson(storageAliases)
		pxClient.StorageAliases = storageAliases
		pxClient.Parent = clusterDatabase
		list = append(list, pxClient)
	}
	pxClients = list

	// Add Storage
	pxClients = queries.AssignStorage(pxClients)
	return pxClients

}

func main() {
	cobra.OnInitialize(InitConfig)
	cmd.Execute()
	os.Exit(0)
}
