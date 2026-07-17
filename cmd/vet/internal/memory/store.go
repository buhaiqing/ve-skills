package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// FailurePatternEntry is one failure pattern in the runtime memory store.
type FailurePatternEntry struct {
	Skill    string `json:"skill"`
	Pattern  string `json:"pattern"`
	Category string `json:"category"`
	Fix      string `json:"fix"`
	Source   string `json:"source"`
	Count    int    `json:"count"`
}

// FailurePatternStore is the on-disk representation.
type FailurePatternStore struct {
	Version  string                `json:"version"`
	Patterns []FailurePatternEntry `json:"patterns"`
}

var mu sync.Mutex

func storePath(root string) string {
	return filepath.Join(root, ".runtime", "memory", "failure-patterns.json")
}

// AppendFailurePattern appends or updates a failure pattern.
// If (skill, pattern) already exists → count++; otherwise → new entry with count=1.
// Uses a package-level sync.Mutex for concurrency safety.
func AppendFailurePattern(root string, entry FailurePatternEntry) error {
	mu.Lock()
	defer mu.Unlock()

	store, err := loadStore(root)
	if err != nil {
		store = &FailurePatternStore{Version: "1.0.0"}
	}

	found := false
	for i := range store.Patterns {
		if store.Patterns[i].Skill == entry.Skill && store.Patterns[i].Pattern == entry.Pattern {
			store.Patterns[i].Count++
			found = true
			break
		}
	}
	if !found {
		entry.Count = 1
		store.Patterns = append(store.Patterns, entry)
	}

	return writeStore(root, store)
}

// LoadFailurePatterns loads all failure patterns from disk.
func LoadFailurePatterns(root string) ([]FailurePatternEntry, error) {
	mu.Lock()
	defer mu.Unlock()

	store, err := loadStore(root)
	if err != nil {
		return nil, nil
	}
	return store.Patterns, nil
}

// GetPatternsBySkill filters patterns by skill and returns the top-N by count descending.
func GetPatternsBySkill(root, skill string, limit int) ([]FailurePatternEntry, error) {
	patterns, err := LoadFailurePatterns(root)
	if err != nil {
		return nil, err
	}

	var filtered []FailurePatternEntry
	for _, p := range patterns {
		if p.Skill == skill {
			filtered = append(filtered, p)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Count > filtered[j].Count
	})

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func loadStore(root string) (*FailurePatternStore, error) {
	data, err := os.ReadFile(storePath(root))
	if err != nil {
		return nil, err
	}

	var store FailurePatternStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func writeStore(root string, store *FailurePatternStore) error {
	dir := filepath.Dir(storePath(root))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(storePath(root), data, 0o644)
}
