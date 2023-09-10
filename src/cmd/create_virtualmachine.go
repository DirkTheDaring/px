/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"px/configmap"
	"px/ignition"
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
	//Size   string
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
func DetermineVmid(machine map[string]interface{}, _type string) int {

	vmid, ok := configmap.GetInt(machine, "vmid")
	if !ok {
		//fmt.Fprintf(os.Stderr, "DetermineVmid() vmid not a number.\n")
		goto skip
	}
	if InProxmoxVmidRange(vmid) {
		return vmid
	}
	return 0
skip:
	var name string
	if _type == "lxc" {
		name = "hostname"
	} else {
		name = "name"
	}
	vmid, err := shared.GetVmidByAttribute(machine, name)
	if err == nil {
		//fmt.Fprintf(os.Stderr, "vmid from name: %v\n", vmid)
		return vmid
	}
	ip, ok := shared.GetIpv4Address(machine, _type)
	if !ok {
		return 0
	}
	vmid, err = shared.DeriveVmidFromIp4Address(ip)
	//fmt.Fprintf(os.Stderr, "vmid from ipv4: %v\n", vmid)
	if err != nil {
		return 0
	}

	return vmid
}
func SetImportFrom(cluster map[string]interface{}, machine map[string]interface{}) {
	selectors, _ := configmap.GetMapEntry(cluster, "selectors")
	newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
	latestContent := shared.ExctractLatest(shared.GlobalPxCluster, newStorageContent)

	storageDrives := configmap.SelectKeys("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$", machine)

	for _, storageDrive := range storageDrives {
		storageData, _ := configmap.GetMapEntry(machine, storageDrive)
		importFrom, ok := configmap.GetString(storageData, "import-from")
		if !ok {
			continue
		}
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
			continue
		}
		storageData["import-from"] = latestAndGreatest
	}
}

func SetOSTemplate(cluster map[string]interface{}, machine map[string]interface{}) {

	selectors, _ := configmap.GetMapEntry(cluster, "selectors")
	newStorageContent := shared.JoinClusterAndSelector(shared.GlobalPxCluster, selectors)
	latestContent := shared.ExctractLatest(shared.GlobalPxCluster, newStorageContent)
	ostemplate, ok := configmap.GetString(machine, "ostemplate")
	if !ok {
		return
	}
	fmt.Printf("ostemplate %s\n", ostemplate)

	var latestAndGreatest string
	for _, storageLatestItem := range latestContent {
		label := storageLatestItem["label"].(string)
		if ostemplate == label {
			latestAndGreatest = storageLatestItem["volid"].(string)
			machine["ostemplate"] = latestAndGreatest
			return
		}
	}

}

func GetIgnitionArgs(cluster map[string]interface{}, node string, ignitionName string) (string, error) {
	//fmt.Fprintf(os.Stderr, "GetIgnitionArgs() NODE: %v\n", node)
	ignitionConfiguration, _ := configmap.GetMapEntry(cluster, "ignition")

	if createVirtualmachineOptions.Dump {
		proxmox.DumpJson(ignitionConfiguration)
	}
	storage, ok := configmap.GetString(ignitionConfiguration, "storage")
	if !ok {
		return "", errors.New("the ignition section does not contain a storage entry.")
	}
	// Fixme you should also check if this storage can be used to upload iso files
	if !query.In(shared.GlobalPxCluster.GetStorageNamesOnNode(node), storage) {
		return "", errors.New(fmt.Sprintf("storage for ignition does not exist: %s\n", storage))
	}

	pxClient := shared.GlobalPxCluster.GetPxClient(node)
	storageEntry := pxClient.GetStorageByName(storage)
	if storageEntry == nil {
		return "", errors.New(fmt.Sprintf("storage for ignition not found: %s\n", storage))
	}
	contentArray := strings.Split(storageEntry["content"].(string), ",")
	if !query.In(contentArray, "iso") {
		return "", errors.New(fmt.Sprintf("storage does not accept iso content: %s\n", storage))
	}
	path := storageEntry["path"].(string)
	ignitionArgs := "-fw_cfg name=opt/com.coreos/config,file=" + path + "/template/iso/" + ignitionName + ".iso"
	return ignitionArgs, nil
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

func MapRootFs(machine map[string]interface{}, aliases map[string]string, storageNames []string) map[string]interface{} {
	rootfs, ok := configmap.GetMapEntry(machine, "rootfs")
	if !ok {
		return machine
	}
	volume, ok := configmap.GetString(rootfs, "volume")
	if !ok {
		return machine
	}
	slice := strings.Split(volume, ":")

	if len(slice) != 2 {
		return machine
	}

	storageName := slice[0]

	alias, ok := aliases[storageName]
	if ok {
		volume = alias
	}
	if !query.In(storageNames, volume) {
		fmt.Fprintf(os.Stderr, "storageName for %s not found: %s\n", "rootfs/volume", volume)
		return machine
	}
	rootfs["volume"] = volume + ":" + slice[1]
	machine["rootfs"] = rootfs
	return machine

}

func CreateCT(machine map[string]interface{}, vars map[string]interface{}, dump bool, node string, cattle string, createIgnition bool, dryRun bool) error {

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
	}
	if createCT {
		fmt.Fprintf(os.Stderr, "CREATE CT %v %v\n", vmid, node)
	} else {
		fmt.Fprintf(os.Stderr, "UPDATE CT %v %v\n", vmid, node)
	}
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

		shared.CreateContainer(node, ct2)
		//shared.WaitForVMUnlock(node, int64(vmid))
	}

	return nil
}

