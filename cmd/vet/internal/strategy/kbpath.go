package strategy

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultKnowledgeBasePath is the repo-relative path for persisted strategy KB.
const DefaultKnowledgeBasePath = ".runtime/strategy/kb.json"

// RelKnowledgeBasePath is an alias kept for Wave A call sites.
const RelKnowledgeBasePath = DefaultKnowledgeBasePath

// LoadKnowledgeBase returns a KB loaded from root/DefaultKnowledgeBasePath.
// Missing file → empty KB. Corrupt/permission errors are logged to stderr;
// ProposeFix still continues with an empty in-memory KB.
func LoadKnowledgeBase(root string) *KnowledgeBase {
	kb := NewKnowledgeBase()
	path := filepath.Join(root, DefaultKnowledgeBasePath)
	if err := kb.Load(path); err != nil {
		// Load already returns nil for os.IsNotExist.
		fmt.Fprintf(os.Stderr, "strategy: load knowledge base %s: %v\n", path, err)
	}
	return kb
}
