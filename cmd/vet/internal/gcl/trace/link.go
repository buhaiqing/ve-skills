package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// incidentTraceLocal is the minimal local view of an incident-loop trace
// (audit-results/incident-trace-*.json). We deliberately do NOT import the
// internal/check/trace package to avoid coupling the two schemas; this struct
// only captures the fields needed to build the request_id index.
type incidentTraceLocal struct {
	TicketID   string `json:"ticket_id"`
	Iterations []struct {
		VeCalls []struct {
			RequestID string `json:"request_id"`
		} `json:"ve_calls"`
	} `json:"iterations"`
}

// LinkedTrace holds the set of trace files (relative to root) that share one
// cloud request_id.
type LinkedTrace struct {
	GclTraces      []string `json:"gcl_traces"`
	IncidentTraces []string `json:"incident_traces"`
}

// UnlinkedSummary records request_ids that appear in only one trace family.
type UnlinkedSummary struct {
	GclOnly      []string `json:"gcl_only"`
	IncidentOnly []string `json:"incident_only"`
}

// LinkCounts summarizes the index build.
type LinkCounts struct {
	Gcl      int `json:"gcl"`
	Incident int `json:"incident"`
	Linked   int `json:"linked"`
}

// LinkResult is the output contract of the link aggregation.
type LinkResult struct {
	TraceSchemaVersion string          `json:"trace_schema_version"` // "v1-link"
	GeneratedAt        string          `json:"generated_at"`
	RequestIndex       map[string]*LinkedTrace `json:"request_index"`
	Unlinked           UnlinkedSummary `json:"unlinked"`
	Counts             LinkCounts      `json:"counts"`
}

// LinkIndex scans audit-results/gcl-trace-*.json and incident-trace-*.json
// under root, building a request_id → LinkedTrace index. Parse failures are
// skipped (not fatal). Only non-empty request_ids participate. It is a pure
// function of root's audit-results directory and returns a LinkResult without
// touching the filesystem beyond reading.
func LinkIndex(root string) (*LinkResult, error) {
	auditDir := filepath.Join(root, "audit-results")
	if _, err := os.Stat(auditDir); err != nil {
		// No audit-results dir → empty result, not an error.
		return &LinkResult{
			TraceSchemaVersion: "v1-link",
			GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
			RequestIndex:       map[string]*LinkedTrace{},
		}, nil
	}

	gclPaths, _ := filepath.Glob(filepath.Join(auditDir, "gcl-trace-*.json"))
	incidentPaths, _ := filepath.Glob(filepath.Join(auditDir, "incident-trace-*.json"))

	idx := map[string]*LinkedTrace{}
	counts := LinkCounts{}

	for _, p := range gclPaths {
		t := ParseTrace(p)
		if t == nil {
			continue
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}
		for _, it := range t.Iterations {
			id := strings.TrimSpace(it.RequestID)
			if id == "" {
				continue
			}
			lt, ok := idx[id]
			if !ok {
				lt = &LinkedTrace{}
				idx[id] = lt
			}
			lt.GclTraces = append(lt.GclTraces, rel)
			counts.Gcl++
		}
	}

	for _, p := range incidentPaths {
		it := parseIncidentTraceLocal(p)
		if it == nil {
			continue
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}
		for _, iter := range it.Iterations {
			for _, call := range iter.VeCalls {
				id := strings.TrimSpace(call.RequestID)
				if id == "" {
					continue
				}
				lt, ok := idx[id]
				if !ok {
					lt = &LinkedTrace{}
					idx[id] = lt
				}
				lt.IncidentTraces = append(lt.IncidentTraces, rel)
				counts.Incident++
			}
		}
	}

	// Linked = request_ids present in both families.
	for _, lt := range idx {
		if len(lt.GclTraces) > 0 && len(lt.IncidentTraces) > 0 {
			counts.Linked++
		}
	}

	unlinked := UnlinkedSummary{}
	for id, lt := range idx {
		if len(lt.GclTraces) > 0 && len(lt.IncidentTraces) > 0 {
			continue
		}
		if len(lt.GclTraces) > 0 {
			unlinked.GclOnly = append(unlinked.GclOnly, id)
		} else if len(lt.IncidentTraces) > 0 {
			unlinked.IncidentOnly = append(unlinked.IncidentOnly, id)
		}
	}
	sort.Strings(unlinked.GclOnly)
	sort.Strings(unlinked.IncidentOnly)

	return &LinkResult{
		TraceSchemaVersion: "v1-link",
		GeneratedAt:        time.Now().UTC().Format(time.RFC3339),
		RequestIndex:       idx,
		Unlinked:           unlinked,
		Counts:             counts,
	}, nil
}

// parseIncidentTraceLocal parses an incident trace with the minimal local
// struct. Returns nil on any read/parse error so callers can skip bad files.
func parseIncidentTraceLocal(path string) *incidentTraceLocal {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var it incidentTraceLocal
	if err := json.Unmarshal(data, &it); err != nil {
		return nil
	}
	return &it
}

// PersistLink writes the LinkResult to
// audit-results/trace-link-YYYYMMDD-HHMMSS.json and returns the path. Mirrors
// PersistTrace's style (UTC timestamp + MarshalIndent).
func PersistLink(root string, r *LinkResult) (string, error) {
	outDir := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102-150405")
	path := filepath.Join(outDir, "trace-link-"+ts+".json")
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// CmdLink runs the trace-link aggregation. Returns the process exit code.
func CmdLink(root string) int {
	r, err := LinkIndex(root)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: link index: %v\n", err)
		return 1
	}
	linkPath := ""
	if r.Counts.Gcl > 0 || r.Counts.Incident > 0 {
		p, perr := PersistLink(root, r)
		if perr != nil {
			fmt.Fprintf(stderr, "ERROR: persist link: %v\n", perr)
			return 1
		}
		linkPath = p
	}
	result := map[string]any{
		"link_path": linkPath,
		"linked":    r.Counts.Linked,
		"gcl":       r.Counts.Gcl,
		"incident":  r.Counts.Incident,
	}
	b, _ := json.Marshal(result)
	fmt.Println(string(b))
	return 0
}