// FIXME Shell are more sophisticated
func ParseArgs(spec map[string]interface{}) []string {
	argString, ok := configmap.GetString(spec, "args")
	if !ok {
		return []string{}
	}
	args := strings.Split(argString, " ")
	return args
}

func HasIgnition(args []string) bool {
	if len(args) < 2 {
		return false
	}
	fwCfgFound := false
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if fwCfgFound {
			subArgs := strings.Split(arg, ",")
			for _, subArg := range subArgs {
				if strings.Index(subArg, "name=opt/com.coreos/config") == 0 {
					return true
				}
			}
			fwCfgFound = false
		} else {
			if strings.Index(arg, "-fw_cfg") == 0 {
				fwCfgFound = true
			}
		}
	}
	return false
}

func CreateVM(spec map[string]interface{}, vars map[string]interface{}, dump bool, node string, cattle string, dryRun bool) error {
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
	}
	if createVM {
		fmt.Fprintf(os.Stderr, "CREATE VM %v %v\n", vmid, node)
	} else {
		fmt.Fprintf(os.Stderr, "UPDATE VM %v %v\n", vmid, node)
	}

	//fmt.Fprintf(os.Stderr, "boot: %v\n", spec["boot"])

	spec["vmid"] = vmid

	_, found := configmap.GetString(spec, "name")
	if !found {
		spec["name"] = "vm" + strconv.Itoa(vmid)
	}

	sshkeys_unescaped, found := configmap.GetString(spec, "sshkeys")
	if found {
		// Absolute QUIRK to satisfy proxmox API
		//sshkeys_escaped := shared.UrlEncode(sshkeys_unescaped, true)
		sshkeys_escaped := url.PathEscape(sshkeys_unescaped)
		sshkeys_escaped = strings.Replace(sshkeys_escaped, "@", "%40", -1)
		sshkeys_escaped = strings.Replace(sshkeys_escaped, "+", "%2B", -1)
		sshkeys_escaped = strings.Replace(sshkeys_escaped, "=", "%3D", -1)

		//sshkeys_escaped = strings.Replace(sshkeys_escaped, ".", "%2E", -1)
		//sshkeys_escaped := sshkeys_unescaped
		//fmt.Fprintf(os.Stderr, "sshkeys old: %v\n", sshkeys_unescaped)
		//fmt.Fprintf(os.Stderr, "sshkeys new: %v\n", sshkeys_escaped)
		spec["sshkeys"] = sshkeys_escaped
	}

	//fmt.Fprintf(os.Stderr, "vmid: %v name: %v\n", vmid, result["name"])
	//os.Exit(1)
	// Figure out naming for machine - also that we can name the ignition file later on

	ignitionFilename := ""
	ignitionArgs := ""

	ignitionName := spec["name"].(string) + ".ign"

	args := ParseArgs(spec)
	createIgnition := HasIgnition(args)

	fmt.Fprintf(os.Stderr, "CreateVM() createIgnition: %v\n", createIgnition)

	if createIgnition {
		//ignitionConfiguration["enabled"] = "true"
		ignitionArgs, err = GetIgnitionArgs(cluster, node, ignitionName)
		if err != nil {
			fmt.Fprintf(os.Stdout, "ignition rendering failed %v\n", err)
			os.Exit(1)
		}
	}

	if ignitionArgs != "" {
		//fmt.Fprintf(os.Stderr, "ignitionArgs: %v\n", ignitionArgs)
		spec["args"] = ignitionArgs
		//fmt.Fprintf(os.Stderr, "---\n")
		//proxmox.DumpJson(result)
		//fmt.Fprintf(os.Stderr, "---\n")
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
		shared.WaitForVMUnlock(node, int64(vmid))
	}
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

	// Read the macaddres from the vmConfigData, and merge it to result
	// then we finally have a fully configured network device section, with
	// macaddress, which can be used in ignition or later maybe in autounattend
	netVars := BuildMacAddressEntries(spec, vmConfigData)
	spec = configmap.MergeMapRecursive(netVars, spec)

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
		/*
		persistent_modules, found := configmap.GetString(ignitionConfiguration, "persistent_modules")
		if found {
			vars["persistent_modules"] = persistent_modules

		}
		*/

		//proxmox.DumpJson(vars)

		if dump {
			//fmt.Fprintf(os.Stderr, "XXX---\n")
			proxmox.DumpJson(vars)
			///fmt.Fprintf(os.Stderr, "XXX---\n")
		}

		ignition.CreateProxmoxIgnitionFile(vars, ignitionName)

		ignitionFilename = "output/" + ignitionName + ".iso"
		if dryRun {
			goto skip
		}
		fmt.Fprintf(os.Stderr, "Upload %v %v\n", node, storage)
		f, _ := os.Open(ignitionFilename)
		shared.Upload(node, storage, "iso", f)
		f.Close()

	}
