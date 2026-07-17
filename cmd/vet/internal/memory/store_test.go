package memory

import (
	"sync"
	"testing"
)

func TestAppend_NewPattern(t *testing.T) {
	root := t.TempDir()

	err := AppendFailurePattern(root, FailurePatternEntry{
		Skill:    "ve-ecs-ops",
		Pattern:  "disk full on /data",
		Category: "disk",
		Fix:      "expand disk",
		Source:   "GCL_MAX_ITER",
	})
	if err != nil {
		t.Fatalf("AppendFailurePattern: %v", err)
	}

	patterns, err := LoadFailurePatterns(root)
	if err != nil {
		t.Fatalf("LoadFailurePatterns: %v", err)
	}
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Count != 1 {
		t.Errorf("expected count=1, got %d", patterns[0].Count)
	}
}

func TestAppend_ExistingPattern(t *testing.T) {
	root := t.TempDir()

	entry := FailurePatternEntry{
		Skill:    "ve-ecs-ops",
		Pattern:  "disk full on /data",
		Category: "disk",
		Fix:      "expand disk",
		Source:   "GCL_MAX_ITER",
	}

	_ = AppendFailurePattern(root, entry)
	_ = AppendFailurePattern(root, entry)

	patterns, err := LoadFailurePatterns(root)
	if err != nil {
		t.Fatalf("LoadFailurePatterns: %v", err)
	}
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Count != 2 {
		t.Errorf("expected count=2, got %d", patterns[0].Count)
	}
}

func TestAppend_DifferentSkill(t *testing.T) {
	root := t.TempDir()

	_ = AppendFailurePattern(root, FailurePatternEntry{
		Skill:   "ve-ecs-ops",
		Pattern: "disk full",
	})
	_ = AppendFailurePattern(root, FailurePatternEntry{
		Skill:   "ve-slb-ops",
		Pattern: "disk full",
	})

	patterns, _ := LoadFailurePatterns(root)
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(patterns))
	}
	// Verify both have count=1 and are independent
	for _, p := range patterns {
		if p.Count != 1 {
			t.Errorf("skill %s: expected count=1, got %d", p.Skill, p.Count)
		}
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	root := t.TempDir()

	patterns, err := LoadFailurePatterns(root)
	if err != nil {
		t.Fatalf("LoadFailurePatterns: %v", err)
	}
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(patterns))
	}
}

func TestGetBySkill_Filter(t *testing.T) {
	root := t.TempDir()

	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p1", Count: 1})
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "b", Pattern: "p2", Count: 1})
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p3", Count: 1})

	results, err := GetPatternsBySkill(root, "a", 10)
	if err != nil {
		t.Fatalf("GetPatternsBySkill: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Skill != "a" {
			t.Errorf("unexpected skill: %s", r.Skill)
		}
	}
}

func TestGetBySkill_Limit(t *testing.T) {
	root := t.TempDir()

	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p1", Count: 1})
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p2", Count: 1})
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p3", Count: 1})
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p1", Count: 1}) // bump p1 to count=2
	_ = AppendFailurePattern(root, FailurePatternEntry{Skill: "a", Pattern: "p1", Count: 1}) // bump p1 to count=3

	results, err := GetPatternsBySkill(root, "a", 2)
	if err != nil {
		t.Fatalf("GetPatternsBySkill: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// p1 should be first (count=3)
	if results[0].Pattern != "p1" || results[0].Count != 3 {
		t.Errorf("expected p1 count=3 first, got %s count=%d", results[0].Pattern, results[0].Count)
	}
}

func TestAppend_Concurrent(t *testing.T) {
	root := t.TempDir()

	entry := FailurePatternEntry{
		Skill:    "ve-ecs-ops",
		Pattern:  "concurrent pattern",
		Category: "concurrency",
	}

	var wg sync.WaitGroup
	const n = 10
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = AppendFailurePattern(root, entry)
		}()
	}
	wg.Wait()

	patterns, err := LoadFailurePatterns(root)
	if err != nil {
		t.Fatalf("LoadFailurePatterns: %v", err)
	}
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Count != n {
		t.Errorf("expected count=%d, got %d", n, patterns[0].Count)
	}
}
