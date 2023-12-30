/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"px/api"
	"px/configmap"
	"px/proxmox"
	"px/queries"
	"px/shared"

	"github.com/spf13/cobra"
)

type SnapshotDeleteOptions struct {
	Match string
}

var snapshotDeleteOptions = &SnapshotDeleteOptions{}

// deleteCmd represents the delete command
var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Fprintf(os.Stderr, "snapshot delete called: %v\n", args)
		checkErr(snapshotDeleteOptions.Validate(args))
		checkErr(snapshotDeleteOptions.Run(args))
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	snapshotDeleteCmd.Flags().StringVar(&snapshotDeleteOptions.Match, "match", "", "match")
	//snapshotDeleteCmd.Flags().StringVar(&snapshotDeleteOptions.Node, "node", "", "node")
	//snapshotDeleteCmd.Flags().IntVar(&snapshotDeleteOptions.Vmid, "vmid", 0, "vmid")
}

func (o *SnapshotDeleteOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New(fmt.Sprintf("please specifiy snapshot name\n"))
	}
	return nil
}

func (o *SnapshotDeleteOptions) Run(args []string) error {
	snapshotName := args[0]

	DoSnapshotDelete(snapshotName, o.Match)

	return nil
}

func DoSnapshotDelete(snapshotName string, match string) {
	machines := GetSnapshotsAll()

	filteredMachines := shared.FilterStringColumns(machines, []string{"name", "snapshot"}, []string{match, snapshotName})
	for _, filteredMachine := range filteredMachines {
		node, ok := configmap.GetString(filteredMachine, "node")
		if !ok {
			continue
		}
		vmid, ok := configmap.GetInt(filteredMachine, "vmid")
		if !ok {
			continue
		}
		_type, ok := configmap.GetString(filteredMachine, "type")
		if !ok {
			continue
		}
		name, ok := configmap.GetString(filteredMachine, "name")
		if !ok {
			continue
		}
		vmidInt64 := int64(vmid)
		//name := filteredMachine["name"].(string)

		/*
			machine, _ := etc.GlobalPxCluster.UniqueMachines[vmid]
			status, _ := configmap.GetString(machine, "status")

				if status != "stopped" {
					fmt.Fprintf(os.Stderr, "ignoring snapshot '%v' on node '%v' for %v: machine is running!\n", snapshotName, node, name)
					continue
				}
		*/
		fmt.Fprintf(os.Stderr, "delete snapshot '%v' on node '%v' for %v\n", snapshotName, node, name)
		if _type == proxmox.PROXMOX_MACHINE_CT {
			api.DeleteContainerSnapshot(node, vmidInt64, snapshotName)
			queries.WaitForContainerUnlock(node, vmidInt64)
		} else {
			api.DeleteVMSnapshot(node, vmidInt64, snapshotName)
			queries.WaitForVMUnlock(node, vmidInt64)
		}
	}
}
