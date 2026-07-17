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
		fmt.Fprintln(os.Stderr, "vet agent: missing subcommand (run|resume|status)")
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
	default:
		fmt.Fprintf(os.Stderr, "vet agent: unknown subcommand %q\n", sub)
		os.Exit(2)
	}
}

func runAgentRun(args []string) {
	fs := flag.NewFlagSet("agent run", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root")
	payloadJSON := fs.String("payload", "", "incident payload as JSON string")
	dryRun := fs.Bool("dry-run", false, "dry run: execute all steps except actual GCL execution")
	fs.Parse(args)

	if *payloadJSON == "" {
		fmt.Fprintln(os.Stderr, "agent run: --payload is required")
		os.Exit(2)
	}

	payload, err := agent.ParseJSON([]byte(*payloadJSON))
	if err != nil {
		// Try natural language parsing as fallback
		payload, err = agent.ParseNaturalLanguage(*payloadJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent run: parse payload failed: %v\n", err)
			os.Exit(1)
		}
	}

	runID := fmt.Sprintf("%08x", time.Now().UnixNano()%0x100000000)

	if *dryRun {
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | dry-run mode | product=%s symptom=%s\n",
			runID, payload.ProductHint, payload.Symptom)

		// Run all steps except EXECUTE
		triage := agent.Triage(payload)
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | triage | primary=%s confidence=%s\n",
			runID, triage.PrimarySkill, triage.Confidence)

		evidence := agent.Diagnose(*root, triage.PrimarySkill, payload)
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | diagnose | findings=%d partial=%v\n",
			runID, len(evidence.Findings), evidence.Partial)

		plan := agent.ProposeFix(evidence, payload)
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | propose | ops=%d\n",
			runID, len(plan.Operations))

		confirm := agent.Confirm(*root, plan)
		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | confirm | decision=%s reason=%s\n",
			runID, confirm.Decision, confirm.Reason)

		fmt.Fprintf(os.Stderr, "[%s] [INFO] agent.run | dry-run complete | would execute %d ops\n",
			runID, len(plan.Operations))
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
