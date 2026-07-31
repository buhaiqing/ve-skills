package agent

import (
	"strings"
	"testing"
)

func TestASKWithoutConfirmedByPausesBeforeExecute(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	payload := &IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "mysql connection pool exhausted",
		TicketID:    "DOPS-ASK-PAUSE",
	}
	res := RunDry(root, payload, "ask-pause-1")
	if res.Success {
		t.Fatal("ASK without ConfirmedBy must not succeed")
	}
	if !strings.Contains(res.Error, "awaiting human confirmation") {
		t.Fatalf("error=%q", res.Error)
	}
	if res.FinalStep != StepConfirm {
		t.Fatalf("FinalStep=%v want Confirm", res.FinalStep)
	}
	st, err := LoadState(root, "ask-pause-1")
	if err != nil || st == nil {
		t.Fatalf("LoadState: %v %#v", err, st)
	}
	if st.CurrentStep != StepConfirm {
		t.Fatalf("CurrentStep=%v want Confirm", st.CurrentStep)
	}
	if st.Confirm == nil || st.Confirm.Decision != "ASK" {
		t.Fatalf("Confirm=%+v", st.Confirm)
	}
	if st.Result != nil {
		t.Fatalf("Result should be nil while awaiting confirm, got %+v", st.Result)
	}
}

func TestASKConfirmResumeReachesExecute(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	payload := &IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "mysql connection pool exhausted",
		TicketID:    "DOPS-ASK-RESUME",
	}
	runID := "ask-resume-1"
	paused := RunDry(root, payload, runID)
	if paused.Success || !strings.Contains(paused.Error, "awaiting human confirmation") {
		t.Fatalf("expected pause, got %+v", paused)
	}

	st, err := LoadState(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	st.Confirm.Decision = "AUTO"
	st.Confirm.ConfirmedBy = "ticket:DOPS-ASK-RESUME|human:test"
	st.CurrentStep = StepExecute
	st.Result = nil
	if err := SaveState(root, st); err != nil {
		t.Fatal(err)
	}

	resumed := RunDry(root, payload, runID)
	if !resumed.Success {
		t.Fatalf("resume dry-run should succeed after confirm, got %+v", resumed)
	}
	st2, err := LoadState(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	if st2.Confirm == nil || st2.Confirm.ConfirmedBy == "" {
		t.Fatalf("ConfirmedBy lost on resume: %+v", st2.Confirm)
	}
	if st2.CurrentStep < StepExecute {
		t.Fatalf("CurrentStep=%v should have reached Execute+", st2.CurrentStep)
	}
}
