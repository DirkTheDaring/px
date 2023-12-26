/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"px/configmap"
	"px/proxmox"
	"px/proxmox/cattles"
	"px/proxmox/query"
	"px/shared"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/DirkTheDaring/px-api-client-go"
	"github.com/DirkTheDaring/px-api-client-internal-go"

)

type CreateVirtualmachineOptions struct {
	Node     string
	Vmid     int
	Name     string
	Cattle   string
	Ignition bool
	Dump     bool

	Ip         string
	Gw         string
	DryRun     bool
	Nameserver string
	Update     bool
}

var createVirtualmachineOptions = &CreateVirtualmachineOptions{}

var PROXMOX_MIN_VMID int = 100
var PROXMOX_MAX_VMID int = 1000000

func checkErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

// virtualmachineCmd represents the virtualmachine command
var createVirtualmachineCmd = &cobra.Command{
	Use:     "virtualmachine",
	Aliases: []string{"vm"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		//fmt.Println("--- Pre run for create virtualmachine")
		pxClients := shared.GetStorageContentAll(shared.GlobalPxCluster.PxClients)
		//shared.GlobalPxCluster = shared.ProcessCluster(pxClients)
		shared.GlobalPxCluster.PxClients = pxClients
		//fmt.Println("--- Pre run end")

	},
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("virtualmachine called", args)
		checkErr(createVirtualmachineOptions.Validate())
		checkErr(createVirtualmachineOptions.Run())
	},
}

func init() {
	createCmd.AddCommand(createVirtualmachineCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// virtualmachineCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// virtualmachineCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//createVirtualmachineCmd.Flags().BoolP("dump", "", false, "dump")
	//createVirtualmachineCmd.Flags().StringSliceVar(&o.FileSources, "from-file", o.FileSources, "Key file can be specified using its file path, in which case file basename will be used as configmap key, or optionally with a key and file path, in which case the given key will be used.  Specifying a directory will iterate each named file in the directory whose basename is a valid configmap key.")
	//createVirtualmachineCmd.Flags().String("foo", "", "A help for foo")
	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Node, "node", "", "node")
	createVirtualmachineCmd.Flags().IntVar(&createVirtualmachineOptions.Vmid, "vmid", 0, "vmid from 100 to 100000. set 0 for autosetting")
	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Cattle, "cattle", "small", "cattle")
	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Name, "name", "", "name")
	createVirtualmachineCmd.Flags().BoolVar(&createVirtualmachineOptions.Ignition, "ignition", false, "ignition")
	createVirtualmachineCmd.Flags().BoolVar(&createVirtualmachineOptions.Dump, "dump", false, "dump")

	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Ip, "ip", "", "ip")
	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Gw, "gw", "", "gw")
	createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Nameserver, "nameserver", "", "nameserver")

	createVirtualmachineCmd.Flags().BoolVar(&createVirtualmachineOptions.DryRun, "dry-run", false, "dry-run")
	createVirtualmachineCmd.Flags().BoolVar(&createVirtualmachineOptions.Update, "update", false, "update")

	//createVirtualmachineCmd.Flags().StringVar(&createVirtualmachineOptions.Size, "size", "15G", "size")

}

// Complete loads data from the command line environment

func (o *CreateVirtualmachineOptions) Complete() error {
	//fmt.Fprintf(os.Stderr, "CreateVirtualmachineOptions o: %v\n", o)
	return nil
}

