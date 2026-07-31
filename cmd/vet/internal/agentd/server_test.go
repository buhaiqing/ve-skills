package agentd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

func setupTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "agentd-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	server := NewServer(tmpDir, ":0", 3)
	return server, tmpDir
}

func createRunState(t *testing.T, tmpDir, runID string, state *agent.RunState) {
	t.Helper()
	runDir := filepath.Join(tmpDir, ".runtime", "agent", "runs", runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	if err := os.WriteFile(filepath.Join(runDir, "state.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHealthHandler(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %v", resp["status"])
	}
	if resp["version"] != "0.2.0" {
		t.Errorf("expected version=0.2.0, got %v", resp["version"])
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetRunHandlerNotFound(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/nonexistent", nil)
	w := httptest.NewRecorder()

	server.getRunHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetRunHandlerSuccess(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "1234567890"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepTriage,
		Payload:     agent.IncidentPayload{ProductHint: "ecs", Symptom: "cpu>90%"},
	}
	createRunState(t, tmpDir, runID, state)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/"+runID, nil)
	w := httptest.NewRecorder()

	server.getRunHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp agent.RunState
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.RunID != runID {
		t.Errorf("expected run_id=%s, got %s", runID, resp.RunID)
	}
	if resp.Payload.ProductHint != "ecs" {
		t.Errorf("expected product=ecs, got %s", resp.Payload.ProductHint)
	}
}

func TestGetRunHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/runs/123", nil)
	w := httptest.NewRecorder()

	server.getRunHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetRunHandlerMissingRunID(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/", nil)
	w := httptest.NewRecorder()

	server.getRunHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetRunHandlerFromCache(t *testing.T) {
	server, _ := setupTestServer(t)

	runID := "555666777"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepExecute,
		Payload:     agent.IncidentPayload{ProductHint: "rds"},
	}
	server.setRunState(runID, state)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/"+runID, nil)
	w := httptest.NewRecorder()

	server.getRunHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp agent.RunState
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.CurrentStep != agent.StepExecute {
		t.Errorf("expected step=EXECUTE, got %s", resp.CurrentStep)
	}
}

func TestCreateIncidentHandler(t *testing.T) {
	server, _ := setupTestServer(t)

	payload := agent.IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "cpu>90%",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["run_id"] == nil || resp["run_id"] == "" {
		t.Error("expected run_id in response")
	}
	if resp["status"] != "queued" {
		t.Errorf("expected status=queued, got %v", resp["status"])
	}
}

func TestCreateIncidentHandlerMissingProduct(t *testing.T) {
	server, _ := setupTestServer(t)

	payload := agent.IncidentPayload{
		Symptom: "cpu>90%",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateIncidentHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestCreateIncidentHandlerInvalidJSON(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateIncidentHandlerSetsDefaults(t *testing.T) {
	server, _ := setupTestServer(t)

	payload := agent.IncidentPayload{
		ProductHint: "redis",
		Symptom:     "memory full",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateIncidentHandlerPoolFull(t *testing.T) {
	server, _ := setupTestServer(t)

	server.pool.activeCount = int64(server.pool.maxConcurrent)

	payload := agent.IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "down",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.createIncidentHandler(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	server.pool.activeCount = 0
}

func TestConfirmRunHandler(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "444555666"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepConfirm,
		Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		Confirm:     &agent.ConfirmResult{Decision: "ASK", Reason: "test"},
	}
	createRunState(t, tmpDir, runID, state)

	body := []byte(`{"confirmed": false, "comment": "not now"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, err := agent.LoadState(tmpDir, runID)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if updated.Confirm.Decision != "REFUSE" {
		t.Errorf("expected decision=REFUSE, got %s", updated.Confirm.Decision)
	}
	if updated.Confirm.Reason != "not now" {
		t.Errorf("expected reason='not now', got %s", updated.Confirm.Reason)
	}
}

func TestConfirmRunHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/123/confirm", nil)
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestConfirmRunHandlerInvalidJSON(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/123/confirm", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConfirmRunHandlerNotFound(t *testing.T) {
	server, _ := setupTestServer(t)

	body := []byte(`{"confirmed": true, "confirmed_by": "DOPS-999"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/nonexistent/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestConfirmRunHandlerMissingConfirmedBy(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "111222333"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepConfirm,
		Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		Confirm:     &agent.ConfirmResult{Decision: "ASK", Reason: "test"},
	}
	createRunState(t, tmpDir, runID, state)

	body := []byte(`{"confirmed": true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfirmRunHandlerPersistsConfirmedBy(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "persist-by-001"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepConfirm,
		Payload:     agent.IncidentPayload{ProductHint: "ecs", TicketID: "DOPS-1"},
		Confirm:     &agent.ConfirmResult{Decision: "ASK", Reason: "stub heal"},
		Result:      &agent.ExecuteResult{Success: false, ErrorClass: "policy_block"},
	}
	createRunState(t, tmpDir, runID, state)

	body := []byte(`{"confirmed":true,"confirmed_by":"ticket:DOPS-1|human:alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	updated, err := agent.LoadState(tmpDir, runID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Confirm.Decision != "AUTO" {
		t.Fatalf("Decision=%s want AUTO", updated.Confirm.Decision)
	}
	if updated.Confirm.ConfirmedBy != "ticket:DOPS-1|human:alice" {
		t.Fatalf("ConfirmedBy=%q", updated.Confirm.ConfirmedBy)
	}
	if updated.CurrentStep != agent.StepExecute {
		t.Fatalf("CurrentStep=%v want Execute", updated.CurrentStep)
	}
	if updated.Result != nil {
		t.Fatalf("Result should be cleared, got %+v", updated.Result)
	}
}

func TestConfirmRunHandlerMissingRunID(t *testing.T) {
	server, _ := setupTestServer(t)

	body := []byte(`{"confirmed": true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs//confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConfirmRunHandlerNilConfirm(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "777888999"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepDiagnose,
		Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		Confirm:     nil,
	}
	createRunState(t, tmpDir, runID, state)

	body := []byte(`{"confirmed": false}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.confirmRunHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestListRunsHandlerEmpty(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs", nil)
	w := httptest.NewRecorder()

	server.listRunsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["total"] != float64(0) {
		t.Errorf("expected total=0, got %v", resp["total"])
	}
}

func TestListRunsHandlerWithData(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	for i := 0; i < 3; i++ {
		runID := fmt.Sprintf("%d", 1000000000000000000+i)
		state := &agent.RunState{
			RunID:       runID,
			CurrentStep: agent.StepDone,
			Payload:     agent.IncidentPayload{ProductHint: "ecs"},
			Triage:      &agent.TriageResult{PrimarySkill: "ve-ecs-ops"},
		}
		createRunState(t, tmpDir, runID, state)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs", nil)
	w := httptest.NewRecorder()

	server.listRunsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["total"] != float64(3) {
		t.Errorf("expected total=3, got %v", resp["total"])
	}
}

func TestListRunsHandlerWithFilter(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	for i, step := range []agent.Step{agent.StepDone, agent.StepIngest} {
		runID := fmt.Sprintf("%d", 2000000000000000000+i)
		state := &agent.RunState{
			RunID:       runID,
			CurrentStep: step,
			Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		}
		createRunState(t, tmpDir, runID, state)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs?status=completed", nil)
	w := httptest.NewRecorder()

	server.listRunsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["total"] != float64(1) {
		t.Errorf("expected filtered total=1, got %v", resp["total"])
	}
}

func TestListRunsHandlerPagination(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs?page=5&limit=2", nil)
	w := httptest.NewRecorder()

	server.listRunsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["page"] != float64(5) {
		t.Errorf("expected page=5, got %v", resp["page"])
	}
	if resp["limit"] != float64(2) {
		t.Errorf("expected limit=2, got %v", resp["limit"])
	}
}

func TestListRunsHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", nil)
	w := httptest.NewRecorder()

	server.listRunsHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestDashboardHandler(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()

	server.dashboardHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected text/html, got %s", contentType)
	}
}

func TestDashboardHandlerMethodNotAllowed(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard", nil)
	w := httptest.NewRecorder()

	server.dashboardHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestPoolSubmit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	pool := NewPool(tmpDir, 2)

	payload := &agent.IncidentPayload{
		ProductHint: "ecs",
		Symptom:     "cpu>90%",
	}

	runID, err := pool.Submit(payload)
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	if runID == "" {
		t.Error("expected non-empty run_id")
	}

	statePath := filepath.Join(tmpDir, ".runtime", "agent", "runs", runID, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("state file not created")
	}
}

func TestPoolActiveCount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool-count-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	pool := NewPool(tmpDir, 5)

	if pool.ActiveCount() != 0 {
		t.Errorf("expected active count 0, got %d", pool.ActiveCount())
	}
}

func TestPoolDrain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool-drain-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	pool := NewPool(tmpDir, 2)

	ctx := context.Background()
	if err := pool.Drain(ctx); err != nil {
		t.Errorf("drain with no tasks failed: %v", err)
	}
}

func TestPoolDrainTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool-drain-timeout-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	pool := NewPool(tmpDir, 1)
	pool.wg.Add(1)
	pool.sem <- struct{}{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = pool.Drain(ctx)
	if err == nil {
		t.Error("expected timeout error")
	}

	<-pool.sem
	pool.wg.Done()
}

func TestExtractRunID(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   string
	}{
		{"/api/v1/runs/12345", "/api/v1/runs/", "12345"},
		{"/api/v1/runs/12345/confirm", "/api/v1/runs/", "12345"},
		{"/api/v1/runs/", "/api/v1/runs/", ""},
	}

	for _, tt := range tests {
		got := extractRunID(tt.path, tt.prefix)
		if got != tt.want {
			t.Errorf("extractRunID(%q, %q) = %q, want %q", tt.path, tt.prefix, got, tt.want)
		}
	}
}

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		state *agent.RunState
		want  string
	}{
		{&agent.RunState{CurrentStep: agent.StepIngest}, "running"},
		{&agent.RunState{CurrentStep: agent.StepDone}, "completed"},
		{&agent.RunState{CurrentStep: agent.StepConfirm, Confirm: nil}, "paused"},
		{&agent.RunState{Error: "something failed"}, "failed"},
	}

	for _, tt := range tests {
		got := determineStatus(tt.state)
		if got != tt.want {
			t.Errorf("determineStatus() = %q, want %q", got, tt.want)
		}
	}
}

func TestExtractTimeFromRunID(t *testing.T) {
	tests := []struct {
		runID   string
		wantNon bool
	}{
		{"1700000000000000000", true},
		{"not-a-number", false},
		{"", false},
	}

	for _, tt := range tests {
		got := extractTimeFromRunID(tt.runID)
		if (got != "") != tt.wantNon {
			t.Errorf("extractTimeFromRunID(%q) = %q, wantNonEmpty=%v", tt.runID, got, tt.wantNon)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{1234, "1234"},
	}

	for _, tt := range tests {
		got := formatInt(tt.n)
		if got != tt.want {
			t.Errorf("formatInt(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		f    float64
		want string
	}{
		{0.0, "0%"},
		{1.0, "100%"},
		{0.85, "85%"},
	}

	for _, tt := range tests {
		got := formatPercent(tt.f)
		if got != tt.want {
			t.Errorf("formatPercent(%f) = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0ms"},
		{500, "500ms"},
		{1000, "1s"},
		{2500, "2s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.ms)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func TestParseRunTimestamp(t *testing.T) {
	ts, err := parseRunTimestamp("1700000000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.IsZero() {
		t.Error("expected non-zero time")
	}

	ts2, err := parseRunTimestamp("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ts2.IsZero() {
		t.Error("expected zero time for non-numeric run ID")
	}
}

func TestRunsTable(t *testing.T) {
	runs := []RunSummary{
		{RunID: "111", Status: "running", Product: "ecs", Step: "TRIAGE", CreatedAt: "2026-01-01"},
		{RunID: "222", Status: "completed", Product: "rds", Step: "DONE", CreatedAt: "2026-01-02"},
	}

	got := runsTable(runs)
	if len(got) == 0 {
		t.Error("expected non-empty table HTML")
	}
	if !strings.Contains(got, "111") || !strings.Contains(got, "222") {
		t.Error("expected both run IDs in table")
	}
}

func TestRunsTableEmpty(t *testing.T) {
	got := runsTable([]RunSummary{})
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSkillTable(t *testing.T) {
	skills := map[string]int{"ve-ecs-ops": 5, "ve-rds-ops": 3}
	got := skillTable(skills)
	if len(got) == 0 {
		t.Error("expected non-empty table HTML")
	}
	if !strings.Contains(got, "ve-ecs-ops") || !strings.Contains(got, "ve-rds-ops") {
		t.Error("expected both skills in table")
	}
}

func TestSkillTableEmpty(t *testing.T) {
	got := skillTable(map[string]int{})
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestRenderRunsTable(t *testing.T) {
	runs := []RunSummary{{RunID: "x", Status: "running", Product: "ecs", Step: "INGEST"}}
	got := renderRunsTable(runs)
	if !strings.Contains(got, "x") {
		t.Error("expected run ID in rendered table")
	}
}

func TestRenderSkillTable(t *testing.T) {
	got := renderSkillTable(map[string]int{"ve-ecs-ops": 1})
	if !strings.Contains(got, "ve-ecs-ops") {
		t.Error("expected skill name in rendered table")
	}
}

func TestRenderDashboard(t *testing.T) {
	stats := &DashboardStats{
		TotalRuns:     10,
		SuccessRate:   0.9,
		AvgDurationMs: 1500,
		ActiveRuns:    2,
		BySkill:       map[string]int{"ve-ecs-ops": 5},
		RecentRuns: []RunSummary{
			{RunID: "123", Status: "completed", Product: "ecs", Step: "DONE", CreatedAt: "2026-01-01"},
		},
	}

	got := renderDashboard(stats)
	if len(got) == 0 {
		t.Error("expected non-empty dashboard HTML")
	}
	if !strings.Contains(got, "Agent Dashboard") {
		t.Error("expected title in dashboard")
	}
}

func TestAggregateStatsWithData(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := fmt.Sprintf("%d", 3000000000000000000)
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepDone,
		Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		Triage:      &agent.TriageResult{PrimarySkill: "ve-ecs-ops", Confidence: "high"},
	}
	createRunState(t, tmpDir, runID, state)

	runID2 := fmt.Sprintf("%d", 3000000000000000001)
	state2 := &agent.RunState{
		RunID:       runID2,
		CurrentStep: agent.StepTriage,
		Payload:     agent.IncidentPayload{ProductHint: "rds"},
	}
	createRunState(t, tmpDir, runID2, state2)

	stats := server.aggregateStats()

	if stats.TotalRuns != 2 {
		t.Errorf("expected total_runs=2, got %d", stats.TotalRuns)
	}
	if stats.ActiveRuns != 1 {
		t.Errorf("expected active_runs=1, got %d", stats.ActiveRuns)
	}
	if stats.BySkill["ve-ecs-ops"] != 1 {
		t.Errorf("expected ve-ecs-ops count=1, got %d", stats.BySkill["ve-ecs-ops"])
	}
	if len(stats.RecentRuns) != 2 {
		t.Errorf("expected 2 recent runs, got %d", len(stats.RecentRuns))
	}
}

func TestAggregateStatsNoRunsDir(t *testing.T) {
	server, _ := setupTestServer(t)

	stats := server.aggregateStats()
	if stats.TotalRuns != 0 {
		t.Errorf("expected 0 runs, got %d", stats.TotalRuns)
	}
}

func TestGetSetRunState(t *testing.T) {
	server, _ := setupTestServer(t)

	if s := server.getRunState("nonexistent"); s != nil {
		t.Errorf("expected nil, got %v", s)
	}

	state := &agent.RunState{RunID: "abc", CurrentStep: agent.StepTriage}
	server.setRunState("abc", state)

	got := server.getRunState("abc")
	if got == nil {
		t.Fatal("expected non-nil state")
	}
	if got.RunID != "abc" {
		t.Errorf("expected run_id=abc, got %s", got.RunID)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "bad request")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != "bad request" {
		t.Errorf("expected error='bad request', got %q", resp["error"])
	}
}

func TestRouteByMethodConfirm(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "999888777"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepConfirm,
		Payload:     agent.IncidentPayload{ProductHint: "ecs"},
		Confirm:     &agent.ConfirmResult{Decision: "ASK", Reason: "destructive"},
	}
	createRunState(t, tmpDir, runID, state)

	body := []byte(`{"confirmed": false, "comment": "rejected"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs/"+runID+"/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.routeByMethod(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "running" {
		t.Errorf("expected status=running, got %v", resp["status"])
	}
}

func TestRouteByMethodGetRun(t *testing.T) {
	server, tmpDir := setupTestServer(t)

	runID := "111222333"
	state := &agent.RunState{
		RunID:       runID,
		CurrentStep: agent.StepTriage,
		Payload:     agent.IncidentPayload{ProductHint: "rds"},
	}
	createRunState(t, tmpDir, runID, state)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/runs/"+runID, nil)
	w := httptest.NewRecorder()

	server.routeByMethod(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
