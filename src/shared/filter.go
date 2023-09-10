package shared

import (
	"strings"
)

func FilterStringColumn(outputs []map[string]interface{}, filterColumn string, filterString string) []map[string]interface{} {

	list := []map[string]interface{}{}

	for _, output := range outputs {
		value, _ := output[filterColumn].(string)
		if !strings.HasPrefix(value, filterString) {
			continue
		}
		list = append(list, output)
	}
	return list
}

func FilterStringColumns(outputs []map[string]interface{}, filterColumns []string, filterStrings []string) []map[string]interface{} {
	list := []map[string]interface{}{}

	for _, output := range outputs {
		match := true
		for i, filterColumn := range filterColumns {
			value, _ := output[filterColumn].(string)
			if strings.HasPrefix(value, filterStrings[i]) {
				continue
			}
			match = false
			break
		}
		if !match {
			continue
		}

		list = append(list, output)
	}
	return list
}
