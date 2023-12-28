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

type StartOptions struct {
	Match string
}

var startOptions = &StartOptions{}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
		machines := etc.GlobalPxCluster.Machines

		if startOptions.Match != "" {
			filteredMachines := shared.FilterStringColumns(machines, []string{"name", "status"}, []string{startOptions.Match, "stopped"})
			for _, filteredMachine := range filteredMachines {
				//fmt.Fprintf(os.Stderr, "%v %v\n", filteredMachine["vmid"], filteredMachine["status"])
				node := filteredMachine["node"].(string)
				vmid := filteredMachine["vmid"].(float64)
				name := filteredMachine["name"].(string)
				_type := filteredMachine["type"].(string)
				vmidInt64 := int64(vmid)
				fmt.Fprintf(os.Stderr, "start: %v %v %v %v\n", node, vmidInt64, name, _type)
				if _type == "lxc" {
					api.StartContainer(node, vmidInt64)
				} else {
					api.StartVM(node, vmidInt64)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	startCmd.Flags().StringVar(&startOptions.Match, "match", "", "match")

}
