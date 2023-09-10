package ignition

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"px/configmap"
	"runtime"
)

// Source:
// https://github.com/coreos/butane/releases

func GetButaneCmd(data map[string]interface{}) string {
	value := os.Getenv("PX_BUTANE_CMD")
	if value != "" {
		return value
	}
	butaneCmd := configmap.GetStringWithDefault(data, "butane_cmd", "")
	if butaneCmd != "" {
		return value
	}

	var arch string
	cmdPostfix := ""
	unknown := "unknown"

	if runtime.GOARCH == "amd64" {
		arch = "x86_64"
	} else {
		arch = runtime.GOARCH
	}

	if runtime.GOOS == "windows" {
		cmdPostfix = ".exe"
		unknown = "pc"
	}
	if runtime.GOOS == "darwin" {
		unknown = "apple"
	}

	/*
		if runtime.GOOS == "windows" {
			butaneCmd = "findstr.exe"
			butaneArgs = []string{"^"}
		}
	*/
	butaneCmd = "butane-" + arch + "-" + unknown + "-" + runtime.GOOS + "-gnu" + cmdPostfix
	return butaneCmd
}

func GetButaneArgs(data map[string]interface{}) []string {

	args := configmap.GetStringSliceWithDefault(data, "butane_args", []string{})

	butaneFilesDir := configmap.GetStringWithDefault(data, "butane_files_dir", "")
	if butaneFilesDir == "" {
		fmt.Fprintf(os.Stderr, "no butane_files_dir set \n")
		return args
	}
	parentDir := filepath.Dir(butaneFilesDir)

	arg := "--files-dir=" + parentDir
	args = append(args, arg)

	return args
}

func Butane(inStream io.Reader, data map[string]interface{}) bytes.Buffer {

	butaneCmd := GetButaneCmd(data)
	butaneArgs := GetButaneArgs(data)
	//fmt.Fprintf(os.Stderr, "EXEC: %v %v\n", butaneCmd, butaneArgs)

	result := pipeThroughShellCmd(butaneCmd, butaneArgs, inStream)
	return result
}
