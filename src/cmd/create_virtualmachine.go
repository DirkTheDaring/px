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
	"px/api"
	"px/configmap"
	"px/etc"
	"px/proxmox"
	"px/proxmox/cattles"
	"px/proxmox/query"
	"px/queries"
	"px/shared"
	"regexp"
	"strconv"
	"strings"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
	"github.com/spf13/cobra"
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
		shared.InitConfig(ClusterName)

		//fmt.Println("--- Pre run for create virtualmachine")
		queries.GetStorageContentAll(etc.GlobalPxCluster.GetPxClients())
		//shared.GlobalPxCluster = shared.ProcessCluster(pxClients)
		//etc.GlobalPxCluster.SetPxClients(pxClients)
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
	if !etc.GlobalPxCluster.HasNode(createVirtualmachineOptions.Node) {
		return errors.New("node does not exist: " + createVirtualmachineOptions.Node + " Possible nodes: " + fmt.Sprintf("%v", etc.GlobalPxCluster.GetNodeNames()))
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
	/*
		// FIXME check if cattle name exists

		if createVirtualmachineOptions.Ignition {
			// FIXME check if target ignition exists
		}
	*/
	//fmt.Fprintf(os.Stderr, "CreateVirtualmachineOptions Validate o: %v\n", o)
	return nil
}
func CreatePxObjectCreateContainerRequest(config map[string]interface{}) (pxapiobject.CreateContainerRequest, error) {
	createContainerRequest := pxapiobject.CreateContainerRequest{}
	txt, err := json.Marshal(config)
	if err != nil {
		return createContainerRequest, err
	}
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
	if err != nil {
		return createVMRequest, err
	}
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
	if err != nil {
		return createContainerRequest, err
	}
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
	if err != nil {
		return createVMRequest, err
	}
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

	newStorageContent := shared.JoinClusterAndSelector(*etc.GlobalPxCluster, selectors)
	latestContent := shared.ExtractLatest(*etc.GlobalPxCluster, newStorageContent)

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
	newStorageContent := shared.JoinClusterAndSelector(*etc.GlobalPxCluster, selectors)
	latestContent := shared.ExtractLatest(*etc.GlobalPxCluster, newStorageContent)
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
	regexCompiled, _ := regexp.Compile(regex)

	netVars := map[string]interface{}{}
	for _, networkAdapter := range networkAdapters {
		net := vmConfigData[networkAdapter].(string)
		array := strings.Split(net, ",")

		for _, item := range array {
			array2 := strings.Split(item, "=")
			if len(array) != 2 {
				continue
			}
			matched := regexCompiled.MatchString(array2[1])
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
	vmid, createCT, err := createVMID("lxc", machine, node)
	if err != nil {
		return fmt.Errorf("error creating VMID: %w", err)
	}

	if !doUpdate && !createCT {
		fmt.Fprintf(os.Stdout, "SKIP: CT %v %v (use --update if required)\n", vmid, node)
		return nil
	}

	fmt.Fprintf(os.Stdout, "%s CT %v %v\n", actionLabel(createCT), vmid, node)

	if err := handleCTCreation(machine, vars, node, vmid, createCT, dryRun); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	if doUpdate {
		return updateCTConfig(machine, node, vmid, dryRun)
	}

	return nil
}

func actionLabel(createCT bool) string {
	if createCT {
		return "CREATE"
	}
	return "UPDATE"
}

func handleCTCreation(machine, vars map[string]interface{}, node string, vmid int, createCT, dryRun bool) error {
	clusterDatabase, err := etc.GlobalPxCluster.PickCluster(ClusterName)
	if err != nil {
		return fmt.Errorf("error picking cluster: %w", err)
	}
	cluster := clusterDatabase.GetCluster()

	prepareMachineForCreation(machine, cluster, node, vmid)

	if createCT && !dryRun {
		return createContainer(node, machine, vmid)
	}
	return nil
}

func prepareMachineForCreation(machine, cluster map[string]interface{}, node string, vmid int) {
	SetOSTemplate(cluster, machine)
	aliases := etc.GlobalPxCluster.GetAliasOnNode(node)

	storageNames := etc.GlobalPxCluster.GetStorageNamesOnNode(node)
	cattles.ProcessStorage(machine, aliases, storageNames)
	machine = MapRootFs(machine, aliases, storageNames)
	machine["vmid"] = vmid
	setDefaultHostnameIfNeeded(machine, vmid)
}

func setDefaultHostnameIfNeeded(machine map[string]interface{}, vmid int) {
	if _, found := configmap.GetString(machine, "hostname"); !found {
		machine["hostname"] = "ct" + strconv.Itoa(vmid)
	}
}

func createContainer(node string, machine map[string]interface{}, vmid int) error {
	containerObject, err := CreatePxObjectCreateContainerRequest(machine)
	if err != nil {
		return fmt.Errorf("error creating container request: %w", err)
	}

	containerFlat, err := CreatePxFlatCreateContainerRequest(containerObject)
	if err != nil {
		return fmt.Errorf("error creating flat container request: %w", err)
	}

	//txt2,_ := json.Marshal(containerFlat)
	//fmt.Fprintf(os.Stderr, "  update: %s\n", txt2)

	res, err := api.CreateContainer(node, containerFlat)
	if err != nil {
		return fmt.Errorf("error creating container: %w", err)
	}

	queries.WaitForUPID(node, res.GetData())
	return nil
}

func updateCTConfig(machine map[string]interface{}, node string, vmid int, dryRun bool) error {
	ctConfig, err := queries.JSONGetCTConfig(node, int64(vmid))
	if err != nil {
		return fmt.Errorf("error getting CT config: %w", err)
	}

	ctConfigData, _ := configmap.GetMapEntry(ctConfig, "data")
	newConfig := createChanges(ctConfigData, machine)
	if len(newConfig) == 0 {
		return nil
	}

	return updateCT(machine, node, vmid, newConfig, dryRun)
}

func updateCT(machine map[string]interface{}, node string, vmid int, newConfig map[string]interface{}, dryRun bool) error {

	updateContainerConfigSyncRequest := pxapiflat.UpdateContainerConfigSyncRequest{}
	txt, _ := json.Marshal(newConfig)
	if err := json.Unmarshal(txt, &updateContainerConfigSyncRequest); err != nil {
		return fmt.Errorf("error unmarshalling update request: %w", err)
	}
	fmt.Fprintf(os.Stderr, "  update: %s\n", txt)
	if dryRun {
		return nil
	}
	_, err := api.UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
	if err != nil {
		return fmt.Errorf("error updating container config: %w", err)
	}
	return nil
}

// FIXME it should be unique over the vmid, which is not deteactable
// if you just look into the hashmap which is node/vmid nowadays
func generateClusterId(node string) (int64, error) {
	/*
		if len(etc.GlobalPxCluster.PxClients) <= 1 {
			return api.GetClusterNextId(node)
		}
	*/
	if !etc.GlobalPxCluster.IsVirtualCluster() {
		return api.GetClusterNextId(node)
	}
	pxClient, _ := etc.GlobalPxCluster.GetPxClient(node)
	formula := 8*100000 + pxClient.OrigIndex*1000

	for offset := 0; offset < 1000; offset++ {
		vmid := formula + offset
		/*
			if _, found := etc.GlobalPxCluster.UniqueMachines[vmid]; !found {
				return int64(vmid), nil
			}
		*/

		if !etc.GlobalPxCluster.Exists(node, int64(vmid)) {
			return int64(vmid), nil
		}
	}

	return 0, errors.New("unable to generate a unique cluster ID")
}

func createVMID(_type string, spec map[string]interface{}, node string) (int, bool, error) {
	createVM := false

	vmid := DetermineVmid(spec, _type)
	//fmt.Fprintf(os.Stderr, "STAGE 1: node=%v vmid=%v create=%v\n", node, vmid, createVM)
	if vmid != 0 {
		if !etc.GlobalPxCluster.Exists(node, int64(vmid)) {
			createVM = true
		}
	}
	//fmt.Fprintf(os.Stderr, "STAGE 2: node=%v vmid=%v create=%v\n", node, vmid, createVM)
	if vmid == 0 {
		createVM = true
		vmid64, err := generateClusterId(node)
		if err == nil {
			vmid = int(vmid64)
			//fmt.Fprintf(os.Stderr, "vmid64: %v\n", vmid64)
			//fmt.Fprintf(os.Stderr, "vmid: %v\n", vmid)
		}
	}

	//fmt.Fprintf(os.Stderr, "STAGE 3: node=%v vmid=%v create=%v\n", node, vmid, createVM)
	if vmid == 0 {
		err := fmt.Errorf("could not create vmid")
		return 0, false, err
	}

	return vmid, createVM, nil
}

func CreateVM(spec map[string]interface{}, vars map[string]interface{}, dump bool, node string, cattle string, dryRun bool, doUpdate bool) error {
	vmid, createVM, err := createVMID("qemu", spec, node)
	if err != nil {
		return fmt.Errorf("create VM ID error: %w", err)
	}

	if !doUpdate && !createVM {
		fmt.Fprintf(os.Stdout, "SKIP: VM %v %v (use --update if required)\n", vmid, node)
		return nil
	}

	logCreationOrUpdate(createVM, vmid, node)
	cluster, err := prepareVMForCreationOrUpdate(spec, node)
	if err != nil {
		return err
	}

	spec["vmid"] = vmid
	setDefaultVMNameIfNotSet(spec, vmid)

	var ignitionName string
	createIgnition := HasIgnition(spec)
	if createIgnition {
		ignitionName = HandleIgnition(spec, cluster, node)
	}

	if err := createOrUpdateVM(spec, vmid, createVM, dryRun, node); err != nil {
		return err
	}

	// also a vm creation needs adjustment if we need to resize the disks

	updateVMConfig(spec, vmid, node, dryRun)

	if createIgnition {
		DoIgnition(spec, cluster, node, createIgnition, vars, dump, dryRun, ignitionName)
	}
	return nil
}

func logCreationOrUpdate(createVM bool, vmid int, node string) {
	action := "UPDATE"
	if createVM {
		action = "CREATE"
	}
	fmt.Fprintf(os.Stdout, "%s VM %v %v\n", action, vmid, node)
}

func prepareVMForCreationOrUpdate(spec map[string]interface{}, node string) (map[string]interface{}, error) {
	aliases := etc.GlobalPxCluster.GetAliasOnNode(node)

	storageNames := etc.GlobalPxCluster.GetStorageNamesOnNode(node)
	//cluster, err := shared.PickCluster(etc.GlobalConfigData, ClusterName)
	clusterDatabase, err := etc.GlobalPxCluster.PickCluster(ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error picking cluster: %w", err)
	}
	cluster := clusterDatabase.GetCluster()

	SetImportFrom(cluster, spec)
	cattles.ProcessStorage(spec, aliases, storageNames)
	return cluster, nil
}

func setDefaultVMNameIfNotSet(spec map[string]interface{}, vmid int) {
	if _, found := configmap.GetString(spec, "name"); !found {
		spec["name"] = "vm" + strconv.Itoa(vmid)
	}
}

func createOrUpdateVM(spec map[string]interface{}, vmid int, createVM, dryRun bool, node string) error {
	vm, err := CreatePxObjectCreateVMRequest(spec)
	if err != nil {
		return fmt.Errorf("error creating VM request: %w", err)
	}

	vm2, err := CreatePxFlatCreateVMRequest(vm)
	if err != nil {
		return fmt.Errorf("error creating flat VM request: %w", err)
	}

	if createVM && !dryRun {
		api.CreateVM(node, vm2)
		queries.WaitForVMUnlock(node, int64(vmid))
	}
	return nil
}

func updateVMConfig(spec map[string]interface{}, vmid int, node string, dryRun bool) error {
	vmConfig, err := queries.JSONGetVMConfig(node, int64(vmid))
	if err != nil {
		return fmt.Errorf("error getting VM config: %w", err)
	}

	vmConfigData, _ := configmap.GetMapEntry(vmConfig, "data")

	if err := processStorageDrives(spec, vmConfigData, dryRun, node, vmid); err != nil {
		return err
	}

	newConfig := createChanges(vmConfigData, spec)
	if len(newConfig) == 0 {
		return nil
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

func calculateNewDriveSizeIncrease(spec map[string]interface{}, vmConfigData map[string]interface{}, node string, vmid int, storageDrive string) (int64, error) {
	storageData, ok := configmap.GetMapEntry(spec, storageDrive)
	if !ok {
		return 0, nil // Skip if storage data is not found
	}

	size, ok := configmap.GetString(storageData, "size")
	if !ok {
		return 0, nil // Skip if size is not specified
	}

	desiredSizeInBytes, ok := shared.CalculateSizeInBytes(size)
	if !ok {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	driveEntry, ok := configmap.GetString(vmConfigData, storageDrive)
	if !ok {
		return 0, nil // Skip if drive entry is not found
	}
	currentSizeInBytes, ok := extractCurrentSize(driveEntry)

	if !ok || currentSizeInBytes == desiredSizeInBytes {
		return 0, nil // Skip if current size is invalid or equal to desired size
	}

	if currentSizeInBytes > desiredSizeInBytes {
		err := fmt.Errorf("drive %s size configuration is smaller than current VM: %s < %s", storageDrive, shared.ToSizeString(desiredSizeInBytes), shared.ToSizeString(currentSizeInBytes))
		return 0, err
	}
	delta_size := desiredSizeInBytes - currentSizeInBytes
	fmt.Fprintf(os.Stdout, "  %s: increase size by %s to: %s\n", storageDrive, shared.ToSizeString(delta_size), shared.ToSizeString(desiredSizeInBytes))
	return delta_size, nil
}

func handleDriveSize(spec map[string]interface{}, vmConfigData map[string]interface{}, dryRun bool, node string, vmid int, storageDrive string) error {
	delta_size, _ := calculateNewDriveSizeIncrease(spec, vmConfigData, node, vmid, storageDrive)

	if delta_size == 0 {
		return nil
	}
	if dryRun {
		return nil
	}

	return queries.ResizeVMDisk(node, vmid, storageDrive, delta_size)

}

func extractCurrentSize(driveEntry string) (int64, bool) {
	sizeStr, _ := GetConfigOptionAndRemaining(driveEntry, "size")
	parts := strings.SplitN(sizeStr, "=", 2)
	if len(parts) < 2 {
		return 0, false
	}
	return shared.CalculateSizeInBytes(parts[1])
}

func updateVMConfiguration(vmConfigData, spec map[string]interface{}, dryRun bool, node string, vmid int) error {
	newConfig := createChanges(vmConfigData, spec)
	if len(newConfig) == 0 {
		return nil
	}
	/*
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
	*/

	return shared.UpdateVMConfiguration(node, int64(vmid), newConfig, dryRun)
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
	result2, _ := etc.GlobalPxCluster.GetPxClient(createVirtualmachineOptions.Node)
	vars := result2.Vars
	return CreateVM(result, vars, createVirtualmachineOptions.Dump, createVirtualmachineOptions.Node, createVirtualmachineOptions.Cattle, createVirtualmachineOptions.DryRun, createVirtualmachineOptions.Update)
}
