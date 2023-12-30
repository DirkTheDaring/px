package shared

import (
	"os"
	"px/etc"
)

func Status(match string) {
	{
		machines := etc.GlobalPxCluster.GetMachines()
		sortedMachines := StringSortMachines(machines, []string{"name"}, []bool{true})
		/*
			for _, m := range sortedMachines {
				fmt.Fprintf(os.Stderr, "%s\n", m["name"])
			}
		*/

		headers := []string{"name", "type", "node", "vmid", "status"}
		//alignments_right := []string{"type", "node"}
		//alignments_right := []string{"vmid", "node"}
		alignments_right := []string{"vmid"}

		RenderOnConsoleNew(sortedMachines, headers, "name", match, alignments_right)
		os.Exit(0)
	}
}
