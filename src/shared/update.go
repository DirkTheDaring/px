package shared

import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
	"log"
	"px/configmap"
        "github.com/DirkTheDaring/px-api-client-go"
        "github.com/DirkTheDaring/px-api-client-internal-go"

)
// splitAtPosition splits a string into two parts at the specified position.
// If the position is outside the string, it returns the original string and an empty string.
func splitAtPosition(str string, position int) (string, string) {
    if position < 0 || position >= len(str) {
        return str, ""
    }
    return str[:position], str[position+1:]
}

func Update(match string, settings []string) {
	{
		machines := GlobalPxCluster.Machines

		char := '='
		myDict := make(map[string]interface{})


		for _, item := range settings {
			position := strings.IndexRune(item, char)
			if position != -1 {
				key,val := splitAtPosition(item, position)
				//fmt.Println(key, val)
				if strings.ToLower(val) == "false" {
					myDict[key] = false
					continue
				}
				if strings.ToLower(val) == "true" {
					myDict[key] = true
					continue
				}
				if val == "0" {
					myDict[key]=false
					continue
				}
				if val == "1" {
					myDict[key]=true
					continue
				}
				
				myDict[key] = val

			}
		}
		filterColumn := "name"
		filterString := match
		for _, output := range machines {
			if filterString != "" {
				value, _ := output[filterColumn].(string)
				if !strings.HasPrefix(value, filterString) {
					continue
				}
				vmid,_  := configmap.GetInt(output, "vmid")
				node,_  := configmap.GetString(output, "node")
				_type,_ := configmap.GetString(output, "type")

				jsonData, err := json.Marshal(myDict)
				if err != nil {
					log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
				}
				//fmt.Println(string(jsonData))

				if _type == "qemu" {
					updateVMConfigRequestObject := pxapiobject.UpdateVMConfigRequest{}
					err = json.Unmarshal(jsonData, &updateVMConfigRequestObject)
					jsonData, err = json.Marshal(updateVMConfigRequestObject)
					//fmt.Println(string(jsonData))
					updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
					CopyUpdateVMConfigRequest(&updateVMConfigRequest, &updateVMConfigRequestObject)
					resp, _ := UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)
					upid := resp.GetData()
					//fmt.Fprintf(os.Stderr, "upid = %s\n", upid)
					//shared.GetNodeTaskStatus(node, upid)
					WaitForUPID(node,upid) 
				} else {
					updateContainerConfigSyncRequestObject := pxapiobject.UpdateContainerConfigSyncRequest{}
					err = json.Unmarshal(jsonData, &updateContainerConfigSyncRequestObject)
					jsonData, err = json.Marshal(updateContainerConfigSyncRequestObject)
					fmt.Println(string(jsonData))

					updateContainerConfigSyncRequest := pxapiflat.UpdateContainerConfigSyncRequest{}

					CopyUpdateContainerConfigSyncRequest(&updateContainerConfigSyncRequest, &updateContainerConfigSyncRequestObject)

					UpdateContainerConfigSync(node, int64(vmid), updateContainerConfigSyncRequest)
					/*
					updateVMConfigRequest := pxapiflat.UpdateVMConfigRequest{}
					CopyUpdateVMConfigRequest(&updateVMConfigRequest, &updateVMConfigRequestObject)
					resp, _ := UpdateVMConfig(node, int64(vmid), &updateVMConfigRequest)
					upid := resp.GetData()
					*/
					//fmt.Fprintf(os.Stderr, "upid = %s\n", upid)
					//shared.GetNodeTaskStatus(node, upid)

					//WaitForUPID(node,upid) 

				}
			}

		}

		os.Exit(0)
	}
}
