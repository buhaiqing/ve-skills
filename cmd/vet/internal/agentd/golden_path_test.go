package agentd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

func writeGoldenPathFixtures(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "execution-risk.md"), []byte(`# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | single | AUTO |
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "domain-allowlist.md"), []byte("# Domain\n\n## 1. Eligible skills\n\n`ve-ecs-ops`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func waitForValueJSON(t *testing.T, server *Server, root, runID string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	valuePath := filepath.Join(root, ".runtime", "agent", "runs", runID, "value.json")

	for time.Now().Before(deadline) {
		if _, err := os.Stat(valuePath); err == nil {
			return
		}

		state, err := agent.LoadState(root, runID)
		if err == nil && state != nil && state.Confirm != nil && state.Confirm.Decision == "ASK" {
			body := []byte(`{"confirmed": true, "confirmed_by": "DOPS-GOLDEN-TEST"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.confirmRunHandler(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("confirm ASK run: status=%d body=%s", w.Code, w.Body.String())
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.pool.Drain(ctx)

	if _, err := os.Stat(valuePath); err != nil {
		state, _ := agent.LoadState(root, runID)
		t.Fatalf("value.json missing after %v: run state=%+v", timeout, state)
	}
}

func TestGoldenPathIncidentToValueJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentd-golden-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	writeGoldenPathFixtures(t, tmpDir)
	server := NewServer(tmpDir, ":0", 2)

	payload := agent.IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "cpu>90%",
		TicketID:    "DOPS-GOLDEN-TEST",
		Source:      "httptest",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.createIncidentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create incident: status=%d body=%s", w.Code, w.Body.String())
	}

	var created map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	runID, _ := created["run_id"].(string)
	if runID == "" {
		t.Fatal("expected run_id in create response")
	}

	statePath := filepath.Join(tmpDir, ".runtime", "agent", "runs", runID, "state.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("run state not created: %v", err)
	}

	waitForValueJSON(t, server, tmpDir, runID, 90*time.Second)

	data, err := os.ReadFile(filepath.Join(tmpDir, ".runtime", "agent", "runs", runID, "value.json"))
	if err != nil {
		t.Fatalf("read value.json: %v", err)
	}

	var metrics agent.ValueMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("parse value.json: %v", err)
	}
	if metrics.RunID != runID {
		t.Errorf("value run_id=%q want %q", metrics.RunID, runID)
	}
	if metrics.TicketID != payload.TicketID {
		t.Errorf("value ticket_id=%q want %q", metrics.TicketID, payload.TicketID)
	}

	dash := server.aggregateStats()
	if dash.Value.Samples < 1 {
		t.Errorf("dashboard value samples=%d want >=1", dash.Value.Samples)
	}
}
