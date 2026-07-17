package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeGclTrace(t *testing.T, dir, name string, reqIDs []string) {
	t.Helper()
	tr := &Trace{
		TraceSchemaVersion: "v1",
		Skill:              "ve-ecs-ops",
		Request:            "test",
		Final:              Final{Status: "PASS"},
	}
	for i, id := range reqIDs {
		tr.Iterations = append(tr.Iterations, Iteration{Iter: i, RequestID: id})
	}
	b, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		t.Fatalf("marshal gcl trace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
		t.Fatalf("write gcl trace: %v", err)
	}
}

func writeIncidentTrace(t *testing.T, dir, name string, reqIDs []string) {
	t.Helper()
	it := incidentTraceLocal{TicketID: "T-1"}
	iter := struct {
		VeCalls []struct {
			RequestID string `json:"request_id"`
		} `json:"ve_calls"`
	}{}
	for _, id := range reqIDs {
		iter.VeCalls = append(iter.VeCalls, struct {
			RequestID string `json:"request_id"`
		}{RequestID: id})
	}
	it.Iterations = append(it.Iterations, iter)
	b, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		t.Fatalf("marshal incident trace: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
		t.Fatalf("write incident trace: %v", err)
	}
}

func TestLinkIndexOneToMany(t *testing.T) {
	root := t.TempDir()
	audit := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(audit, 0o755); err != nil {
		t.Fatal(err)
	}
	shared := "req-shared-123"
	writeGclTrace(t, audit, "gcl-trace-a.json", []string{shared, "req-only-gcl"})
	writeGclTrace(t, audit, "gcl-trace-b.json", []string{shared})
	writeIncidentTrace(t, audit, "incident-trace-x.json", []string{shared, "req-only-incident"})

	r, err := LinkIndex(root)
	if err != nil {
		t.Fatalf("LinkIndex: %v", err)
	}

	lt, ok := r.RequestIndex[shared]
	if !ok {
		t.Fatalf("request %q missing from index", shared)
	}
	if len(lt.GclTraces) != 2 {
		t.Errorf("GclTraces len = %d, want 2", len(lt.GclTraces))
	}
	if len(lt.IncidentTraces) != 1 {
		t.Errorf("IncidentTraces len = %d, want 1", len(lt.IncidentTraces))
	}
	if r.Counts.Linked != 1 {
		t.Errorf("Counts.Linked = %d, want 1", r.Counts.Linked)
	}
	if r.Counts.Gcl != 3 {
		t.Errorf("Counts.Gcl = %d, want 3", r.Counts.Gcl)
	}
	if r.Counts.Incident != 2 {
		t.Errorf("Counts.Incident = %d, want 2", r.Counts.Incident)
	}

	// Unlinked.
	if !containsStr(r.Unlinked.GclOnly, "req-only-gcl") {
		t.Errorf("GclOnly = %v, want req-only-gcl", r.Unlinked.GclOnly)
	}
	if !containsStr(r.Unlinked.IncidentOnly, "req-only-incident") {
		t.Errorf("IncidentOnly = %v, want req-only-incident", r.Unlinked.IncidentOnly)
	}
}

func TestLinkIndexMissingAuditDir(t *testing.T) {
	root := t.TempDir() // no audit-results created
	r, err := LinkIndex(root)
	if err != nil {
		t.Fatalf("LinkIndex: %v", err)
	}
	if len(r.RequestIndex) != 0 {
		t.Errorf("RequestIndex len = %d, want 0", len(r.RequestIndex))
	}
	if r.Counts.Gcl != 0 || r.Counts.Incident != 0 || r.Counts.Linked != 0 {
		t.Errorf("Counts = %+v, want all zero", r.Counts)
	}
}

func TestLinkIndexSkipsCorruptFiles(t *testing.T) {
	root := t.TempDir()
	audit := filepath.Join(root, "audit-results")
	if err := os.MkdirAll(audit, 0o755); err != nil {
		t.Fatal(err)
	}
	// Corrupt gcl trace (invalid JSON).
	if err := os.WriteFile(filepath.Join(audit, "gcl-trace-bad.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Corrupt incident trace (invalid JSON).
	if err := os.WriteFile(filepath.Join(audit, "incident-trace-bad.json"), []byte("@@@"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Valid gcl trace with a request id.
	writeGclTrace(t, audit, "gcl-trace-good.json", []string{"req-ok"})

	r, err := LinkIndex(root)
	if err != nil {
		t.Fatalf("LinkIndex: %v", err)
	}
	if r.Counts.Gcl != 1 {
		t.Fatalf("Counts.Gcl = %d, want 1 (corrupt files skipped)", r.Counts.Gcl)
	}
	if _, ok := r.RequestIndex["req-ok"]; !ok {
		t.Errorf("valid request req-ok missing from index")
	}
}

func TestPersistLink(t *testing.T) {
	root := t.TempDir()
	r := &LinkResult{
		TraceSchemaVersion: "v1-link",
		GeneratedAt:        "2026-01-01T00:00:00Z",
		RequestIndex:       map[string]*LinkedTrace{},
		Counts:             LinkCounts{Gcl: 0, Incident: 0, Linked: 0},
	}
	path, err := PersistLink(root, r)
	if err != nil {
		t.Fatalf("PersistLink: %v", err)
	}
	if !filepath.HasPrefix(path, filepath.Join(root, "audit-results", "trace-link-")) {
		t.Errorf("unexpected path %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("PersistLink did not write file: %v", err)
	}
	// Round-trip: re-read and verify schema version.
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got LinkResult
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal persisted link: %v", err)
	}
	if got.TraceSchemaVersion != "v1-link" {
		t.Errorf("TraceSchemaVersion = %q, want v1-link", got.TraceSchemaVersion)
	}
}

func containsStr(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
