package strategy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKnowledgeBaseMissingIsEmpty(t *testing.T) {
	kb := LoadKnowledgeBase(t.TempDir())
	if kb.Query("cpu spike") != nil {
		t.Fatal("expected empty KB when file missing")
	}
}

func TestLoadKnowledgeBaseCorruptFallsBackEmpty(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, DefaultKnowledgeBasePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	kb := LoadKnowledgeBase(root)
	if kb.Query("anything") != nil {
		t.Fatal("corrupt KB must fall back to empty, not panic")
	}
}
