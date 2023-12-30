package shared

import (
	"errors"
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
	"px/proxmox"
	"px/proxmox/query"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func GetClusterIndex(configData map[string]interface{}, name string) (int, error) {
	clusters, err := configmap.GetInterfaceSliceValue(configData, "clusters")
	if err != nil {
		return -1, err
	}

	max := len(clusters)
	if max == 0 {
		return -1, errors.New("no clusers defined.")
	}

	number64, err := strconv.ParseInt(name, 10, 32)
	if err == nil {
		pos := int(number64)
		if pos >= 0 && pos < max {
			return pos, nil
		} else {
			return -1, errors.New("cluster index not found: " + strconv.Itoa(pos))
		}
	}

	for i, cluster := range clusters {
		clusterName, ok := cluster["name"]
		// if cluster has no name
		if !ok {
			continue
		}
		if clusterName == name {
			return i, nil
		}
	}
	return -1, errors.New("cluster name not found: " + name)
}
func PickCluster(configData map[string]interface{}, name string) (map[string]interface{}, error) {

	clusters, err := configmap.GetInterfaceSliceValue(configData, "clusters")
	if err != nil {
		return nil, err
	}

	max := len(clusters)
	if max == 0 {
		return nil, errors.New("no clusters defined.")
	}

	number64, err := strconv.ParseInt(name, 10, 32)
	if err == nil {
		pos := int(number64)
		if pos >= 0 && pos < max {
			return clusters[pos], nil
		} else {
			return nil, errors.New("cluster index not found: " + strconv.Itoa(pos))
		}
	}

	for i, cluster := range clusters {
		clusterName, ok := cluster["name"]
		// if cluster has no name
		if !ok {
			continue
		}
		if clusterName == name {
			return clusters[i], nil
		}

	}
	return nil, errors.New("cluster name not found: " + name)
}

func StringSortMachines(machines []map[string]interface{}, fieldNames []string, ascending []bool) []map[string]interface{} {
	//fmt.Fprintf(os.Stderr, "StringSortMachines machines: %v fieldNames: %v\n", machines, fieldNames)

	count := 0
	newMachines := []map[string]interface{}{}

	for _, item := range machines {
		found := true
		for k := 0; k < len(fieldNames); k++ {
			if item[fieldNames[k]] == nil {
				found = false
				break
			}
		}
		if !found {
			continue
		}
		count++
		//fmt.Fprintf(os.Stderr, "StringSortMachines() k: %v\n", item[fieldNames[0]])
		newMachines = append(newMachines, item)
	}

	sort.Slice(newMachines, func(i, j int) bool {
		k := 0
		/*
			for k = 0; k < count-1; k++ {
				//fmt.Fprintf(os.Stderr, "StringSortMachines() k: %v\n", k)

				a := newMachines[i][fieldNames[k]].(string)
				b := newMachines[j][fieldNames[k]].(string)
				if a != b {
					break
				}
			}
		*/

		if ascending[k] == true {
			//fmt.Fprintf(os.Stderr, "StringSortMachines() k: %T\n", machines[i][fieldNames[k]])
			a := newMachines[i][fieldNames[k]].(string)
			//fmt.Fprintf(os.Stderr, "StringSortMachines() k: %T\n", machines[j][fieldNames[k]])
			b := newMachines[j][fieldNames[k]].(string)
			return a < b
		} else {
			a := newMachines[i][fieldNames[k]].(string)
			b := newMachines[j][fieldNames[k]].(string)
			return a > b
		}
	})
	return newMachines
}
func JoinClusterAndSelector(pxCluster etc.PxCluster, selectors map[string]interface{}) []map[string]interface{} {
	storageNames := pxCluster.GetStorageNames()
	storageContent := pxCluster.GetStorageContent()
	//fmt.Fprintf(os.Stdout, "selectors: %v\n", selectors)

	newStorageContent := []map[string]interface{}{}
	for storage, values := range selectors {
		//fmt.Fprintf(os.Stdout, "selectors storage: %v\n", storage)
		if !query.In(storageNames, storage) {
			fmt.Fprintf(os.Stderr, "storage does not exist: %v\n", storage)
			continue
		}
		//fmt.Fprintf(os.Stderr, "storageNames:  %v\n", storageNames)
		//fmt.Fprintf(os.Stdout, "selectors value: %T %v\n", value, value)
		selectorMaps, ok := values.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stdout, "Failed to cast: %T\n", values)
			continue
		}

		for key, matchValue := range selectorMaps {
			match, _ := matchValue.(string)

			//fmt.Fprintf(os.Stdout, "valid selectors: %v %v %v\n", key, storage, match)
			r, _ := regexp.Compile(match)

			for _, item := range storageContent {
				storage2 := item["storage"].(string)
				if storage2 != storage {
					continue
				}
				volid := item["volid"].(string)
				array := strings.Split(volid, "/")
				str := array[len(array)-1]
				if r.MatchString(str) {
					//fmt.Fprintf(os.Stdout, "match: %v %v %v\n", key, str, item["node"])
					//item["label"] = storage + ":" + key
					item["label"] = key
					newStorageContent = append(newStorageContent, item)
				}
			}
		}
	}
	sort.Slice(newStorageContent, func(i, j int) bool {

		if newStorageContent[i]["label"].(string) == newStorageContent[j]["label"].(string) {
			if newStorageContent[i]["node"].(string) == newStorageContent[j]["node"].(string) {
				if newStorageContent[i]["volid"].(string) > newStorageContent[j]["volid"].(string) {
					return true
				}
			}
		}
		if newStorageContent[i]["label"].(string) == newStorageContent[j]["label"].(string) {
			if newStorageContent[i]["node"].(string) < newStorageContent[j]["node"].(string) {
				return true
			}
		}
		if newStorageContent[i]["label"].(string) < newStorageContent[j]["label"].(string) {
			return true
		}

		return false
	})
	return newStorageContent
}
func ExtractLatest(pxCluster etc.PxCluster, newStorageContent []map[string]interface{}) []map[string]interface{} {

	nodeLength := len(pxCluster.Nodes)
	nodeLookup := map[string]int{}

	for i, name := range pxCluster.Nodes {
		nodeLookup[name] = i
	}

	volidMap := map[string][]bool{}

	// the boolArray contains on which host the volid exists
	// if the boolArray only contains true, then this volid exists on all
	for _, item := range newStorageContent {
		volid := item["volid"].(string)
		if boolArray, ok := volidMap[volid]; ok {
			//fmt.Printf("exists in the map: %s\n", volid)
			index, found := nodeLookup[item["node"].(string)]
			if !found {
				continue
			}
			boolArray[index] = true
			volidMap[volid] = boolArray
		} else {
			//fmt.Println("foo does not exist in the map!")
			nodeName := item["node"].(string)
			index, found := nodeLookup[nodeName]
			if !found {
				if nodeName == "*" || nodeName == "" {
					boolArray := make([]bool, nodeLength)
					for i := 0; i < nodeLength; i++ {
						boolArray[i] = true
					}
					volidMap[volid] = boolArray
					continue
				}
				//fmt.Printf("node not found: %s\n", nodeName)
				continue
			}
			boolArray := make([]bool, nodeLength)
			for i := 0; i < nodeLength; i++ {
				boolArray[i] = false
			}
			boolArray[index] = true
			volidMap[volid] = boolArray
		}
	}

	// Logical And summary of all bools
	volidMapSummary := map[string]bool{}
	for key, boolArray := range volidMap {
		//fmt.Printf("pre: %s\t%v\n", key, boolArray)

		boolValue := boolArray[0]
		for i := 1; i < len(boolArray); i++ {
			boolValue = boolValue && boolArray[i]
		}
		volidMapSummary[key] = boolValue
		//fmt.Printf("result: %t\t%s\n", boolValue, key)
		//fmt.Println("foo does not exist in the map!")
	}

	newStorageContent20 := []map[string]interface{}{}
	wasProcessed := map[string]bool{}

	for _, item := range newStorageContent {
		volid := item["volid"].(string)
		isAvailableEverywhere, ok := volidMapSummary[volid]
		if !ok {
			continue
		}
		if !isAvailableEverywhere {
			continue
		}
		_, ok = wasProcessed[volid]
		if ok {
			continue
		}
		item["node"] = "*"
		newStorageContent20 = append(newStorageContent20, item)
		wasProcessed[volid] = true
	}

	/*
		for _, item := range newStorageContent20 {
			volid := item["volid"].(string)
			fmt.Printf("result: %s\n", volid)
		}
	*/

	newStorageContent2 := []map[string]interface{}{}
	prevNode := ""
	prevLabel := ""

	// loop through storage context
	for _, item := range newStorageContent20 {
		node := item["node"].(string)
		label := item["label"].(string)
		if prevNode != node {
			item["label"] = item["storage"].(string) + ":" + label
			newStorageContent2 = append(newStorageContent2, item)
			goto forward1
		}
		if prevNode == node && prevLabel != label {
			item["label"] = item["storage"].(string) + ":" + label
			newStorageContent2 = append(newStorageContent2, item)
			goto forward2
		}
	forward1:
		prevNode = node
	forward2:
		prevLabel = label
	}

	// Beauty: Remove duplicate labels, but only those which
	// reference the same amount as len(pxCluster.Nodes), which should
	// mean, that they are available on every node.
	// Not sure if this heuristic will bite me.

	sort.Slice(newStorageContent2, func(i, j int) bool {

		if newStorageContent2[i]["label"].(string) == newStorageContent2[j]["label"].(string) {
			if newStorageContent2[i]["node"].(string) == newStorageContent2[j]["node"].(string) {
				if newStorageContent2[i]["volid"].(string) > newStorageContent2[j]["volid"].(string) {
					return true
				}
			}
		}
		if newStorageContent2[i]["label"].(string) == newStorageContent2[j]["label"].(string) {
			if newStorageContent2[i]["node"].(string) < newStorageContent2[j]["node"].(string) {
				return true
			}
		}
		if newStorageContent2[i]["label"].(string) < newStorageContent2[j]["label"].(string) {
			return true
		}

		return false
	})

	return newStorageContent2

}
func StringFilter(mapList []map[string]interface{}, key string, value string) []map[string]interface{} {
	//fmt.Fprintf(os.Stderr, "StringFilter %v\n", len(mapList))
	list := []map[string]interface{}{}
	for _, mapItem := range mapList {
		/*
			_, ok := mapItem["maxcpu"]
			if ok {
				//os.Exit(1)
				fmt.Fprintf(os.Stderr, "--- ERROR ---\n")
			}
			*&
		*/
		tmp, ok := configmap.GetString(mapItem, key)
		if !ok {
			fmt.Fprintf(os.Stderr, "StringFilter() key not found: %v\n", key)
			proxmox.DumpJson(mapItem)
		}
		if tmp == value {
			list = append(list, mapItem)
		}
	}
	return list

}
func GetMachinesByName(name string) []map[string]interface{} {
	return StringFilter(etc.GlobalPxCluster.GetMachines(), "name", name)
}

