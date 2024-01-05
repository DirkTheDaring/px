package main

import (
	"os"
	"px/authentication"
	"px/cmd"
	"px/configmap"
	"px/etc"
	"px/queries"
	"px/shared"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func InitConfig() {
	//fmt.Fprintf(os.Stderr, "initConfig() clustername: %v\n", cmd.ClusterName)
	etc.InitGlobalConfigData()
	cluster := etc.GetClusterByName(cmd.ClusterName)

	//fmt.Fprintf(os.Stderr, "initConfig() cluster: %v\n", cluster)

	clusterNodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")
	var timeout time.Duration = time.Millisecond * time.Duration(configmap.GetIntWithDefault(cluster, "timeout", 500))

	var pm authentication.PasswordManager = authentication.NewSimplePasswordManager(clusterNodes)

	pxClients := authentication.LoginClusterNodes(clusterNodes, pm.GetCredentials, timeout)

	pxClients = queries.AddClusterResources(pxClients)

	list := []etc.PxClient{}

	clusterVars := configmap.GetMapEntryWithDefault(cluster, "vars", map[string]interface{}{})
	clusterIgnition := configmap.GetMapEntryWithDefault(cluster, "ignition", map[string]interface{}{})
	clusterAliases := configmap.GetMapEntryWithDefault(cluster, "aliases", map[string]interface{}{})

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
		list = append(list, pxClient)
	}
	pxClients = list

	// Add Storage
	pxClients = queries.AssignStorage(pxClients)

	// All vmids are assigned (and duplicates excluded)
	etc.GlobalPxCluster = etc.ProcessCluster(pxClients)
	//fmt.Fprintf(os.Stderr, "******************** Init ended\n")
}
func main() {
	cobra.OnInitialize(InitConfig)
	cmd.Execute()
	os.Exit(0)
}
