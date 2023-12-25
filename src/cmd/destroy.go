/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"px/configmap"
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
		fmt.Println("destroy called")
		machines := shared.GlobalPxCluster.Machines
		if destroyOptions.Match != "" {
			filteredMachines := shared.FilterStringColumns(machines, []string{"name", "status"}, []string{destroyOptions.Match, "stopped"})
			for _, filteredMachine := range filteredMachines {
				//fmt.Fprintf(os.Stderr, "%v %v\n", filteredMachine["vmid"], filteredMachine["status"])

				node, ok := configmap.GetString(filteredMachine, "node")
				if !ok {
					continue
				}
				vmid, ok := configmap.GetInt(filteredMachine, "vmid")
				if !ok {
					continue
				}
				_type, ok := configmap.GetString(filteredMachine, "type")
				if !ok {
					continue
				}
				name, ok := configmap.GetString(filteredMachine, "name")
				if !ok {
					continue
				}
				vmidInt64 := int64(vmid)

				fmt.Fprintf(os.Stderr, "destroy: %v %v %v %v\n", node, vmidInt64, name, _type)
				if _type == "lxc" {
					shared.DeleteContainer(node, vmidInt64)
				} else {
					shared.DeleteVM(node, vmidInt64)
				}
			}
		}
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