func (o *CreateVirtualmachineOptions) Validate() error {
	if o.Vmid == 0 {
		goto skip
	}
	if !(o.Vmid >= PROXMOX_MIN_VMID && o.Vmid <= PROXMOX_MAX_VMID) {
		return errors.New("vmid not in range from " + strconv.Itoa(PROXMOX_MIN_VMID) + " to " + strconv.Itoa(PROXMOX_MAX_VMID))
	}
skip:
	if !query.In(shared.GlobalPxCluster.Nodes, createVirtualmachineOptions.Node) {
		return errors.New("node does not exist: " + createVirtualmachineOptions.Node + " Possible nodes: " + fmt.Sprintf("%v", shared.GlobalPxCluster.Nodes))
	}

	if createVirtualmachineOptions.Ip != "" {
		_, _, err := net.ParseCIDR(createVirtualmachineOptions.Ip)
		if err != nil {
			return err
		}
	}
	if createVirtualmachineOptions.Gw != "" {
		if net.ParseIP(createVirtualmachineOptions.Gw) == nil {
			return errors.New("Invalid IP in gw: " + createVirtualmachineOptions.Gw)
		}
	}
	if createVirtualmachineOptions.Nameserver != "" {
		if net.ParseIP(createVirtualmachineOptions.Nameserver) == nil {
			return errors.New("Invalid IP in dns: " + createVirtualmachineOptions.Nameserver)
		}
	}
	if createVirtualmachineOptions.Name != "" {
		hostnameRegexRFC1123 := "^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$"
		matched, _ := regexp.MatchString(hostnameRegexRFC1123, createVirtualmachineOptions.Name)
		if !matched {
			return errors.New("name is not compliant with RFC1123: " + createVirtualmachineOptions.Name)
		}
	}
	// Fixme check if cattle name exists

	if createVirtualmachineOptions.Ignition {
		// Fixme check if target ignition exists
	}

	//fmt.Fprintf(os.Stderr, "CreateVirtualmachineOptions Validate o: %v\n", o)
	return nil
}
func CreatePxObjectCreateContainerRequest(config map[string]interface{}) (pxapiobject.CreateContainerRequest, error) {
	createContainerRequest := pxapiobject.CreateContainerRequest{}
	txt, err := json.Marshal(config)
	if createVirtualmachineOptions.Dump {
		fmt.Fprintf(os.Stderr, "%v\n", string(txt))
	}
	err = json.Unmarshal(txt, &createContainerRequest)
	if err != nil {
		return createContainerRequest, err
	}
	return createContainerRequest, err
}

func CreatePxObjectCreateVMRequest(config map[string]interface{}) (pxapiobject.CreateVMRequest, error) {
	createVMRequest := pxapiobject.CreateVMRequest{}
	txt, err := json.Marshal(config)
	if createVirtualmachineOptions.Dump {
		fmt.Fprintf(os.Stderr, "%v\n", string(txt))
	}
	err = json.Unmarshal(txt, &createVMRequest)
	if err != nil {
		return createVMRequest, err
	}
	return createVMRequest, err
}
func CreatePxFlatCreateContainerRequest(ct pxapiobject.CreateContainerRequest) (pxapiflat.CreateContainerRequest, error) {
	createContainerRequest := pxapiflat.CreateContainerRequest{}
	shared.CopyContainer(&createContainerRequest, &ct)

	txt, err := json.Marshal(createContainerRequest)
	if createVirtualmachineOptions.Dump {
		fmt.Fprintf(os.Stderr, "%v\n", string(txt))
	}
	err = json.Unmarshal(txt, &createContainerRequest)
	if err != nil {
		return createContainerRequest, err
	}
	return createContainerRequest, err
}

func CreatePxFlatCreateVMRequest(vm pxapiobject.CreateVMRequest) (pxapiflat.CreateVMRequest, error) {
	createVMRequest := pxapiflat.CreateVMRequest{}
	shared.CopyVM(&createVMRequest, &vm)

	//fmt.Fprintf(os.Stderr, "CreatePxFlatCreateVMRequest: %v\n", createVMRequest)

	txt, err := json.Marshal(createVMRequest)
	if createVirtualmachineOptions.Dump {
		fmt.Fprintf(os.Stderr, "%v\n", string(txt))
	}
	err = json.Unmarshal(txt, &createVMRequest)
	if err != nil {
		return createVMRequest, err
	}
	return createVMRequest, err
}
func GetConfigOptionAndRemaining(config string, option string) (string, string) {
	list := []string{}
	array := strings.Split(config, ",")
	resultOption := ""
	search := option + "="
	for _, item := range array {
		if strings.HasPrefix(item, search) {
			resultOption = item
			continue
		}
		list = append(list, item)
	}
	remainingString := strings.Join(list, ",")
	return resultOption, remainingString
}
func InProxmoxVmidRange(vmid int) bool {
	return vmid >= PROXMOX_MIN_VMID && vmid <= PROXMOX_MAX_VMID
}

