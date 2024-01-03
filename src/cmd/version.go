/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed VERSION.txt
var versionFS embed.FS

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		DoVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// DoVersion reads the version number from a file and prints it.
func DoVersion() error {
	// Read the version file
	version, err := fs.ReadFile(versionFS, "VERSION.txt")
	if err != nil {
		// Return the error instead of panicking
		return err
	}

	// Split the version string and get the first line
	versionArray := strings.Split(string(version), "\n")
	firstLine := versionArray[0]

	// Print the version number
	fmt.Println("px version:", firstLine)

	return nil
}
