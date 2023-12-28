/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"px/etc"
	"px/queries"
	"px/shared"

	"github.com/spf13/cobra"
)

// storageContentCmd represents the list command
var storageContentCmd = &cobra.Command{
	Use:   "content",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		pxClients, _ := queries.GetStorageContentAll(etc.GlobalPxCluster.PxClients)
		etc.GlobalPxCluster.PxClients = pxClients

	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("storage content called")
		storageContent := etc.GlobalPxCluster.GetStorageContent()
		headers := []string{"storage", "node", "type", "content", "volid"}
		shared.RenderOnConsole(storageContent, headers, "", "")

	},
}

func init() {
	storageCmd.AddCommand(storageContentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storageContentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storageContentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