// DetermineVmid extracts the VM ID from the provided machine map based on the given type.
// It tries different methods to find the VM ID and returns 0 if it cannot be determined.
func DetermineVmid(machine map[string]interface{}, _type string) int {
    if vmid, ok := configmap.GetInt(machine, "vmid"); ok && InProxmoxVmidRange(vmid) {
        return vmid
    }

    // Get the VMID by name or hostname depending on the type
    nameKey := getNameKey(_type)
    if vmid, err := shared.GetVmidByAttribute(machine, nameKey); err == nil {
        return vmid
    }

    // Attempt to get the VMID from IP address
    if ip, ok := shared.GetIpv4Address(machine, _type); ok {
        if vmid, err := shared.DeriveVmidFromIp4Address(ip); err == nil {
            return vmid
        }
    }

    // Return 0 if no VMID could be determined
    return 0
}

// getNameKey returns the appropriate key for the name or hostname based on the type.
func getNameKey(_type string) string {
    if _type == "lxc" {
        return "hostname"
    }
    return "name"
}


// SetImportFrom updates the 'import-from' field in machine's storage drives based on the cluster configuration.
func SetImportFrom(cluster, machine map[string]interface{}) {
    selectors, ok := configmap.GetMapEntry(cluster, "selectors")
    if !ok {
        // Handle error, log it, or return if necessary
        return
    }

    newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
    latestContent := shared.ExtractLatest(shared.GlobalPxCluster, newStorageContent)

    updateStorageDrives(machine, latestContent)
}

// updateStorageDrives updates the 'import-from' fields in the machine's storage drives.
func updateStorageDrives(machine map[string]interface{}, latestContent []map[string]interface{}) {
    storageDrivePattern := regexp.MustCompile("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$")

    for driveKey, driveData := range machine {
        if storageDrivePattern.MatchString(driveKey) {
            if storageData, ok := driveData.(map[string]interface{}); ok {
                updateImportFrom(storageData, latestContent)
            }
        }
    }
}

// updateImportFrom updates the 'import-from' field in a given storage data.
func updateImportFrom(storageData map[string]interface{}, latestContent []map[string]interface{}) {
    importFrom, ok := configmap.GetString(storageData, "import-from")
    if !ok {
        return
    }

    for _, latestItem := range latestContent {
        if latestItemLabel, ok := latestItem["label"].(string); ok && importFrom == latestItemLabel {
            if latestItemVolid, ok := latestItem["volid"].(string); ok {
                storageData["import-from"] = latestItemVolid
                break
            }
        }
    }
}


func SetOSTemplate(cluster map[string]interface{}, machine map[string]interface{}) {

	selectors, _ := configmap.GetMapEntry(cluster, "selectors")
	newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
	latestContent := shared.ExtractLatest(shared.GlobalPxCluster, newStorageContent)
	ostemplate, ok := configmap.GetString(machine, "ostemplate")
	if !ok {
		return
	}
	//fmt.Printf("ostemplate %s\n", ostemplate)

	var latestAndGreatest string
	for _, storageLatestItem := range latestContent {
		label := storageLatestItem["label"].(string)
		if ostemplate == label {
			latestAndGreatest = storageLatestItem["volid"].(string)
			machine["ostemplate"] = latestAndGreatest
			fmt.Fprintf(os.Stderr, "  ostemplate: map alias %s to %s\n", ostemplate, latestAndGreatest)
			return
		}
	}

}


func BuildMacAddressEntries(result map[string]interface{}, vmConfigData map[string]interface{}) map[string]interface{} {
	networkAdapters := configmap.SelectKeys("^(net)[0-9]+$", result)
	regex := "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"
	netVars := map[string]interface{}{}
	for _, networkAdapter := range networkAdapters {
		net := vmConfigData[networkAdapter].(string)
		array := strings.Split(net, ",")

		for _, item := range array {
			array2 := strings.Split(item, "=")
			if len(array) != 2 {
				continue
			}
			matched, _ := regexp.MatchString(regex, array2[1])
			if !matched {
				continue
			}
			macaddrMap := map[string]interface{}{}
			macaddrMap["model"] = array2[0]
			macaddrMap["macaddr"] = array2[1]

			netVars[networkAdapter] = macaddrMap
			break
		}
		//fmt.Fprintf(os.Stderr, "%v: %v\n", networkAdapter, netVars)
	}
	return netVars
}

// MapRootFs updates the 'rootfs' field in the machine map with an alias and ensures it's in the provided storage names.
func MapRootFs(machine map[string]interface{}, aliases map[string]string, storageNames []string) map[string]interface{} {
    rootfs, ok := configmap.GetMapEntry(machine, "rootfs")
    if !ok {
        return machine
    }

    volume, ok := configmap.GetString(rootfs, "volume")
    if !ok {
        return machine
    }

    updatedVolume := getUpdatedVolume(volume, aliases, storageNames)
    if updatedVolume == "" {
        fmt.Fprintf(os.Stderr, "storageName for %s not found: %s\n", "rootfs/volume", volume)
        return machine
    }

    rootfs["volume"] = updatedVolume
    machine["rootfs"] = rootfs
    return machine
}

