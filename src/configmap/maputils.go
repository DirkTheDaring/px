package configmap

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

// Copy Map

func CopyMapRecursive(src map[string]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	/*
	   for key, value := range src {
	           if reflect.TypeOf(value).Kind() == reflect.Map {
	                   value2, ok := value.(map[string]interface{})
	                   if ok {
	                           newMap[key] = CopyMapRecursive(value2)
	                   }
	           } else {
	                   newMap[key] = value
	           }
	   }
	*/
	for key, value := range src {
		mapValue, hasMap := value.(map[string]interface{})
		if hasMap {
			newMap[key] = CopyMapRecursive(mapValue)
		} else {
			newMap[key] = value
		}
	}
	return newMap
}

// Merge two maps into a new map
// slices are not copied but referenced - but as we use it "readonly" that does not matter
func MergeMapRecursive(dst map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	newMap := CopyMapRecursive(dst)
	for key, value := range src {
		// only keys of type string, don't bother with anythign else
		if reflect.TypeOf(key).Kind() != reflect.String {
			continue
		}
		mapValue, hasMap := value.(map[string]interface{})
		if !hasMap {
			newMap[key] = value
			continue
		}
		tmp, dstHasValue := dst[key]
		if dstHasValue {
			dstMapValue, dstValueIsMap := tmp.(map[string]interface{})
			if dstValueIsMap {
				newMap[key] = MergeMapRecursive(dstMapValue, mapValue)
				continue
			}
		}
		newMap[key] = CopyMapRecursive(mapValue)
	}
	return newMap
}

func GetString(data map[string]interface{}, name string) (string, bool) {
	value, ok := data[name]
	if !ok {
		return "", false
	}
	valueStr, ok := value.(string)
	if !ok {
		return "", false
	}
	return valueStr, true
}

func GetStringWithDefault(data map[string]interface{}, name string, defaultValue string) string {
	value, ok := GetString(data, name)
	if !ok {
		return defaultValue
	}
	return value
}

func SetString(data map[string]interface{}, name string, value string) {
	iface := make([]interface{}, 1)
	iface[0] = value
	data[name] = iface
}

func GetInt(data map[string]interface{}, name string) (int, bool) {
	value, ok := data[name]
	if !ok {
		return 0, false
	}

	valueInt, ok := value.(int)
	if ok {
		return valueInt, true
	}

	valueFloat64, ok := value.(float64)
	if ok {
		valueInt := int(valueFloat64)
		return valueInt, true
	}

	valueStr, ok := value.(string)
	valueInt, err := strconv.Atoi(valueStr)
	if err == nil {
		return valueInt, true
	}

	return 0, false
}
func GetBoolWithDefault(data map[string]interface{}, name string, defaultValue bool) bool {
	value, ok := data[name]
	if !ok {
		return defaultValue
	}
	boolValue, ok := value.(bool)
	if ok {
		return boolValue
	}
	return defaultValue
}
func GetIntWithDefault(data map[string]interface{}, name string, defaultValue int) int {
	intValue, ok := GetInt(data, name)
	if !ok {
		return defaultValue
	}
	return intValue
}
func GetStringSliceWithDefault(data map[string]interface{}, name string, defaultValue []string) []string {

	value, ok := data[name]
	//fmt.Fprintf(os.Stderr, "GetStringSliceWithDefault() %v %T\n", ok, value)
	valueSlice, ok := value.([]string)
	if ok {
		return valueSlice
	}

	//fmt.Fprintf(os.Stderr, "GetStringSliceWithDefault() 1 interface\n")

	valueSliceInterface, ok := value.([]interface{})
	if !ok {
		return defaultValue
	}
	stringSlice := []string{}
	//fmt.Fprintf(os.Stderr, "GetStringSliceWithDefault() 2 interface\n")
	for _, v := range valueSliceInterface { // <-- fails
		itemValue, ok := v.(string)
		if !ok {
			return defaultValue
		}
		stringSlice = append(stringSlice, itemValue)
	}
	return stringSlice
}

func GetStringValueWithDefault(data map[string]interface{}, name string, defaultValue string) string {
	value, ok := data[name]
	if !ok {
		return defaultValue
	}
	valueStr, ok := value.(string)
	if !ok {
		return defaultValue
	}
	return valueStr
}
func GetInterfaceSliceValueWithDefault(data map[string]interface{}, name string, defaultValue []interface{}) []interface{} {
	value, ok := data[name]
	fmt.Fprintf(os.Stderr, "GetStringSliceValueWithDefault() %v %v (%T)\n", name, ok, value)
	//fmt.Fprintf(os.Stderr, "len = %v\n", len(value))

	if !ok {
		return defaultValue
	}
	valueInterfaceSlice, ok := value.([]interface{})

	fmt.Fprintf(os.Stderr, "GetStringSliceValueWithDefault() valueStr %v %v\n", ok, valueInterfaceSlice)

	if !ok {
		return defaultValue
	}
	fmt.Fprintf(os.Stderr, "len = %v\n", len(valueInterfaceSlice))
	return valueInterfaceSlice
}

func GetMapEntryWithDefault(data map[string]interface{}, name string, _default map[string]interface{}) map[string]interface{} {
	value, ok := data[name]
	if !ok {
		return _default
	}
	myMap, ok := value.(map[string]interface{})
	if !ok {
		return _default
	}
	return myMap
}

func GetMapEntry(data map[string]interface{}, name string) (map[string]interface{}, bool) {
	result := map[string]interface{}{}
	value, ok := data[name]
	if !ok {
		//fmt.Fprintf(os.Stderr, "not found: %s\n", name)
		return result, false
	}
	myMap, ok := value.(map[string]interface{})
	if !ok {
		//fmt.Fprintf(os.Stderr, "not found: \n")
		return result, false
	}
	return myMap, true
}
func SelectKeys(match string, data map[string]interface{}) []string {
	regex, _ := regexp.Compile(match)
	list := []string{}
	for key, _ := range data {
		if !regex.MatchString(key) {
			continue
		}
		//fmt.Fprintf(os.Stderr, "%v %v\n", key, value)
		list = append(list, key)
	}
	//fmt.Fprintf(os.Stderr, "SelectKeys() list: %v\n", list)
	return list
}
