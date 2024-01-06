/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/api"
	"px/etc"
	"px/shared"

	"github.com/spf13/cobra"
)

type StopOptions struct {
	Match string
}

var stopOptions = &StopOptions{}

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		DoStep(stopOptions.Match)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	stopCmd.Flags().StringVar(&stopOptions.Match, "match", "", "match")
}

func DoStep(match string) {
	if match == "" {
		return
	}
	machines := etc.GlobalPxCluster.GetMachines()

	// Handle match as it is special
	machines = shared.SelectMachines(machines, match)
	//filteredMachines := shared.FilterStringColumns(machines, []string{"name", "status"}, []string{match, "running"})
	filteredMachines := shared.FilterStringColumns(machines, []string{"status"}, []string{"running"})
	for _, filteredMachine := range filteredMachines {
		//fmt.Fprintf(os.Stderr, "%v %v\n", filteredMachine["vmid"], filteredMachine["status"])
		node := filteredMachine["node"].(string)
		vmid := filteredMachine["vmid"].(float64)
		name := filteredMachine["name"].(string)
		_type := filteredMachine["type"].(string)
		vmidInt64 := int64(vmid)
		fmt.Fprintf(os.Stderr, "stop: %v %v %v %v\n", node, vmidInt64, name, _type)
		if _type == "lxc" {
			api.StopContainer(node, vmidInt64)
		} else {
			api.StopVM(node, vmidInt64)
		}
	}

}
