package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TicketStat holds per-ticket aggregate counters.
type TicketStat struct {
	VeCalls        int    `json:"ve_calls"`
	RequestIDs     int    `json:"request_ids"` // non-empty request_ids within this ticket (internal, not cross-link)
	PolicyDecision string `json:"policy_decision"`
}

// IncidentSummary is the output contract of the incident aggregation.
type IncidentSummary struct {
	TraceSchemaVersion string                `json:"trace_schema_version"` // "v1-incident"
	GeneratedAt        string                `json:"generated_at"`
	Totals             map[string]int        `json:"totals"` // tickets, ve_calls, request_ids_covered
	ByTicket           map[string]TicketStat `json:"by_ticket"`
	PolicyDecision     map[string]int        `json:"policy_decision"` // AUTO/ASK/REFUSE counts
}

// AggregateIncident scans audit-results/incident-trace-*.json under root and
// aggregates them by ticket_id. Corrupt files are skipped (not fatal). It is a
// pure function of root's audit-results directory and returns an
// IncidentSummary without touching the filesystem beyond reading. When no
// audit-results dir exists it returns a zero-valued summary (not an error).
func AggregateIncident(root string) *IncidentSummary {
	s := &IncidentSummary{
		TraceSchemaVersion: "v1-incident",
		GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
		Totals:             map[string]int{},
		ByTicket:           map[string]TicketStat{},
		PolicyDecision:     map[string]int{},
	}

	auditDir := filepath.Join(root, "audit-results")
	paths, _ := filepath.Glob(filepath.Join(auditDir, "incident-trace-*.json"))

	for _, p := range paths {
		it := parseIncidentTraceLocal(p)
		if it == nil {
			fmt.Fprintf(stderr, "WARN: skip %s\n", p)
			continue
		}
		tid := strings.TrimSpace(it.TicketID)
		if tid == "" {
			fmt.Fprintf(stderr, "WARN: skip %s (empty ticket_id)\n", p)
			continue
		}

		stat := s.ByTicket[tid]
		for _, iter := range it.Iterations {
			stat.VeCalls += len(iter.VeCalls)
			for _, call := range iter.VeCalls {
				if strings.TrimSpace(call.RequestID) != "" {
					stat.RequestIDs++
				}
			}
		}
		if strings.TrimSpace(it.PolicyDecision) != "" {
			stat.PolicyDecision = strings.TrimSpace(it.PolicyDecision)
			s.PolicyDecision[strings.TrimSpace(it.PolicyDecision)]++
		}
		s.ByTicket[tid] = stat
	}

	// Totals computed from the aggregated per-ticket stats.
	s.Totals["tickets"] = len(s.ByTicket)
	for _, st := range s.ByTicket {
		s.Totals["ve_calls"] += st.VeCalls
		s.Totals["request_ids_covered"] += st.RequestIDs
	}
	return s
}

// PersistIncident writes the IncidentSummary to
// audit-results/incident-summary-YYYYMMDD-HHMMSS.json and returns the path.
// Mirrors PersistTrace's style (UTC timestamp + MarshalIndent).
func PersistIncident(root string, s *IncidentSummary) (string, error) {
	outDir := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102-150405")
	path := filepath.Join(outDir, "incident-summary-"+ts+".json")
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// CmdIncident runs the incident-trace aggregation. Returns the process exit
// code. When audit-results is absent it prints an empty result and exits 0.
func CmdIncident(root string) int {
	s := AggregateIncident(root)
	incidentPath := ""
	if len(s.ByTicket) > 0 {
		p, perr := PersistIncident(root, s)
		if perr != nil {
			fmt.Fprintf(stderr, "ERROR: persist incident summary: %v\n", perr)
			return 1
		}
		incidentPath = p
	}
	result := map[string]any{
		"incident_path": incidentPath,
		"tickets":       len(s.ByTicket),
	}
	b, _ := json.Marshal(result)
	fmt.Println(string(b))
	return 0
}
