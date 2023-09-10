package ignition

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func CreateShellCmd(name string, args []string, outstream io.Writer, instream io.Reader) *exec.Cmd {
	// use args array as ellipse
	cmd := exec.Command(name, args...)
	cmd.Stdin = instream
	cmd.Stdout = outstream
	cmd.Stderr = os.Stderr
	return cmd
}
func pipeThroughShellCmd(name string, args []string, inStream io.Reader) bytes.Buffer {

	var outStream io.Writer
	buffer := bytes.Buffer{}
	outStream = &buffer

	cmd := CreateShellCmd(name, args, outStream, inStream)
	error := cmd.Run()
	if error != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", error)
		return bytes.Buffer{}
	}
	return buffer
}
