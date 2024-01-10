package queries

import (
	"context"
	"errors"
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
	"sort"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

/*
{
  "data": [
    {
      "cpu": 0.013426763498155,
      "disk": 0,
      "diskread": 12506161605120,
      "diskwrite": 149433603584,
      "id": "qemu/100",
      "maxcpu": 2,
      "maxdisk": 53687091200,
      "maxmem": 17179869184,
      "mem": 15037038592,
      "name": "vm-test-01",
      "netin": 13957404640,
      "netout": 5775371384,
      "node": "pve",
      "status": "running",
      "template": 0,
      "type": "qemu",
      "uptime": 579398,
      "vmid": 100
    },
	{
      "cpu": 0,
      "disk": 0,
      "diskread": 0,
      "diskwrite": 0,
      "id": "lxc/103",
      "maxcpu": 1,
      "maxdisk": 8589934592,
      "maxmem": 1073741824,
      "mem": 0,
      "name": "ct-test-01",
      "netin": 0,
      "netout": 0,
      "node": "pve",
      "status": "stopped",
      "template": 0,
      "type": "lxc",
      "uptime": 0,
      "vmid": 103
    },
    {
      "cgroup-mode": 2,
      "cpu": 0.0749163879598662,
      "disk": 64552714240,
      "id": "node/pve",
      "level": "",
      "maxcpu": 6,
      "maxdisk": 101918625792,
      "maxmem": 33364914176,
      "mem": 17379082240,
      "node": "pve",
      "status": "online",
      "type": "node",
      "uptime": 579546
    },
	{
      "content": "rootdir,images",
      "disk": 585857903886,
      "id": "storage/pve/local-lvm",
      "maxdisk": 853398257664,
      "node": "pve",
      "plugintype": "lvmthin",
      "shared": 0,
      "status": "available",
      "storage": "local-lvm",
      "type": "storage"
    },
    {
      "content": "rootdir,iso,backup,images,vztmpl,snippets",
      "disk": 2516818722816,
      "id": "storage/pve/shared",
      "maxdisk": 9596038807552,
      "node": "pve",
      "plugintype": "nfs",
      "shared": 1,
      "status": "available",
      "storage": "shared",
      "type": "storage"
    },
    {
      "content": "images,backup,iso,vztmpl",
      "disk": 64552718336,
      "id": "storage/pve/local",
      "maxdisk": 101918625792,
      "node": "pve",
      "plugintype": "dir",
      "shared": 0,
      "status": "available",
      "storage": "local",
      "type": "storage"
    },
    {
      "id": "sdn/pve/localnetwork",
      "node": "pve",
      "sdn": "localnetwork",
      "status": "ok",
      "type": "sdn"
    }
  ]
}
*/

func GetClusterResources(apiClient *pxapiflat.APIClient, context context.Context) map[string]interface{} {
	_, r, err := apiClient.ClusterAPI.GetClusterResources(context).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ClusterApi.GetClusterResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil
	}
	//resources := clusterResourcesResponse.GetData()
	restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)
	return restResponse
}

func AddClusterResource(pxClient *etc.PxClient) error {
	// get the cluster resource and filter out only the qemu and lxc
	// there is also a type=="storage" in the answer, but it doesn't
	// contain sufficient information (path missing) for the storage
	// therefore this is not evaluated
	clusterResources := GetClusterResources(pxClient.ApiClient, pxClient.Context)
	if clusterResources == nil {
		return errors.New("")
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

	return nil
}

/*
	func AddClusterResources(pxClients []etc.PxClient) []etc.PxClient {
		list := []etc.PxClient{}
		for _, pxClient := range pxClients {
			pxClient, err := AddClusterResource(pxClient)
			if err == nil {
				list = append(list, pxClient)
			}
		}
		return list
	}
*/
func AddClusterResources(pxClients []*etc.PxClient) {

	for _, pxClient := range pxClients {
		err := AddClusterResource(pxClient)
		if err == nil {
			// FIXME her wee need to give an error status
		}
	}

}
