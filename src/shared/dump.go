package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"px/etc"
)

func DumpNodes(pxCluster *etc.PxCluster) {
	fmt.Println("dump nodes called")
	//result := etc.GlobalPxCluster.GetPxClientLookup()
	result := pxCluster.GetPxClientLookup()
	for key, value := range result {
		fmt.Fprintf(os.Stderr, "%v %v\n", key, value)
	}
	os.Exit(0)
}

func DumpAny(data any) {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "<NIL>\n")
		return
	}
	fmt.Fprintf(os.Stderr, "%v\n", json)
}
