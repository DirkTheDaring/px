package main

import (
	"fmt"
	"os"
	"px/authentication"
	"px/cmd"
	"px/configmap"
	"px/ignition"
	"px/proxmox/clusters"
	"px/shared"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"px/queries"

	"github.com/spf13/cobra"
)

/*
InOrSkipIfEmpty dos there is one command which does the same  cat
echo hello | findstr.exe "^"
*/
func Test() {
	testdata := ignition.LoadEmbeddedYamlFile("defaults/test.yaml")
	//result := ignition.CreateIgnition(testdata)

	//result := ignition.CreateProxmoxIgnition(testdata)
	//fmt.Fprintf(os.Stderr, result.String())
	ignition.CreateProxmoxIgnitionFile(testdata, "prod1-master1"+".ign")
}

func DumpSystem(configData map[string]interface{}, clusterName string) {
	json := configmap.DataToJSON(configData)
	fmt.Fprintf(os.Stdout, "%s\n", json)
}



func AssignStorage(pxClients []shared.PxClient) []shared.PxClient {
	list := []shared.PxClient{}
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

func AddClusterResources(pxClients []shared.PxClient) []shared.PxClient {
	list := []shared.PxClient{}
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

func renderOnConsole(outputs []map[string]interface{}, headers []string, filterColumn string, filterString string) {

	if len(headers) == 0 && len(outputs) > 0 {
		list := []string{}
		item := outputs[0]
		for key := range item {
			list = append(list, key)
		}
		headers = list
	}

	rows := [][]any{}

	maxColSizes := make([]int, len(headers))
	for i := range maxColSizes {
		maxColSizes[i] = len(headers[i])
	}
	for _, output := range outputs {
		if filterString != "" {
			value, _ := output[filterColumn].(string)
			if !strings.HasPrefix(value, filterString) {
				continue
			}
		}

		cols := []any{}
		for i, header := range headers {
			value, ok := output[header]
			if !ok {
				value = ""
			}
			valueString, ok := value.(string)
			if ok {
				if len(valueString) > maxColSizes[i] {
					maxColSizes[i] = len(valueString)
				}
				cols = append(cols, valueString)
				continue
			}

			valueFloat64, ok := value.(float64)
			if !ok {
				cols = append(cols, "")
				continue
			}
			valueInt := int(valueFloat64)
			valueString = strconv.Itoa(valueInt)
			if len(valueString) > maxColSizes[i] {
				maxColSizes[i] = len(valueString)
			}
			cols = append(cols, valueString)

		}
		//fmt.Fprintf(os.Stderr, "%v\n", cols)
		rows = append(rows, cols)
	}

	format := "%-" + strconv.Itoa(maxColSizes[0]) + "s"
	for i := 1; i < len(maxColSizes); i++ {
		format = format + " %-" + strconv.Itoa(maxColSizes[i]) + "s"
	}
	format = format + "\n"

	headers2 := []any{}
	for _, header := range headers {
		headers2 = append(headers2, strings.ToUpper(header))
	}

	fmt.Fprintf(os.Stdout, format, headers2...)

	for _, cols := range rows {
		fmt.Fprintf(os.Stdout, format, cols...)
	}

	//colSize = len(headers)
	//rowSize = len(outputs)
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
func GetKey() {

}
func initConfig() {

	//fmt.Fprintf(os.Stderr, "initConfig() clustername: %v\n", cmd.ClusterName)
	shared.GlobalConfigData = clusters.GetSystemConfiguration()
	clusterIndex, err := shared.GetClusterIndex(shared.GlobalConfigData, cmd.ClusterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find a cluster: %s\n", cmd.ClusterName)
		os.Exit(1)
	}
	clusters, _ := configmap.GetInterfaceSliceValue(shared.GlobalConfigData, "clusters")
	cluster := clusters[clusterIndex]
	clusterNodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")

	//fmt.Fprintf(os.Stdout, "result: %v\n", cluster["name"])
	var timeout time.Duration = time.Millisecond * time.Duration(configmap.GetIntWithDefault(cluster, "timeout", 500))
	//fmt.Fprintf(os.Stderr, "timeout: %v\n", timeout)

	pxClients := authentication.LoginClusterNodes(clusterNodes, timeout)
	pxClients = AddClusterResources(pxClients)

	list := []shared.PxClient{}

	clusterVars := configmap.GetMapEntryWithDefault(cluster, "vars", map[string]interface{}{})
	clusterIgnition := configmap.GetMapEntryWithDefault(cluster, "ignition", map[string]interface{}{})
	clusterAliases := configmap.GetMapEntryWithDefault(cluster, "aliases", map[string]interface{}{})

	for _, pxClient := range pxClients {
		//fmt.Fprintf(os.Stderr, "cluster nodes: %v\n", pxClient.Nodes)
		//fmt.Fprintf(os.Stderr, "initConfig() nodeConfig: %v\n", pxClient.OrigIndex)
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
		//fmt.Fprintf(os.Stderr, "initConfig() storageResponse: %v\n", storageResponse)
		storageSlice, _ := configmap.GetInterfaceSliceValue(storageResponse, "data")
		pxClient.Storage = storageSlice

		storageAliases := map[string]interface{}{}
		for _, node := range pxClient.Nodes {
			//fmt.Fprintf(os.Stderr, "Process Node: %v\n", node)
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
	//GetClusterNodes(pxClients)

	// Add Storage
	pxClients = AssignStorage(pxClients)

	// Add Storage Alias Mappings
	//pxClients = ProcessAliases(pxClients, clusterNodes)

	// All vmids are assigned (and duplicates excluded)
	shared.GlobalPxCluster = shared.ProcessCluster(pxClients)

	//fmt.Fprintf(os.Stderr, "******************** Init ended\n")
}

func main() {
	cobra.OnInitialize(initConfig)
	cmd.Execute()
	os.Exit(0)
}
