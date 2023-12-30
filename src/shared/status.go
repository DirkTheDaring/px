package shared

import (
	"os"
	"px/etc"
)

func Status(match string) {
	{
		machines := etc.GlobalPxCluster.Machines
		headers := []string{"name", "type", "node", "vmid", "status"}
		//alignments_right := []string{"type", "node"}
		//alignments_right := []string{"vmid", "node"}
		alignments_right := []string{"vmid"}

		RenderOnConsoleNew(machines, headers, "name", match, alignments_right)
		os.Exit(0)
	}
}
