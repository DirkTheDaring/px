package machine

import (
	"fmt"
	"os"
	"px/configmap"
)

func LoadMachine(data map[string]interface{}) map[string]interface{} {
	machineData := map[string]interface{}{}
	filepath := "files/settings/machine.yaml"
	err := configmap.LoadEmbeddedYamlFile(machineData, files, filepath)
	fmt.Fprintf(os.Stderr, "LoadMachine() error: %v\n", err)
	machineData = configmap.MergeMapRecursive(machineData, data)
	return machineData
}
