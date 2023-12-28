package shared

// FIXME
// (1) Validate type by using reflection
// (2) report non-matching type and skip
// (3) report on existing key and skip

/*
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
*/
