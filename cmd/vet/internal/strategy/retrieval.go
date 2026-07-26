package strategy

import (
	"strings"
	"sync"
	"time"
)

type FailurePattern struct {
	Pattern    string
	Skill      string
	Solution   string
	HitCount   int
	LastHit    time.Time
	Confidence float64
}

type KnowledgeBase struct {
	patterns []FailurePattern
	graphs   map[string]*StrategyGraph
	mu       sync.RWMutex
}

func NewKnowledgeBase() *KnowledgeBase {
	return &KnowledgeBase{
		patterns: make([]FailurePattern, 0),
		graphs:   make(map[string]*StrategyGraph),
	}
}

func (kb *KnowledgeBase) Query(symptom string) *FailurePattern {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	var best *FailurePattern
	for i := range kb.patterns {
		p := &kb.patterns[i]
		if strings.Contains(strings.ToLower(symptom), strings.ToLower(p.Pattern)) {
			if best == nil || p.Confidence > best.Confidence {
				best = p
			}
		}
	}
	return best
}

func (kb *KnowledgeBase) Learn(pattern FailurePattern) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	now := time.Now()
	for i := range kb.patterns {
		if kb.patterns[i].Pattern == pattern.Pattern && kb.patterns[i].Skill == pattern.Skill {
			kb.patterns[i].HitCount++
			kb.patterns[i].LastHit = now
			kb.patterns[i].Confidence = float64(kb.patterns[i].HitCount) / float64(kb.patterns[i].HitCount+1)
			if pattern.Solution != "" {
				kb.patterns[i].Solution = pattern.Solution
			}
			return
		}
	}

	pattern.HitCount = 1
	pattern.LastHit = now
	pattern.Confidence = 0.5
	kb.patterns = append(kb.patterns, pattern)
}

func (kb *KnowledgeBase) AddGraph(name string, graph *StrategyGraph) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.graphs[name] = graph
}

func (kb *KnowledgeBase) GetGraph(name string) *StrategyGraph {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return kb.graphs[name]
}