func GetVmidByAttribute(machine map[string]interface{}, attribute string) (int, error) {

	machineNameStr, ok := configmap.GetString(machine, attribute)
	if !ok {
		return 0, errors.New(fmt.Sprintf("GetVmidByAttribute(): attribute not found: ", attribute))
	}
	//fmt.Fprintf(os.Stderr, "STEP 2\n")
	machines := GetMachinesByName(machineNameStr)
	//fmt.Fprintf(os.Stderr, "STEP 3\n")
	if len(machines) > 1 {
		//fmt.Fprintf(os.Stderr, "machineName is not unique: %s\n", machineNameStr)
		return 0, errors.New(fmt.Sprintf("machineName is not unique: %s", machineNameStr))
		//fmt.Fprintf(os.Stderr, "machineName is not unique: %s\n", machineNameStr)
	}
	//fmt.Fprintf(os.Stderr, "STEP 4\n")
	if len(machines) == 0 {
		return 0, errors.New("no machine with this name found:" + machineNameStr)
	}
	//fmt.Fprintf(os.Stderr, "STEP 5\n")
	vmid, ok := machines[0]["vmid"]
	//fmt.Fprintf(os.Stderr, "STEP 6\n")
	if !ok {
		fmt.Fprintf(os.Stderr, "from machine 0 = %v\n", vmid)
		return 0, errors.New("machine has no vmid set: " + machineNameStr)
	}
	//fmt.Fprintf(os.Stderr, "STEP 7\n")
	vmidFloat64, ok := vmid.(float64)
	if !ok {
		return 0, errors.New("could not convert vmid to int")
	}
	valueInt := int(vmidFloat64)
	return valueInt, nil

}
func GetIpv4Address(machine map[string]interface{}, _type string) (string, bool) {
	var name string
	if _type == "lxc" {
		name = "net0"
	} else {
		name = "ipconfig0"
	}
	entry, ok := configmap.GetMapEntry(machine, name)
	if !ok {
		return "", false
	}
	ip, ok := configmap.GetString(entry, "ip")
	if !ok {
		return "", false
	}
	return ip, true
}

