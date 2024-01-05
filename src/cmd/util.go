package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"px/api"
	"px/configmap"
	"px/etc"
	"px/proxmox"
	"px/proxmox/cattles"
	"px/queries"
	"px/shared"
	"strings"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
)

func GetAttribute(configValue string, name string) (string, bool) {
	attributes := strings.Split(configValue, ",")
	name = name + "="
	for _, attribute := range attributes {
		if strings.HasPrefix(attribute, name) {
			return attribute[len(name):], true
		}
	}
	return "", false
}

func GetDeployedDriveSize(vmConfigData map[string]interface{}, storageDrive string) int64 {
	configValue, ok := configmap.GetString(vmConfigData, storageDrive)
	if !ok {
		return 0
	}
	configValue, ok = GetAttribute(configValue, "size")
	if !ok {
		return 0
	}
	currentSizeInBytes, ok := shared.CalculateSizeInBytes(configValue)
	if !ok {
		return 0
	}
	return currentSizeInBytes
}

func GetDriveSize(spec map[string]interface{}, storageDrive string) int64 {
	storageData, ok := configmap.GetMapEntry(spec, storageDrive)
	if !ok {
		return 0
	}
	size, ok := configmap.GetString(storageData, "size")
	//fmt.Fprintf(os.Stderr, "importFrom: %v %v\n", ok, importFrom)
	if !ok {
		return 0
	}
	sizeInBytes, ok := shared.CalculateSizeInBytes(size)
	//fmt.Fprintf(os.Stderr, "importFrom: %v %v\n", ok, importFrom)
	if !ok {
		return 0
	}
	return sizeInBytes
}

func CalcSizeDifference(vmConfigData map[string]interface{}, spec map[string]interface{}, storageDrive string) int64 {
	currentSize := GetDeployedDriveSize(vmConfigData, storageDrive)
	if currentSize == 0 {
		return 0
	}

	size := GetDriveSize(spec, storageDrive)
	if size == 0 {
		return 0
	}
	if size < currentSize {
		return 0
	}
	return size - currentSize
}

// GetVars extracts the "vars" map from the given object's "metadata" field.
// Returns an empty map if "metadata" or "vars" is not present.
func GetVars(object map[string]interface{}) map[string]interface{} {
	metadata, found := configmap.GetMapEntry(object, "metadata")
	if !found {
		return make(map[string]interface{})
	}

	vars, found := configmap.GetMapEntry(metadata, "vars")
	if !found {
		return make(map[string]interface{})
	}

	return vars
}

// ResolveList processes a list of items, merging each item's vars with the list's vars.
// Returns a new list with updated vars for each item.
func ResolveList(list map[string]interface{}) []map[string]interface{} {
	listVars := GetVars(list)
	items, _ := configmap.GetInterfaceSliceValue(list, "items")

	var newList []map[string]interface{}
	for _, item := range items {
		itemVars := GetVars(item)

		// Merge item vars with list vars. Each iteration uses an unmodified copy of listVars.
		mergedVars := configmap.MergeMapRecursive(listVars, itemVars)

		if metadata, found := configmap.GetMapEntry(item, "metadata"); found {
			metadata["vars"] = mergedVars
		}

		newList = append(newList, item)
	}

	return newList
}

// inherit allows to inherit from a given "cattle" template
// FIXME not at all used, because of bad documentation

func Inherit(object map[string]interface{}) map[string]interface{} {
	kind, ok := configmap.GetString(object, "kind")
	if !ok {
		return object
	}
	metadata, ok := configmap.GetMapEntry(object, "metadata")
	if !ok {
		return object
	}
	inherit, ok := configmap.GetString(metadata, "inherit")
	if !ok {
		return object
	}
	spec, ok := configmap.GetMapEntry(object, "spec")
	if !ok {
		return object
	}
	templateData := map[string]interface{}{}

	if strings.ToLower(kind) == "lxc" {
		template := cattles.CreateCattle("ct", inherit, templateData)
		newSpec := configmap.MergeMapRecursive(template, spec)
		object["spec"] = newSpec
		return object
	}
	if strings.ToLower(kind) == "virtualmachine" {
		template := cattles.CreateCattle("vm", inherit, templateData)
		newSpec := configmap.MergeMapRecursive(template, spec)
		object["spec"] = newSpec
		return object
	}

	return templateData
}

