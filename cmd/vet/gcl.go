package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/gcl/gate"
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
	default:
		fmt.Fprintf(os.Stderr, "vet gcl %s: unknown subcommand (run|gate|trace)\n", sub)
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
	fs.Parse(args)

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

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