// getUpdatedVolume processes the volume string and returns an updated volume if applicable.
func getUpdatedVolume(volume string, aliases map[string]string, storageNames []string) string {
    parts := strings.Split(volume, ":")
    if len(parts) != 2 {
        return ""
    }

    storageName, storagePath := parts[0], parts[1]
    if alias, ok := aliases[storageName]; ok {
        storageName = alias
    }

    if !query.In(storageNames, storageName) {
        return ""
    }

    return storageName + ":" + storagePath
}



func CreateCT(machine map[string]interface{}, vars map[string]interface{}, dump bool, node string, cattle string, createIgnition bool, dryRun bool, doUpdate bool) error {
	createCT := false

	vmid := DetermineVmid(machine, "lxc")
	//fmt.Fprintf(os.Stderr, "DetermineVmid() vmid: %v\n", vmid)
	if vmid != 0 {
		_, found := shared.GlobalPxCluster.UniqueMachines[vmid]
		if !found {
			createCT = true
		}
	}
	// if vmid is still zero
	if vmid == 0 {
		createCT = true
		vmid64, err := shared.GenerateClusterId(node)
		if err == nil {
			vmid = int(vmid64)
			fmt.Fprintf(os.Stderr, "vmid64: %v\n", vmid64)
			fmt.Fprintf(os.Stderr, "vmid: %v\n", vmid)
		}
	}

	if vmid == 0 {
		// Fixme fail if still 0
		fmt.Fprintf(os.Stderr, "Could not create CT.\n")
		return nil
	}

	if (!doUpdate && !createCT) {
		//fmt.Fprintf(os.Stderr, "SKIP: VM %v %v %v %v\n", vmid, node,doUpdate, createVM)
		fmt.Fprintf(os.Stderr, "SKIP: CT %v %v (use --update if required)\n", vmid, node)
		return nil
	}

	if createCT {
		fmt.Fprintf(os.Stderr, "CREATE CT %v %v\n", vmid, node)
	} else {
		fmt.Fprintf(os.Stderr, "UPDATE CT %v %v\n", vmid, node)
	}
	aliases := shared.GlobalPxCluster.GetAliasOnNode(node)
	storageNames := shared.GlobalPxCluster.GetStorageNamesOnNode(node)
	cluster, err := shared.PickCluster(shared.GlobalConfigData, ClusterName)
	if err != nil {
		return err
	}
	//SetImportFrom(cluster, machine)
	SetOSTemplate(cluster, machine)

	cattles.ProcessStorage(machine, aliases, storageNames)
	machine = MapRootFs(machine, aliases, storageNames)
	//fmt.Fprintf(os.Stderr, "NODE: %v\n", node)

	machine["vmid"] = vmid

	_, found := configmap.GetString(machine, "hostname")
	if !found {
		machine["hostname"] = "ct" + strconv.Itoa(vmid)
	}

	ct, err := CreatePxObjectCreateContainerRequest(machine)

	if err != nil {
		return err
	}
	/*
		fmt.Fprintf(os.Stderr, "%v\n", vm)
		fmt.Fprintf(os.Stderr, "vmid: %v\n", vm.GetVmid())
		fmt.Fprintf(os.Stderr, "args: %v\n", vm.GetArgs())
		fmt.Fprintf(os.Stderr, "memory: %v\n", vm.GetMemory())
	*/
	ct2, err := CreatePxFlatCreateContainerRequest(ct)

	if createCT && !dryRun { // only on create
		//fmt.Fprintf(os.Stderr, "%v\n", ct2)
                // FIXME wait for UPID?
		res, err := shared.CreateContainer(node, ct2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "createContainer error: %v\n", err)
			return nil
		}
		upid := res.GetData()
		shared.WaitForUPID(node,upid)
		//shared.WaitForVMUnlock(node, int64(vmid))
	}

	if !doUpdate {
		return nil
	}
	ctConfig, err := shared.JSONGetCTConfig(node,int64(vmid))
	ctConfigData,_ := configmap.GetMapEntry(ctConfig, "data")

	//fmt.Fprintf(os.Stderr, "ctConfig: %v %v\n", ctConfigData, err)

	newConfig := createChanges(ctConfigData, machine)

	//fmt.Fprintf(os.Stderr, "ctConfig: %v %v\n", ctConfigData, err)
	// No changes
	if len(newConfig) == 0 {
		return nil
	}

	txt, _ := json.Marshal(newConfig)

	//fmt.Fprintf(os.Stderr, "  update: %s\n", txt)

	updateContainerConfigSyncRequest := pxapiflat.UpdateContainerConfigSyncRequest{}
	err = json.Unmarshal(txt, &updateContainerConfigSyncRequest)
	if true {
		txt2,_ := json.Marshal(updateContainerConfigSyncRequest)
		fmt.Fprintf(os.Stderr, "  update: %s\n", txt2)
	}

	if dryRun {
		return nil
	}

	_, err = shared.UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)

	if err != nil {
		fmt.Fprintf(os.Stderr, "update container config for %v on node %s error: %v\n", vmid, node, err)
		return nil
	}
	//upid := resp.GetData()
	//shared.WaitForUPID(node,upid)

	return nil
}



