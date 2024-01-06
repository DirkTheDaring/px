package shared

import (
	"px/api"
	"px/etc"
)

// ATTENTION!
// this initializes global variables!

func InitConfig(clusterName string) {
	pxCluster := InitPxCluster(clusterName)
	connections := BuildMappingTable(pxCluster.GetPxClients())
	api.GlobalSimpleApi = api.NewSimpleAPI(connections)
	etc.GlobalPxCluster = pxCluster
}