/*
func DoFlash(kind string, spec map[string]interface{}, node string) {

	//pxClient := shared.GlobalPxCluster.GetPxClient(node)
	if strings.ToLower(kind) == "lxc" {
		return
	}

	vmid := DetermineVmid(spec, "qemu")
	if vmid != 0 {
		_, found := shared.GlobalPxCluster.UniqueMachines[vmid]
		if !found {
			//fmt.Fprintf(os.Stderr, "vmid not found\n")
			return
		}
	}

	//pxClient := shared.GlobalPxCluster.GetPxClient(node)

	vmConfig, err := shared.JSONGetVMConfig(node, int64(vmid))
	if err != nil {
		return
	}
	vmConfigData, ok := configmap.GetMapEntry(vmConfig, "data")
	if !ok {
		return
	}
	//fmt.Fprintf(os.Stderr, "vmConfigData = %v %v\n", vmConfigData, err)

	name, ok := configmap.GetString(vmConfigData, "name")

	if !ok {
		return
	}
	if len(flashOptions.Match) != 0 {
		// FIXME:
		// FIXME: MATCH

		if !strings.HasPrefix(name, flashOptions.Match) {
			return
		}
		fmt.Fprintf(os.Stderr, "match = %v\n", flashOptions.Match)
	}
	fmt.Fprintf(os.Stderr, "vmid %v\n", vmid)
	fmt.Fprintf(os.Stderr, "name = %v\n", name)

	boot, ok := configmap.GetString(vmConfigData, "boot")
	if !ok {
		fmt.Fprintf(os.Stderr, "no boot entry for vmid %v\n", vmid)
		return
	}
	//fmt.Fprintf(os.Stderr, "boot = %v\n", boot)
	if !strings.HasPrefix(boot, "order=") {
		return
	}
	boot = boot[len("order="):]
	array := strings.Split(boot, ";")
	storageDrive := array[0]
	fmt.Fprintf(os.Stderr, "boot = %v\n", storageDrive)

	disk, ok := configmap.GetString(vmConfigData, storageDrive)
	if !ok {
		return
	}
	//fmt.Fprintf(os.Stderr, "disk = %v\n", disk)
	array = strings.Split(disk, ",")
	diskValue := array[0]
	fmt.Fprintf(os.Stderr, "diskValue = %v\n", diskValue)

	//result := pxClient.ApiClient.NodesApi.GetVMConfig(pxClient.Context, node, int64(vmid))
	//fmt.Fprintf(os.Stderr, "result = %v\n", result)

	cluster, err := shared.PickCluster(shared.GlobalConfigData, ClusterName)
	if err != nil {
		return
	}

	aliases := shared.GlobalPxCluster.GetAliasOnNode(node)
	fmt.Fprintf(os.Stderr, "aliases = %v\n", aliases)
	storageNames := shared.GlobalPxCluster.GetStorageNamesOnNode(node)
	spec = cattles.ProcessStorage(spec, aliases, storageNames)

	storageData, _ := configmap.GetMapEntry(spec, storageDrive) // spec???

	importFrom, ok := configmap.GetString(storageData, "import-from")
	if !ok {
		return
	}

	pxClients := shared.GetStorageContentAll(shared.GlobalPxCluster.PxClients)
	shared.GlobalPxCluster.PxClients = pxClients

	selectors, _ := configmap.GetMapEntry(cluster, "selectors")
	newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
	latestContent := shared.ExtractLatest(shared.GlobalPxCluster, newStorageContent)

	//fmt.Fprintf(os.Stderr, "newStorageContent = %v\n", newStorageContent)
	found := false
	var latestAndGreatest string
	for _, storageLatestItem := range latestContent {
		label := storageLatestItem["label"].(string)
		if importFrom == label {
			latestAndGreatest = storageLatestItem["volid"].(string)
			found = true
			break
		}
	}
	if !found {
		return
	}
	fmt.Fprintf(os.Stderr, "latestAndGreatest = %v\n", latestAndGreatest)

	newConfig := map[string]interface{}{}
	storageData["import-from"] = latestAndGreatest
	newConfig[storageDrive] = storageData

	//fmt.Fprintf(os.Stderr, "%v\n", newConfig)

	json_txt, _ := json.Marshal(newConfig)
	//fmt.Fprintf(os.Stderr, "%s\n", json_txt)

	updateVMConfigRequestObject := pxapiobject.UpdateVMConfigRequest{}
	err = json.Unmarshal(json_txt, &updateVMConfigRequestObject)

	updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
	shared.CopyUpdateVMConfigRequest(&updateVMConfigRequest, &updateVMConfigRequestObject)

	fmt.Fprintf(os.Stderr, "%s\n", *updateVMConfigRequest.Virtio0)

	resp, err := shared.UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)

	upid := resp.GetData()
	fmt.Fprintf(os.Stderr, "upid = %s\n", upid)
	//shared.GetNodeTaskStatus(node, upid)
	shared.WaitForUPID(node, upid)

	//fmt.Fprintf(os.Stderr, "%v %v\n", res, err)
	//shared.WaitForStatus(node,int64(vmid), "test", 60)
	//shared.Wait(node,int64(vmid))

	//time.Sleep(2 * time.Second)
	//shared.WaitForVMUnlock(node, int64(vmid))

	diff := CalcSizeDifference(vmConfigData, spec, storageDrive)
	fmt.Fprintf(os.Stderr, "diff %v\n", diff)
	if diff == 0 {
		return
	}
	res, err := shared.ResizeVMDisk(node, int64(vmid), storageDrive, "+"+strconv.FormatInt(diff, 10))
	upid = res.GetData()
	fmt.Fprintf(os.Stderr, "upid = %s\n", upid)
	shared.WaitForUPID(node, upid)

	//fmt.Fprintf(os.Stderr, "%v %v\n", res, err)
	//time.Sleep(2 * time.Second)
	//shared.WaitForVMUnlock(node, int64(vmid))

}
*/