func CreateVM(spec map[string]interface{}, vars map[string]interface{}, dump bool, node string, cattle string, dryRun bool, doUpdate bool) error {
	//fmt.Fprintf(os.Stderr, "NODE: %v\n", node)
	var ok bool
	createVM := false

	vmid := DetermineVmid(spec, "qemu")
	//fmt.Fprintf(os.Stderr, "vmid from GetVmid: %v\n", vmid)
	if vmid != 0 {
		_, found := shared.GlobalPxCluster.UniqueMachines[vmid]
		if !found {
			createVM = true
		}
	}
	// if vmid is still zero
	if vmid == 0 {
		createVM = true
		vmid64, err := shared.GenerateClusterId(node)
		if err == nil {
			vmid = int(vmid64)
			//fmt.Fprintf(os.Stderr, "vmid64: %v\n", vmid64)
			//fmt.Fprintf(os.Stderr, "vmid: %v\n", vmid)
		}
	}
	if vmid == 0 {
		// Fixme fail if still 0
		fmt.Fprintf(os.Stderr, "Could not create VM.\n")
		return nil
	}

	if (!doUpdate && !createVM) {
		//fmt.Fprintf(os.Stderr, "SKIP: VM %v %v %v %v\n", vmid, node,doUpdate, createVM)
		fmt.Fprintf(os.Stderr, "SKIP: VM %v %v (use --update if required)\n", vmid, node)
		return nil
	}
	if createVM {
		fmt.Fprintf(os.Stderr, "CREATE VM %v %v\n", vmid, node)
	} else {
		fmt.Fprintf(os.Stderr, "UPDATE VM %v %v\n", vmid, node)
	}

	// Map Aliases to real drives
	aliases := shared.GlobalPxCluster.GetAliasOnNode(node)

	//fmt.Fprintf(os.Stderr, "CreateVM() map=%v\n", shared.GlobalPxCluster.PxClientLookup)
	//fmt.Fprintf(os.Stderr, "CreateVM() aliases=%v\n", aliases)
	storageNames := shared.GlobalPxCluster.GetStorageNamesOnNode(node)
	cluster, err := shared.PickCluster(shared.GlobalConfigData, ClusterName)
	if err != nil {
		return err
	}
	SetImportFrom(cluster, spec)

	cattles.ProcessStorage(spec, aliases, storageNames)


	spec["vmid"] = vmid

	_, found := configmap.GetString(spec, "name")
	if !found {
		spec["name"] = "vm" + strconv.Itoa(vmid)
	}

	//fmt.Fprintf(os.Stderr, "vmid: %v name: %v\n", vmid, result["name"])
	//os.Exit(1)
	// Figure out naming for machine - also that we can name the ignition file later on

	var ignitionFilename string = ""
	var ignitionName string = ""

	createIgnition := HasIgnition(spec)
	if createIgnition {
		ignitionName = HandleIgnition(spec,cluster,node)
	}

	//proxmox.DumpJson(spec)
	vm, err := CreatePxObjectCreateVMRequest(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vm err: %v\n", err)
		return err
	}
	//fmt.Fprintf(os.Stderr, "RNG0: %v\n", vm.GetRng0())
	/*
		fmt.Fprintf(os.Stderr, "%v\n", vm)
		fmt.Fprintf(os.Stderr, "vmid: %v\n", vm.GetVmid())
		fmt.Fprintf(os.Stderr, "args: %v\n", vm.GetArgs())
		fmt.Fprintf(os.Stderr, "memory: %v\n", vm.GetMemory())
	*/
	vm2, err := CreatePxFlatCreateVMRequest(vm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "createFlat err: %v\n", err)
		return err
	}
	// quit here it is only a dry run

	if createVM && !dryRun { // only on create
		shared.CreateVM(node, vm2)
		// FIXME no UPID???
		shared.WaitForVMUnlock(node, int64(vmid))
	}
	// FIXME -- dry run
	if dryRun {
		return nil
	}

	// After creation of VM we need to read the created config
	// to get information like mac address
	vmConfig, err := shared.JSONGetVMConfig(node, int64(vmid))
	if err != nil {
		return err
	}
	vmConfigData, ok := configmap.GetMapEntry(vmConfig, "data")
	if !ok {
		return nil
	}


        // Sometimes updates are required (e.g. size change after image flash)
	DoVMUpdate(vmConfigData, spec, dryRun, node, vmid)


	// Read the macaddres from the vmConfigData, and merge it to result
	// then we finally have a fully configured network device section, with
	// macaddress, which can be used in ignition or later maybe in autounattend
	netVars := BuildMacAddressEntries(spec, vmConfigData)
	spec = configmap.MergeMapRecursive(netVars, spec)

	if createIgnition {
		DoIgnition(spec, cluster, node, createIgnition,vars,dump, dryRun,ignitionName, ignitionFilename)
	}
	return nil
}

