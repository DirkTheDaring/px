package etc

import (
	"errors"
	"fmt"
	"os"
	"px/configmap"
	"px/proxmox/clusters"
	"strconv"
)

func InitGlobalConfigData() {
	GlobalConfigData = clusters.GetSystemConfiguration()
}

func GetClusterIndex(configData map[string]interface{}, name string) (int, error) {
	clusters, err := configmap.GetInterfaceSliceValue(configData, "clusters")
	if err != nil {
		return -1, err
	}

	max := len(clusters)
	if max == 0 {
		return -1, errors.New("no clusters defined")
	}

	number64, err := strconv.ParseInt(name, 10, 32)
	if err == nil {
		pos := int(number64)
		if pos >= 0 && pos < max {
			return pos, nil
		} else {
			return -1, errors.New("cluster index not found: " + strconv.Itoa(pos))
		}
	}

	for i, cluster := range clusters {
		clusterName, ok := cluster["name"]
		// if cluster has no name
		if !ok {
			continue
		}
		if clusterName == name {
			return i, nil
		}
	}
	return -1, errors.New("cluster name not found: " + name)
}

func GetClusterByName(clusterName string) map[string]interface{} {
	clusterIndex, err := GetClusterIndex(GlobalConfigData, clusterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find a cluster: %s\n", clusterName)
		os.Exit(1)
	}
	clusters, _ := configmap.GetInterfaceSliceValue(GlobalConfigData, "clusters")
	cluster := clusters[clusterIndex]
	return cluster
	//clusterNodes, _ := configmap.GetInterfaceSliceValue(cluster, "nodes")
}
