package ignition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"px/configmap"
)

// butane cannot handle Symlinks to dirs, therefore we convert the butane_files_dir setting
// if there is a symlink contained
func convertSymlinkToPath(configData map[string]interface{}, butaneFilesDir string) string {
	filesDir, err := filepath.EvalSymlinks(butaneFilesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ""
	}
	return filesDir
}
func dataToJSON(data map[string]interface{}) []byte {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return json
}

func CreateIgnition(data map[string]interface{}, name string) bytes.Buffer {

	buffer := bytes.Buffer{}
	var outStream io.Writer
	outStream = &buffer

	debugLevel := configmap.GetIntWithDefault(data, "debug_level", 0)
	//fmt.Fprintf(os.Stderr, "DEBUGL LEVEL: %v\n", debugLevel)

	if debugLevel > 0 {
		json := dataToJSON(data)
		WriteFile(name+".vars.json", string(json))
	}

	// Quirk for butane, if butane_files-dir contains a symlink, it fails
	// therefore this section converts an symlink to an absolute path
	varname := "butane_files_dir"
	butaneFilesDir := configmap.GetStringWithDefault(data, varname, "butane")
	butaneFilesDir = convertSymlinkToPath(data, butaneFilesDir)
	data[varname] = butaneFilesDir

	baseName := filepath.Base(butaneFilesDir)
	//parentDir := filepath.Dir(butaneFilesDir)

	configmap.SetString(data, "trees", baseName)

	RenderIgnitionYaml(outStream, data)

	if debugLevel > 0 {
		//fmt.Fprintf(os.Stderr, "Writing file: %v\n", name+".yaml")
		WriteFile(name+".yaml", buffer.String())
	}

	//fmt.Fprintf(os.Stderr, "%s\n", buffer.String())

	var inStream io.Reader
	inStream = &buffer
	output := Butane(inStream, data)
	if debugLevel > 0 {
		WriteFile(name+".json", output.String())
	}

	return output
}