// DoVMUpdate handles the virtual machine update process based on the provided parameters.
func DoVMUpdate(vmConfigData, spec map[string]interface{}, dryRun bool, node string, vmid int) error {
    if err := processStorageDrives(spec, vmConfigData, dryRun, node, vmid); err != nil {
        return err
    }

    if err := updateVMConfiguration(vmConfigData, spec, dryRun, node, vmid); err != nil {
        return err
    }

    return nil
}

func processStorageDrives(spec, vmConfigData map[string]interface{}, dryRun bool, node string, vmid int) error {
    storageDrives := configmap.SelectKeys("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$", spec)

    for _, storageDrive := range storageDrives {
        if err := handleDriveSize(spec, vmConfigData, dryRun, node, vmid, storageDrive); err != nil {
            fmt.Fprintf(os.Stderr, "%v\n", err)
        }
    }
    return nil
}

func handleDriveSize(spec, vmConfigData map[string]interface{}, dryRun bool, node string, vmid int, storageDrive string) error {
    storageData, ok := configmap.GetMapEntry(spec, storageDrive)
    if !ok {
        return nil // Skip if storage data is not found
    }

    size, ok := configmap.GetString(storageData, "size")
    if !ok {
        return nil // Skip if size is not specified
    }

    sizeInBytes, ok := shared.CalculateSizeInBytes(size)
    if !ok {
        return fmt.Errorf("invalid size format: %s\n", size)
    }

    driveEntry, ok := configmap.GetString(vmConfigData, storageDrive)
    if !ok {
        return nil // Skip if drive entry is not found
    }
    currentSizeInBytes, ok := extractCurrentSize(driveEntry)

    if !ok || currentSizeInBytes == sizeInBytes {
        return nil // Skip if current size is invalid or equal to desired size
    }

    if currentSizeInBytes > sizeInBytes {
        fmt.Fprintf(os.Stderr, "Drive %s size configuration is smaller than current VM: %s < %s\n", storageDrive, shared.ToSizeString(sizeInBytes), shared.ToSizeString(currentSizeInBytes))
        return nil
    }
    delta_size := sizeInBytes-currentSizeInBytes
    fmt.Fprintf(os.Stderr, "  %s increase size by %s to: %s\n", storageDrive, shared.ToSizeString(delta_size), shared.ToSizeString(sizeInBytes))

    if dryRun {
	    return nil
    }
    return resizeDrive(node, vmid, storageDrive, delta_size)
}

func extractCurrentSize(driveEntry string) (int64, bool) {
    sizeStr, _ := GetConfigOptionAndRemaining(driveEntry, "size")
    parts := strings.SplitN(sizeStr, "=", 2)
    if len(parts) < 2 {
        return 0, false
    }
    return shared.CalculateSizeInBytes(parts[1])
}

