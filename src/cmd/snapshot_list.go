/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"px/configmap"
	"px/proxmox"
	"px/shared"

	"github.com/spf13/cobra"
)

type SnapshotListOptions struct {
	Match string
}

var snapshotListOptions = &SnapshotListOptions{}

func GetSnapshotsAll() []map[string]interface{} {
	machines := shared.GlobalPxCluster.Machines

	//fmt.Fprintf(os.Stderr, "CALL\n")
	list := []map[string]interface{}{}
	for _, machine := range machines {
		node, ok := configmap.GetString(machine, "node")
		if !ok {
			continue
		}
		vmid, ok := configmap.GetInt(machine, "vmid")
		if !ok {
			continue
		}
		_type, ok := configmap.GetString(machine, "type")
		if !ok {
			continue
		}
		vmidInt64 := int64(vmid)
		if _type == proxmox.PROXMOX_MACHINE_CT {
			snapshots, err := shared.GetContainerSnapshots(node, vmidInt64)
			if err != nil {
				continue
			}
			dataList := snapshots.GetData()
			for _, data := range dataList {
				if *data.Name == "current" {
					continue
				}
				txt, _ := json.Marshal(data)
				//fmt.Fprintf(os.Stderr, "%v\n", string(txt))
				item := map[string]interface{}{}
				json.Unmarshal(txt, &item)
				item["type"] = _type
				item["node"] = node
				item["snapshot"] = item["name"]
				item["vmid"] = machine["vmid"]
				item["name"] = machine["name"]
				list = append(list, item)
			}
		} else {
			snapshots, err := shared.GetVMSnapshots(node, vmidInt64)
			if err != nil {
				continue
			}
			dataList := snapshots.GetData()
			for _, data := range dataList {
				if *data.Name == "current" {
					continue
				}
				txt, _ := json.Marshal(data)
				item := map[string]interface{}{}
				json.Unmarshal(txt, &item)
				item["type"] = _type
				item["node"] = node
				item["snapshot"] = item["name"]
				item["vmid"] = machine["vmid"]
				item["name"] = machine["name"]
				list = append(list, item)
			}
		}
	}
	return list
}

// createCmd represents the create command
var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("snapshot list called")

		snapshots := GetSnapshotsAll()
		headers := []string{"snapshot", "type", "parent", "node", "vmid", "name", "snaptime"}
		snapshots = shared.StringSortMachines(snapshots, []string{"snapshot", "node", "name"}, []bool{true, true, true})

		shared.RenderOnConsole(snapshots, headers, "name", snapshotListOptions.Match)
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	snapshotListCmd.Flags().StringVar(&snapshotListOptions.Match, "match", "", "match")
}
