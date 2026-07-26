package alarm

import (
	"sort"
	"time"
)

type Alarm struct {
	ID           string
	Product      string
	Metric       string
	Value        float64
	Threshold    float64
	ResourceID   string
	ResourceName string
	Region       string
	Severity     string
	At           time.Time
	Tags         map[string]string
}

type AlarmGroup struct {
	RootCause  *Alarm
	Alarms     []*Alarm
	Count      int
	Products   []string
	Metrics    []string
	IsStorm    bool
}

type StormConfig struct {
	Window        time.Duration
	CountThreshold int
	SeverityRate  float64
}

func NewStormConfig() *StormConfig {
	return &StormConfig{
		Window:         5 * time.Minute,
		CountThreshold: 10,
		SeverityRate:   0.5,
	}
}

type dedupKey struct {
	product    string
	resourceID string
	metric     string
	slot       int64
}

func keyFor(a *Alarm, windowMillis int64) dedupKey {
	return dedupKey{
		a.Product,
		a.ResourceID,
		a.Metric,
		a.At.UnixMilli() / windowMillis,
	}
}

func Merge(alarms []*Alarm, cfg *StormConfig) []*AlarmGroup {
	if cfg == nil {
		cfg = NewStormConfig()
	}
	if len(alarms) == 0 {
		return nil
	}

	sort.Slice(alarms, func(i, j int) bool {
		return alarms[i].At.Before(alarms[j].At)
	})

	windowMillis := cfg.Window.Milliseconds()
	groups := make(map[dedupKey]*AlarmGroup)

	for _, a := range alarms {
		k := keyFor(a, windowMillis)
		if groups[k] == nil {
			groups[k] = &AlarmGroup{
				RootCause: a,
				Alarms:    []*Alarm{},
				Products:  []string{},
				Metrics:   []string{},
			}
		}
		groups[k].Alarms = append(groups[k].Alarms, a)
		groups[k].Count++
		if a.At.Before(groups[k].RootCause.At) {
			groups[k].RootCause = a
		}
	}

	var result []*AlarmGroup
	for _, g := range groups {
		g.IsStorm = g.Count > cfg.CountThreshold

		seenProduct := make(map[string]bool)
		for _, a := range g.Alarms {
			if !seenProduct[a.Product] {
				g.Products = append(g.Products, a.Product)
				seenProduct[a.Product] = true
			}
		}

		seenMetric := make(map[string]bool)
		for _, a := range g.Alarms {
			if !seenMetric[a.Metric] {
				g.Metrics = append(g.Metrics, a.Metric)
				seenMetric[a.Metric] = true
			}
		}

		result = append(result, g)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsStorm != result[j].IsStorm {
			return result[i].IsStorm
		}
		return result[i].RootCause.At.Before(result[j].RootCause.At)
	})

	return result
}
