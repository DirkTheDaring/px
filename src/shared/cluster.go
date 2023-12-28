package shared

import (
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
)

func ClusterList() {
	//fmt.Fprintf(os.Stderr, "cluster: %s\n", commandLineInterface.Cluster)
	clusters, _ := configmap.GetInterfaceSliceValue(etc.GlobalConfigData, "clusters")
	for _, cluster := range clusters {
		fmt.Fprintf(os.Stdout, "%s\n", cluster["name"])
	}
	os.Exit(0)
}
