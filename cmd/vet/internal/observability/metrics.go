package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Counter struct {
	name   string
	labels map[string]string
	value  int64
}

func (c *Counter) Inc() {
	c.value++
}

func (c *Counter) Add(n int64) {
	c.value += n
}

func (c *Counter) Value() int64 {
	return c.value
}

type Gauge struct {
	name   string
	labels map[string]string
	value  float64
}

func (g *Gauge) Set(v float64) {
	g.value = v
}

func (g *Gauge) Inc() {
	g.value++
}

func (g *Gauge) Dec() {
	g.value--
}

func (g *Gauge) Value() float64 {
	return g.value
}

type Histogram struct {
	name    string
	labels  map[string]string
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
}

func (h *Histogram) Observe(v float64) {
	h.sum += v
	h.count++
	for i, bound := range h.buckets {
		if v <= bound {
			h.counts[i]++
		}
	}
}

type MetricsSnapshot struct {
	Timestamp  time.Time           `json:"timestamp"`
	Counters   map[string]int64    `json:"counters"`
	Gauges     map[string]float64  `json:"gauges"`
	Histograms map[string]map[string]interface{} `json:"histograms"`
}

type MetricsCollector struct {
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	mu         sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
	}
}

func (mc *MetricsCollector) Counter(name string, labels map[string]string) *Counter {
	key := metricKey(name, labels)
	mc.mu.RLock()
	c, ok := mc.counters[key]
	mc.mu.RUnlock()
	if ok {
		return c
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if c, ok = mc.counters[key]; ok {
		return c
	}
	c = &Counter{name: name, labels: labels}
	mc.counters[key] = c
	return c
}

func (mc *MetricsCollector) Gauge(name string, labels map[string]string) *Gauge {
	key := metricKey(name, labels)
	mc.mu.RLock()
	g, ok := mc.gauges[key]
	mc.mu.RUnlock()
	if ok {
		return g
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if g, ok = mc.gauges[key]; ok {
		return g
	}
	g = &Gauge{name: name, labels: labels}
	mc.gauges[key] = g
	return g
}

func (mc *MetricsCollector) Histogram(name string, labels map[string]string, buckets []float64) *Histogram {
	key := metricKey(name, labels)
	mc.mu.RLock()
	h, ok := mc.histograms[key]
	mc.mu.RUnlock()
	if ok {
		return h
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if h, ok = mc.histograms[key]; ok {
		return h
	}
	counts := make([]int64, len(buckets))
	h = &Histogram{name: name, labels: labels, buckets: buckets, counts: counts}
	mc.histograms[key] = h
	return h
}

func (mc *MetricsCollector) Snapshot() MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	snap := MetricsSnapshot{
		Timestamp:  time.Now(),
		Counters:   make(map[string]int64),
		Gauges:     make(map[string]float64),
		Histograms: make(map[string]map[string]interface{}),
	}

	for key, c := range mc.counters {
		snap.Counters[key] = c.value
	}
	for key, g := range mc.gauges {
		snap.Gauges[key] = g.value
	}
	for key, h := range mc.histograms {
		snap.Histograms[key] = map[string]interface{}{
			"sum":   h.sum,
			"count": h.count,
		}
	}
	return snap
}

func (mc *MetricsCollector) Persist(path string) error {
	snap := mc.Snapshot()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func metricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	key := name + "{"
	first := true
	for k, v := range labels {
		if !first {
			key += ","
		}
		key += k + "=" + v
		first = false
	}
	key += "}"
	return key
}