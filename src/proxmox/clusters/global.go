package clusters

import "embed"

var (
	pxConfigFileDir     string = ".px"
	pxConfigFileName    string = "config.yaml"
	pxConfigFilenameVar string = "PX_CONFIG_FILE_PATH"
	//go:embed "files/*"
	files embed.FS
)
