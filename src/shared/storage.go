package shared

import (
	"os"
	"px/etc"
)

func StorageList() {
	types := []string{}
	storage := etc.GlobalPxCluster.GetStorage(types)

	headers := []string{"storage", "type", "path", "node"}
	storage = StringSortMachines(storage, []string{"storage"}, []bool{true})
	RenderOnConsole(storage, headers, "", "")
	os.Exit(0)
}
