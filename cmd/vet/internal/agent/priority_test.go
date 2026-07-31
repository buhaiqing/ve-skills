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
