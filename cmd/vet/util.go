package main

import (
	"os"
	"path/filepath"
)

func up(p string) string  { return filepath.Dir(p) }
func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
