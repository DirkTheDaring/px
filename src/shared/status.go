package shared

import (
	"os"
)

func Status(match string) {
	{
		machines := GlobalPxCluster.Machines
		headers := []string{"name", "type", "node", "vmid", "status"}

		RenderOnConsole(machines, headers, "name", match)
		os.Exit(0)
	}
}
