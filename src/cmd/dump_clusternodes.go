/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"px/etc"
	"px/queries"
	"px/shared"

	"github.com/spf13/cobra"
)

var dumpClusterNodesCmd = &cobra.Command{
	Use:   "clusternodes",
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
		queries.DumpClusterNodes(etc.GlobalPxCluster.GetPxClients())
	},
}

func init() {
	dumpCmd.AddCommand(dumpClusterNodesCmd)
}
