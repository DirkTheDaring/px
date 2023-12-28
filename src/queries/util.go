package queries

import (
	"fmt"
	"os"
	"px/configmap"
	"px/etc"
	"strings"
)

func getSubset(filterString string, machine_type string) []map[string]interface{} {
	machines := etc.GlobalPxCluster.Machines

	filterColumn := "name"

	var newlist = []map[string]interface{}{}

	for _, machine := range machines {
		value, ok := machine[filterColumn].(string)
		if !ok {
			continue
		}
		if !strings.HasPrefix(value, filterString) {
			continue
		}
		_type, ok := configmap.GetString(machine, "type")
		if !ok {
			continue
		}
		if _type != machine_type {
			continue
		}
		newlist = append(newlist, machine)
	}

	return newlist
}

func Apply(match string, settings []string) {

	if match == "" {
		fmt.Fprintf(os.Stdout, "you must set --match not to an empty string.\n")
		os.Exit(1)
	}
	if len(settings) == 0 {
		fmt.Fprintf(os.Stdout, "no settings given.\n")
		os.Exit(1)
	}

	qemu_list := getSubset(match, "qemu")
	if len(qemu_list) != 0 {
		ApplyQemu(qemu_list, settings)
	}

	lxc_list := getSubset(match, "lxc")
	if len(lxc_list) != 0 {
		ApplyLxc(lxc_list, settings)
	}
	os.Exit(0)
}
