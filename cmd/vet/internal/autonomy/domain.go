package autonomy

import (
	"sync"
)

type L4Level string

const (
	L1Basic        L4Level = "L1"
	L2Autonomous   L4Level = "L2"
	L3Semantic     L4Level = "L3"
	L4SelfEvolving L4Level = "L4"
)

type Policy struct {
	Name      string
	Condition string
	Action    string
	Priority  int
	Enabled   bool
}

type Action struct {
	Type      string
	Target    string
	Params    map[string]interface{}
	ApprovedBy string
}

type AutonomousDomain struct {
	Name       string
	Level      L4Level
	Skills     []string
	Policies   []Policy
	Actions    []Action
	Knowledge  map[string]string
	Children   []*AutonomousDomain
	Parent     *AutonomousDomain
	mu         sync.RWMutex
}

func NewAutonomousDomain(name string) *AutonomousDomain {
	return &AutonomousDomain{
		Name:      name,
		Level:     L1Basic,
		Skills:    make([]string, 0),
		Policies:  make([]Policy, 0),
		Actions:   make([]Action, 0),
		Knowledge: make(map[string]string),
		Children:  make([]*AutonomousDomain, 0),
	}
}

func (d *AutonomousDomain) AutoExpand(newSkill string, maturityScore float64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, s := range d.Skills {
		if s == newSkill {
			return false
		}
	}

	d.Skills = append(d.Skills, newSkill)

	if maturityScore >= 0.7 {
		if d.canExpandTo(nextLevel(d.Level)) {
			d.Level = nextLevel(d.Level)
			return true
		}
	}

	return false
}

func (d *AutonomousDomain) canExpandTo(target L4Level) bool {
	switch target {
	case L2Autonomous:
		return len(d.Policies) >= 1 && len(d.Actions) >= 1
	case L3Semantic:
		return len(d.Policies) >= 3 && len(d.Actions) >= 3 && len(d.Knowledge) >= 3
	case L4SelfEvolving:
		return len(d.Policies) >= 5 && len(d.Actions) >= 5 && len(d.Knowledge) >= 5
	default:
		return false
	}
}

func nextLevel(current L4Level) L4Level {
	switch current {
	case L1Basic:
		return L2Autonomous
	case L2Autonomous:
		return L3Semantic
	case L3Semantic:
		return L4SelfEvolving
	default:
		return current
	}
}

func (d *AutonomousDomain) EvaluateMaturity() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	policyScore := 0.0
	if len(d.Policies) >= 5 {
		policyScore = 0.3
	} else {
		policyScore = float64(len(d.Policies)) * 0.06
	}

	actionScore := 0.0
	if len(d.Actions) >= 5 {
		actionScore = 0.4
	} else {
		actionScore = float64(len(d.Actions)) * 0.08
	}

	knowledgeScore := 0.0
	if len(d.Knowledge) >= 5 {
		knowledgeScore = 0.3
	} else {
		knowledgeScore = float64(len(d.Knowledge)) * 0.06
	}

	return policyScore + actionScore + knowledgeScore
}

func (d *AutonomousDomain) CrossCoordinate(other *AutonomousDomain) []Action {
	d.mu.RLock()
	defer d.mu.RUnlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	overlap := make([]string, 0)
	for _, s1 := range d.Skills {
		for _, s2 := range other.Skills {
			if s1 == s2 {
				overlap = append(overlap, s1)
				break
			}
		}
	}

	if len(overlap) == 0 {
		return nil
	}

	actions := make([]Action, 0, len(overlap)*2)
	for _, skill := range overlap {
		actions = append(actions, Action{
			Type:   "heal",
			Target: skill,
			Params: map[string]interface{}{
				"source_domain": d.Name,
				"target_domain":  other.Name,
				"skill":          skill,
			},
			ApprovedBy: "cross_coordinate",
		})
		actions = append(actions, Action{
			Type:   "notify",
			Target: skill,
			Params: map[string]interface{}{
				"source_domain": d.Name,
				"target_domain":  other.Name,
				"skill":          skill,
			},
			ApprovedBy: "cross_coordinate",
		})
	}

	return actions
}

func (d *AutonomousDomain) Reconcile() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	anomalies := make([]string, 0)

	actionTargets := make(map[string]bool)
	for _, a := range d.Actions {
		actionTargets[a.Target] = true
	}

	policyActions := make(map[string]bool)
	for _, p := range d.Policies {
		policyActions[p.Action] = true
	}

	for _, s := range d.Skills {
		if !actionTargets[s] && !policyActions[s] {
			anomalies = append(anomalies, "orphan skill: "+s)
		}
	}

	for _, p := range d.Policies {
		if !p.Enabled {
			continue
		}
		hasAction := false
		for _, a := range d.Actions {
			if a.Type == p.Action {
				hasAction = true
				break
			}
		}
		if !hasAction {
			anomalies = append(anomalies, "policy without action: "+p.Name)
		}
	}

	return anomalies
}