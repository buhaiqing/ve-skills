package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/agent"
)

func runAgent(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "vet agent: missing subcommand (run|resume|status|eval-report)")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "run":
		runAgentRun(rest)
	case "resume":
		runAgentResume(rest)
	case "status":
		runAgentStatus(rest)
	case "eval-report":
		runAgentEvalReport(rest)
	default:
		fmt.Fprintf(os.Stderr, "vet agent: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func runAgentRun(args []string) {
	fs := flag.NewFlagSet("agent run", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	payloadJSON := fs.String("payload", "", "incident payload as JSON string")
	region := fs.String("region", "", "region override (default: cn-beijing)")
	dryRun := fs.Bool("dry-run", false, "dry run: execute all steps except actual GCL execution")
	fs.Parse(args)

	if *payloadJSON == "" {
		fmt.Fprintln(os.Stderr, "agent run: --payload is required")
		os.Exit(2)
	}

	payload, err := agent.ParseJSON([]byte(*payloadJSON))
	if err != nil {
		payload, err = agent.ParseNaturalLanguage(*payloadJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent run: parse payload failed: %v\n", err)
			os.Exit(1)
		}
	}

	if *region != "" {
		payload.Region = *region
	}

	runID := fmt.Sprintf("%d", time.Now().UnixNano())

	if *dryRun {
		dryRunResult := agent.RunDry(*root, payload, runID)
		if !dryRunResult.Success {
			fmt.Fprintf(os.Stderr, "[%s] [ERROR] agent.run | dry-run failed | step=%s error=%s\n",
				runID, dryRunResult.FinalStep.String(), dryRunResult.Error)
			os.Exit(1)
		}
		return
	}

	result := agent.Run(*root, payload, runID)
	if !result.Success {
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] agent.run | failed | step=%s error=%s\n",
			runID, result.FinalStep.String(), result.Error)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | success | step=%s\n", runID, result.FinalStep.String())
}

func runAgentResume(args []string) {
	fs := flag.NewFlagSet("agent resume", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	runID := fs.String("run-id", "", "run ID to resume")
	confirmed := fs.Bool("confirmed", false, "confirm ASK-class operations")
	fs.Parse(args)

	if *runID == "" {
		fmt.Fprintln(os.Stderr, "agent resume: --run-id is required")
		os.Exit(2)
	}

	state, err := agent.LoadState(*root, *runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent resume: load state failed: %v\n", err)
		os.Exit(1)
	}
	if state == nil {
		fmt.Fprintf(os.Stderr, "agent resume: run %s not found\n", *runID)
		os.Exit(1)
	}

	if *confirmed {
		// Resume from CONFIRM step with confirmation
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.resume | confirmed | resuming from %s\n",
			*runID, state.CurrentStep.String())
		result := agent.Run(*root, &state.Payload, *runID)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.resume | status | current_step=%s\n",
		*runID, state.CurrentStep.String())
}

func runAgentStatus(args []string) {
	fs := flag.NewFlagSet("agent status", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	runID := fs.String("run-id", "", "run ID to check")
	fs.Parse(args)

	if *runID == "" {
		fmt.Fprintln(os.Stderr, "agent status: --run-id is required")
		os.Exit(2)
	}

	state, err := agent.LoadState(*root, *runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent status: load state failed: %v\n", err)
		os.Exit(1)
	}
	if state == nil {
		fmt.Fprintf(os.Stderr, "agent status: run %s not found\n", *runID)
		os.Exit(1)
	}

	fmt.Printf("Run ID:      %s\n", state.RunID)
	fmt.Printf("Step:        %s\n", state.CurrentStep.String())
	fmt.Printf("Product:     %s\n", state.Payload.ProductHint)
	fmt.Printf("Symptom:     %s\n", state.Payload.Symptom)
	if state.Triage != nil {
		fmt.Printf("Primary:     %s (%s)\n", state.Triage.PrimarySkill, state.Triage.Confidence)
	}
	if state.Confirm != nil {
		fmt.Printf("Decision:    %s (%s)\n", state.Confirm.Decision, state.Confirm.Reason)
	}
	if state.Result != nil {
		fmt.Printf("Exec Result: success=%v\n", state.Result.Success)
	}
	if state.Error != "" {
		fmt.Printf("Error:       %s\n", state.Error)
	}
}

func runAgentEvalReport(args []string) {
	fs := flag.NewFlagSet("agent eval-report", flag.ExitOnError)
	inPath := fs.String("in", "", "path to EvalReportInput JSON")
	outPath := fs.String("out", "audit-results/eval-report.json", "output report path")
	fs.Parse(args)
	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "agent eval-report: --in is required")
		os.Exit(2)
	}
	in, err := agent.LoadEvalSamples(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent eval-report: load: %v\n", err)
		os.Exit(1)
	}
	rep := agent.BuildEvalReport(in)
	if err := agent.WriteEvalReport(*outPath, rep); err != nil {
		fmt.Fprintf(os.Stderr, "agent eval-report: write: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "eval-report written to %s (samples=%d)\n", *outPath, rep.Samples)
}
