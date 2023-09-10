package shared

import (
	"fmt"
	"os"
	"px/configmap"
)

func ClusterList() {
	//fmt.Fprintf(os.Stderr, "cluster: %s\n", commandLineInterface.Cluster)
	clusters, _ := configmap.GetInterfaceSliceValue(GlobalConfigData, "clusters")
	for _, cluster := range clusters {
		fmt.Fprintf(os.Stdout, "%s\n", cluster["name"])
	}
	os.Exit(0)
}
