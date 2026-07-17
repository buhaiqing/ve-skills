package heal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// DefaultLogPath is where self-healing events are appended (framework §6.2).
// Relative to the repo root (the cwd when `vet` runs) so the telemetry lands
// under audit-results/ and shows up in the GCL quality aggregate. Overridable
// via `vet gcl heal-stats --log`.
const DefaultLogPath = "audit-results/ve-self-healing.log"

// Event is the framework §6.2 log row:
//
//	<ISO> | <event_type> | <error_code> | <action> | <result> | <duration>
//
// error_code carries the *real* `ve` CLI signal (the heal class name or the
// matched excerpt) — NOT the framework installer codes (NET_*/PERM_*/GO_*),
// which are out of scope for this package (see T09/T11 spec §2.1).
type Event struct {
	ISO        string
	EventType  string
	ErrorCode  string
	Action     string
	Result     string // "ok" | "fail"
	DurationMs int64
}

// Format renders the §6.2 pipe-delimited row.
func (e Event) Format() string {
	return fmt.Sprintf("%s | %s | %s | %s | %s | %d",
		e.ISO, e.EventType, e.ErrorCode, e.Action, e.Result, e.DurationMs)
}

// Validate enforces the sanity checks from spec §3: a row is only worth
// persisting if it has a positive duration and a known result.
func (e Event) Validate() error {
	if e.DurationMs <= 0 {
		return fmt.Errorf("heal: invalid duration %d (must be > 0)", e.DurationMs)
	}
	if e.Result != "ok" && e.Result != "fail" {
		return fmt.Errorf("heal: invalid result %q (want ok|fail)", e.Result)
	}
	return nil
}

// AppendEvent writes one §6.2 row, rejecting invalid rows before they touch
// the file. The writer is typically an *os.File opened append; callers may
// pass any io.Writer (e.g. a buffer in tests).
func AppendEvent(w io.Writer, e Event) error {
	if err := e.Validate(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, e.Format())
	return err
}

// ParseFile reads a §6.2 log file, keeping only rows whose ISO timestamp is
// at or after `since`. Malformed or pre-sanity rows are skipped and counted so
// the aggregator can report (not fail on) dirty data. Returns the parsed
// events and the number skipped.
func ParseFile(path string, since time.Time) ([]Event, int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	defer f.Close()

	var events []Event
	skipped := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		e, ok := parseRow(line)
		if !ok {
			skipped++
			continue
		}
		ts, err := time.Parse(time.RFC3339, e.ISO)
		if err != nil {
			skipped++
			continue
		}
		if ts.Before(since) {
			continue
		}
		events = append(events, e)
	}
	if err := sc.Err(); err != nil {
		return nil, skipped, err
	}
	return events, skipped, nil
}

// parseRow splits a §6.2 pipe row into an Event. Returns ok=false on the
// wrong field count or an invalid result so the caller can skip it.
func parseRow(line string) (Event, bool) {
	parts := strings.Split(line, "|")
	if len(parts) != 6 {
		return Event{}, false
	}
	e := Event{
		ISO:        strings.TrimSpace(parts[0]),
		EventType:  strings.TrimSpace(parts[1]),
		ErrorCode:  strings.TrimSpace(parts[2]),
		Action:     strings.TrimSpace(parts[3]),
		Result:     strings.TrimSpace(parts[4]),
		DurationMs: 0,
	}
	if e.Result != "ok" && e.Result != "fail" {
		return Event{}, false
	}
	var d int64
	if _, err := fmt.Sscan(strings.TrimSpace(parts[5]), &d); err != nil || d <= 0 {
		return Event{}, false
	}
	e.DurationMs = d
	return e, true
}

// LoadEvents is a convenience that parses the default log path since `since`.
func LoadEvents(since time.Time) ([]Event, int, error) {
	return ParseFile(DefaultLogPath, since)
}
