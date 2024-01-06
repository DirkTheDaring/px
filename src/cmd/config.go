/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"px/configmap"
	"px/etc"
	"px/queries"
	"px/shared"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
)

type ConfigOptions struct {
	Match string
}

var configOptions = &ConfigOptions{}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
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
		DoConfig(configOptions.Match)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	configCmd.Flags().StringVar(&configOptions.Match, "match", "", "match")
}

func DoConfig(match string) {
	// Retrieve the list of machines from the global cluster.
	machines := etc.GlobalPxCluster.GetMachines()

	// Handle match as it is special
	filteredMachines := shared.SelectMachines(machines, match)

	var lines []map[string]interface{}
	var disks map[string]bool = make(map[string]bool)

	regex, _ := regexp.Compile("^(virtio|scsi)[0-9]+$")

	for _, machine := range filteredMachines {
		node, okNode := configmap.GetString(machine, "node")
		vmid, okVMID := configmap.GetInt(machine, "vmid")
		machineType, okType := configmap.GetString(machine, "type")
		//name, okName := configmap.GetString(machine, "name")

		// Skip this iteration if any required details are missing.
		if !okNode || !okVMID || !okType {
			continue
		}

		machineConfig, _ := getMachineConfig(node, int64(vmid), machineType)

		result, _ := configmap.GetMapEntry(machineConfig, "data")

		result["node"] = node
		result["vmid"] = vmid
		result["type"] = machineType

		if machineType == "lxc" {
			name, ok := result["hostname"]
			if ok {
				result["name"] = name
			}
		}
		for key, _ := range result {

			if regex.MatchString(key) {
				disks[key] = true
				diskStr, _ := configmap.GetString(result, key)
				//valueStr, _ := configmap.GetString(diskStr, "size")
				valueStr, _ := GetConfigOptionAndRemaining(diskStr, "size")

				if len(valueStr) > 5 {
					result[key] = valueStr[5:]

				}
				/*
					else {
						fmt.Fprintf(os.Stderr, "valueStr: %v\n", valueStr)
					}
				*/

			}
		}
		lines = append(lines, result)

	}

	disk_headers := []string{}
	for key, _ := range disks {
		disk_headers = append(disk_headers, key)
	}
	sort.Strings(disk_headers)

	headers := []string{"node", "vmid", "type", "name", "memory", "cores", "cpu"}
	headers = append(headers, disk_headers...)

	shared.RenderOnConsoleNew(lines, headers, []string{"memory", "virtio0", "virtio1", "cores"})

}
func getMachineConfig(node string, vmid int64, machineType string) (map[string]interface{}, error) {
	if machineType == "qemu" {
		return queries.JSONGetVMConfig(node, vmid)
	}
	if machineType == "lxc" {
		return queries.JSONGetCTConfig(node, vmid)
	}
	return nil, fmt.Errorf("machineType not supported: %v", machineType)
}
