package agent

import "testing"

func TestBuildValuePriorityReportOrdersFalseRefuseFirst(t *testing.T) {
	samples := []EvalSample{
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-redis-ops", LabeledSkill: "ve-redis-ops", PolicyDecision: "AUTO", ShouldRefuse: false},
	}
	rep := BuildValuePriorityReport(samples, nil)
	if len(rep.Top) == 0 || rep.Top[0].Key != "ve-ecs-ops" || rep.Top[0].Reason != "high_false_refuse" {
		t.Fatalf("%+v", rep.Top)
	}
}

func TestBuildValuePriorityReportReasonTierWithValues(t *testing.T) {
	samples := []EvalSample{
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", PolicyDecision: "REFUSE", ShouldRefuse: false},
	}
	values := []ValueMetrics{
		{Success: false, MTTAMs: 9000},
		{Success: true, MTTAMs: 100},
	}
	rep := BuildValuePriorityReport(samples, values)
	if len(rep.Top) < 3 {
		t.Fatalf("want ≥3 items, got %+v", rep.Top)
	}
	if rep.Top[0].Reason != "high_false_refuse" {
		t.Fatalf("tier0=%s want high_false_refuse; full=%+v", rep.Top[0].Reason, rep.Top)
	}
	if rep.Top[1].Reason != "low_success" {
		t.Fatalf("tier1=%s want low_success; full=%+v", rep.Top[1].Reason, rep.Top)
	}
	if rep.Top[2].Reason != "high_mtta" {
		t.Fatalf("tier2=%s want high_mtta; full=%+v", rep.Top[2].Reason, rep.Top)
	}
}