func DeriveVmidFromIp4Address(ip string) (int, error) {
	// Split the IP address and subnet mask
	parts := strings.Split(ip, "/")
	if len(parts) != 2 {
		return 0, errors.New("invalid IP format: expected IPv4 and subnet mask")
	}

	// Extract the IPv4 part
	ipv4 := parts[0]

	// Split the IPv4 address into its components
	octets := strings.Split(ipv4, ".")
	if len(octets) != 4 {
		return 0, errors.New("invalid IPv4 format: expected four octets")
	}

	// Convert the third and fourth octets to integers
	a, err := strconv.Atoi(octets[2])
	if err != nil {
		return 0, errors.New("invalid third octet: not an integer")
	}

	b, err := strconv.Atoi(octets[3])
	if err != nil {
		return 0, errors.New("invalid fourth octet: not an integer")
	}

	// Calculate the new VM ID
	newVmid := a*1000 + b
	return newVmid, nil
}

func CalculateSizeInBytes(sizestring string) (int64, bool) {

	if sizestring == "" {
		return 0, false
	}
	c := sizestring[len(sizestring)-1:]
	index := strings.Index("0123456789KMGT", c)
	if index == -1 {
		return 0, false
	}
	sizestring = sizestring[:len(sizestring)-1]
	n, err := strconv.ParseInt(sizestring, 10, 64)
	if err != nil {
		return 0, false
	}

	var power int
	if index < 10 {
		power = 0
	} else {
		power = index - 9
	}

	for i := 0; i < power; i++ {
		n = n * 1024
	}
	//fmt.Fprintf(os.Stderr, "calculateSizeInBytes(): %s %v %v\n", c, power, n)
	return n, true
}

func ToSizeString(size int64) string {

	//fmt.Fprintf(os.Stderr, "ToSizeString(%v)\n", size)
	if size == 0 {
		return "0"
	}

	var i int64
	for i = 4; size >= 0; i-- {
		if size >= Pow(1024, i) {
			break
		}
	}
	if i == 0 {
		return strconv.FormatInt(size, 10)
	}

	sizePostfix := " KMGT"
	result := size / Pow(1024, i)

	str := strconv.FormatInt(result, 10)
	str = str + string(sizePostfix[i])
	return str
}
