/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
	"px/queries"
	"px/shared"
	"strings"

	"github.com/spf13/cobra"
)

// dumpMacaddrCmd represents the dump command
var dumpMacaddrCmd = &cobra.Command{
	Use:   "macaddr",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		dump_macaddr()
	},
}

func init() {
	dumpCmd.AddCommand(dumpMacaddrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpMacaddrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpMacaddrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Function to get network adapters of a VM
func getNetworkAdapters(vmConfig map[string]interface{}) []map[string]interface{} {
	var networkAdaptersList []map[string]interface{}

	networkAdapters := configmap.SelectKeys("^net[0-9]+$", vmConfig)
	for _, networkAdapter := range networkAdapters {
		netConfig, _ := configmap.GetString(vmConfig, networkAdapter)
		virtioStr, _ := GetConfigOptionAndRemaining(netConfig, "virtio")
		macAddress := getMacAddress(virtioStr)

		if macAddress != "" {
			adapterInfo := map[string]interface{}{
				"net":     networkAdapter,
				"macaddr": macAddress,
			}
			networkAdaptersList = append(networkAdaptersList, adapterInfo)
		}
	}

	return networkAdaptersList
}

// Function to extract MAC address from the virtio string
func getMacAddress(virtioStr string) string {
	array := strings.Split(virtioStr, "=")
	if len(array) == 2 {
		return array[1]
	}
	return ""
}

// Function to process each machine
func processMachine(machine map[string]interface{}, list *[]map[string]interface{}) {
	vmType, _ := configmap.GetString(machine, "type")
	if vmType == "lxc" {
		return
	}
	//proxmox.DumpJson(machine)
	vmid, _ := configmap.GetInt(machine, "vmid")

	//fmt.Fprintf(os.Stderr, "%v\n", vmid)
	node, _ := configmap.GetString(machine, "node")
	vmConfig, err := queries.JSONGetVMConfig(node, int64(vmid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting VM config: %v\n", err)
		return
	}

	vmConfigData, ok := configmap.GetMapEntry(vmConfig, "data")
	if !ok {
		fmt.Fprintf(os.Stderr, "VM config data not found\n")
		return
	}

	networkAdapters := getNetworkAdapters(vmConfigData)
	for _, adapter := range networkAdapters {
		adapter["node"] = node
		adapter["vmid"] = int64(vmid)
		adapter["name"] = vmConfigData["name"]
		*list = append(*list, adapter)
	}
}

func dump_macaddr() {
	var machineList []map[string]interface{}
	for _, machine := range etc.GlobalPxCluster.GetMachines() {
		processMachine(machine, &machineList)
	}

	headers := []string{"macaddr", "net", "node", "vmid", "name"}
	shared.RenderOnConsoleNew(machineList, headers, nil)

}
