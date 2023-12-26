package cmd

import (
	"fmt"
	"os"
	"px/configmap"
	"px/proxmox"
	"px/proxmox/cattles"
	"px/shared"
	"strings"
	"encoding/json"
	"strconv"
	"github.com/DirkTheDaring/px-api-client-go"
	"github.com/DirkTheDaring/px-api-client-internal-go"
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
	currentSize := GetDeployedDriveSize(vmConfigData,storageDrive)
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

func GetVars(object map[string]interface{}) map[string]interface{} {
	metadata, ok := configmap.GetMapEntry(object, "metadata")
	if !ok {
		return map[string]interface{}{}
	}
	vars, ok := configmap.GetMapEntry(metadata, "vars")
	if !ok {
		return map[string]interface{}{}
	}
	return vars
}

func ResolveList(list map[string]interface{}) []map[string]interface{} {
	newList := []map[string]interface{}{}
	listVars := GetVars(list)
	items, _ := configmap.GetInterfaceSliceValue(list, "items")
	for _, item := range items {
		vars := GetVars(item)
		newVars := configmap.MergeMapRecursive(listVars, vars)
		metadata, _ := configmap.GetMapEntry(item, "metadata")
		metadata["vars"] = newVars
		newList = append(newList, item)
	}
	return newList
}
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

func ProcessSection(object map[string]interface{}, cmd string) {

	kind, ok := configmap.GetString(object, "kind")
	//fmt.Fprintf(os.Stderr, "ProcessSection %v\n", kind)
	if !ok {
		return
	}
	if strings.ToLower(kind) == "list" {


		// Resolve List also inherits var data
		list := ResolveList(object)
		//fmt.Fprintf(os.Stderr, "ProcessSection() list len:  %v\n", len(list))
		for _, item := range list {
			//metadata := configmap.GetMapEntryWithDefault(item, "metadata", map[string]interface{}{})
			//vars := configmap.GetMapEntryWithDefault(metadata, "vars", map[string]interface{}{})
			//proxmox.DumpJson(vars)
			ProcessSection(item, cmd)
		}
		return
	}

	object = Inherit(object)
	//proxmox.DumpJson(object)

	//fmt.Fprintf(os.Stderr, "STEP 01\n")
	spec, ok := configmap.GetMapEntry(object, "spec")
	if !ok {
		return
	}

	//fmt.Fprintf(os.Stderr, "STEP 02\n")
	metadata, ok := configmap.GetMapEntry(object, "metadata")
	if !ok {
		return
	}

	//fmt.Fprintf(os.Stderr, "STEP 03\n")
	node, ok := configmap.GetString(metadata, "node")
	if !ok {
		return
	}

	if cmd == "create" {
		if strings.ToLower(kind) == "lxc" {
			/*
				hostname, ok := configmap.GetString(spec, "hostname")
				if !ok {
					hostname = ""
				}
				fmt.Fprintf(os.Stderr, "%v %v %v\n", kind, hostname, node)
			*/
			if shared.GlobalPxCluster.HasNode(node) {
				pxClient := shared.GlobalPxCluster.GetPxClient(node)
				itemVars := configmap.GetMapEntryWithDefault(metadata, "vars", map[string]interface{}{})
				vars := configmap.MergeMapRecursive(pxClient.Vars, itemVars)
				//fmt.Fprintf(os.Stderr, "DEBUG LXC: %v %v\n", kind, node)
				CreateCT(spec, vars, createOptions.Dump, node, "small", false, createOptions.DryRun, createOptions.Update )
			} else {
				fmt.Fprintf(os.Stderr, "node does not exist: %v\n", node)
			}
		} else {
			/*
				name, ok := configmap.GetString(spec, "name")
				if !ok {
					name = ""
				}
				fmt.Fprintf(os.Stderr, "VM: %v %v %v\n", kind, name, node)
			*/

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
		return
	}
	if cmd == "flash" {
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
		/*
		SetImportFrom(cluster, spec)
		*/

	        aliases := shared.GlobalPxCluster.GetAliasOnNode(node)
		fmt.Fprintf(os.Stderr, "aliases = %v\n", aliases)
	        storageNames := shared.GlobalPxCluster.GetStorageNamesOnNode(node)
		spec = cattles.ProcessStorage(spec, aliases, storageNames)

		storageData, _ := configmap.GetMapEntry(spec, storageDrive) // spec???

		importFrom, ok := configmap.GetString(storageData, "import-from")
		if !ok {
			return
		}

		/*
		importFrom, ok := configmap.GetString(storageData, "import-from")
		if !ok {
			return
		}
		//fmt.Fprintf(os.Stderr, "import-from = %v\n", importFrom)
		*/


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
                newConfig[storageDrive] =  storageData

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
		shared.WaitForUPID(node,upid)

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
		shared.WaitForUPID(node,upid)

		//fmt.Fprintf(os.Stderr, "%v %v\n", res, err)
                //time.Sleep(2 * time.Second)
		//shared.WaitForVMUnlock(node, int64(vmid))

	}
}

func ProcessFiles(filenames []string, cmd string) {
	//fmt.Fprintf(os.Stderr, "ProcessFiles\n")
	for _, filename := range filenames {
		//fmt.Fprintf(os.Stderr, "ProcessFiles 1\n")
		sections, err := shared.ReadYAMLWithDashDashDashSingle(filename)
		//fmt.Fprintf(os.Stderr, "%v %v %v\n", filename, len(sections), err)
		//fmt.Fprintf(os.Stderr, "ProcessFiles 2\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v %v\n", filename, err)
			continue
		}
		//fmt.Fprintf(os.Stderr, "ProcessFiles 3\n")
		for _, section := range sections {
			//fmt.Fprintf(os.Stderr, "ProcessFiles() content: %v\n", section)
			ProcessSection(section, cmd)
		}

	}
}
