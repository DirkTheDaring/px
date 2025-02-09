package queries

import (
	"context"
	"fmt"
	"os"
	"px/configmap"
	"px/etc"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
)

func GetStorage(apiClient *pxapiflat.APIClient, context context.Context) (map[string]interface{}, error) {
	//{
	//  "data": [
	//    {
	//      "content": "rootdir,images",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "shared": 0,
	//      "storage": "BALDUR1",
	//      "type": "lvm",
	//      "vgname": "BALDUR1"
	//    },
	//    {
	//      "content": "iso,backup,images,snippets,vztmpl,rootdir",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "export": "/volume2/proxmox-shared",
	//      "options": "vers=3,soft",
	//      "path": "/mnt/pve/myshared",
	//      "server": "192.168.178.249",
	//      "shared": 1,
	//      "storage": "myshared",
	//      "type": "nfs"
	//    },
	//    {
	//      "content": "images,rootdir",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "storage": "samsung-ssd-500GB",
	//      "thinpool": "samsung-ssd-500GB",
	//      "type": "lvmthin",
	//      "vgname": "samsung-ssd-500GB"
	//    },
	//    {
	//      "content": "backup,iso,vztmpl,images",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "path": "/var/lib/vz",
	//      "storage": "local",
	//      "type": "dir"
	//    },
	//    {
	//      "content": "rootdir,images",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "storage": "local-lvm",
	//      "thinpool": "data",
	//      "type": "lvmthin",
	//      "vgname": "pve"
	//    },
	//    {
	//      "content": "rootdir,images",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "shared": 0,
	//      "storage": "BALDUR2",
	//      "type": "lvm",
	//      "vgname": "BALDUR2"
	//    },
	//    {
	//      "content": "iso",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "path": "/etc/pve/ignition",
	//      "prune-backups": "keep-all=1",
	//      "shared": 0,
	//      "storage": "ignition",
	//      "type": "dir"
	//    },
	//    {
	//      "content": "images,vztmpl,snippets,rootdir,iso,backup",
	//      "digest": "717a0a0cdefeb95b8de458fda15770dc4603253b",
	//      "export": "/volume2/proxmox-shared",
	//      "options": "vers=3,soft",
	//      "path": "/mnt/pve/shared",
	//      "server": "192.168.178.249",
	//      "shared": 1,
	//      "storage": "shared",
	//      "type": "nfs"
	//    }
	//  ]
	//}
	_, r, err := apiClient.StorageApi.GetStorage(context).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageApi.GetStorage``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, err
	}
	//resources := clusterResourcesResponse.GetData()
	restResponse, _ := ConvertJsonHttpResponseBodyToMap(r)
	//fmt.Fprintf(os.Stderr, "resp: %v %v\n", len(resources), restResponse["data"])
	//fmt.Fprintf(os.Stderr, "resp: %v\n", restResponse["data"])
	//fmt.Fprintf(os.Stderr, "resp: %v\n", r)

	//json := configmap.DataToJSON(restResponse)
	//fmt.Fprintf(os.Stdout, "%s\n", json)

	return restResponse, nil
}
func AssignStorage(pxClients []etc.PxClient) []etc.PxClient {
	list := []etc.PxClient{}
	for _, pxClient := range pxClients {
		storageResponse, _ := GetStorage(pxClient.ApiClient, pxClient.Context)
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
