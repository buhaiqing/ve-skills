package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/gate"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/heal"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/run"
	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/trace"
)

// runGCL dispatches `vet gcl <run|gate|trace>`.
func runGCL(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet gcl: missing subcommand (run|gate|trace)")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "run":
		runGCLRun(rest)
	case "gate":
		runGCLGate(rest)
	case "trace":
		runGCLTrace(rest)
	case "heal-stats":
		runGCLHealStats(rest)
	default:
		fmt.Fprintf(os.Stderr, "vet gcl %s: unknown subcommand (run|gate|trace|heal-stats)\n", sub)
		os.Exit(3)
	}
}

func runGCLRun(args []string) {
	fs := flag.NewFlagSet("gcl run", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	skill := fs.String("skill", "", "skill id, e.g. ve-ecs-ops")
	request := fs.String("request", "", "sanitized user request (stored in trace)")
	command := fs.String("command", "", "shell command for Generator")
	maxIter := fs.Int("max-iter", 0, "max loop iterations (0 = skill default)")
	timeout := fs.Int("timeout", 120, "generator command timeout (s)")
	structural := fs.Bool("structural-critic-only", false, "rule-based structural critic (CI/dry-run)")
	criticJSON := fs.String("critic-json", "", "external Critic JSON file")
	criticStdin := fs.Bool("critic-stdin", false, "read Critic JSON from stdin")
	criticCmd := fs.String("critic-command", "", "isolated Critic command (separate process)")
	confirmed := fs.Bool("confirmed", false, "vouch for ASK-class operations (otherwise treated as blocked in non-interactive mode)")
	confirmedBy := fs.String("confirmed-by", "", "provenance of --confirmed (ticket id / human handle from the Step 5 {{user.confirm}} gate); recorded in trace for audit")
	heal := fs.String("heal", "smart", "retry strategy: 'smart' (error-classification-driven L4 retry) or 'none' (legacy fixed-count loop)")
	fs.Parse(args)

	if *skill == "" || *request == "" || *command == "" {
		fmt.Fprintln(os.Stderr, "vet gcl run: --skill, --request, --command are required")
		os.Exit(2)
	}
	code := run.Run(run.Options{
		Root:           *root,
		Skill:          *skill,
		Request:        *request,
		Command:        *command,
		MaxIter:        *maxIter,
		Timeout:        *timeout,
		StructuralOnly: *structural,
		CriticJSON:     *criticJSON,
		CriticStdin:    *criticStdin,
		CriticCommand:  *criticCmd,
		Confirmed:      *confirmed,
		ConfirmedBy:    *confirmedBy,
		Heal:           *heal,
	}).ExitCode
	os.Exit(code)
}

func runGCLGate(args []string) {
	fs := flag.NewFlagSet("gcl gate", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	skills := fs.String("skills", "", "comma-separated skill names (default: all)")
	skipIncident := fs.Bool("skip-incident-loop", false, "skip incident-loop-agent")
	fs.Parse(args)

	var skillList []string
	if *skills != "" {
		for _, s := range splitComma(*skills) {
			skillList = append(skillList, s)
		}
	}
	code := gate.Run(*root, skillList, *skipIncident, *jsonOut)
	os.Exit(code)
}

func runGCLTrace(args []string) {
	fs := flag.NewFlagSet("gcl trace", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	input := fs.String("input", "", "comma-separated trace file(s)/glob (default: audit-results/gcl-trace-*.json)")
	sinceHours := fs.String("since-hours", "", "only traces modified within N hours")
	link := fs.Bool("link", false, "correlate gcl-trace and incident-trace by request_id")
	incident := fs.Bool("incident", false, "aggregate incident-trace-*.json by ticket")
	fs.Parse(args)

	if *link {
		code := trace.CmdLink(*root)
		os.Exit(code)
	}

	if *incident {
		code := trace.CmdIncident(*root)
		os.Exit(code)
	}

	var inputs []string
	if *input != "" {
		inputs = splitComma(*input)
	}
	var hours *int
	if *sinceHours != "" {
		if n, err := strconv.Atoi(*sinceHours); err == nil {
			hours = &n
		}
	}
	code := trace.CmdAggregate(*root, inputs, hours)
	os.Exit(code)
}

func runGCLHealStats(args []string) {
	fs := flag.NewFlagSet("gcl heal-stats", flag.ExitOnError)
	since := fs.String("since", "7d", "lookback window: Nd / Nw / Nm (days/weeks/months)")
	logPath := fs.String("log", heal.DefaultLogPath, "self-healing log path (§6.2 format)")
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	fs.Parse(args)

	sinceDur, err := parseSince(*since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vet gcl heal-stats: %v\n", err)
		os.Exit(2)
	}
	events, skipped, err := heal.ParseFile(*logPath, time.Now().Add(-sinceDur))
	if err != nil {
		fmt.Fprintf(os.Stderr, "vet gcl heal-stats: failed to read %s: %v\n", *logPath, err)
		os.Exit(1)
	}
	var m heal.Metrics
	for _, e := range events {
		m.Record(heal.HealEvent{
			ISO: e.ISO, EventType: e.EventType, ErrorCode: e.ErrorCode,
			Action: e.Action, Result: e.Result, DurationMs: e.DurationMs,
		})
	}
	if *jsonOut {
		b, _ := json.MarshalIndent(map[string]any{
			"total": m.TotalCount, "success": m.SuccessCount,
			"success_rate": m.SuccessRate(), "avg_duration_ms": m.AvgDurationMs(),
			"user_intervention_rate": m.UserInterventionRate(), "fallback_rate": m.FallbackRate(),
			"skipped": skipped,
		}, "", "  ")
		fmt.Println(string(b))
		return
	}
	printSLO := func(label string, value float64, target float64, lowerBetter bool) {
		pass := value >= target
		if lowerBetter {
			pass = value <= target
		}
		mark := "✅"
		if !pass {
			mark = "❌"
		}
		fmt.Printf("%-22s %.1f%% (target: %s %.0f%%) %s\n",
			label, value*100, cmpWord(lowerBetter), target*100, mark)
	}
	printDuration := func(label string, valueMs float64, targetMs float64) {
		pass := valueMs <= targetMs
		mark := "✅"
		if !pass {
			mark = "❌"
		}
		fmt.Printf("%-22s %.1fs (target: %s %.0fs) %s\n",
			label, valueMs/1000.0, cmpWord(true), targetMs/1000.0, mark)
	}
	fmt.Printf("Self-healing stats (since %s, events=%d, skipped=%d):\n", *since, m.TotalCount, skipped)
	printSLO("Success rate", m.SuccessRate(), heal.TargetSuccessRate, false)
	printDuration("Avg duration", m.AvgDurationMs(), heal.TargetAvgDurationMs)
	printSLO("User intervention", m.UserInterventionRate(), heal.TargetUserIntervention, true)
	printSLO("Fallback usage", m.FallbackRate(), heal.TargetFallback, true)
}

// parseSince parses Nd / Nw / Nm into a duration.
func parseSince(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid --since %q (want e.g. 7d)", s)
	}
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return 0, fmt.Errorf("invalid --since %q: %v", s, err)
	}
	switch s[len(s)-1] {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case 'm':
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid --since %q: unit must be d/w/m", s)
	}
}

func cmpWord(lowerBetter bool) string {
	if lowerBetter {
		return "<"
	}
	return ">"
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
