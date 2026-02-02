package main

import (
	"embed"
	"io/fs"
)

//go:embed web/dist/*
var webFS embed.FS

// GetWebFS returns the embedded web filesystem.
func GetWebFS() (fs.FS, error) {
	return fs.Sub(webFS, "web/dist")
}
