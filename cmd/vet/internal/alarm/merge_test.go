package alarm

import (
	"testing"
	"time"
)

var nowFunc = time.Now

func alarm(id, product, metric, resourceID string, value float64, minutesAgo int) *Alarm {
	return &Alarm{
		ID:         id,
		Product:    product,
		Metric:     metric,
		Value:      value,
		ResourceID: resourceID,
		At:         nowFunc().Add(-time.Duration(minutesAgo) * time.Minute),
		Severity:   "high",
	}
}

func alarmWithTime(id, product, metric, resourceID string, value float64, at time.Time) *Alarm {
	return &Alarm{
		ID:         id,
		Product:    product,
		Metric:     metric,
		Value:      value,
		ResourceID: resourceID,
		At:         at,
		Severity:   "high",
	}
}

func TestMerge_T1_AlarmStorm(t *testing.T) {
	// Freeze time so all alarms share the same "now" — 5 alarms spanning 4 min → same 5-min slot.
	base := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	oldNow := nowFunc
	nowFunc = func() time.Time { return base }
	defer func() { nowFunc = oldNow }()

	var alarms []*Alarm
	for i := 0; i < 5; i++ {
		alarms = append(alarms, alarmWithTime("a"+string(rune('0'+i)), "ECS", "CPU", "i-xxx", 95, base.Add(-time.Duration(5-i)*time.Minute)))
	}
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	groups := Merge(alarms, cfg)
	t.Logf("DEBUG T1: len=%d Count=%d Threshold=%d IsStorm=%v",
		len(groups), groups[0].Count, cfg.CountThreshold, groups[0].IsStorm)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group (all in same 5-min window), got %d", len(groups))
	}
	if groups[0].Count != 5 {
		t.Errorf("expected Count=5, got %d", groups[0].Count)
	}
	// IsStorm = Count > CountThreshold → 5 > 10 = false
	if groups[0].IsStorm {
		t.Errorf("expected IsStorm=false, got true (Count=%d Threshold=%d)",
			groups[0].Count, cfg.CountThreshold)
	}
}

func TestMerge_T2_SameInstanceDifferentMetrics(t *testing.T) {
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	alarms := []*Alarm{
		alarm("a", "ECS", "CPU", "i-xxx", 95, 5),
		alarm("b", "ECS", "Memory", "i-xxx", 90, 4),
		alarm("c", "ECS", "DiskUsage", "i-xxx", 85, 3),
	}

	groups := Merge(alarms, cfg)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups (different metrics deduplicated separately), got %d", len(groups))
	}
}

func TestMerge_T3_SameResourceCrossProduct(t *testing.T) {
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	// Same ResourceID but different products → still separate groups (key includes product)
	alarms := []*Alarm{
		alarm("a", "ECS", "CPU", "resource-abc", 95, 5),
		alarm("b", "RDS", "CPU", "resource-abc", 88, 4),
	}

	groups := Merge(alarms, cfg)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (different products), got %d", len(groups))
	}
}

func TestMerge_T4_DifferentResources(t *testing.T) {
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	alarms := []*Alarm{
		alarm("a", "ECS", "CPU", "i-xxx", 95, 5),
		alarm("b", "ECS", "CPU", "i-yyy", 92, 4),
	}

	groups := Merge(alarms, cfg)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (different resources), got %d", len(groups))
	}
}

func TestMerge_T5_SingleAlarm(t *testing.T) {
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	alarms := []*Alarm{alarm("a", "ECS", "CPU", "i-xxx", 95, 5)}

	groups := Merge(alarms, cfg)
	t.Logf("DEBUG T1: len=%d Count=%d Threshold=%d IsStorm=%v",
		len(groups), groups[0].Count, cfg.CountThreshold, groups[0].IsStorm)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Count != 1 {
		t.Errorf("expected Count=1, got %d", groups[0].Count)
	}
	if groups[0].IsStorm {
		t.Error("expected IsStorm=false for Count=1")
	}
}

func TestMerge_T6_StormBoundary(t *testing.T) {
	cfg := NewStormConfig()
	cfg.CountThreshold = 10

	var alarms []*Alarm
	for i := 0; i < 9; i++ {
		alarms = append(alarms, alarm("a"+string(rune('0'+i)), "ECS", "CPU", "i-xxx", 95, 9-i))
	}

	groups := Merge(alarms, cfg)
	if groups[0].IsStorm {
		t.Error("expected IsStorm=false for Count=9 with threshold=10")
	}
}
