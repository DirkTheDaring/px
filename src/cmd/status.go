/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"px/etc"
	"px/shared"

	"github.com/spf13/cobra"
)

type StatusOptions struct {
	Match string
}

var statusOptions = &StatusOptions{}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		DoStatus(ClusterName, statusOptions.Match)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	statusCmd.Flags().StringVar(&statusOptions.Match, "match", "", "match")
	//statusCmd.Flags().StringVar
}

func DoStatus(clusterName string, match string) {
	//fmt.Println("status called")
	//fmt.Fprintf(os.Stderr, "BUG?: >%v<\n", statusOptions.Match)
	shared.Status(etc.GlobalPxCluster, match)
}
