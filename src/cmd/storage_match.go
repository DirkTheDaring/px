/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"px/configmap"
	"px/etc"
	"px/queries"
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
		shared.InitConfig(ClusterName)
		pxClients, _ := queries.GetStorageContentAll(etc.GlobalPxCluster.GetPxClients())
		etc.GlobalPxCluster.SetPxClients(pxClients)

	},
	Run: func(cmd *cobra.Command, args []string) {
		DoStorageMatch(ClusterName)
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
func DoStorageMatch(clusterName string) {

	clusterDatabase, _ := etc.GlobalPxCluster.PickCluster(clusterName)
	cluster := clusterDatabase.GetCluster()
	selectors, _ := configmap.GetMapEntry(cluster, "selectors")
	newStorageContent := shared.JoinClusterAndSelector(*etc.GlobalPxCluster, selectors)
	headers := []string{"label", "volid", "node"}
	shared.RenderOnConsoleNew(newStorageContent, headers, nil)

}
