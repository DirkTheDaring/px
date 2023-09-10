package console

import (
	"github.com/alecthomas/kong"
)

func InitCommandline() (CommandLineInterface, *kong.Context) {
	commandLineInterface := CommandLineInterface{}
	context := kong.Parse(&commandLineInterface)
	return commandLineInterface, context
}
