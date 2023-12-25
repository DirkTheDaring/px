/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"px/shared"

	"github.com/spf13/cobra"
)

type ShutdownOptions struct {
	Match string
}

var shutdownOptions = &ShutdownOptions{}

// startCmd represents the start command
var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown a virtual machine or container gracefully",
	Long: `The shutdown command is used to gracefully shut down a virtual machine (VM) or container (LXC)
running on Proxmox Virtual Environment (PVE). This ensures that the VM is powered off safely, 
allowing for the proper termination of processes and file system operations.`,
	Run: func(cmd *cobra.Command, args []string) {

		machines := shared.GlobalPxCluster.Machines

		if shutdownOptions.Match != "" {
			filteredMachines := shared.FilterStringColumns(machines, []string{"name", "status"}, []string{shutdownOptions.Match, "running"})
			for _, filteredMachine := range filteredMachines {
				//fmt.Fprintf(os.Stderr, "%v %v\n", filteredMachine["vmid"], filteredMachine["status"])
				node := filteredMachine["node"].(string)
				vmid := filteredMachine["vmid"].(float64)
				name := filteredMachine["name"].(string)
				_type := filteredMachine["type"].(string)
				vmidInt64 := int64(vmid)
				fmt.Fprintf(os.Stderr, "shutdown: %v %v %v %v\n", node, vmidInt64, name, _type)
				if _type == "lxc" {
					shared.ShutdownContainer(node, vmidInt64)
				} else {
					shared.ShutdownVM(node, vmidInt64)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(shutdownCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	shutdownCmd.Flags().StringVar(&shutdownOptions.Match, "match", "", "match")
}
