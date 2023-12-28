package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
)

func DumpSystem(configData map[string]interface{}) {
	json := configmap.DataToJSON(configData)
	fmt.Fprintf(os.Stdout, "%s\n", json)
}

func DumpNodes(configData map[string]interface{}) {
	fmt.Println("dump nodes called")

	for key, value := range etc.GlobalPxCluster.PxClientLookup {
		fmt.Fprintf(os.Stderr, "%v %v\n", key, value)
	}
	os.Exit(0)
}
func DumpAny(data any) {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "<NIL>\n", json)
		return
	}
	fmt.Fprintf(os.Stderr, "%v\n", json)
}
