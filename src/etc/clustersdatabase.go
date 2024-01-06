package etc

import (
	"errors"
	"fmt"
	"os"
	"px/configmap"
	"strconv"
)

type ClustersDatabase struct {
	table *map[string]interface{}
}

func NewClustersDatabase(table *map[string]interface{}) *ClustersDatabase {
	simpleConfigDatabase := ClustersDatabase{table: table}
	return &simpleConfigDatabase
}

func (clustersDatabase *ClustersDatabase) getClusterIndex(clusters []map[string]interface{}, name string) (int, error) {
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

func (clustersDatabase *ClustersDatabase) getClusters() []map[string]interface{} {
	clusters, err := configmap.GetInterfaceSliceValueByPtr(clustersDatabase.table, "clusters")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find a clusters setting in config")
		os.Exit(1)
	}
	return clusters
}

func (clustersDatabase *ClustersDatabase) GetClusterDatabaseByName(clusterName string) (*ClusterDatabase, error) {
	clusters := clustersDatabase.getClusters()
	clusterIndex, err := clustersDatabase.getClusterIndex(clusters, clusterName)
	if err != nil {
		return nil, fmt.Errorf("could not find a cluster: %s", clusterName)
	}
	cluster := clusters[clusterIndex]

	cluster2 := ClusterDatabase{table: &cluster}

	return &cluster2, nil

}
