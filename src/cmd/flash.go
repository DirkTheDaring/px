/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

type FlashOptions struct {
	Match  string
	File   []string
	DryRun bool
	Dump   bool
}

var flashOptions = &FlashOptions{}

// flashCmd represents the flash command
var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("flash called")
		checkErr(flashOptions.Validate())
		checkErr(flashOptions.Run())
	},
}

func init() {
	rootCmd.AddCommand(flashCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flashCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// flashCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flashCmd.Flags().StringVar(&flashOptions.Match, "match", "", "match")
	flashCmd.Flags().StringSliceVarP(&flashOptions.File, "from-file", "f", createOptions.File, "create from file")
	flashCmd.Flags().BoolVar(&flashOptions.DryRun, "dry-run", false, "dry-run")
	flashCmd.Flags().BoolVar(&flashOptions.Dump, "dump", false, "dump")
}

func (o *FlashOptions) Validate() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	return nil
}

func (o *FlashOptions) Run() error {
	//fmt.Fprintf(os.Stderr, "o: %v\n", o)
	ProcessFiles(flashOptions.File, "flash")
	return nil
}
