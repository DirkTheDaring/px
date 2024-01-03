package main

import (
	"fmt"
	"os"
	"px/authentication"
	"px/cmd"
	"px/configmap"
	"px/etc"
	"px/proxmox/clusters"
	"px/queries"
	"px/shared"
	"regexp"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func AssignStorage(pxClients []etc.PxClient) []etc.PxClient {
	list := []etc.PxClient{}
	for _, pxClient := range pxClients {
		storageResponse := queries.GetStorage(pxClient.ApiClient, pxClient.Context)
		if storageResponse == nil {
			continue
		}
		storageSlice, _ := configmap.GetInterfaceSliceValue(storageResponse, "data")
		pxClient.Storage = storageSlice
		//fmt.Fprintf(os.Stderr, "%v\n", pxClient.Storage)
		list = append(list, pxClient)
	}
	return list
}

func AddClusterResources(pxClients []etc.PxClient) []etc.PxClient {
	list := []etc.PxClient{}
	for _, pxClient := range pxClients {
		// get the cluster resource and filter out only the qemu and lxc
		// there is also a type=="storage" in the answer, but it doesn't
		// contain sufficient information (path missing) for the storage
		// therefore this is not evaluated
		clusterResources := queries.GetClusterResources(pxClient.ApiClient, pxClient.Context)
		if clusterResources == nil {
			continue
		}
		clusterResourcesSlice, _ := configmap.GetInterfaceSliceValue(clusterResources, "data")
		machines := []map[string]interface{}{}
		storage := []map[string]interface{}{}
		for _, clusterResource := range clusterResourcesSlice {
			_type := clusterResource["type"]
			if _type == "qemu" || _type == "lxc" {
				//DumpSystem(clusterResource)
				machines = append(machines, clusterResource)
				continue
			}
			if _type == "storage" {
				storage = append(storage, clusterResource)
				continue
			}
		}

		/* Unfortunately the cluster function to get all available nodes, does not work on
		   nodes which have not actually joined a cluster. So Fallback to the heuristic,
		   that only where storage is a vm/lxc container is possible. Therefore, derive the
		   nodes in the cluster from storage items, which always the nodes where they are on assigned
		   Hint you cannot use ClusterResources, as a node without a virtual machine/lxc does not show
		   up in the list.
		   Works for clusters and unjoined nodes then
		*/

		nodeList := []string{}
		for _, storageItem := range storage {

			nodeName := storageItem["node"].(string)

			for _, item := range nodeList {
				if item == nodeName {
					goto Skip
				}
			}
			nodeList = append(nodeList, nodeName)
		Skip:
		}

		sort.Strings(nodeList)
		//fmt.Fprintf(os.Stderr, "nodeList: %v\n", nodeList)
		pxClient.Nodes = nodeList
		pxClient.Machines = machines
		list = append(list, pxClient)
	}
	return list
}

func GetPriorityMatch(prioritylist []string, list []string) string {
	for _, priorityItem := range prioritylist {
		for _, item := range list {
			match, _ := regexp.MatchString(priorityItem, item)
			if match {
				return item
			}
		}
	}
	return ""
}

func FilterStorageByNodeName(storageSlice []map[string]interface{}, nodeName string) []string {
	localStorage := []string{}
	for _, storageItem := range storageSlice {
		storage, _ := configmap.GetString(storageItem, "storage")
		storageNodes, found := configmap.GetString(storageItem, "nodes")
		if found && storageNodes != nodeName {
			//fmt.Fprintf(os.Stderr, "storageNodes: %T %v %v\n", storageNodes, storageNodes, storage)
			continue
		}
		localStorage = append(localStorage, storage)
	}
	return localStorage
}

func initConfig() {
	//fmt.Fprintf(os.Stderr, "initConfig() clustername: %v\n", cmd.ClusterName)
	etc.GlobalConfigData = clusters.GetSystemConfiguration()
	clusterIndex, err := shared.GetClusterIndex(etc.GlobalConfigData, cmd.ClusterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find a cluster: %s\n", cmd.ClusterName)
		os.Exit(1)
	}
	clusters, _ := configmap.GetInterfaceSliceValue(etc.GlobalConfigData, "clusters")
	cluster := clusters[clusterIndex]
	clusterNodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")

	var timeout time.Duration = time.Millisecond * time.Duration(configmap.GetIntWithDefault(cluster, "timeout", 500))

	pxClients := authentication.LoginClusterNodes(clusterNodes, timeout)
	pxClients = AddClusterResources(pxClients)

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
			localStorage := FilterStorageByNodeName(storageSlice, node)
			sort.Strings(localStorage)
			keys := []string{}
			for aliasName := range pxClient.Aliases {
				keys = append(keys, aliasName)
			}
			sort.Strings(keys)
			storageAliases2 := map[string]string{}
			for _, key := range keys {
				aliasValues := configmap.GetStringSliceWithDefault(pxClient.Aliases, key, []string{})
				result := GetPriorityMatch(aliasValues, localStorage)
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
	pxClients = AssignStorage(pxClients)

	// All vmids are assigned (and duplicates excluded)
	etc.GlobalPxCluster = etc.ProcessCluster(pxClients)
	//fmt.Fprintf(os.Stderr, "******************** Init ended\n")
}

func main() {
	cobra.OnInitialize(initConfig)
	cmd.Execute()
	os.Exit(0)
}
