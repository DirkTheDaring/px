package shared

import (
	"os"
	"px/etc"
	"strings"
)

// matchColumn checks if the given item matches the filter criteria.
func matchColumn(item map[string]interface{}, filterColumns, filterStrings []string) bool {
	for i, filterColumn := range filterColumns {
		value, ok := item[filterColumn].(string)
		if !ok || !strings.HasPrefix(value, filterStrings[i]) {
			return false
		}
	}
	return true
}

// filterStringColumns filters the outputs based on the provided column and string filters.
func filterStringColumns(outputs []map[string]interface{}, filterColumns, filterStrings []string) []map[string]interface{} {
	var filteredOutputs []map[string]interface{}
	for _, output := range outputs {
		if matchColumn(output, filterColumns, filterStrings) {
			filteredOutputs = append(filteredOutputs, output)
		}
	}
	return filteredOutputs
}

// selectMachines filters machines based on the provided match criteria.
func SelectMachines(machines []map[string]interface{}, match string) []map[string]interface{} {

	parts := strings.SplitN(match, ":", 2)

	switch len(parts) {
	case 1:
		return filterStringColumns(machines, []string{"name"}, []string{parts[0]})
	case 2:
		return filterStringColumns(machines, []string{"name", "node"}, []string{parts[1], parts[0]})
	default:
		return machines
	}
}

// Status displays the status of machines filtered by the match criteria.
func Status(match string) {
	machines := etc.GlobalPxCluster.GetMachines()
	machines = SelectMachines(machines, match)
	sortedMachines := StringSortMachines(machines, []string{"name"}, []bool{true})

	headers := []string{"name", "type", "node", "vmid", "status"}
	alignmentsRight := []string{"vmid"}

	//RenderOnConsoleNew(sortedMachines, headers, "name", "", alignmentsRight)
	RenderOnConsoleNew(sortedMachines, headers, alignmentsRight)
	os.Exit(0)
}
