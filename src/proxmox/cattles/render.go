package cattles

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"px/configmap"
	"px/proxmox/aliases"
	"px/proxmox/query"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

func Dump(files embed.FS, filename string) {
	data := map[string]interface{}{}
	configmap.LoadEmbeddedYamlFile(data, files, filename)
	result := configmap.DataToJSON(data)
	fmt.Fprintf(os.Stderr, "%s\n", result)
}

func LoadYamlFile(data map[string]interface{}, filename string) bool {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return false
	}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		return false
	}
	return true
}

func GetenvWithDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

func PrepareEmbeddedTemplate(files embed.FS, fileglob []string, funcMap template.FuncMap) (*template.Template, error) {

	filename := fileglob[0]
	basename := path.Base(filename)

	t := template.New(basename).Funcs(funcMap)
	tmpl, err := t.ParseFS(files, fileglob...)
	return tmpl, err
}

func ProcessStorage(data map[string]interface{}, aliases map[string]string, storageNames []string) map[string]interface{} {
	//fmt.Fprintf(os.Stderr, "Aliases %v\n", aliases)

	list := configmap.SelectKeys("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$", data)
	for _, storageDrive := range list {
		storageData, _ := configmap.GetMapEntry(data, storageDrive)
		file, ok := configmap.GetString(storageData, "file")
		if !ok {
			continue
		}
		slice := strings.Split(file, ":")

		if len(slice) != 2 {
			continue
		}

		storageName := slice[0]
		var newStorageName string

		alias, ok := aliases[storageName]
		if ok {
			newStorageName = alias
			if !query.In(storageNames, alias) {
				fmt.Fprintf(os.Stderr, "  storageName %s (alias %s) not found: %s\n",  storageName, alias, file)
				continue
			}

		} else {
			newStorageName = storageName
			if !query.In(storageNames, storageName) {
				fmt.Fprintf(os.Stderr, "  storageName %s not found: %s\n",  alias, file)
				continue
			}
		}

		newFile := newStorageName + ":" + slice[1]

		if newFile != file {
			fmt.Fprintf(os.Stderr, "  %s: %s (%s: %s)\n", storageDrive, newFile, storageDrive, file)
		}
		storageData["file"] = newFile


		// Same semantic as proxmox, import-from is only parsed if there is "0" string
		if slice[1] != "0" {
			continue
		}

		data[storageDrive] = storageData

	}

	return data
}

func CreateCattle(typeName string, cattleName string, data map[string]interface{}) map[string]interface{} {

	internalDir := "files"
	externalDir := GetenvWithDefault("PX_CATTLE_DIR", internalDir)
	filename := typeName + ".yaml"

	configData := map[string]interface{}{}
	// the next two lines merge internal config and external config, if it exists
	err := configmap.LoadEmbeddedYamlFile(configData, files, internalDir+"/types/"+filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateCattler() could not load yaml file: %v\n", err)
	}
	ok := LoadYamlFile(configData, externalDir+"/types/"+filename)

	cattleMap, ok := configmap.GetMapEntry(configData, cattleName)
	if !ok {
		return configData
	}

	filename = configmap.GetStringValueWithDefault(cattleMap, "setting", "small.yaml")

	settingsData := map[string]interface{}{}
	ok = LoadYamlFile(settingsData, externalDir+"/settings/"+filename)
	if !ok {
		err = configmap.LoadEmbeddedYamlFile(settingsData, files, internalDir+"/settings/"+filename)
	}
	settingsData = configmap.MergeMapRecursive(settingsData, data)
	filename = configmap.GetStringValueWithDefault(cattleMap, "template", typeName+"-cattle.gotext")

	// Try to load template from externalDir first, then fallback to internal
	buffer := bytes.Buffer{}
	outputStream := &buffer
	// This either a windows or linux format but PrepareFileTemplate doesn't like windows style pathes
	filePath := externalDir + "/templates/" + filename
	tmpl, err := configmap.PrepareFileTemplate(filePath)
	// fallback to internal dir
	if err != nil {
		//fmt.Fprintf(os.Stderr, "CreateCattle(): filePath: %s %v\n", filePath, err)
		// for internal use UNIX filepath
		filePath = internalDir + "/templates/" + filename
		tmpl, err = PrepareEmbeddedTemplate(files, []string{filePath}, nil)
	}
	err = tmpl.Execute(outputStream, settingsData)

	if err != nil {
		return settingsData
	}
	/*
		if err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", buffer.String())
		}
	*/
	data = map[string]interface{}{}
	err = yaml.Unmarshal(buffer.Bytes(), data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
	}
	return data
}

