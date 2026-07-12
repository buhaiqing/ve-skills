package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRepoRootResolvesToRepo confirms repoRoot() finds the ve-skills root
// (cmd/vet + AGENTS.md) instead of falling back to "/" or the cwd when the
// binary lives outside the repo.
func TestRepoRootResolvesToRepo(t *testing.T) {
	root := repoRoot()
	if !isRepoRoot(root) {
		t.Fatalf("repoRoot()=%q is not a repo root (missing cmd/vet or AGENTS.md)", root)
	}
	if root == "/" {
		t.Fatal("repoRoot() returned \"/\" — must not default to filesystem root")
	}
	if filepath.Base(root) != "ve-skills" {
		t.Fatalf("repoRoot()=%q: expected repo dir named ve-skills", root)
	}
}

// TestIsRepoRoot verifies the marker check used by repoRoot.
func TestIsRepoRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// cwd is cmd/vet, which is NOT the repo root itself.
	if isRepoRoot(cwd) {
		t.Fatal("cmd/vet should not be detected as repo root")
	}
	// two levels up is the repo root.
	repo := filepath.Dir(filepath.Dir(cwd))
	if !isRepoRoot(repo) {
		t.Fatalf("%q should be detected as repo root", repo)
	}
}
