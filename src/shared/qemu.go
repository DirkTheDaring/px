package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"px/api"
	"px/configmap"
	"px/queries"
	"reflect"
	"strconv"
	"strings"

	pxapiflat "github.com/DirkTheDaring/px-api-client-go"
	pxapiobject "github.com/DirkTheDaring/px-api-client-internal-go"
)

func UpdateVMConfiguration(node string, vmid int64, config map[string]interface{}, dryRun bool) error {
	jsonTxt, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
	if err = json.Unmarshal(jsonTxt, &updateVMConfigRequest); err != nil {
		return fmt.Errorf("failed to unmarshal VM config: %v", err)
	}
	fmt.Fprintf(os.Stderr, "  update: %s\n", jsonTxt)
	if dryRun {
		return nil
	}
	return UpdateVMConfigurationSync(node, vmid, updateVMConfigRequest)
}
func UpdateVMConfigurationSync(node string, vmid int64, updateVMConfigRequest pxapiflat.UpdateVMConfigRequest) error {
	resp, err := api.UpdateVMConfig(node, vmid, &updateVMConfigRequest)
	if err != nil {
		return fmt.Errorf("failed to update VM config: %v", err)
	}
	queries.WaitForUPID(node, resp.GetData())
	return nil
}

func getAttributeTypeDict(mystruct any) map[string]string {

	myDict := make(map[string]string)

	elements := reflect.ValueOf(mystruct).Elem()

	for i := 0; i < elements.NumField(); i++ {
		field := elements.Type().Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		jsonFieldName := strings.Split(tag, ",")[0]
		_type := field.Type.String()
		//fmt.Println(field.Name, jsonFieldName, _type)
		myDict[jsonFieldName] = _type
	}
	return myDict
}
func processSettings(settings []string, attributeTypeDict map[string]string) map[string]interface{} {
	char := '='
	validTypes := []string{"*string", "*bool", "*int64"}

	myDict := make(map[string]interface{})

	for _, item := range settings {
		position := strings.IndexRune(item, char)
		if position == -1 {
			fmt.Fprintf(os.Stderr, "attribute %s ignored, no value\n", item)
			continue
		}
		key, val := splitAtPosition(item, position)
		_type, ok := attributeTypeDict[key]

		if !ok {
			fmt.Fprintf(os.Stderr, "attribute %s ignored: key does not exist in target\n", key)
			continue
		}

		if !stringInSlice(_type, validTypes) {
			fmt.Fprintf(os.Stderr, "attribute %s type not supported: %s\n", key, _type)
			continue
		}

		switch _type {
		case "*bool":
			parsedBool, err := strconv.ParseBool(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "attribute %s has invalid boolean value: %s\n", key, val)
				continue
			}
			myDict[key] = parsedBool

		case "*int64":
			valInt64, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "attribute %s could not be converted to int64: %s\n", key, val)
				continue
			}
			myDict[key] = valInt64

		case "*string":
			myDict[key] = val

		default:
			fmt.Fprintf(os.Stderr, "unsupported attribute type: %s\n", _type)
		}
	}

	return myDict
}

// splitAtPosition splits a string into two parts at the specified position.
// If the position is outside the string, it returns the original string and an empty string.
func splitAtPosition(str string, position int) (string, string) {
	if position < 0 || position >= len(str) {
		return str, ""
	}
	return str[:position], str[position+1:]
}

// Function to check if a string is in a string array
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func ApplyQemu(machines []map[string]interface{}, settings []string) {

	updateVMConfigRequestObject := pxapiobject.UpdateVMConfigRequest{}
	attributeTypeDict := getAttributeTypeDict(&updateVMConfigRequestObject)
	myDict := processSettings(settings, attributeTypeDict)

	jsonData, err := json.Marshal(myDict)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}
	err = json.Unmarshal(jsonData, &updateVMConfigRequestObject)
	//fmt.Println(string(jsonData))

	for _, machine := range machines {
		vmid, _ := configmap.GetInt(machine, "vmid")
		node, _ := configmap.GetString(machine, "node")

		fmt.Fprintf(os.Stdout, "Apply Virtual Machine %v on %s\n", vmid, node)

		updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
		CopyUpdateVMConfigRequest(&updateVMConfigRequest, &updateVMConfigRequestObject)
		/*
			resp, _ := api.UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)
			upid := resp.GetData()
			//fmt.Fprintf(os.Stderr, "upid = %s\n", upid)
			//shared.GetNodeTaskStatus(node, upid)
			WaitForUPID(node, upid)
		*/

		UpdateVMConfigurationSync(node, int64(vmid), updateVMConfigRequest)
	}
}
