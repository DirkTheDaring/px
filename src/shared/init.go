package shared

import (
	"px/configmap"
	"regexp"
)

func GetPriorityMatch(prioritylist []string, list []string) string {
	for _, priorityItem := range prioritylist {
		for _, item := range list {
			match, _ := regexp.MatchString(priorityItem, item)
			if match {
				return item
			}
		}
	}
	return ""
}

func FilterStorageByNodeName(storageSlice []map[string]interface{}, nodeName string) []string {
	localStorage := []string{}
	for _, storageItem := range storageSlice {
		storage, _ := configmap.GetString(storageItem, "storage")
		storageNodes, found := configmap.GetString(storageItem, "nodes")
		if found && storageNodes != nodeName {
			//fmt.Fprintf(os.Stderr, "storageNodes: %T %v %v\n", storageNodes, storageNodes, storage)
			continue
		}
		localStorage = append(localStorage, storage)
	}
	return localStorage
}
