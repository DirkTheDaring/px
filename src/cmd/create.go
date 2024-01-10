/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/documents"
	"px/etc"
	"px/queries"
	"px/shared"

	"github.com/spf13/cobra"
)

type CreateOptions struct {
	File   []string
	DryRun bool
	Dump   bool
	Update bool
}

var createOptions = &CreateOptions{}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)
		//fmt.Println("--- Pre run for create virtualmachine")
		pxClients, _ := queries.GetStorageContentAll(etc.GlobalPxCluster.GetPxClients())
		//shared.GlobalPxCluster = shared.ProcessCluster(pxClients)
		etc.GlobalPxCluster.SetPxClients(pxClients)

		//fmt.Println("--- Pre run end")

	},
	Run: func(cmd *cobra.Command, args []string) {
		checkErr(createOptions.Validate())
		checkErr(createOptions.Run())
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	createCmd.Flags().StringSliceVarP(&createOptions.File, "from-file", "f", createOptions.File, "create from file")
	createCmd.Flags().BoolVar(&createOptions.DryRun, "dry-run", false, "dry-run")
	createCmd.Flags().BoolVar(&createOptions.Dump, "dump", false, "dump")
	createCmd.Flags().BoolVarP(&createOptions.Update, "update", "u", false, "update settings (cores, memory, disksize)")
}

// Complete loads data from the command line environment

func (o *CreateOptions) Complete() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *CreateOptions) Validate() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *CreateOptions) Run() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	//proxmox.Test()
	///ProcessFiles(createOptions.File, "create")
	err := documents.ProcessFiles("create", ProcessSection, createOptions.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	return nil
}
