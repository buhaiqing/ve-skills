package agent

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

const priorityTopN = 5

// PriorityItem is one ranked triage target for the weekly value report.
type PriorityItem struct {
	Key    string  `json:"key"`
	Reason string  `json:"reason"`
	Score  float64 `json:"score"`
}

// ValuePriorityReport ranks skills/paths by operational urgency.
type ValuePriorityReport struct {
	Top []PriorityItem `json:"top"`
}

// BuildValuePriorityReport ranks high false-refuse, low success, and high MTTA.
// Top N=5; ties break by Key lexicographic order.
func BuildValuePriorityReport(samples []EvalSample, values []ValueMetrics) ValuePriorityReport {
	var items []PriorityItem

	type refuseStat struct {
		refuse, falseRefuse int
	}
	bySkill := map[string]*refuseStat{}
	for _, s := range samples {
		skill := s.PredictedSkill
		if skill == "" {
			skill = s.LabeledSkill
		}
		if skill == "" {
			continue
		}
		st := bySkill[skill]
		if st == nil {
			st = &refuseStat{}
			bySkill[skill] = st
		}
		if s.PolicyDecision == "REFUSE" {
			st.refuse++
			if !s.ShouldRefuse {
				st.falseRefuse++
			}
		}
	}
	for skill, st := range bySkill {
		if st.refuse == 0 {
			continue
		}
		items = append(items, PriorityItem{
			Key:    skill,
			Reason: "high_false_refuse",
			Score:  float64(st.falseRefuse) / float64(st.refuse),
		})
	}

	if len(values) > 0 {
		fail := 0
		mttas := make([]int64, len(values))
		for i, v := range values {
			if !v.Success {
				fail++
			}
			mttas[i] = v.MTTAMs
		}
		items = append(items, PriorityItem{
			Key:    "_global",
			Reason: "low_success",
			Score:  float64(fail) / float64(len(values)),
		})
		items = append(items, PriorityItem{
			Key:    "_global",
			Reason: "high_mtta",
			Score:  float64(p50Int64(mttas)),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		return items[i].Key < items[j].Key
	})
	if len(items) > priorityTopN {
		items = items[:priorityTopN]
	}
	return ValuePriorityReport{Top: items}
}

// LoadValueMetricsJSONL reads one ValueMetrics object per non-empty line.
func LoadValueMetricsJSONL(path string) ([]ValueMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []ValueMetrics
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var m ValueMetrics
		if err := json.Unmarshal(line, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, sc.Err()
}

// WriteValuePriorityReport writes the priority report as indented JSON.
func WriteValuePriorityReport(path string, r ValuePriorityReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
