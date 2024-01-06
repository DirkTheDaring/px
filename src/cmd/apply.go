/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"px/shared"

	"github.com/spf13/cobra"
)

type ApplyOptions struct {
	Match string
	Set   []string
}

var applyOptions = &ApplyOptions{}

// applyCmd represents the set command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Update the configuration of virtual machines and containers",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

px apply --set onboot=0 --match emea-prod`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("get called", args)
		applyOptions.Run()
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//applyCmd.Flags().Bool("dump", false, "dump")
	//applyCmd.Flags().BoolP("dump", "", false, "dump")
	//applyCmd.Flags().StringSliceVarP(&applyOptions.FileSources, "from-file", "f", applyOptions.FileSources, "set")
	applyCmd.Flags().StringVar(&applyOptions.Match, "match", "", "match")
	applyCmd.Flags().StringSliceVarP(&applyOptions.Set, "set", "s", applyOptions.Set, "set option")
}

// Complete loads data from the command line environment

func (o *ApplyOptions) Complete() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *ApplyOptions) Validate() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *ApplyOptions) Run() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	shared.Apply(applyOptions.Match, applyOptions.Set)
	return nil
}
