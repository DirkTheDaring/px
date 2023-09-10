/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"px/configmap"
	"px/shared"

	"github.com/spf13/cobra"
)

// storageMatchCmd represents the list command
var storageMatchCmd = &cobra.Command{
	Use:   "match",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		pxClients := shared.GetStorageContentAll(shared.GlobalPxCluster.PxClients)
		shared.GlobalPxCluster.PxClients = pxClients

	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("storage match called")
		cluster, _ := shared.PickCluster(shared.GlobalConfigData, ClusterName)

		selectors, _ := configmap.GetMapEntry(cluster, "selectors")
		newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
		headers := []string{"label", "volid", "node"}
		shared.RenderOnConsole(newStorageContent, headers, "", "")
	},
}

func init() {
	storageCmd.AddCommand(storageMatchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storageMatchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storageMatchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
