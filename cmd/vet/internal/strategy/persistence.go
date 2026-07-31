package strategy

import (
	"encoding/json"
	"os"
	"time"
)

type patternState struct {
	Pattern    string  `json:"pattern"`
	Skill      string  `json:"skill"`
	Solution   string  `json:"solution"`
	HitCount   int     `json:"hit_count"`
	LastHit    string  `json:"last_hit"`
	Confidence float64 `json:"confidence"`
}

type nodeState struct {
	ID          string   `json:"id"`
	Symptom     string   `json:"symptom"`
	Operation   string   `json:"operation"`
	Condition   string   `json:"condition"`
	Priority    int      `json:"priority"`
	Enabled     bool     `json:"enabled"`
	ChildrenIDs []string `json:"children_ids"`
}

type kbState struct {
	Patterns []patternState      `json:"patterns"`
	Graphs   map[string][]nodeState `json:"graphs"`
}

func patternsToState(ps []FailurePattern) []patternState {
	ss := make([]patternState, len(ps))
	for i, p := range ps {
		ss[i] = patternState{
			Pattern:    p.Pattern,
			Skill:      p.Skill,
			Solution:   p.Solution,
			HitCount:   p.HitCount,
			LastHit:    p.LastHit.Format(time.RFC3339Nano),
			Confidence: p.Confidence,
		}
	}
	return ss
}

func stateToPatterns(ss []patternState) []FailurePattern {
	ps := make([]FailurePattern, len(ss))
	for i, s := range ss {
		t, err := time.Parse(time.RFC3339Nano, s.LastHit)
		if err != nil {
			t = time.Time{}
		}
		ps[i] = FailurePattern{
			Pattern:    s.Pattern,
			Skill:      s.Skill,
			Solution:   s.Solution,
			HitCount:   s.HitCount,
			LastHit:    t,
			Confidence: s.Confidence,
		}
	}
	return ps
}

func walkNodes(root *StrategyNode) []*StrategyNode {
	visited := make(map[string]bool)
	var result []*StrategyNode
	var dfs func(node *StrategyNode)
	dfs = func(node *StrategyNode) {
		if node == nil || visited[node.ID] {
			return
		}
		visited[node.ID] = true
		result = append(result, node)
		for _, child := range node.Children {
			dfs(child)
		}
	}
	dfs(root)
	return result
}

func graphsToState(graphs map[string]*StrategyGraph) map[string][]nodeState {
	result := make(map[string][]nodeState)
	for name, g := range graphs {
		g.mu.RLock()
		nodes := walkNodes(g.root)
		nodeStates := make([]nodeState, len(nodes))
		for i, n := range nodes {
			childrenIDs := make([]string, len(n.Children))
			for j, c := range n.Children {
				childrenIDs[j] = c.ID
			}
			nodeStates[i] = nodeState{
				ID:          n.ID,
				Symptom:     n.Symptom,
				Operation:   n.Operation,
				Condition:   n.Condition,
				Priority:    n.Priority,
				Enabled:     n.Enabled,
				ChildrenIDs: childrenIDs,
			}
		}
		g.mu.RUnlock()
		result[name] = nodeStates
	}
	return result
}

func stateToGraphs(saved map[string][]nodeState) map[string]*StrategyGraph {
	result := make(map[string]*StrategyGraph)
	for name, nodeStates := range saved {
		g := NewStrategyGraph()
		nodesMap := make(map[string]*StrategyNode)
		incoming := make(map[string]int)

		for _, ns := range nodeStates {
			node := &StrategyNode{
				ID:        ns.ID,
				Symptom:   ns.Symptom,
				Operation: ns.Operation,
				Condition: ns.Condition,
				Priority:  ns.Priority,
				Enabled:   ns.Enabled,
				Children:  nil,
			}
			nodesMap[ns.ID] = node
			incoming[ns.ID] = 0
		}

		for _, ns := range nodeStates {
			parent := nodesMap[ns.ID]
			if parent == nil {
				continue
			}
			children := make([]*StrategyNode, 0, len(ns.ChildrenIDs))
			for _, cid := range ns.ChildrenIDs {
				if child, ok := nodesMap[cid]; ok {
					children = append(children, child)
					incoming[cid]++
				}
			}
			parent.Children = children
		}

		var root *StrategyNode
		for _, ns := range nodeStates {
			if incoming[ns.ID] == 0 {
				root = nodesMap[ns.ID]
				break
			}
		}
		if root == nil && len(nodeStates) > 0 {
			root = nodesMap[nodeStates[0].ID]
		}

		g.nodes = nodesMap
		if root != nil {
			g.root = root
		}
		result[name] = g
	}
	return result
}

func (kb *KnowledgeBase) Save(path string) error {
	kb.mu.RLock()
	state := kbState{
		Patterns: patternsToState(kb.patterns),
		Graphs:   graphsToState(kb.graphs),
	}
	kb.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func (kb *KnowledgeBase) Load(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state kbState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	patterns := stateToPatterns(state.Patterns)
	graphs := stateToGraphs(state.Graphs)

	kb.mu.Lock()
	kb.patterns = patterns
	kb.graphs = graphs
	kb.mu.Unlock()

	return nil
}
