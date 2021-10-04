package services

import (
	"embed"
)

//go:embed *.md
var Docs embed.FS