skip:
	// Some post processing, after Virtual Machine creation
	storageDrives := configmap.SelectKeys("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$", spec)

	for _, storageDrive := range storageDrives {
		storageData, _ := configmap.GetMapEntry(spec, storageDrive)
		//fmt.Fprintf(os.Stderr, "storageData %s %v\n", storageDrive, storageData)
		size, ok := configmap.GetString(storageData, "size")
		//fmt.Fprintf(os.Stderr, "importFrom: %v %v\n", ok, importFrom)
		if !ok {
			continue
		}
		sizeInBytes, ok := shared.CalculateSizeInBytes(size)
		if !ok {
			// Fixme this should have been validated all along already
			fmt.Fprintf(os.Stderr, "sizeInBytes: %v\n", sizeInBytes)
			continue
		}

		driveEntry, ok := configmap.GetString(vmConfigData, storageDrive)
		if !ok {
			continue
		}
		//fmt.Fprintf(os.Stderr, "driveEntry: %v\n", driveEntry)

		sizeStr, remainingDriveEntry := GetConfigOptionAndRemaining(driveEntry, "size")
		//fmt.Fprintf(os.Stderr, "sizeStr: %v remainingDriveEntry: %v\n", sizeStr, remainingDriveEntry)

		tmpArray := strings.Split(sizeStr, "=")
		if len(tmpArray) < 2 {
			continue
		}
		currentSizeInBytes, ok := shared.CalculateSizeInBytes(tmpArray[1])
		if !ok {
			continue
		}

		//fmt.Fprintf(os.Stderr, "currentSizeInBytes %v\n", currentSizeInBytes)
		if currentSizeInBytes >= sizeInBytes {
			continue
		}
		//fmt.Fprintf(os.Stderr, "RESIZE: %v\n", storageDrive)
		driveEntry = remainingDriveEntry + ",size=" + strconv.FormatInt(sizeInBytes, 10)
		//fmt.Fprintf(os.Stderr, "%v: %v\n", storageDrive, driveEntry)
		if false {
			newConfig := map[string]interface{}{}
			newConfig[storageDrive] = driveEntry
			txt, _ := json.Marshal(newConfig)
			if dump {
				//fmt.Fprintf(os.Stderr, "---\n")
				fmt.Fprintf(os.Stderr, "%v\n", string(txt))
				//fmt.Fprintf(os.Stderr, "---\n")
			}
			updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
			err = json.Unmarshal(txt, &updateVMConfigRequest)
			//fmt.Fprintf(os.Stderr, "vmid: %v %v\n", uint64(vmid), updateVMConfigRequest.GetVirtio0())
			shared.UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)
			shared.WaitForVMUnlock(node, int64(vmid))
		}

		delta_size := sizeInBytes - currentSizeInBytes
		shared.ResizeVMDisk(node, int64(vmid), storageDrive, "+"+strconv.FormatInt(delta_size, 10))
		shared.WaitForVMUnlock(node, int64(vmid))
	}
	return nil
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
	return CreateVM(result, vars, createVirtualmachineOptions.Dump, createVirtualmachineOptions.Node, createVirtualmachineOptions.Cattle, createVirtualmachineOptions.DryRun)
}
