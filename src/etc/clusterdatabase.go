package etc

import (
	"fmt"
	"os"
	"px/configmap"
)

type ClusterDatabase struct {
	table *map[string]interface{}
}

func (clusterDatabase *ClusterDatabase) GetNodes() []map[string]interface{} {
	nodes, err := configmap.GetInterfaceSliceValueByPtr(clusterDatabase.table, "nodes")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find a cluster/node setting in config")
		os.Exit(1)
	}
	return nodes
}
func (clusterDatabase *ClusterDatabase) GetTimeout() int {
	return configmap.GetIntWithDefault(*clusterDatabase.table, "timeout", 500)
}
func (clusterDatabase *ClusterDatabase) GetVars() map[string]interface{} {
	return configmap.GetMapEntryWithDefault(*clusterDatabase.table, "vars", map[string]interface{}{})
}
func (clusterDatabase *ClusterDatabase) GetIgnition() map[string]interface{} {
	return configmap.GetMapEntryWithDefault(*clusterDatabase.table, "ignition", map[string]interface{}{})
}
func (clusterDatabase *ClusterDatabase) GetAliases() map[string]interface{} {
	return configmap.GetMapEntryWithDefault(*clusterDatabase.table, "aliases", map[string]interface{}{})
}
func (clusterDatabase *ClusterDatabase) GetCluster() map[string]interface{} {
	return *clusterDatabase.table
}
