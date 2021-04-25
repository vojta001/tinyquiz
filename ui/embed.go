package ui

import "embed"

//go:embed html/*.tmpl.html
var HTMLTemplates embed.FS

//go:embed static
var StaticFiles embed.FS
