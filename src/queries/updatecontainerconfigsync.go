package queries

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"px/api"
	"px/configmap"
	"px/shared"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
)

func UpdateContainerConfigSync(node string, vmid int64, updateContainerConfigSyncRequest pxapiflat.UpdateContainerConfigSyncRequest) (*pxapiflat.CreateVM200Response, error) {
	resp, err := api.UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
	return resp, err
}

func ApplyLxc(machines []map[string]interface{}, settings []string) {

	updateContainerConfigSyncRequestObject := pxapiobject.UpdateContainerConfigSyncRequest{}
	attributeTypeDict := getAttributeTypeDict(&updateContainerConfigSyncRequestObject)
	myDict := processSettings(settings, attributeTypeDict)

	jsonData, err := json.Marshal(myDict)

	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	err = json.Unmarshal(jsonData, &updateContainerConfigSyncRequestObject)
	jsonData, err = json.Marshal(updateContainerConfigSyncRequestObject)
	//fmt.Println(string(jsonData))

	for _, machine := range machines {
		vmid, _ := configmap.GetInt(machine, "vmid")
		node, _ := configmap.GetString(machine, "node")

		fmt.Fprintf(os.Stdout, "Apply Container %v on %s\n", vmid, node)

		updateContainerConfigSyncRequest := pxapiflat.UpdateContainerConfigSyncRequest{}
		shared.CopyUpdateContainerConfigSyncRequest(&updateContainerConfigSyncRequest, &updateContainerConfigSyncRequestObject)
		UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
	}
}
