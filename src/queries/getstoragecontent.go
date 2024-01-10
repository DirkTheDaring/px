package queries

import (
	"encoding/json"
	"fmt"
	"net/http"

	//"os"
	"px/configmap"
	"px/etc"
)

func ConvertJsonHttpResponseBodyToMap(r *http.Response) (map[string]interface{}, error) {
	if r == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer r.Body.Close()

	var restResponse map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&restResponse); err != nil {
		return nil, err
	}
	return restResponse, nil
}

func GetJsonStorageContent(pxClient etc.PxClient, node, storage string) (map[string]interface{}, error) {
	_, r, err := pxClient.ApiClient.NodesAPI.GetStorageContent(pxClient.Context, node, storage).Execute()
	if err != nil {
		return nil, fmt.Errorf("error calling GetStorageContent: %v", err)
	}

	return ConvertJsonHttpResponseBodyToMap(r)
}

func GetStorageContentAll(pxClients []etc.PxClient) ([]etc.PxClient, error) {
	var updatedClients []etc.PxClient

	for _, pxClient := range pxClients {
		storageContent, _ := GetClusterStorageContent(pxClient)

		//jsonBinary, _ := json.Marshal(storageContent)
		//fmt.Fprintf(os.Stdout, "host %v = %v\n", pxClient.Nodes[0], string(jsonBinary))

		pxClient.StorageContent = storageContent
		updatedClients = append(updatedClients, pxClient)
	}

	return updatedClients, nil
}

func GetStorageContentAll2(pxClients []*etc.PxClient) error {
	for _, pxClient := range pxClients {
		storageContent, _ := GetClusterStorageContent(*pxClient)
		pxClient.StorageContent = storageContent
	}
	return nil

}

// The following function transforms this into a lookup table, with the machine name ("pve") at the root
// {
//   "pve": {
//     "ignition": [
//       {
//         "content": "iso",
//         "ctime": 1660132400,
//         "format": "iso",
//         "size": 16384,
//         "volid": "ignition:iso/prod1-master2.ign.iso"
//       }
//     ],
//     "local": [
//       {
//         "content": "images",
//         "ctime": 1699740058,
//         "format": "qcow2",
//         "parent": null,
//         "size": 10737418240,
//         "used": 1638281216,
//         "vmid": 0,
//         "volid": "local:000/fedora-coreos-38.20231027.3.2-qemu.x86_64.qcow2"
//       }
//     ],
//     "shared": [
//       {
//         "content": "images",
//         "ctime": 1659204723,
//         "format": "qcow2",
//         "parent": null,
//         "size": 5368709120,
//         "used": 460128256,
//         "vmid": 0,
//         "volid": "shared:000/Fedora-Cloud-Base-36-1.5.aarch64.qcow2"
//       }
//     ]
//   }
// }}

func GetClusterStorageContent(pxClient etc.PxClient) (map[string]interface{}, error) {

	storageContent := make(map[string]interface{})

	for _, node := range pxClient.Nodes {
		nodeStorageLookup, _ := GetNodeStorageContent(node, pxClient.Storage, pxClient)
		storageContent[node] = nodeStorageLookup
	}

	return storageContent, nil
}

// We get back content list of a specific storage
//    {
//      "size": 2581094400,
//      "content": "iso",
//      "volid": "myshared:iso/windows-software.iso",
//      "ctime": 1617474936,
//      "format": "iso"
//    },
//    {
//      "ctime": 1640352024,
//      "volid": "myshared:snippets/cloud-config-ubuntu-vendor-data",
//      "content": "snippets",
//      "size": 182,
//      "format": "snippet"
//    },
//    {
//      "format": "txz",
//      "content": "vztmpl",
//      "size": 48226864,
//      "ctime": 1659189753,
//      "volid": "myshared:vztmpl/Fedora-Container-Base-36-20220719.0-sshd.x86_64.tar.xz"
//    },

func GetNodeStorageContent(node string, pxClientStorage []map[string]interface{}, pxClient etc.PxClient) (map[string]interface{}, error) {

	nodeStorageLookup := make(map[string]interface{})

	for _, item := range pxClientStorage {
		storageType, ok := item["type"].(string)
		if !ok || !(storageType == "dir" || storageType == "nfs") {
			continue
		}

		storage, ok := item["storage"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid storage type")
		}

		result, err := GetJsonStorageContent(pxClient, node, storage)
		if err != nil {
			return nil, err
		}

		data, err := configmap.GetInterfaceSliceValue(result, "data")
		if err != nil {
			return nil, err
		}

		nodeStorageLookup[storage] = data
	}

	return nodeStorageLookup, nil
}
