package ignition

import (
	"bytes"
	"fmt"
	"os"
	"px/configmap"
)
// PadBuffer pads the buffer with spaces until it reaches the specified length.
// It modifies the buffer in place.
func PadBuffer(buffer *bytes.Buffer, length int) {
    size := length - buffer.Len()
    if size > 0 {
        payload := bytes.Repeat([]byte{' '}, size)
        buffer.Write(payload)
    }
}

// A proxmox ignition is padded to 16384 if it is smaller than 16384, otherwise
// the upload function in proxmox does not wor (QUIRK!)
func CreateProxmoxIgnition(configData map[string]interface{}, name string) bytes.Buffer {
	result, err := CreateIgnition(configData, name)
	if err != nil {
		var buffer bytes.Buffer
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return buffer
	}
	PadBuffer(&result, 16384)
	return result
}

// WriteFile creates a file with the given name and writes the content to it.
// It returns an error if any operation fails.
func WriteFile(name string, content string) error {
    // Create or truncate the file
    f, err := os.Create(name)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write the content to the file
    _, err = f.WriteString(content)
    if err != nil {
        return err
    }

    // Ensure that the content is actually written to disk
    err = f.Sync()
    if err != nil {
        return err
    }

    return nil
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