func extractDiskValue(disk string) string {
	return strings.Split(disk, ",")[0]
}

func prepareSpecForStorage(node string, spec map[string]interface{}, storageDrive string) {
	aliases := etc.GlobalPxCluster.GetAliasOnNode(node)
	storageNames := etc.GlobalPxCluster.GetStorageNamesOnNode(node)
	cattles.ProcessStorage(spec, aliases, storageNames)
}

func getLatestAndGreatestImport(cluster, spec map[string]interface{}, storageDrive string) (string, error) {
	storageData, ok := configmap.GetMapEntry(spec, storageDrive)
	if !ok {
		return "", fmt.Errorf("storage data not found")
	}

	importFrom, ok := configmap.GetString(storageData, "import-from")
	if !ok {
		return "", fmt.Errorf("import-from not found")
	}

	newStorageContent := shared.ExtractLatest(etc.GlobalPxCluster, shared.JoinClusterAndSelector(etc.GlobalPxCluster, cluster["selectors"].(map[string]interface{})))

	for _, storageLatestItem := range newStorageContent {
		if label, ok := storageLatestItem["label"].(string); ok && importFrom == label {
			return storageLatestItem["volid"].(string), nil
		}
	}
	return "", fmt.Errorf("latest and greatest not found")
}

func updateVMConfig2(node string, vmid int, storageDrive, latestAndGreatest string) {
	newConfig := map[string]interface{}{
		storageDrive: map[string]interface{}{
			"import-from": latestAndGreatest,
		},
	}

	jsonTxt, err := json.Marshal(newConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling new config: %v\n", err)
		return
	}

	var updateVMConfigRequestObject pxapiobject.UpdateVMConfigRequest
	if err := json.Unmarshal(jsonTxt, &updateVMConfigRequestObject); err != nil {
		fmt.Fprintf(os.Stderr, "error unmarshaling to update VM config request object: %v\n", err)
		return
	}

	var updateVMConfigRequest pxapiflat.UpdateVMConfigRequest
	shared.CopyUpdateVMConfigRequest(&updateVMConfigRequest, &updateVMConfigRequestObject)
	fmt.Fprintf(os.Stderr, "%s\n", *updateVMConfigRequest.Virtio0)

	if _, err := api.UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest); err != nil {
		fmt.Fprintf(os.Stderr, "error updating VM config: %v\n", err)
		return
	}
}

