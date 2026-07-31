package heal

import (
	"context"
	"fmt"
	"os/exec"
)

// ProbeRunner executes a read-only probe via argv (no shell).
type ProbeRunner func(ctx context.Context, argv []string) (stdout string, err error)

// DefaultProbeRunner runs argv[0] with argv[1:] via exec.CommandContext (P1: no sh -c).
func DefaultProbeRunner(ctx context.Context, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("empty probe argv")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	out, err := cmd.Output()
	return string(out), err
}

// RunProbe runs a probe with the given runner. Empty argv returns an error.
func RunProbe(ctx context.Context, runner ProbeRunner, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("empty probe argv")
	}
	if runner == nil {
		runner = DefaultProbeRunner
	}
	_, err := runner(ctx, argv)
	return err
}
