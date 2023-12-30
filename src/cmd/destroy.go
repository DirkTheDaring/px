/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/api"
	"px/configmap"
	"px/etc"
	"px/shared"

	"github.com/spf13/cobra"
)

type DestroyOptions struct {
	Match string
}

var destroyOptions = &DestroyOptions{}

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		DoDestroy(destroyOptions.Match)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	destroyCmd.Flags().StringVar(&destroyOptions.Match, "match", "", "match")

}

// DoDestroy removes machines that match the specified criteria.
// It prints the details of the machines being destroyed.
func DoDestroy(match string) {
	// Early exit if match is empty, as no machines would be matched.
	if match == "" {
		return
	}

	// Retrieve the list of machines from the global cluster.
	machines := etc.GlobalPxCluster.GetMachines()

	// Filter out machines that match the criteria and are not stopped.
	// This also reports machines that are excluded because they are stopped.
	filteredMachines := shared.FilterStringColumns(machines, []string{"name", "status"}, []string{match, "stopped"})

	// Iterate over the filtered machines to destroy them.
	for _, machine := range filteredMachines {
		// Extract necessary details from the machine's config.
		node, okNode := configmap.GetString(machine, "node")
		vmid, okVMID := configmap.GetInt(machine, "vmid")
		machineType, okType := configmap.GetString(machine, "type")
		name, okName := configmap.GetString(machine, "name")

		// Skip this iteration if any required details are missing.
		if !okNode || !okVMID || !okType || !okName {
			continue
		}

		// Log the details of the machine being destroyed.
		fmt.Fprintf(os.Stderr, "destroy: %v %v %v %v\n", node, int64(vmid), name, machineType)

		// Destroy the machine based on its type.
		if machineType == "lxc" {
			api.DeleteContainer(node, int64(vmid))
		} else {
			api.DeleteVM(node, int64(vmid))
		}
	}
}
