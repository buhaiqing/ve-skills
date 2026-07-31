package strategy

import "path/filepath"

const RelKnowledgeBasePath = ".runtime/strategy/kb.json"

func LoadKnowledgeBase(root string) *KnowledgeBase {
	kb := NewKnowledgeBase()
	_ = kb.Load(filepath.Join(root, RelKnowledgeBasePath))
	return kb
}
