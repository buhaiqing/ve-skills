package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EvalSample struct {
	PredictedSkill string `json:"predicted_skill"`
	LabeledSkill   string `json:"labeled_skill"`
	GCLFirstPass   bool   `json:"gcl_first_pass"`
	PolicyDecision string `json:"policy_decision"`
	ShouldRefuse   bool   `json:"should_refuse"`
}

type EvalReportInput struct {
	Samples []EvalSample `json:"samples"`
}

type EvalReport struct {
	TriageTop1Accuracy float64 `json:"triage_top1_accuracy"`
	GCLFirstPassRate   float64 `json:"gcl_first_pass_rate"`
	FalseRefuseRate    float64 `json:"false_refuse_rate"`
	Samples            int     `json:"samples"`
}

func BuildEvalReport(in EvalReportInput) EvalReport {
	r := EvalReport{Samples: len(in.Samples)}
	var triageOK, triageN, gclOK, refuseN, falseRefuse int
	for _, s := range in.Samples {
		if s.LabeledSkill != "" {
			triageN++
			if s.PredictedSkill == s.LabeledSkill {
				triageOK++
			}
		}
		if s.GCLFirstPass {
			gclOK++
		}
		if s.PolicyDecision == "REFUSE" {
			refuseN++
			if !s.ShouldRefuse {
				falseRefuse++
			}
		}
	}
	if triageN > 0 {
		r.TriageTop1Accuracy = float64(triageOK) / float64(triageN)
	}
	if r.Samples > 0 {
		r.GCLFirstPassRate = float64(gclOK) / float64(r.Samples)
	}
	if refuseN > 0 {
		r.FalseRefuseRate = float64(falseRefuse) / float64(refuseN)
	}
	return r
}

func WriteEvalReport(path string, r EvalReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadEvalSamples(path string) (EvalReportInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EvalReportInput{}, err
	}
	var in EvalReportInput
	if err := json.Unmarshal(data, &in); err != nil {
		return EvalReportInput{}, err
	}
	return in, nil
}
