package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func ConvertJsonHttpResponseToMap(r *http.Response) map[string]interface{} {
	restResponse := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&restResponse)
	return restResponse
}

func GetJsonStorageContent(pxClient PxClient, node string, storage string) map[string]interface{} {
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
	_, r, err := pxClient.ApiClient.NodesApi.GetStorageContent(pxClient.Context, node, storage).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NodesApi.GetStorageContent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil
	}
	restResponse := ConvertJsonHttpResponseToMap(r)
	return restResponse
}
