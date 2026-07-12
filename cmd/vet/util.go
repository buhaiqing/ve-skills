package main

import (
	"os"
	"path/filepath"
)

func up(p string) string { return filepath.Dir(p) }

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// repoRoot returns the repository root: two levels up from the vet binary
// (cmd/vet/vet -> repo root), falling back to ".".
func repoRoot() string {
	exe, err := os.Executable()
	if err == nil {
		if p := up(up(up(exe))); dirExists(p) {
			return p
		}
	}
	return "."
}