func resizeDrive(node string, vmid int, storageDrive string, deltaSize int64) error {
    res, err := shared.ResizeVMDisk(node, int64(vmid), storageDrive, "+"+strconv.FormatInt(deltaSize, 10))
    if err != nil {
        return fmt.Errorf("failed to resize disk: %v", err)
    }

    shared.WaitForUPID(node, res.GetData())
    return nil
}

func updateVMConfiguration(vmConfigData, spec map[string]interface{}, dryRun bool, node string, vmid int) error {
    newConfig := createChanges(vmConfigData, spec)
    if len(newConfig) == 0 {
        return nil
    }
    txt, _ := json.Marshal(newConfig)


    updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
    if err := json.Unmarshal(txt, &updateVMConfigRequest); err != nil {
        return fmt.Errorf("failed to unmarshal VM config: %v", err)
    }

    if dryRun {
        fmt.Fprintf(os.Stderr, "Update VM Config (Dry Run): %v\n", updateVMConfigRequest)
        return nil
    }
    fmt.Fprintf(os.Stderr, "  update: %s\n", txt)

    resp, err := shared.UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)
    if err != nil {
        return fmt.Errorf("failed to update VM config: %v", err)
    }

    shared.WaitForUPID(node, resp.GetData())
    return nil
}

// createChanges generates a new configuration based on the differences
// between the existing vmConfigData and the desired spec.
func createChanges(vmConfigData, spec map[string]interface{}) map[string]interface{} {
    newConfig := make(map[string]interface{})

    // Update configuration if there are changes.
    updateConfig(newConfig, vmConfigData, spec, "memory")
    updateConfig(newConfig, vmConfigData, spec, "balloon")
    updateConfig(newConfig, vmConfigData, spec, "cores")
    updateConfig(newConfig, vmConfigData, spec, "onboot")
    //fmt.Errorf("txt: %v\n", newConfig)
    //fmt.Fprintf(os.Stderr, "createChanges() %v\n", newConfig)

    return newConfig
}

// updateConfig checks and updates the configuration for a given key if needed.
func updateConfig(newConfig, vmConfigData, spec map[string]interface{}, key string) {
    if newValue, ok := configmap.GetInt64(spec, key); ok {
        if currentValue, ok := configmap.GetInt64(vmConfigData, key); ok && currentValue != newValue {
            newConfig[key] = newValue
	    return
        }
    }
    if newBoolValue, ok := configmap.GetBool(spec, key); ok {
	var flag float64 = 0
	if newBoolValue {
		flag = 1
	}
        if currentValue, ok := configmap.GetFloat64(vmConfigData, key); ok && currentValue != flag {
		newConfig[key] = flag

	}
    }
}


func (o *CreateVirtualmachineOptions) Run() error {
	//fmt.Fprintf(os.Stderr, "CreateVirtualmachineOptions o: %v\n", o)
	//fmt.Fprintf(os.Stderr, "clusterName: %s\n", ClusterName)
	templateData := map[string]interface{}{}
	templateData["node"] = createVirtualmachineOptions.Node
	templateData["vmid"] = createVirtualmachineOptions.Vmid
	templateData["name"] = createVirtualmachineOptions.Name
	templateData["nameserver"] = createVirtualmachineOptions.Nameserver

	if createVirtualmachineOptions.Ip != "" {
		ipconfig0 := map[string]interface{}{}
		ipconfig0["ip"] = createVirtualmachineOptions.Ip
		if createVirtualmachineOptions.Gw != "" {
			ipconfig0["gw"] = createVirtualmachineOptions.Gw
		}
		templateData["ipconfig0"] = ipconfig0
	}
	//fmt.Fprintf(os.Stderr, "NODE: %v\n", node)
	result := cattles.CreateCattle("vm", createVirtualmachineOptions.Cattle, templateData)
	if createVirtualmachineOptions.Dump {
		fmt.Fprintf(os.Stderr, "---\n")
		proxmox.DumpJson(result)
		fmt.Fprintf(os.Stderr, "---\n")
	}
	vars := shared.GlobalPxCluster.GetPxClient(createVirtualmachineOptions.Node).Vars
	return CreateVM(result, vars, createVirtualmachineOptions.Dump, createVirtualmachineOptions.Node, createVirtualmachineOptions.Cattle, createVirtualmachineOptions.DryRun, createVirtualmachineOptions.Update)
}
