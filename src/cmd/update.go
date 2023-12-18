/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
//	"fmt"
//	"os"

	"github.com/spf13/cobra"
	"px/shared"
)

type UpdateOptions struct {
	Match string
	Set []string
}

var updateOptions = &UpdateOptions{}

// updateCmd represents the set command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

px update --set onboot=0 --match emea-prod`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("get called", args)
		updateOptions.Run()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//updateCmd.Flags().Bool("dump", false, "dump")
	//updateCmd.Flags().BoolP("dump", "", false, "dump")
	//updateCmd.Flags().StringSliceVarP(&updateOptions.FileSources, "from-file", "f", updateOptions.FileSources, "set")
	updateCmd.Flags().StringVar(&updateOptions.Match, "match", "", "match")
	updateCmd.Flags().StringSliceVarP(&updateOptions.Set, "set", "s", updateOptions.Set, "set option")
}

// Complete loads data from the command line environment

func (o *UpdateOptions) Complete() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *UpdateOptions) Validate() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *UpdateOptions) Run() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	shared.Update(updateOptions.Match, updateOptions.Set)
	return nil
}
