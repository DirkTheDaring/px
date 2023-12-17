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

// CreateIgnition generates an Ignition configuration based on provided data.
func CreateIgnition(data map[string]interface{}, name string) (bytes.Buffer, error) {
    var buffer bytes.Buffer

    if err := handleDebugMode(data, name); err != nil {
        return buffer, err
    }

    if err := processButaneFilesDir(data); err != nil {
        return buffer, err
    }

    if err := renderIgnitionYaml(&buffer, data); err != nil {
        return buffer, err
    }

    return generateButaneOutput(&buffer, data, name)
}

// handleDebugMode handles debug mode operations.
func handleDebugMode(data map[string]interface{}, name string) error {
    debugLevel := configmap.GetIntWithDefault(data, "debug_level", 0)
    if debugLevel > 0 {
	json, err := json.Marshal(data)
        if err != nil {
            return err
        }
        return WriteFile(name+".vars.json", string(json))
    }
    return nil
}

// processButaneFilesDir processes the 'butane_files_dir' directory.
func processButaneFilesDir(data map[string]interface{}) error {
    butaneFilesDir := configmap.GetStringWithDefault(data, "butane_files_dir", "butane")
    convertedDir:= convertSymlinkToPath(data, butaneFilesDir)
    //convertedDir, err := convertSymlinkToPath(data, butaneFilesDir)
    //if err != nil {
    //    return err
    //}
    data["butane_files_dir"] = convertedDir
    configmap.SetString(data, "trees", filepath.Base(convertedDir))
    return nil
}

// renderIgnitionYaml renders Ignition YAML configuration.
func renderIgnitionYaml(outStream io.Writer, data map[string]interface{}) error {
    RenderIgnitionYaml(outStream, data)
    return nil
}

// generateButaneOutput generates output for Butane.
func generateButaneOutput(inStream io.Reader, data map[string]interface{}, name string) (bytes.Buffer, error) {
    output :=  Butane(inStream, data)

    debugLevel := configmap.GetIntWithDefault(data, "debug_level", 0)
    if debugLevel > 0 {
        if err := WriteFile(name+".json", output.String()); err != nil {
            return output, err
        }
    }
    return output, nil
}
