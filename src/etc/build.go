package etc

import (
	"fmt"
	"os"
	"px/configmap"
	"sort"
	"strconv"
)

func buildMappingTable(pxClients []*PxClient) (map[string]int, []string) {
	nodeIndexMap := make(map[string]int)
	var nodeList []string

	for index, client := range pxClients {
		for _, nodeName := range client.Nodes {
			if existingIndex, exists := nodeIndexMap[nodeName]; exists {
				fmt.Fprintf(os.Stderr, "WARN: Duplicate node '%v' found in clusters %v and %v\n", nodeName, index, existingIndex)
				continue
			}
			nodeIndexMap[nodeName] = index
			nodeList = append(nodeList, nodeName)
		}
	}
	return nodeIndexMap, nodeList
}

func buildMappingTableForMachines(pxClients []*PxClient) (map[string]map[string]interface{}, []map[string]interface{}) {
	vmidMachineMap := make(map[string]map[string]interface{})
	vmidMachineMapInternal := make(map[int]map[string]interface{})
	var machineList []map[string]interface{}

	for _, client := range pxClients {
		for _, machine := range client.Machines {
			nodeName, okNode := configmap.GetString(machine, "node")
			vmid, okVMID := configmap.GetInt(machine, "vmid")

			if !okNode || !okVMID {
				fmt.Fprintf(os.Stderr, "Error: 'node' or 'vmid' missing for machine: %v\n", machine)
				continue
			}

			if existingMachine, exists := vmidMachineMapInternal[vmid]; exists {
				if true {
					//fmt.Fprintf(os.Stderr, "WARN: VMID %v conflict between nodes '%v' and '%v' adding anyway\n", vmid, nodeName, existingMachine["node"])
				} else {
					fmt.Fprintf(os.Stderr, "WARN: VMID %v conflict between nodes '%v' and '%v'\n", vmid, nodeName, existingMachine["node"])
					continue
				}

			}
			vmidStr := strconv.FormatInt(int64(vmid), 10)
			keyStr := fmt.Sprintf("%s/%s", nodeName, vmidStr)

			vmidMachineMapInternal[vmid] = machine
			vmidMachineMap[keyStr] = machine
			machineList = append(machineList, machine)
		}
	}
	return vmidMachineMap, machineList
}

func InitCluster(pxCluster *PxCluster, pxClients []*PxClient) {

	/*
		var oldPxClients []PxClient
		for _, client := range pxClients {
			oldPxClients = append(oldPxClients, *client)
		}
		pxCluster.pxClients = oldPxClients
	*/

	pxCluster.pxClients2 = pxClients

	nodeIndexMap, nodeList := buildMappingTable(pxCluster.pxClients2)
	pxCluster.pxClientLookup = nodeIndexMap

	newMap := make(map[string]*PxClient)
	for key, value := range nodeIndexMap {
		newMap[key] = pxClients[value]
	}
	//pxCluster.api = api.NewSSimpleAPI(newMap)

	sort.Strings(nodeList)
	pxCluster.nodes = nodeList

	vmidMachineMap, machines := buildMappingTableForMachines(pxCluster.pxClients2)
	pxCluster.uniqueMachines = vmidMachineMap

	//StringSortMachines(machines, []string{"name"}, []bool{true})
	pxCluster.machines = machines // not unique

}
