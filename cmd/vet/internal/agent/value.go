package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultBaselineManualMin is the assumed manual MTTA+MTTR baseline in minutes.
const DefaultBaselineManualMin = 30.0

// ValueMetrics holds MTTA/MTTR/labor-saved telemetry for one agent run.
type ValueMetrics struct {
	RunID             string  `json:"run_id"`
	TicketID          string  `json:"ticket_id,omitempty"`
	Success           bool    `json:"success"`
	PolicyDecision    string  `json:"policy_decision,omitempty"`
	AlertedAt         string  `json:"alerted_at"`
	StartedAt         string  `json:"started_at"`
	ResolvedAt        string  `json:"resolved_at"`
	MTTAMs            int64   `json:"mtta_ms"`
	MTTRMs            int64   `json:"mttr_ms"`
	AgentDurationMs   int64   `json:"agent_duration_ms"`
	LaborMinutesSaved float64 `json:"labor_minutes_saved"`
	BaselineManualMin float64 `json:"baseline_manual_min"`
}

// ValueInput is the clock inputs for ComputeValue.
type ValueInput struct {
	RunID             string
	TicketID          string
	PolicyDecision    string
	Success           bool
	AlertedAt         time.Time
	StartedAt         time.Time
	ResolvedAt        time.Time
	BaselineManualMin float64 // 0 → DefaultBaselineManualMin
}

// ComputeValue derives MTTA/MTTR/labor metrics from timestamps.
func ComputeValue(in ValueInput) ValueMetrics {
	baseline := in.BaselineManualMin
	if baseline <= 0 {
		baseline = DefaultBaselineManualMin
	}

	mtta := in.StartedAt.Sub(in.AlertedAt).Milliseconds()
	if mtta < 0 {
		mtta = 0
	}
	mttr := in.ResolvedAt.Sub(in.AlertedAt).Milliseconds()
	if mttr < 0 {
		mttr = 0
	}
	agentMs := in.ResolvedAt.Sub(in.StartedAt).Milliseconds()
	if agentMs < 0 {
		agentMs = 0
	}

	labor := 0.0
	if in.Success {
		labor = baseline - float64(agentMs)/60000.0
		if labor < 0 {
			labor = 0
		}
	}

	return ValueMetrics{
		RunID:             in.RunID,
		TicketID:          in.TicketID,
		Success:           in.Success,
		PolicyDecision:    in.PolicyDecision,
		AlertedAt:         in.AlertedAt.UTC().Format(time.RFC3339Nano),
		StartedAt:         in.StartedAt.UTC().Format(time.RFC3339Nano),
		ResolvedAt:        in.ResolvedAt.UTC().Format(time.RFC3339Nano),
		MTTAMs:            mtta,
		MTTRMs:            mttr,
		AgentDurationMs:   agentMs,
		LaborMinutesSaved: labor,
		BaselineManualMin: baseline,
	}
}

// PersistValue writes value.json under the run dir and appends a JSONL audit line.
func PersistValue(root string, m ValueMetrics) error {
	dir := runDir(root, m.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "value.json"), data, 0o644); err != nil {
		return err
	}

	auditDir := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(m)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(auditDir, "value-metrics.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// FormatValueComment builds a human-readable ticket comment body.
func FormatValueComment(m ValueMetrics) string {
	return fmt.Sprintf(
		"## Agent Value Metrics\n\n"+
			"- RunID: `%s`\n"+
			"- Success: %v\n"+
			"- Policy: %s\n"+
			"- MTTA: %d ms\n"+
			"- MTTR: %d ms\n"+
			"- Agent duration: %d ms\n"+
			"- Labor minutes saved: %.2f (baseline %.0f min)\n",
		m.RunID, m.Success, m.PolicyDecision,
		m.MTTAMs, m.MTTRMs, m.AgentDurationMs,
		m.LaborMinutesSaved, m.BaselineManualMin,
	)
}

// TicketWriter writes value metrics back to a ticket sink.
type TicketWriter interface {
	WriteValueComment(ticketID, body string) error
}

// FileTicketWriter writes <Dir>/<ticketID>.md.
type FileTicketWriter struct {
	Dir string
}

// WriteValueComment writes the comment body to Dir/ticketID.md.
func (w FileTicketWriter) WriteValueComment(ticketID, body string) error {
	if err := os.MkdirAll(w.Dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.Dir, ticketID+".md"), []byte(body), 0o644)
}
