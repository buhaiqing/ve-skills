package critic

import "testing"

func TestStructuralCritic(t *testing.T) {
	// success, no leak → all 1.0 except idempotency 0.5
	ok := StructuralCritic(GeneratorResult{Command: "ve ecs DescribeInstances", ExitCode: 0, ResultExcerpt: "ok"})
	if ok.Scores["correctness"] != 1.0 {
		t.Errorf("correctness want 1.0 got %v", ok.Scores["correctness"])
	}
	if ok.Scores["safety"] != 1.0 {
		t.Errorf("safety want 1.0 got %v", ok.Scores["safety"])
	}
	if ok.Blocking {
		t.Error("should not be blocking on clean success")
	}

	// leak → safety 0.0 + blocking
	leak := StructuralCritic(GeneratorResult{Command: "x", ExitCode: 0, ResultExcerpt: "SecretKey=EXAMPLE_VALUE"})
	if leak.Scores["safety"] != 0.0 {
		t.Errorf("safety want 0.0 got %v", leak.Scores["safety"])
	}
	if !leak.Blocking {
		t.Error("leak must be blocking")
	}

	// non-zero exit → correctness 0.0 + blocking
	fail := StructuralCritic(GeneratorResult{Command: "ve x", ExitCode: 1, ResultExcerpt: "err"})
	if fail.Scores["correctness"] != 0.0 {
		t.Errorf("correctness want 0.0 got %v", fail.Scores["correctness"])
	}
	if !fail.Blocking {
		t.Error("non-zero exit must be blocking")
	}
}

func TestDecide(t *testing.T) {
	if Decide(Scores{"safety": 0, "correctness": 1, "idempotency": 1, "traceability": 1, "spec_compliance": 1}) != "SAFETY_FAIL" {
		t.Error("safety 0 → SAFETY_FAIL")
	}
	if Decide(Scores{"safety": 1, "correctness": 0.5, "idempotency": 0.5, "traceability": 0.5, "spec_compliance": 0.5}) != "PASS" {
		t.Error("all at threshold → PASS")
	}
	if Decide(Scores{"safety": 1, "correctness": 0, "idempotency": 0.5, "traceability": 0.5, "spec_compliance": 0.5}) != "RETRY" {
		t.Error("correctness below threshold → RETRY")
	}
}

func TestValidatePayload(t *testing.T) {
	good := CriticResult{Scores: Scores{"correctness": 1, "safety": 1, "idempotency": 0.5, "traceability": 0.5, "spec_compliance": 0.5}, Suggestions: []string{"ok"}, Blocking: false}
	if errs := ValidatePayload(good); len(errs) != 0 {
		t.Errorf("expected valid payload, got %v", errs)
	}
	bad := CriticResult{Scores: Scores{"correctness": 0.7}}
	if errs := ValidatePayload(bad); len(errs) == 0 {
		t.Error("expected errors for out-of-range score + missing dims")
	}
}
