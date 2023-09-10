package proxmox

import (
	"fmt"
	"os"
	"px/configmap"
	"px/proxmox/cattles"
	"px/proxmox/machine"
	"px/proxmox/machines"
)

func DumpJson(data map[string]interface{}) {
	json := configmap.DataToJSON(data)
	fmt.Fprintf(os.Stdout, "%s\n", json)
}

func Test2() {

}

func Test() {

	// Phase 1 - Cattle Configuration
	// Phase 1 load  data which can override cattle settings
	filename := "test/test.yaml"
	testData := map[string]interface{}{}
	configmap.LoadEmbeddedYamlFile(testData, test, filename)

	/*
		data := cattles.CreateCattle("ct", "small", testData)
		json := configmap.DataToJSON(data)
		fmt.Fprintf(os.Stderr, "Test() json: %s\n", json)
	*/

	// Phase 2 - Cattle Customization
	// now merge cattle with configuration of user
	filename = "test/virtualmachine.yaml"
	//configData := map[string]interface{}{}
	vmData := map[string]interface{}{}
	configmap.LoadEmbeddedYamlFile(vmData, test, filename)

	node := "mn35"
	data := cattles.Configure(vmData, node, testData)
	DumpJson(data)

	machineTestData := map[string]interface{}{}
	result := machine.LoadMachine(machineTestData)
	DumpJson(result)

	machinesTestData := map[string]interface{}{}
	result2 := machines.LoadMachines(machinesTestData)
	DumpJson(result2)

	// FIXME now add args line for ignition if needed
	// then create virtual machine
	// collect data from created machine (MACADRESS)
	// render ignition file and upload

}
