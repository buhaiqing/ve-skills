package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/buhaiqing/ve-skills/cmd/vet/internal/validate"
)

// runValidate dispatches `vet validate`. Faithful Go port of validate_local.py.
func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	root := fs.String("root", repoRoot(), "repo root to scan")
	listOnly := fs.Bool("list", false, "list steps without executing")
	jsonOut := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Parse(args)

	if *listOnly {
		fmt.Printf("Validation steps (%d):\n", len(validate.StepNames()))
		for _, s := range validate.StepNames() {
			fmt.Printf("  - %s\n", s)
		}
		return
	}
	failing, _, total := validate.Run(*root, *listOnly)
	if *jsonOut {
		b, _ := json.MarshalIndent(map[string]any{
			"ok":          len(failing) == 0,
			"steps_total": total,
			"errors":      failing,
		}, "", "  ")
		fmt.Println(string(b))
		return
	}
	if len(failing) > 0 {
		fmt.Printf("FAIL: %d step(s) failed\n", len(failing))
		for name, errs := range failing {
			fmt.Printf("  FAIL %s:\n", name)
			for _, e := range errs {
				fmt.Printf("    - %s\n", e)
			}
		}
		os.Exit(1)
	}
	fmt.Printf("OK: %d steps passed\n", total)
}
