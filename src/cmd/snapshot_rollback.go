/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"px/configmap"
	"px/proxmox"
	"px/shared"
)

type SnapshotRollbackOptions struct {
	Match string
}

var snapshotRollbackOptions = &SnapshotRollbackOptions{}

// snapshotRollbackCmd represents the rollback command
var snapshotRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("rollback called")
		checkErr(snapshotRollbackOptions.Validate(args))
		checkErr(snapshotRollbackOptions.Run(args))
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotRollbackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// snapshotRollbackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// snapshotRollbackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	snapshotRollbackCmd.Flags().StringVar(&snapshotRollbackOptions.Match, "match", "", "match")
}
func (o *SnapshotRollbackOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New(fmt.Sprintf("please specifiy snapshot name\n"))
	}
	return nil
}

func (o *SnapshotRollbackOptions) Run(args []string) error {
	snapshotName := args[0]

	machines := GetSnapshotsAll()

	filteredMachines := shared.FilterStringColumns(machines, []string{"name", "snapshot"}, []string{o.Match, snapshotName})
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
			machine, _ := shared.GlobalPxCluster.UniqueMachines[vmid]

				status, _ := configmap.GetString(machine, "status")

					if status != "stopped" {
						fmt.Fprintf(os.Stderr, "ignoring snapshot '%v' on node '%v' for %v: machine is running!\n", snapshotName, node, name)
						continue
					}

		*/
		fmt.Fprintf(os.Stderr, "rollback snapshot '%v' on node '%v' for %v\n", snapshotName, node, name)
		if _type == proxmox.PROXMOX_MACHINE_CT {
			shared.RollbackContainerSnapshot(node, vmidInt64, snapshotName)
			shared.WaitForCTUnlock(node, vmidInt64)

		} else {
			shared.RollbackVMSnapshot(node, vmidInt64, snapshotName)
			shared.WaitForVMUnlock(node, vmidInt64)
		}
	}
	return nil
}
