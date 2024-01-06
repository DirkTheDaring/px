package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"px/api"
	"px/configmap"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
)

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
		CopyUpdateContainerConfigSyncRequest(&updateContainerConfigSyncRequest, &updateContainerConfigSyncRequestObject)
		//UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
		//resp, err := api.UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
		api.UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
	}
}
