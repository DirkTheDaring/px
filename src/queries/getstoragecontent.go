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

func GetStorageContentAll(pxClients []etc.PxClient) ([]etc.PxClient, error) {
	var updatedClients []etc.PxClient

	for _, pxClient := range pxClients {
		storageContent := make(map[string]interface{})

		for _, node := range pxClient.Nodes {
			nodeStorageLookup := make(map[string]interface{})

			for _, item := range pxClient.Storage {
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

			storageContent[node] = nodeStorageLookup
		}

		pxClient.StorageContent = storageContent
		updatedClients = append(updatedClients, pxClient)
	}

	return updatedClients, nil
}
