/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"px/api"
	"px/configmap"
	"px/etc"
	"px/shared"

	"github.com/spf13/cobra"
)

type SnapshotCreateOptions struct {
	Match string
	Node  string
	Vmid  int
}

var snapshotCreateOptions = &SnapshotCreateOptions{}

// createCmd represents the create command
var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Fprintf(os.Stderr, "snapshot create called: %v\n", args)
		checkErr(snapshotCreateOptions.Validate(args))
		checkErr(snapshotCreateOptions.Run(args))
	},
}

func (o *SnapshotCreateOptions) Validate(args []string) error {
	if o.Node != "" && !etc.GlobalPxCluster.HasNode(o.Node) {
		return fmt.Errorf("node does not exist: %v", o.Node)
	}
	if o.Vmid != 0 && !InProxmoxVmidRange(o.Vmid) {
		return fmt.Errorf("vmid not in range: %v", o.Vmid)
	}
	/* FIXME now in the context of node/vmid
	if o.Vmid != 0 && etc.GlobalPxCluster.UniqueMachines[o.Vmid] == nil {
		return errors.New(fmt.Sprintf("vmid does not exist: %v\n", o.Vmid))
	}
	*/
	if len(args) == 0 {
		return fmt.Errorf("please specifiy snapshot name")
	}
	if len(args[0]) < 2 {
		return fmt.Errorf("snapshotname must at least have 2 characters")
	}
	return nil
}
func (o *SnapshotCreateOptions) Run(args []string) error {
	snapshotName := args[0]
	DoSnapshotCreate(snapshotName, o.Match)
	return nil
}
func init() {
	snapshotCmd.AddCommand(snapshotCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	snapshotCreateCmd.Flags().StringVar(&snapshotCreateOptions.Match, "match", "", "match")
	snapshotCreateCmd.Flags().StringVar(&snapshotCreateOptions.Node, "node", "", "node")
	snapshotCreateCmd.Flags().IntVar(&snapshotCreateOptions.Vmid, "vmid", 0, "vmid")
}

func DoSnapshotCreate(snapshotName string, match string) {
	if match == "" {
		return
	}
	machines := etc.GlobalPxCluster.GetMachines()

	// Handle match as it is special
	filteredMachines := shared.SelectMachines(machines, match)

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
		fmt.Fprintf(os.Stderr, "create snapshot '%v' on node '%v' for %v\n", snapshotName, node, name)
		if _type == "lxc" {
			api.CreateContainerSnapshot(node, vmidInt64, snapshotName)
		} else {
			api.CreateVMSnapshot(node, vmidInt64, snapshotName)
		}
	}

}
