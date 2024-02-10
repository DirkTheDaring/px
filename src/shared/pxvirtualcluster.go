package shared

import (
	"fmt"
	"os"
	/*
		"px/authentication"
	*/
	"px/configmap"
	"px/etc"
	"strings"
	/*
		"fmt"
		"os"
		"px/authentication"
	*///	"px/shared"
)

/* The intention of this struct is to provide a unified access to a cluster. As we offer the option to join indepedent nodes to a cluster, without using the proxmox cluster technology we call this virtual cluster
 *
 * As a virtual cluster needs to care what are his currrent working nodes, he need to query the different real nodes for them and give this information out.

 *  A realNode can contain 1..n nodes, which means it is either a cluster or a single node which has not joined a cluster
 */

/* a joined cluster has list of nodes which are working */
/* This already requires authenticatin, in order to get the list of cluster nodes in a cluster */
/* And a concept of "failed" nodes which are at temporarily not available */

/****************************************************************************/
type NodeConfig struct {
	enabled            bool
	urls               []string
	insecureskipverify bool
	domain             string
	username           string
	failed             bool // true if node is not reachable
}

func NewNodeConfig(node map[string]interface{}) (*NodeConfig, error) {
	var domain string = ""
	var username string = "root@pem"
	var url string = "https://localhost:8006"

	enabled := configmap.GetBoolWithDefault(node, "enabled", true)
	insecureskipverify := configmap.GetBoolWithDefault(node, "insecureskipverify", false)

	username = configmap.GetStringWithDefault(node, "username", username)
	result := strings.Split(username, "@")
	username = result[0]
	if len(result) > 1 {
		domain = result[1]
	}

	domain = configmap.GetStringWithDefault(node, "domain", domain)
	urls := configmap.GetStringSliceWithDefault(node, "urls", []string{})

	if len(urls) == 0 {
		url := configmap.GetStringWithDefault(node, "url", url)
		urls = append(urls, url)
	}

	newNodeConfig := NodeConfig{
		enabled:            enabled,
		urls:               urls,
		insecureskipverify: insecureskipverify,
		domain:             domain,
		username:           username,
		failed:             false,
	}
	return &newNodeConfig, nil
}
func (nodeConfig *NodeConfig) Dump() {
	fmt.Fprintf(os.Stderr, "NodeConfig.enabled:            %v\n", nodeConfig.enabled)
	fmt.Fprintf(os.Stderr, "NodeConfig.urls:               %v\n", nodeConfig.urls)
	fmt.Fprintf(os.Stderr, "NodeConfig.insecureskipverify: %v\n", nodeConfig.insecureskipverify)
	fmt.Fprintf(os.Stderr, "NodeConfig.domain:             %v\n", nodeConfig.domain)
	fmt.Fprintf(os.Stderr, "NodeConfig.username:           %v\n", nodeConfig.username)
}

/****************************************************************************/

type PxJoinedCluster struct {
	//nodes map[string]interface{}
}

func NewPxJoinedCluster(clusterDatabase *etc.ClusterDatabase) (*PxJoinedCluster, error) {
	PxJoinedCluster := PxJoinedCluster{}

	clusterNodes := clusterDatabase.GetNodes()

	for _, clusterNode := range clusterNodes {
		result, _ := NewNodeConfig(clusterNode)
		//result.Dump()
	}
	/*

		var simplePasswordManager authentication.PasswordManager = authentication.NewSimplePasswordManager(clusterNodes)

		logins, err := AuthenticateClusterNodes(clusterDatabase, &simplePasswordManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)

		}
		fmt.Fprintf(os.Stderr, "%v %v", logins, err)
	*/
	/*
		pxClients, err := GeneratePxClientSlice(logins)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)

		}
	*/
	return &PxJoinedCluster, nil
}

func (pxJoinedCluster *PxJoinedCluster) GetMachine(node string, id int64) {
}

func (pxJoinedCluster *PxJoinedCluster) GetNodes() {
}

/****************************************************************************/

type PxVirtualCluster struct {
	pxJoinedCluster *PxJoinedCluster
}

func NewPxVirtualCluster(pxJoinedCluster *PxJoinedCluster) *PxVirtualCluster {
	pxVirtualCluster := PxVirtualCluster{pxJoinedCluster: pxJoinedCluster}
	return &pxVirtualCluster
}

func (vc *PxVirtualCluster) GetNodes() {
}

// id is a string here, because we can have something like pve/112 and pve2/112, different nodes, same id
func (vc *PxVirtualCluster) GetMachine(id string) {
}

/****************************************************************************/
