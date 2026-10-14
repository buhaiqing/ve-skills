package duprefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckDir_FindsDups(t *testing.T) {
	root, _ := os.MkdirTemp("", "dup")
	defer os.RemoveAll(root)
	block := "line one that repeats\nline two repro\nthird line same\nfourth appears many places now"
	for _, name := range []string{"ve-a-ops", "ve-b-ops", "ve-c-ops", "ve-d-ops", "ve-e-ops", "ve-f-ops", "ve-g-ops"} {
		d := filepath.Join(root, name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(block+"\n"+block+"\n"), 0o644)
	}
	findings, err := CheckDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected findings, got 0")
	}
}

func TestCheckDir_NoErrorOnMissing(t *testing.T) {
	root, _ := os.MkdirTemp("", "dup2")
	defer os.RemoveAll(root)
	if _, err := CheckDir(root); err != nil {
		t.Fatal(err)
	}
}
