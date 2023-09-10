package ignition

import "embed"

var (
	//go:embed "templates/*"
	templates embed.FS

	//go:embed "defaults/*"
	defaults embed.FS
)