func EmbeddedFileExists(embedFS embed.FS, name string) bool {
	file, err := embedFS.Open(name)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

func RenderArgString(input string, node string, name string) string {
	//input = "{{.Count}} items are made of {{.Material}}"

	input = "-fw_cfg name=opt/com.coreos/config,file={{storage}}"

	buffer := bytes.Buffer{}
	outStream := &buffer

	tmpl, err := template.New("test").Funcs(map[string]interface{}{
		"storage": func(storageName string) string {
			result := node + ":" + storageName + "//" + name
			return result
		}}).Parse(input)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ""
	}
	data := map[string]interface{}{}
	data["name"] = name
	data["default"] = "ignition"

	err = tmpl.Execute(outStream, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ""
	}
	return outStream.String()
}
func Configure(vmData map[string]interface{}, node string, cattleOverrideData map[string]interface{}) map[string]interface{} {

	// Phase 1 - Cattle Configuration
	// Phase 1 load  data which can override cattle setti^ngs

	cattleData := CreateCattle("vm", "small", cattleOverrideData)

	// Phase 2 - Cattle Customization
	// now merge cattle with configuration of user

	data := configmap.MergeMapRecursive(cattleData, vmData)

	// Phase 3 - put Cattle on farm  :)  (host system)
	// FIXME - currently decision is taken by the user

	// Phase 4 -  Adapt Local conditions
	// now get local-lvm or local-zvm and whater ever storage is configured and map the alias
	// Also apply the naming policy
	//aliases := GetAliases()
	// available storages on a specific host (including nfs shares)
	// 	list := configmap.SelectKeys("^(net|ipconfig|ip6config)[0-9]+$", data)
	available_storages := []string{"local-lvm"}
	list := configmap.SelectKeys("^(virtio|scsi|ide|sata|efidisk|tpmstate)[0-9]+$", data)

	for _, item := range list {
		fmt.Fprintf(os.Stderr, "%s\n", item)
		entry, _ := configmap.GetMapEntry(data, item)
		fmt.Fprintf(os.Stderr, "%v\n", entry)

		file := configmap.GetStringValueWithDefault(entry, "file", "")
		if file == "" {
			continue
		}
		fmt.Fprintf(os.Stderr, "file: %s\n", file)
		pos := strings.Index(file, ":")
		if pos == -1 {
			continue
		}
		alias := file[:pos]
		postfix := file[pos:]
		fmt.Fprintf(os.Stderr, "alias: %s postfix: %s\n", alias, postfix)
		replacement := aliases.LookupAlias(alias, available_storages)
		entry["file"] = replacement + postfix

		importFrom := configmap.GetStringValueWithDefault(entry, "import-from", "")
		if importFrom == "" {
			continue
		}
		fmt.Fprintf(os.Stderr, "importFrom: %s\n", importFrom)
		if importFrom == "LATEST" {
			entry["import-from"] = "fedora-fcos-latest.qcow2"
		}
	}

	// Phase 6
	// Now the strategy for vmid, which is when not existing determined by a policy

	// Policy 1 :  Derive VMID from IP
	// Policy 2:  Don't care about VMID, let system decide, but make name unique
	// Policy :  Handle existing VMID

	vmid := configmap.GetIntWithDefault(data, "vmid", 0)
	if IsValidVmid(vmid) {
		return data
	}
	ipv4, ok := GetIpv4(data, "ipconfig0")
	fmt.Fprintf(os.Stderr, "ipconfig: %v %v\n", ipv4, ok)
	if !ok {
		goto next_try
	}
	// 2 = last 2  number of ipv4
	vmid = ToVmid(ipv4, 2)
	fmt.Fprintf(os.Stderr, "vmid: %v\n", vmid)
	if vmid > 0 {
		data["vmid"] = vmid
		return data
	}

next_try:
	guestname, ok := configmap.GetString(data, "name")
	if !ok {
		fmt.Fprintf(os.Stderr, "no name found\n")
	} else {
		fmt.Fprintf(os.Stderr, "lookup: %s\n", guestname)
	}
	return data
}

func GetIpv4(data map[string]interface{}, ipconfigName string) (string, bool) {
	ipconfigMap, ok := configmap.GetMapEntry(data, ipconfigName)
	fmt.Fprintf(os.Stderr, "ipconfig: %v %v\n", ipconfigMap, ok)

	ip, ok := configmap.GetString(ipconfigMap, "ip")
	if !ok {
		return "", false
	}
	pos := strings.Index(ip, "/")
	if pos == -1 {
		return "", false
	}
	ipv4 := ip[:pos]
	ipv4addr := net.ParseIP(ipv4)
	if ipv4addr == nil {
		return ipv4, false
	}
	fmt.Fprintf(os.Stderr, "ipv4: %v\n", ipv4)
	return ipv4, true
}

func ToVmid(ipv4 string, n int) int {
	fmt.Fprintf(os.Stderr, "ipv4: %v\n", ipv4)

	array := strings.Split(ipv4, ".")
	if len(array) != 4 {
		return 0
	}
	intArray := [4]int{}
	for i, item := range array {
		intVar, err := strconv.Atoi(item)
		if err != nil {
			return 0
		}
		intArray[i] = intVar
	}

	result := 0
	if n == 1 {
		result = intArray[3]
	}
	if n == 2 {
		result = intArray[2]*1000 + intArray[3]
	}
	if n == 3 {
		result = intArray[1]*1000000 + intArray[2]*1000 + intArray[3]
	}
	if !IsValidVmid(result) {
		return 0
	}
	return result
}
func IsValidVmid(vmid int) bool {
	// According to proxmox gui min is 1 and max is 999999999 < 2,147,483,647 which is the maximum signed 32bit integer
	// MAGIC CONSTANT
	if vmid > 0 && vmid < 999999999 {
		return true
	}
	return false
}
