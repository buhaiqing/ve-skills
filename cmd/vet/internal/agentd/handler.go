package agentd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

// healthHandler returns the health status of the server.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := map[string]interface{}{
		"status":         "healthy",
		"version":        "0.2.0",
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
		"active_runs":    s.pool.ActiveCount(),
	}
	writeJSON(w, http.StatusOK, resp)
}

// getRunHandler returns the state of a specific run.
func (s *Server) getRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	runID := extractRunID(r.URL.Path, "/api/v1/runs/")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "missing run_id")
		return
	}

	// Try in-memory first
	state := s.getRunState(runID)
	if state == nil {
		// Try loading from disk
		var err error
		state, err = agent.LoadState(s.root, runID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load state")
			return
		}
	}

	if state == nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// confirmRunHandler handles human confirmation for ASK-class operations.
func (s *Server) confirmRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	runID := extractRunID(r.URL.Path, "/api/v1/runs/")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "missing run_id")
		return
	}

	var req struct {
		Confirmed   bool   `json:"confirmed"`
		ConfirmedBy string `json:"confirmed_by"`
		Comment     string `json:"comment,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Confirmed && strings.TrimSpace(req.ConfirmedBy) == "" {
		writeError(w, http.StatusBadRequest, "confirmed_by required when confirmed=true")
		return
	}

	// Load state
	state, err := agent.LoadState(s.root, runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load state")
		return
	}
	if state == nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	// Update confirm result and reset for Execute resume (ADR-0006).
	if state.Confirm != nil {
		if req.Confirmed {
			state.Confirm.Decision = "AUTO"
			state.Confirm.ConfirmedBy = req.ConfirmedBy
			state.Confirm.Reason = "human confirmed"
			state.CurrentStep = agent.StepExecute
			state.Result = nil
		} else {
			state.Confirm.Decision = "REFUSE"
			state.Confirm.Reason = req.Comment
		}
	} else if req.Confirmed {
		writeError(w, http.StatusConflict, "run has no confirm decision to authorize")
		return
	}

	// Save updated state
	if err := agent.SaveState(s.root, state); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save state")
		return
	}

	// Resume execution if confirmed
	if req.Confirmed {
		s.pool.wg.Add(1)
		go func() {
			defer s.pool.wg.Done()
			s.pool.sem <- struct{}{}
			defer func() { <-s.pool.sem }()
			agent.Run(s.root, &state.Payload, runID)
		}()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run_id":  runID,
		"status":  "running",
		"message": "confirmation received",
	})
}

// createIncidentHandler creates a new agent run from an incident.
func (s *Server) createIncidentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload agent.IncidentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate required fields
	if payload.ProductHint == "" {
		writeError(w, http.StatusBadRequest, "product_hint is required")
		return
	}

	// Set defaults
	if payload.Source == "" {
		payload.Source = "api"
	}
	if payload.RawInput == "" {
		payload.RawInput = fmt.Sprintf("product=%s symptom=%s", payload.ProductHint, payload.Symptom)
	}

	// Check if pool is full (simple queue check)
	if s.pool.ActiveCount() >= s.pool.maxConcurrent {
		w.Header().Set("Retry-After", "5")
		writeError(w, http.StatusTooManyRequests, "server busy, try again later")
		return
	}

	// Submit the incident
	runID, err := s.pool.Submit(&payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create run: %v", err))
		return
	}

	// Cache the state
	state, _ := agent.LoadState(s.root, runID)
	if state != nil {
		s.setRunState(runID, state)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"run_id":  runID,
		"status":  "queued",
		"message": "incident received, run created",
	})
}

// listRunsHandler lists all runs with optional filtering.
func (s *Server) listRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Parse query parameters
	q := r.URL.Query()
	statusFilter := q.Get("status")
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Scan runs directory
	runDir := filepath.Join(s.root, ".runtime", "agent", "runs")
	entries, err := os.ReadDir(runDir)
	if err != nil {
		// Directory might not exist yet
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"runs":  []interface{}{},
			"total": 0,
			"page":  page,
			"limit": limit,
		})
		return
	}

	var allRuns []map[string]interface{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runID := entry.Name()
		state, err := agent.LoadState(s.root, runID)
		if err != nil || state == nil {
			continue
		}

		// Apply status filter
		runStatus := determineStatus(state)
		if statusFilter != "" && runStatus != statusFilter {
			continue
		}

		allRuns = append(allRuns, map[string]interface{}{
			"run_id":       runID,
			"status":       runStatus,
			"current_step": state.CurrentStep.String(),
			"product":      state.Payload.ProductHint,
			"created_at":   extractTimeFromRunID(runID),
		})
	}

	// Apply pagination
	total := len(allRuns)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"runs":  allRuns[start:end],
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// determineStatus infers the run status from state.
func determineStatus(state *agent.RunState) string {
	if state.Error != "" {
		return "failed"
	}
	if state.CurrentStep == agent.StepDone {
		return "completed"
	}
	if state.CurrentStep == agent.StepConfirm && state.Confirm == nil {
		return "paused"
	}
	return "running"
}

// extractTimeFromRunID extracts an approximate time from a nanosecond timestamp run ID.
func extractTimeFromRunID(runID string) string {
	ts, err := strconv.ParseInt(runID, 10, 64)
	if err != nil {
		return ""
	}
	t := time.Unix(0, ts)
	return t.UTC().Format(time.RFC3339)
}

// extractRunID extracts the run ID from a URL path.
func extractRunID(path, prefix string) string {
	rest := strings.TrimPrefix(path, prefix)
	// Run ID is everything up to the next / or end of string
	if idx := strings.Index(rest, "/"); idx != -1 {
		return rest[:idx]
	}
	return rest
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
