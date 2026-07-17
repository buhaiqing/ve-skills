package heal

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMetrics_RecordAndRates(t *testing.T) {
	var m Metrics
	// 8 ok + 2 fail = 10 total → 80% success.
	for i := 0; i < 8; i++ {
		m.Record(HealEvent{Result: "ok", DurationMs: 1000})
	}
	for i := 0; i < 2; i++ {
		m.Record(HealEvent{Result: "fail", DurationMs: 2000})
	}
	if m.TotalCount != 10 {
		t.Errorf("TotalCount = %d, want 10", m.TotalCount)
	}
	if m.SuccessCount != 8 {
		t.Errorf("SuccessCount = %d, want 8", m.SuccessCount)
	}
	if got := m.SuccessRate(); got != 0.8 {
		t.Errorf("SuccessRate = %v, want 0.8", got)
	}
	// (8*1000 + 2*2000)/10 = 1200ms
	if got := m.AvgDurationMs(); got != 1200 {
		t.Errorf("AvgDurationMs = %v, want 1200", got)
	}
}

func TestMetrics_ZeroDivision(t *testing.T) {
	var m Metrics
	if got := m.SuccessRate(); got != 0.0 {
		t.Errorf("SuccessRate on empty = %v, want 0.0", got)
	}
	if got := m.AvgDurationMs(); got != 0.0 {
		t.Errorf("AvgDurationMs on empty = %v, want 0.0", got)
	}
	if got := m.UserInterventionRate(); got != 0.0 {
		t.Errorf("UserInterventionRate on empty = %v, want 0.0", got)
	}
	if got := m.FallbackRate(); got != 0.0 {
		t.Errorf("FallbackRate on empty = %v, want 0.0", got)
	}
}

func TestMetrics_InterventionAndFallbackRates(t *testing.T) {
	// Rates are computed over TotalCount; set the counters directly without
	// Record (which would also increment TotalCount).
	var m Metrics
	m.TotalCount = 10
	m.UserInterventions = 1
	m.FallbackUsed = 1
	if got := m.UserInterventionRate(); got != 0.1 {
		t.Errorf("UserInterventionRate = %v, want 0.1", got)
	}
	if got := m.FallbackRate(); got != 0.1 {
		t.Errorf("FallbackRate = %v, want 0.1", got)
	}
}

func TestMetrics_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/heal.jsonl"
	m := &Metrics{}
	if err := m.Persist(path, HealEvent{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "retryable-retry", Result: "ok", DurationMs: 1200}); err != nil {
		t.Fatalf("Persist: %v", err)
	}
	events, skipped, err := ParseFile(path, mustParseTime("2026-07-01T00:00:00Z"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}
	if len(events) != 1 || events[0].Result != "ok" || events[0].DurationMs != 1200 {
		t.Errorf("unexpected parsed events: %+v", events)
	}
}

func TestLog_Format(t *testing.T) {
	e := Event{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "retryable-retry", Result: "ok", DurationMs: 1200}
	got := e.Format()
	want := "2026-07-17T00:00:00Z | retry | retryable | retryable-retry | ok | 1200"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestLog_AppendEventSanity(t *testing.T) {
	var buf bytes.Buffer
	// Valid row.
	if err := AppendEvent(&buf, Event{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "x", Result: "ok", DurationMs: 100}); err != nil {
		t.Errorf("valid AppendEvent errored: %v", err)
	}
	// duration=0 rejected.
	if err := AppendEvent(&buf, Event{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "x", Result: "ok", DurationMs: 0}); err == nil {
		t.Error("expected error for duration=0")
	}
	// result invalid rejected.
	if err := AppendEvent(&buf, Event{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "x", Result: "maybe", DurationMs: 100}); err == nil {
		t.Error("expected error for invalid result")
	}
	// Only the one valid row should be present.
	if strings.Count(buf.String(), "\n") != 1 {
		t.Errorf("expected exactly 1 logged row, got:\n%s", buf.String())
	}
}

func TestLog_ParseFileSinceAndSkip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/heal.log"
	var buf bytes.Buffer
	// Old row (before since window).
	_ = AppendEvent(&buf, Event{ISO: "2026-01-01T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "x", Result: "ok", DurationMs: 100})
	// Recent row.
	_ = AppendEvent(&buf, Event{ISO: "2026-07-17T00:00:00Z", EventType: "retry", ErrorCode: "retryable", Action: "x", Result: "ok", DurationMs: 200})
	// Malformed row (wrong field count) — skipped on parse.
	buf.WriteString("garbage row without pipes\n")
	// Invalid-result row is rejected by AppendEvent, so it never reaches the
	// file; the only skip source is the malformed line above.
	if err := writeFile(path, buf.String()); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	events, skipped, err := ParseFile(path, mustParseTime("2026-07-01T00:00:00Z"))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	// 1 malformed row skipped; 1 old + 1 recent valid kept (old filtered by since).
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	if len(events) != 1 || events[0].DurationMs != 200 {
		t.Errorf("expected 1 recent valid event (200ms), got %+v", events)
	}
}

func mustParseTime(s string) time.Time {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return tm
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
