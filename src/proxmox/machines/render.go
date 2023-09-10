package machines

import (
	"embed"
	"fmt"
	"os"
	"px/configmap"
)

func LoadMachines(data map[string]interface{}) map[string]interface{} {
	filepath := "files/settings/machines.yaml"
	return LoadEmbeddedYamlFile(data, files, filepath)
}
func LoadEmbeddedYamlFile(data map[string]interface{}, files embed.FS, filepath string) map[string]interface{} {
	machinesData := map[string]interface{}{}
	err := configmap.LoadEmbeddedYamlFile(machinesData, files, filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadEmbeddedYamlFile() error: %v\n", err)
	}
	machinesData = configmap.MergeMapRecursive(machinesData, data)
	return machinesData
}
