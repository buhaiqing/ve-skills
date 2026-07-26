package observability

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewRootTrace(t *testing.T) {
	tc := NewRootTrace("test-svc", "test-op")
	if tc.TraceID == "" {
		t.Error("TraceID should not be empty")
	}
	if tc.SpanID == "" {
		t.Error("SpanID should not be empty")
	}
	if tc.Service != "test-svc" {
		t.Errorf("Service = %q, want %q", tc.Service, "test-svc")
	}
	if tc.Operation != "test-op" {
		t.Errorf("Operation = %q, want %q", tc.Operation, "test-op")
	}
	if tc.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
	if tc.Attributes == nil {
		t.Error("Attributes should not be nil")
	}
}

func TestStartSpan(t *testing.T) {
	parent := NewRootTrace("parent-svc", "parent-op")
	span := StartSpan(parent, "child-svc", "child-op")

	if span.TraceID != parent.TraceID {
		t.Errorf("child TraceID = %q, want %q", span.TraceID, parent.TraceID)
	}
	if span.ParentID != parent.SpanID {
		t.Errorf("child ParentID = %q, want %q", span.ParentID, parent.SpanID)
	}
	if span.SpanID == "" {
		t.Error("child SpanID should not be empty")
	}
	if span.SpanID == parent.SpanID {
		t.Error("child SpanID should differ from parent SpanID")
	}
}

func TestSpanEnd(t *testing.T) {
	parent := NewRootTrace("svc", "op")

	span1 := StartSpan(parent, "svc", "ok-op")
	span1.End(nil)
	if span1.Status != "ok" {
		t.Errorf("Status = %q, want %q", span1.Status, "ok")
	}

	span2 := StartSpan(parent, "svc", "err-op")
	testErr := errors.New("test error")
	span2.End(testErr)
	if span2.Status != "error" {
		t.Errorf("Status = %q, want %q", span2.Status, "error")
	}
	if span2.Err != testErr {
		t.Errorf("Err = %v, want %v", span2.Err, testErr)
	}
}

func TestSpanDuration(t *testing.T) {
	parent := NewRootTrace("svc", "op")
	span := StartSpan(parent, "svc", "sleep-op")
	time.Sleep(15 * time.Millisecond)
	d := span.Duration()
	if d <= 10*time.Millisecond {
		t.Errorf("Duration = %v, want > 10ms", d)
	}
}

func TestWithTraceFromContext(t *testing.T) {
	tc := NewRootTrace("svc", "op")
	ctx := WithTrace(context.Background(), tc)

	got := FromContext(ctx)
	if got == nil {
		t.Fatal("FromContext returned nil")
	}
	if got.TraceID != tc.TraceID {
		t.Errorf("TraceID = %q, want %q", got.TraceID, tc.TraceID)
	}
	if got.SpanID != tc.SpanID {
		t.Errorf("SpanID = %q, want %q", got.SpanID, tc.SpanID)
	}

	// FromContext on empty context should return nil
	emptyCtx := context.Background()
	if FromContext(emptyCtx) != nil {
		t.Error("FromContext on empty context should return nil")
	}
}

func TestMetricsCounter(t *testing.T) {
	mc := NewMetricsCollector()
	labels := map[string]string{"region": "cn-hangzhou"}

	c := mc.Counter("requests", labels)
	c.Inc()
	c.Add(5)
	if c.Value() != 6 {
		t.Errorf("Counter value = %d, want 6", c.Value())
	}

	c2 := mc.Counter("requests", labels)
	if c2 != c {
		t.Error("same labels should return same counter instance")
	}
	c2.Add(3)
	if c.Value() != 9 {
		t.Errorf("Counter value after Add on returned instance = %d, want 9", c.Value())
	}

	c3 := mc.Counter("requests", nil)
	if c3 == c {
		t.Error("different labels should return different counter")
	}
}

func TestMetricsGauge(t *testing.T) {
	mc := NewMetricsCollector()
	labels := map[string]string{"host": "server-1"}

	g := mc.Gauge("temperature", labels)
	g.Set(42.5)
	if g.Value() != 42.5 {
		t.Errorf("Gauge value = %v, want 42.5", g.Value())
	}

	g.Inc()
	if g.Value() != 43.5 {
		t.Errorf("Gauge after Inc = %v, want 43.5", g.Value())
	}

	g.Dec()
	if g.Value() != 42.5 {
		t.Errorf("Gauge after Dec = %v, want 42.5", g.Value())
	}
}

func TestMetricsSnapshot(t *testing.T) {
	mc := NewMetricsCollector()

	c := mc.Counter("total_requests", nil)
	c.Add(100)

	g := mc.Gauge("memory_usage", nil)
	g.Set(75.0)

	snap := mc.Snapshot()
	if snap.Timestamp.IsZero() {
		t.Error("Snapshot Timestamp should not be zero")
	}
	if snap.Counters["total_requests"] != 100 {
		t.Errorf("Counter snapshot = %d, want 100", snap.Counters["total_requests"])
	}
	if snap.Gauges["memory_usage"] != 75.0 {
		t.Errorf("Gauge snapshot = %v, want 75.0", snap.Gauges["memory_usage"])
	}
}