package shared

import (
	"fmt"
	"os"
	"px/configmap"
)

func DumpSystem(configData map[string]interface{}) {
	json := configmap.DataToJSON(configData)
	fmt.Fprintf(os.Stdout, "%s\n", json)
}

func DumpNodes(configData map[string]interface{}) {
	fmt.Println("dump nodes called")

	for key, value := range GlobalPxCluster.PxClientLookup {
		fmt.Fprintf(os.Stderr, "%v %v\n", key, value)
	}
	os.Exit(0)
}