func DoFlash(kind string, spec map[string]interface{}, node string) {
	if isLXC(kind) {
		return
	}

	vmid := DetermineVmid(spec, "qemu")
	if etc.GlobalPxCluster.Exists(node, int64(vmid)) {
		return
	}

	vmConfigData, err := getVMConfigData(node, vmid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting VM config data: %v\n", err)
		return
	}

	name, boot, storageDrive, err := extractVMConfigDetails(vmConfigData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error extracting VM config details: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "boot = %v (%v)\n", storageDrive, boot)

	if !nameMatchesOption(name) {
		return
	}

	disk, err := getDiskFromConfig(vmConfigData, storageDrive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting disk from config: %v\n", err)
		return
	}

	diskValue := extractDiskValue(disk)
	fmt.Fprintf(os.Stderr, "diskValue = %v\n", diskValue)

	cluster, err := shared.PickCluster(etc.GlobalConfigData, ClusterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error picking cluster: %v\n", err)
		return
	}

	prepareSpecForStorage(node, spec, storageDrive)

	latestAndGreatest, err := getLatestAndGreatestImport(cluster, spec, storageDrive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting latest and greatest import: %v\n", err)
		return
	}

	updateVMConfig2(node, vmid, storageDrive, latestAndGreatest)

	diff := CalcSizeDifference(vmConfigData, spec, storageDrive)
	if diff == 0 {
		return
	}

	queries.ResizeVMDisk(node, vmid, storageDrive, diff)
}

func getVMConfigData(node string, vmid int) (map[string]interface{}, error) {
	vmConfig, err := queries.JSONGetVMConfig(node, int64(vmid))
	if err != nil {
		return nil, err
	}
	vmConfigData, ok := configmap.GetMapEntry(vmConfig, "data")
	if !ok {
		return nil, fmt.Errorf("data entry not found in VM config")
	}
	return vmConfigData, nil
}

func extractVMConfigDetails(vmConfigData map[string]interface{}) (name, boot, storageDrive string, err error) {
	// Extract name
	name, ok := configmap.GetString(vmConfigData, "name")
	if !ok {
		return "", "", "", fmt.Errorf("name not found in VM config data")
	}

	// Extract boot and storage drive
	boot, ok = configmap.GetString(vmConfigData, "boot")
	if !ok {
		return "", "", "", fmt.Errorf("boot entry not found in VM config data")
	}
	if !strings.HasPrefix(boot, "order=") {
		return "", "", "", fmt.Errorf("invalid boot order format")
	}
	boot = boot[len("order="):]
	storageDrive = strings.Split(boot, ";")[0]

	return name, boot, storageDrive, nil
}

func nameMatchesOption(name string) bool {
	if len(flashOptions.Match) != 0 && !strings.HasPrefix(name, flashOptions.Match) {
		return false
	}
	fmt.Fprintf(os.Stderr, "match = %v\n", flashOptions.Match)
	return true
}

func getDiskFromConfig(vmConfigData map[string]interface{}, storageDrive string) (string, error) {
	disk, ok := configmap.GetString(vmConfigData, storageDrive)
	if !ok {
		return "", fmt.Errorf("disk entry not found for storage drive %s", storageDrive)
	}
	return disk, nil
}

func isLXC(kind string) bool {
	return strings.ToLower(kind) == "lxc"
}

