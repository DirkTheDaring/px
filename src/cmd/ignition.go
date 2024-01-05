package cmd

import (
	"fmt"
	"os"
	"px/api"
	"px/configmap"
	"px/etc"
	"px/ignition"
	"px/proxmox"
	"px/proxmox/query"
	"px/shared"
	"strings"
)

/*
func DoIgnition(spec, cluster map[string]interface{}, node string, createIgnition bool,vars map[string]interface{}, dump bool, dryRun bool, ignitionName string,ignitionFilename string ) {


	if createIgnition {
		ignitionConfiguration := configmap.GetMapEntryWithDefault(cluster, "ignition", map[string]interface{}{})
		//fmt.Fprintf(os.Stderr, "--- IGNITION\n")
		//proxmox.DumpJson(ignitionConfiguration)
		//fmt.Fprintf(os.Stderr, "--- IGNITION\n")
		storage := configmap.GetStringWithDefault(ignitionConfiguration, "storage", "ignition")

		sshAuthorizedKeysUrl, found := configmap.GetString(ignitionConfiguration, "sshkeys_url")
		sshkeys2 := []string{}
		if found {
			sshkeys2 = shared.GetSSHResource(sshAuthorizedKeysUrl)
			// External Ssh keys first in the list. "Static ssh keys are the fallback"
		}
		//pxClient := shared.GlobalPxCluster.GetPxClient(node)
		//vars := pxClient.Vars

		sshkeys := configmap.GetStringSliceWithDefault(vars, "ssh_authorized_keys", []string{})
		//fmt.Fprintf(os.Stderr, "--- AAAAAAAAAAAAAAAAAAAAAAAA --- \n")

		//fmt.Fprintf(os.Stderr, "XXXX sshkeys: %v\n", sshkeys2)
		//sshkeys = append(sshkeys, sshkeys2...)
		vars["ssh_authorized_keys"] = sshkeys
		vars["ssh_authorized_keys2"] = sshkeys2
		vars["self"] = spec
		//proxmox.DumpJson(vars)

		if dump {
			//fmt.Fprintf(os.Stderr, "XXX---\n")
			proxmox.DumpJson(vars)
			///fmt.Fprintf(os.Stderr, "XXX---\n")
		}

		ignition.CreateProxmoxIgnitionFile(vars, ignitionName)

		ignitionFilename = "output/" + ignitionName + ".iso"
		if dryRun {
			return
		}
		//fmt.Fprintf(os.Stderr, "Upload %v %v\n", node, storage)
		f, _ := os.Open(ignitionFilename)
		shared.Upload(node, storage, "iso", f)
		f.Close()

	}
}
*/

// DoIgnition handles the ignition process based on the provided parameters.
func DoIgnition(spec, cluster map[string]interface{}, node string, createIgnition bool, vars map[string]interface{}, dump bool, dryRun bool, ignitionName string) {
	if !createIgnition {
		return
	}

	ignitionConfiguration := configmap.GetMapEntryWithDefault(cluster, "ignition", map[string]interface{}{})
	if dump {
		proxmox.DumpJson(ignitionConfiguration)
	}

	storage, sshkeys, sshkeys2 := getIgnitionConfig(ignitionConfiguration, vars)
	vars["ssh_authorized_keys"] = sshkeys
	vars["ssh_authorized_keys2"] = sshkeys2
	vars["self"] = spec

	if dump {
		proxmox.DumpJson(vars)
	}

	ignition.CreateProxmoxIgnitionFile(vars, ignitionName)

	if dryRun {
		return
	}
	uploadIgnitionFile(node, storage, ignitionName)
}

// getIgnitionConfig extracts the storage and SSH keys information from the ignition configuration.
func getIgnitionConfig(ignitionConfiguration, vars map[string]interface{}) (string, []string, []string) {
	storage := configmap.GetStringWithDefault(ignitionConfiguration, "storage", "default_ignition_storage")

	sshAuthorizedKeysUrl, found := configmap.GetString(ignitionConfiguration, "sshkeys_url")
	var sshkeys2 []string
	if found {
		sshkeys2 = shared.GetSSHResource(sshAuthorizedKeysUrl)
	}

	sshkeys := configmap.GetStringSliceWithDefault(vars, "ssh_authorized_keys", []string{})
	return storage, sshkeys, sshkeys2
}

// uploadIgnitionFile handles the uploading of the ignition file to the specified storage.
func uploadIgnitionFile(node, storage, ignitionName string) {
	ignitionFilename := "output/" + ignitionName + ".iso"
	f, err := os.Open(ignitionFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open ignition file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	api.Upload(node, storage, "iso", f)
}

// HandleIgnition processes the ignition parameters for a given virtual machine.
func HandleIgnition(spec, cluster map[string]interface{}, node string) string {
	if !HasIgnition(spec) {
		return ""
	}

	ignitionName := spec["name"].(string) + ".ign"
	ignitionArgs, err := GetIgnitionArgs(cluster, node, ignitionName)
	if err != nil {
		fmt.Fprintf(os.Stdout, "ignition rendering failed: %v\n", err)
		os.Exit(1)
	}

	if ignitionArgs != "" {
		spec["args"] = ignitionArgs
	}
	return ignitionName
}

// HasIgnition checks if the '-fw_cfg' followed by 'name=opt/com.coreos/config' is present in the spec's arguments.
func HasIgnition(spec map[string]interface{}) bool {
	argString, ok := configmap.GetString(spec, "args")
	if !ok {
		return false
	}

	return containsIgnitionConfig(strings.Fields(argString))
}

// containsIgnitionConfig checks if the provided arguments contain the ignition configuration.
func containsIgnitionConfig(args []string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-fw_cfg" && hasIgnitionConfig(args[i+1]) {
			return true
		}
	}
	return false
}

// hasIgnitionConfig checks if the string contains the ignition configuration.
func hasIgnitionConfig(arg string) bool {
	subArgs := strings.Split(arg, ",")
	for _, subArg := range subArgs {
		if strings.HasPrefix(subArg, "name=opt/com.coreos/config") {
			return true
		}
	}
	return false
}

func GetIgnitionArgs(cluster map[string]interface{}, node string, ignitionName string) (string, error) {
	//fmt.Fprintf(os.Stderr, "GetIgnitionArgs() NODE: %v\n", node)
	ignitionConfiguration, _ := configmap.GetMapEntry(cluster, "ignition")

	if createVirtualmachineOptions.Dump {
		proxmox.DumpJson(ignitionConfiguration)
	}
	storage, ok := configmap.GetString(ignitionConfiguration, "storage")
	if !ok {
		return "", fmt.Errorf("the ignition section does not contain a storage entry")
	}
	// Fixme you should also check if this storage can be used to upload iso files
	if !query.In(etc.GlobalPxCluster.GetStorageNamesOnNode(node), storage) {
		return "", fmt.Errorf(fmt.Sprintf("storage for ignition does not exist: %s", storage))
	}

	pxClient, _ := etc.GlobalPxCluster.GetPxClient(node)
	storageEntry := pxClient.GetStorageByName(storage)
	if storageEntry == nil {
		return "", fmt.Errorf(fmt.Sprintf("storage for ignition not found: %s", storage))
	}
	contentArray := strings.Split(storageEntry["content"].(string), ",")
	if !query.In(contentArray, "iso") {
		return "", fmt.Errorf(fmt.Sprintf("storage does not accept iso content: %s", storage))
	}
	path := storageEntry["path"].(string)
	ignitionArgs := "-fw_cfg name=opt/com.coreos/config,file=" + path + "/template/iso/" + ignitionName + ".iso"
	return ignitionArgs, nil
}
