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

// storageLatestCmd represents the list command
var storageLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)
		queries.GetStorageContentAll(etc.GlobalPxCluster.GetPxClients())
		//etc.GlobalPxCluster.SetPxClients(pxClients)
	},
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("storage latest called")

		//cluster, _ := shared.PickCluster(etc.GlobalConfigData, ClusterName)
		clusterDatabase, _ := etc.GlobalPxCluster.PickCluster(ClusterName)
		cluster := clusterDatabase.GetCluster()

		selectors, _ := configmap.GetMapEntry(cluster, "selectors")

		newStorageContent := shared.JoinClusterAndSelector(*etc.GlobalPxCluster, selectors)
		//fmt.Fprintf(os.Stderr, "Storage latest called: %v\n", newStorageContent)
		//jsonBinary, _ := json.Marshal(newStorageContent)
		//fmt.Fprintf(os.Stdout, "%v\n", string(jsonBinary))
		newStorageContent = shared.ExtractLatest(*etc.GlobalPxCluster, newStorageContent)
		headers := []string{"label", "volid", "node"}

		//fmt.Fprintf(os.Stderr, "Storage latest called: %v\n", newStorageContent)

		shared.RenderOnConsoleNew(newStorageContent, headers, nil)
	},
}

func init() {
	storageCmd.AddCommand(storageLatestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storageLatestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storageLatestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