/*
func DoCreate(kind string, spec map[string]interface{}, node string, metadata map[string]interface{}) {

	if isLXC(kind) {

		if shared.GlobalPxCluster.HasNode(node) {
			pxClient := shared.GlobalPxCluster.GetPxClient(node)
			itemVars := configmap.GetMapEntryWithDefault(metadata, "vars", map[string]interface{}{})
			vars := configmap.MergeMapRecursive(pxClient.Vars, itemVars)
			//fmt.Fprintf(os.Stderr, "DEBUG LXC: %v %v\n", kind, node)
			CreateCT(spec, vars, createOptions.Dump, node, "small", false, createOptions.DryRun, createOptions.Update)
		} else {
			fmt.Fprintf(os.Stderr, "node does not exist: %v\n", node)
		}
	} else {

		if shared.GlobalPxCluster.HasNode(node) {
			pxClient := shared.GlobalPxCluster.GetPxClient(node)

			if createOptions.Dump {
				fmt.Fprintf(os.Stderr, "--- ignition ---\n")
				proxmox.DumpJson(pxClient.Ignition)
			}

			//fmt.Fprintf(os.Stderr, "--- pxClient.Vars ---\n")
			//proxmox.DumpJson(pxClient.Vars)

			//fmt.Fprintf(os.Stderr, "--- itemVars ---")
			itemVars := configmap.GetMapEntryWithDefault(metadata, "vars", map[string]interface{}{})
			//proxmox.DumpJson(itemVars)
			//fmt.Fprintf(os.Stderr, "--- vars ---\n")
			vars := configmap.MergeMapRecursive(pxClient.Vars, itemVars)
			//proxmox.DumpJson(vars)
			//fmt.Fprintf(os.Stderr, "---------------------\n")
			CreateVM(spec, vars, createOptions.Dump, node, "small", createOptions.DryRun, createOptions.Update)

		} else {
			fmt.Fprintf(os.Stderr, "node does not exist: %v\n", node)
		}
	}
}
*/

func DoCreate(kind string, spec map[string]interface{}, node string, metadata map[string]interface{}) {
	if !etc.GlobalPxCluster.HasNode(node) {
		fmt.Fprintf(os.Stderr, "node does not exist: %v\n", node)
		return
	}

	pxClient, _ := etc.GlobalPxCluster.GetPxClient(node)
	itemVars := configmap.GetMapEntryWithDefault(metadata, "vars", map[string]interface{}{})
	vars := configmap.MergeMapRecursive(pxClient.Vars, itemVars)

	if isLXC(kind) {
		handleLXC(spec, vars, node)
	} else {
		handleVM(spec, vars, node, pxClient.Ignition)
	}
}

func handleLXC(spec, vars map[string]interface{}, node string) {
	CreateCT(spec, vars, createOptions.Dump, node, "small", false, createOptions.DryRun, createOptions.Update)
}

func handleVM(spec, vars map[string]interface{}, node string, ignition map[string]interface{}) {
	if createOptions.Dump {
		fmt.Fprintf(os.Stderr, "--- ignition ---\n")
		proxmox.DumpJson(ignition)
	}
	CreateVM(spec, vars, createOptions.Dump, node, "small", createOptions.DryRun, createOptions.Update)
}

func ProcessSection(cmd string, object map[string]interface{}) {
	kind, ok := configmap.GetString(object, "kind")
	if !ok {
		fmt.Fprintf(os.Stderr, "ProcessSection  kind %v\n", ok)
		return
	}
	if strings.ToLower(kind) == "list" {
		// Resolve List also inherits var data
		list := ResolveList(object)
		//fmt.Fprintf(os.Stderr, "ProcessSection() list len:  %v\n", len(list))
		for _, item := range list {
			ProcessSection(cmd, item)

		}
		return
	}

	object = Inherit(object)
	//proxmox.DumpJson(object)

	spec, ok := configmap.GetMapEntry(object, "spec")
	if !ok {
		return
	}

	metadata, ok := configmap.GetMapEntry(object, "metadata")
	if !ok {
		return
	}

	node, ok := configmap.GetString(metadata, "node")
	if !ok {
		return
	}

	if cmd == "create" {
		DoCreate(kind, spec, node, metadata)
	}

	if cmd == "flash" {
		DoFlash(kind, spec, node)
	}
}
