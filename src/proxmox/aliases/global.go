package aliases

import "embed"

var (
	//go:embed files/*
	px        embed.FS
	pxAliases map[string]interface{} = nil
)
