/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/shared"

	"github.com/spf13/cobra"
)

type GetOptions struct {
	FileSources []string
}

var getOptions = &GetOptions{}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly get a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called", args)
		getOptions.Run()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//getCmd.Flags().Bool("dump", false, "dump")
	//getCmd.Flags().BoolP("dump", "", false, "dump")
	//getCmd.Flags().StringSliceVarP(&getOptions.FileSources, "from-file", "f", getOptions.FileSources, "get from file")

}

// Complete loads data from the command line environment

func (o *GetOptions) Complete() error {
	fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *GetOptions) Validate() error {
	fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *GetOptions) Run() error {
	fmt.Fprintf(os.Stderr, "o: %v\n", o)
	//proxmox.Test()
	return nil
}
