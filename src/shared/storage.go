package shared

import "os"

func StorageList() {
	types := []string{}
	storage := GlobalPxCluster.GetStorage(types)

	headers := []string{"storage", "type", "path", "node"}
	storage = StringSortMachines(storage, []string{"storage"}, []bool{true})
	RenderOnConsole(storage, headers, "", "")
	os.Exit(0)
}
