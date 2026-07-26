package costgate

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func RunCostGateCmd(args []string) int {
	fs := flag.NewFlagSet("vet gcl cost-gate", flag.ExitOnError)
	skill := fs.String("skill", "", "skill name (e.g. ve-ecs-ops)")
	command := fs.String("command", "", "ve CLI command to estimate cost impact")
	jsonOut := fs.Bool("json", false, "machine-readable JSON output")
	fs.Parse(args)

	if *skill == "" || *command == "" {
		fmt.Fprintln(os.Stderr, "vet gcl cost-gate: --skill and --command are required")
		fmt.Fprintln(os.Stderr, "Usage: vet gcl cost-gate --skill ve-ecs-ops --command <escaped_command>")
		return 1
	}

	cost := EstimateCostImpact(*skill, *command)

	if *jsonOut {
		if cost == nil {
			fmt.Println("{}")
			return 0
		}
		data, _ := json.MarshalIndent(cost, "", "  ")
		fmt.Println(string(data))
		return 0
	}

	if cost == nil {
		fmt.Printf("Operation %s (%s) has no billing impact (read-only or unrecognised)\n", *command, *skill)
		return 0
	}

	fmt.Printf("Billing Impact Estimate\n")
	fmt.Printf("  Operation:       %s\n", cost.Operation)
	fmt.Printf("  Billing Model:   %s\n", cost.BillingModel)
	fmt.Printf("  Est. Monthly:    %.2f CNY\n", cost.EstMonthlyCost)
	fmt.Printf("  Refund on Delete: %.2f CNY\n", cost.RefundOnDelete)
	fmt.Printf("  New Monthly:     %.2f CNY\n", cost.NewMonthlyCost)
	fmt.Printf("  Net Delta:       %.2f CNY\n", cost.NetMonthlyDelta)
	if cost.Warning != "" {
		fmt.Printf("  \033[33mWARNING:\033[0m %s\n", cost.Warning)
	}
	return 0
}
