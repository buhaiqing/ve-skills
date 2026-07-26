package strategy

import (
	"testing"
	"time"
)

func TestStrategyGraph(t *testing.T) {
	g := NewStrategyGraph()

	node1 := &StrategyNode{
		ID:        "n1",
		Symptom:   "CPU high",
		Operation: "scale up",
		Priority:  10,
	}
	node2 := &StrategyNode{
		ID:        "n2",
		Symptom:   "memory leak",
		Operation: "restart service",
		Priority:  5,
	}
	g.AddNode(node1)
	g.AddNode(node2)

	path := g.GetPath("CPU")
	if len(path) == 0 {
		t.Fatal("expected path for 'CPU' symptom, got empty")
	}
	if path[len(path)-1].ID != "n1" {
		t.Fatalf("expected last node to be n1, got %s", path[len(path)-1].ID)
	}
}

func TestStrategyGraphEmpty(t *testing.T) {
	g := NewStrategyGraph()

	path := g.GetPath("anything")
	if len(path) != 0 {
		t.Fatalf("expected empty path, got %d nodes", len(path))
	}
}

func TestGeneratePlan(t *testing.T) {
	g := NewStrategyGraph()

	g.AddNode(&StrategyNode{
		ID:        "n1",
		Symptom:   "CPU high",
		Operation: "scale up",
		Priority:  10,
	})
	g.AddNode(&StrategyNode{
		ID:        "n2",
		Symptom:   "memory leak",
		Operation: "restart",
		Priority:  5,
	})
	g.AddNode(&StrategyNode{
		ID:        "n3",
		Symptom:   "disk full",
		Operation: "cleanup",
		Priority:  20,
	})

	plan := g.GeneratePlan([]string{"CPU", "memory", "disk"})
	if len(plan) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(plan))
	}
	if plan[0] != "cleanup" {
		t.Fatalf("expected first operation to be 'cleanup' (highest priority), got '%s'", plan[0])
	}
	if plan[1] != "scale up" {
		t.Fatalf("expected second operation to be 'scale up', got '%s'", plan[1])
	}
	if plan[2] != "restart" {
		t.Fatalf("expected third operation to be 'restart', got '%s'", plan[2])
	}
}

func TestKnowledgeBase(t *testing.T) {
	kb := NewKnowledgeBase()

	kb.Learn(FailurePattern{
		Pattern:  "timeout",
		Skill:    "network",
		Solution: "check connectivity",
	})

	result := kb.Query("connection timeout")
	if result == nil {
		t.Fatal("expected to find pattern for 'connection timeout'")
	}
	if result.Solution != "check connectivity" {
		t.Fatalf("expected solution 'check connectivity', got '%s'", result.Solution)
	}
}

func TestKnowledgeBaseLearn(t *testing.T) {
	kb := NewKnowledgeBase()

	kb.Learn(FailurePattern{
		Pattern:  "timeout",
		Skill:    "network",
		Solution: "check connectivity",
	})

	result := kb.Query("timeout")
	if result == nil {
		t.Fatal("expected to find pattern")
	}
	if result.HitCount != 1 {
		t.Fatalf("expected HitCount 1, got %d", result.HitCount)
	}
	if result.Confidence <= 0 {
		t.Fatalf("expected positive confidence, got %f", result.Confidence)
	}
}

func TestKnowledgeBaseLearnUpdate(t *testing.T) {
	kb := NewKnowledgeBase()

	before := time.Now().Add(-1 * time.Hour)
	kb.Learn(FailurePattern{
		Pattern:  "timeout",
		Skill:    "network",
		Solution: "check connectivity",
	})

	kb.Learn(FailurePattern{
		Pattern:  "timeout",
		Skill:    "network",
		Solution: "check connectivity again",
	})

	result := kb.Query("timeout")
	if result == nil {
		t.Fatal("expected to find pattern")
	}
	if result.HitCount != 2 {
		t.Fatalf("expected HitCount 2, got %d", result.HitCount)
	}
	if !result.LastHit.After(before) {
		t.Fatal("expected LastHit to be updated")
	}
	if result.Confidence <= 0.5 {
		t.Fatalf("expected confidence to be boosted above 0.5, got %f", result.Confidence)
	}
}