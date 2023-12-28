package queries

import (
	"context"
	"fmt"
	"os"

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
	restResponse, _ := ConvertJsonHttpResponseToMap2(r)
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)
	return restResponse
}
