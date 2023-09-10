package ignition

import (
	"bytes"
	"fmt"
	"os"
	"px/configmap"
)

func PadBuffer(buffer bytes.Buffer, length int) bytes.Buffer {
	size := length - buffer.Len()
	if size > 0 {
		payload := make([]byte, size)
		for i := 0; i < size; i++ {
			payload[i] = 32
		}
		buffer.Write(payload)
	}
	return buffer
}

func CreateProxmoxIgnition(configData map[string]interface{}, name string) bytes.Buffer {
	result := CreateIgnition(configData, name)
	result = PadBuffer(result, 16384)
	return result
}
func WriteFile(name string, content string) {
	f, _ := os.Create(name)
	f.WriteString(content)
	f.Close()
}

func CreateProxmoxIgnitionFile(configData map[string]interface{}, name string) {

	// Merge settings and the configData
	defaultData := LoadEmbeddedYamlFile("defaults/settings.yaml")
	data := configmap.MergeMapRecursive(defaultData, configData)

	outputDir := configmap.GetStringWithDefault(data, "output_dir", "")

	//fmt.Fprintf(os.Stderr, "outputDir: %v\n", outputDir)

	if outputDir != "" {
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			error := os.Mkdir(outputDir, 0700)
			if error != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", error)
			}
		}
		name = outputDir + string(os.PathSeparator) + name
		//fmt.Fprintf(os.Stderr, "filename: %v\n", name)
	}

	content := CreateProxmoxIgnition(data, name)
	WriteFile(name+".iso", content.String())
}

func AssemblePartialPath(name string, storageType string) string {
	result := ""
	// This based on heuristic
	if storageType == "vztmpl" {
		result = "/template/cache/" + name + ".ign.tar.gz"
	} else {
		// iso
		result = "/template/iso/" + name + ".ign.iso"
	}
	return result
}
func UploadFile2(storage string, storageType string, name string) string {
	/*
		filename := "@@IGNITION:storage=ignition,format=iso,template=name,tags=master@@"
		value := "-fw_cfg name=opt/com.coreos/config,file=" + filename
		valueNew := "-fw_cfg name=opt/com.coreos/config,file={{ AssemblePath "ignition" "iso" getName() }}"
	*/

	rootDir := ""
	result := rootDir + AssemblePartialPath(name, storageType)
	return result
}
