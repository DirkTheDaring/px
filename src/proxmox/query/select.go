package query

func FromSelectJoin(aTable []map[string]interface{}, bTable []map[string]interface{}, aColumns []string, bColumns []string, onColumns [][]string) []map[string]interface{} {

	list := []map[string]interface{}{}
	for _, aObject := range aTable {
		for _, bObject := range bTable {
			match := true
			// These keys need to match for the join
			var aColumnName string
			var bColumnName string
			for _, onColumn := range onColumns {
				if len(onColumn) == 0 {
					break
				}
				if len(onColumn) == 1 {
					aColumnName = onColumn[0]
					bColumnName = aColumnName
				} else {
					aColumnName = onColumn[0]
					bColumnName = onColumn[1]
				}
				if aObject[aColumnName] != bObject[bColumnName] {
					match = false
					break
				}
			}
			if !match {
				continue
			}
			row := map[string]interface{}{}
			if len(aColumns) != 0 {
				if aColumns[0] == "*" {
					row = aObject
				} else {
					for _, aColumn := range aColumns {
						row[aColumn] = aObject[aColumn]
					}
				}
			}
			if len(bColumns) != 0 {
				if bColumns[0] == "*" {
					for bKey, bValue := range bObject {
						_, found := row[bKey]
						if found {
							continue
						}
						row[bKey] = bValue
					}
				} else {
					for _, bColumn := range bColumns {
						row[bColumn] = aObject[bColumn]
					}
				}
			}
			list = append(list, row)
		}
	}
	return list
}
func FromSelectLeftJoin(aTable []map[string]interface{}, bTable []map[string]interface{}, aColumns []string, bColumns []string, onColumns [][]string) []map[string]interface{} {

	list := []map[string]interface{}{}
	for _, aObject := range aTable {
		leftJoinMatch := false
		for _, bObject := range bTable {
			match := true
			// These keys need to match for the join
			var aColumnName string
			var bColumnName string
			for _, onColumn := range onColumns {
				if len(onColumn) == 0 {
					break
				}
				if len(onColumn) == 1 {
					aColumnName = onColumn[0]
					bColumnName = aColumnName
				} else {
					aColumnName = onColumn[0]
					bColumnName = onColumn[1]
				}
				if aObject[aColumnName] != bObject[bColumnName] {
					match = false
					break
				}
			}
			if !match {
				continue
			}
			leftJoinMatch = true
			row := map[string]interface{}{}
			if len(aColumns) != 0 {
				if aColumns[0] == "*" {
					row = aObject
				} else {
					for _, aColumn := range aColumns {
						row[aColumn] = aObject[aColumn]
					}
				}
			}
			if len(bColumns) != 0 {
				if bColumns[0] == "*" {
					for bKey, bValue := range bObject {
						_, found := row[bKey]
						if found {
							continue
						}
						row[bKey] = bValue
					}
				} else {
					for _, bColumn := range bColumns {
						row[bColumn] = aObject[bColumn]
					}
				}
			}
			list = append(list, row)
		}
		if leftJoinMatch == false {
			list = append(list, aObject)
		}
	}
	return list
}
