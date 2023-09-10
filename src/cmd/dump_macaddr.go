/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"px/configmap"
	"px/shared"
	"strings"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stderr, "dump macaddr\n")
		list := []map[string]interface{}{}
		for _, machine := range shared.GlobalPxCluster.Machines {
			_type, _ := configmap.GetString(machine, "type")
			if _type == "lxc" {
				continue
			}
			vmid, _ := configmap.GetInt(machine, "vmid")
			node, _ := configmap.GetString(machine, "node")

			vmConfig, err := shared.JSONGetVMConfig(node, int64(vmid))
			if err != nil {
				return
			}
			vmConfigData, ok := configmap.GetMapEntry(vmConfig, "data")
			if !ok {
				return
			}
			networkAdapters := configmap.SelectKeys("^net[0-9]+$", vmConfigData)
			for _, networkAdapter := range networkAdapters {
				net, _ := configmap.GetString(vmConfigData, networkAdapter)
				virtioStr, _ := GetConfigOptionAndRemaining(net, "virtio")
				array := strings.Split(virtioStr, "=")
				if len(array) == 2 {
					item := map[string]interface{}{}
					item["node"] = node
					item["vmid"] = machine["vmid"]
					item["name"] = vmConfigData["name"]
					item["net"] = networkAdapter
					item["macaddr"] = array[1]
					list = append(list, item)
				}
			}

		}
		headers := []string{"macaddr", "net", "node", "vmid", "name"}
		shared.RenderOnConsole(list, headers, "name", "")
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
