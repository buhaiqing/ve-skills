package heal

import "fmt"

// BuiltInPromotions maps incident types to per-step ProbeArgv (read-only).
func BuiltInPromotions() map[string][][]string {
	return map[string][][]string{
		"cpu_high": {
			{"ve", "ecs", "DescribeInstances"},
			{"ve", "ecs", "DescribeInstances"},
		},
		"redis_slow_query": {
			{"ve", "redis", "DescribeDBInstances"},
			{"ve", "redis", "DescribeDBInstances"},
			{"ve", "redis", "DescribeDBInstances"},
		},
	}
}

// Promote marks plan steps non-stub and attaches ProbeArgv.
// probes length MUST equal len(plan.Steps); each argv non-empty.
func (o *Orchestrator) Promote(name string, probes [][]string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.promoteUnderLock(name, probes)
}

func (o *Orchestrator) promoteUnderLock(name string, probes [][]string) error {
	plan, ok := o.plans[name]
	if !ok {
		return fmt.Errorf("heal: unknown plan %q", name)
	}
	if len(probes) != len(plan.Steps) {
		return fmt.Errorf("heal: promote %q: need %d probes, got %d", name, len(plan.Steps), len(probes))
	}
	for i := range plan.Steps {
		if len(probes[i]) == 0 {
			return fmt.Errorf("heal: promote %q: empty probe at step %d", name, i)
		}
		plan.Steps[i].Stub = false
		plan.Steps[i].ProbeArgv = append([]string(nil), probes[i]...)
		plan.Steps[i].CheckFn = nil
	}
	return nil
}

// ApplyBuiltInPromotions promotes cpu_high and redis_slow_query.
func (o *Orchestrator) ApplyBuiltInPromotions() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for name, probes := range BuiltInPromotions() {
		if err := o.promoteUnderLock(name, probes); err != nil {
			return err
		}
	}
	return nil
}
