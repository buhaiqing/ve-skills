package main

import (
	"os"
	"path/filepath"
)

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// isRepoRoot reports whether p looks like the ve-skills repo root
// (has both AGENTS.md and a cmd/vet directory).
func isRepoRoot(p string) bool {
	return dirExists(filepath.Join(p, "cmd", "vet")) && fileExists(filepath.Join(p, "AGENTS.md"))
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// repoRoot walks up from the current working directory looking for the
// ve-skills repo root; falls back to "." so callers scan the cwd by default.
// Using cwd (not the binary path) keeps the default correct when the binary
// is installed outside the repo (e.g. /usr/local/bin).
func repoRoot() string {
	start, err := os.Getwd()
	if err != nil {
		return "."
	}
	for p := start; ; p = filepath.Dir(p) {
		if isRepoRoot(p) {
			return p
		}
		if p == "/" || p == "." {
			break
		}
	}
	return "."
}
