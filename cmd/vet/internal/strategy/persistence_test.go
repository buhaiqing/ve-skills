package strategy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	kbPath := filepath.Join(tmpDir, "kb.json")

	kb := NewKnowledgeBase()
	kb.Learn(FailurePattern{Pattern: "error timeout", Skill: "network", Solution: "restart network service"})
	kb.Learn(FailurePattern{Pattern: "out of memory", Skill: "memory", Solution: "increase memory limit"})
	kb.Learn(FailurePattern{Pattern: "disk full", Skill: "storage", Solution: "clean up disk space"})

	g := NewStrategyGraph()
	node1 := &StrategyNode{
		ID:        "node-1",
		Symptom:   "timeout",
		Operation: "check network",
		Priority:  10,
		Enabled:   true,
	}
	node2 := &StrategyNode{
		ID:        "node-2",
		Symptom:   "memory",
		Operation: "check memory usage",
		Priority:  5,
		Enabled:   true,
	}
	g.AddNode(node1)
	g.AddNode(node2)
	kb.AddGraph("diag", g)

	origResult := kb.Query("service error timeout connection")
	if origResult == nil {
		t.Fatal("expected query result, got nil")
	}
	if origResult.Solution != "restart network service" {
		t.Errorf("expected solution 'restart network service', got %q", origResult.Solution)
	}

	if err := kb.Save(kbPath); err != nil {
		t.Fatalf("failed to save kb: %v", err)
	}

	kb2 := NewKnowledgeBase()
	if err := kb2.Load(kbPath); err != nil {
		t.Fatalf("failed to load kb: %v", err)
	}

	loadedResult := kb2.Query("service error timeout connection")
	if loadedResult == nil {
		t.Fatal("expected loaded query result, got nil")
	}
	if loadedResult.Solution != origResult.Solution {
		t.Errorf("loaded solution mismatch: expected %q, got %q", origResult.Solution, loadedResult.Solution)
	}
	if loadedResult.Skill != origResult.Skill {
		t.Errorf("loaded skill mismatch: expected %q, got %q", origResult.Skill, loadedResult.Skill)
	}

	loadedGraph := kb2.GetGraph("diag")
	if loadedGraph == nil {
		t.Fatal("expected loaded graph, got nil")
	}
	path := loadedGraph.GetPath("timeout")
	if len(path) == 0 {
		t.Error("expected path with timeout symptom, got empty")
	}
}

func TestSaveLoadEmptyKB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kb-test-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	kbPath := filepath.Join(tmpDir, "kb-empty.json")

	kb := NewKnowledgeBase()
	if err := kb.Save(kbPath); err != nil {
		t.Fatalf("failed to save empty kb: %v", err)
	}

	kb2 := NewKnowledgeBase()
	if err := kb2.Load(kbPath); err != nil {
		t.Fatalf("failed to load empty kb: %v", err)
	}

	kb2.mu.RLock()
	defer kb2.mu.RUnlock()
	if len(kb2.patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(kb2.patterns))
	}
	if len(kb2.graphs) != 0 {
		t.Errorf("expected 0 graphs, got %d", len(kb2.graphs))
	}
}

func TestLoadNoFile(t *testing.T) {
	kb := NewKnowledgeBase()
	err := kb.Load("no-such-file-xyzabc.json")
	if err != nil {
		t.Errorf("expected nil error for non-existent file, got %v", err)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kb-test-corrupt-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	corruptPath := filepath.Join(tmpDir, "corrupt.json")

	if err := os.WriteFile(corruptPath, []byte("{broken}"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	kb := NewKnowledgeBase()
	err = kb.Load(corruptPath)
	if err == nil {
		t.Error("expected error for corrupt file, got nil")
	}
}

func TestSaveBadDir(t *testing.T) {
	kb := NewKnowledgeBase()
	err := kb.Save("/nonexistent-dir-12345/kb.json")
	if err == nil {
		t.Error("expected error for bad directory, got nil")
	}
}

func TestGraphRebuildChildren(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kb-test-children-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	kbPath := filepath.Join(tmpDir, "kb-children.json")

	kb := NewKnowledgeBase()
	g := NewStrategyGraph()
	child1 := &StrategyNode{
		ID:        "child-1",
		Symptom:   "symptom-a",
		Operation: "op-a",
		Priority:  1,
		Enabled:   true,
	}
	child2 := &StrategyNode{
		ID:        "child-2",
		Symptom:   "symptom-b",
		Operation: "op-b",
		Priority:  2,
		Enabled:   true,
	}
	g.AddNode(child1)
	g.AddNode(child2)
	kb.AddGraph("test-graph", g)

	if err := kb.Save(kbPath); err != nil {
		t.Fatalf("failed to save kb: %v", err)
	}

	kb2 := NewKnowledgeBase()
	if err := kb2.Load(kbPath); err != nil {
		t.Fatalf("failed to load kb: %v", err)
	}

	loadedGraph := kb2.GetGraph("test-graph")
	if loadedGraph == nil {
		t.Fatal("expected loaded graph, got nil")
	}

	loadedGraph.mu.RLock()
	defer loadedGraph.mu.RUnlock()
	root := loadedGraph.root
	if root == nil {
		t.Fatal("expected root node, got nil")
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected root to have 2 children, got %d", len(root.Children))
	}

	childIDs := make(map[string]bool)
	for _, c := range root.Children {
		childIDs[c.ID] = true
	}
	if !childIDs["child-1"] {
		t.Error("missing child-1 in root children")
	}
	if !childIDs["child-2"] {
		t.Error("missing child-2 in root children")
	}
}
