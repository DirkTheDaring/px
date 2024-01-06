/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"px/api"
	"px/configmap"
	"px/etc"
	"px/proxmox"
	"px/shared"
	"time"

	"github.com/spf13/cobra"
)

type SnapshotListOptions struct {
	Match string
}

var snapshotListOptions = &SnapshotListOptions{}

// createCmd represents the create command
var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		shared.InitConfig(ClusterName)

	},
	Run: func(cmd *cobra.Command, args []string) {
		DoSnapshotList(snapshotListOptions.Match)
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
func DoSnapshotList(match string) {
	//fmt.Println("snapshot list called")

	machines := etc.GlobalPxCluster.GetMachines()
	// Handle match as it is special
	machines = shared.SelectMachines(machines, match)

	snapshots := GetSnapshotsAll(machines)

	headers := []string{"snapshot", "type", "parent", "node", "vmid", "name", "snaptime"}
	snapshots = shared.StringSortMachines(snapshots, []string{"snapshot", "node", "name"}, []bool{true, true, true})

	shared.RenderOnConsoleWithFilter(snapshots, headers, "name", snapshotListOptions.Match)

}

func ConvertEpochToDateTime(epoch int64) string {
	t := time.Unix(epoch, 0) // The second parameter is nanoseconds
	humanReadable := t.Format("2006-01-02 15:04:05")
	return humanReadable

}

func GetSnapshotsAll(machines []map[string]interface{}) []map[string]interface{} {

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
			snapshots, err := api.GetContainerSnapshots(node, vmidInt64)
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
				item["snaptime"] = ConvertEpochToDateTime(int64(item["snaptime"].(float64)))
				item["type"] = _type
				item["node"] = node
				item["snapshot"] = item["name"]
				item["vmid"] = machine["vmid"]
				item["name"] = machine["name"]

				list = append(list, item)
			}
		} else {
			snapshots, err := api.GetVMSnapshots(node, vmidInt64)
			//{"data":[{"description":"","name":"k8sprod1","snaptime":1703015930,"vmstate":0},{"description":"You are here!","name":"current","parent":"k8sprod1"}]}
			//jsonData, err := json.Marshal(snapshots)
			//fmt.Println(string(jsonData))

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

				/// convert snaptime to a human readable date
				item["snaptime"] = ConvertEpochToDateTime(int64(item["snaptime"].(float64)))

				item["type"] = _type
				item["node"] = node
				item["snapshot"] = item["name"]
				item["vmid"] = machine["vmid"]
				item["name"] = machine["name"]

				//jsonData, _ := json.Marshal(machine)
				//fmt.Println(string(jsonData))
				list = append(list, item)
			}
		}
	}
	return list
}
