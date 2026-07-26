package strategy

import (
	"sort"
	"strings"
	"sync"
)

type StrategyNode struct {
	ID        string
	Symptom   string
	Operation string
	Condition string
	Priority  int
	Children  []*StrategyNode
	Enabled   bool
}

type StrategyGraph struct {
	nodes map[string]*StrategyNode
	root  *StrategyNode
	mu    sync.RWMutex
}

func NewStrategyGraph() *StrategyGraph {
	root := &StrategyNode{
		ID:      "root",
		Enabled: true,
	}
	return &StrategyGraph{
		nodes: make(map[string]*StrategyNode),
		root:  root,
	}
}

func (g *StrategyGraph) AddNode(node *StrategyNode) {
	g.mu.Lock()
	defer g.mu.Unlock()
	node.Enabled = true
	g.nodes[node.ID] = node
	g.root.Children = append(g.root.Children, node)
}

func (g *StrategyGraph) GetPath(symptom string) []*StrategyNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var path []*StrategyNode
	var dfs func(node *StrategyNode) bool
	dfs = func(node *StrategyNode) bool {
		if node.matchesSymptom(symptom) {
			path = append(path, node)
			return true
		}
		for _, child := range node.Children {
			if dfs(child) {
				path = append([]*StrategyNode{node}, path...)
				return true
			}
		}
		return false
	}
	dfs(g.root)
	return path
}

func (g *StrategyGraph) GeneratePlan(symptoms []string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var nodes []*StrategyNode
	for _, symptom := range symptoms {
		var dfs func(node *StrategyNode)
		dfs = func(node *StrategyNode) {
			if node.matchesSymptom(symptom) && node.Enabled {
				nodes = append(nodes, node)
			}
			for _, child := range node.Children {
				dfs(child)
			}
		}
		dfs(g.root)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Priority > nodes[j].Priority
	})

	var plan []string
	for _, n := range nodes {
		if n.Operation != "" {
			plan = append(plan, n.Operation)
		}
	}
	return plan
}

func (n *StrategyNode) matchesSymptom(symptom string) bool {
	return strings.Contains(strings.ToLower(n.Symptom), strings.ToLower(symptom))
}