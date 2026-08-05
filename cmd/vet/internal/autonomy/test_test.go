package autonomy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunNConsecutiveIncidents(t *testing.T) {
	// Create temporary envelope file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test-envelope.yaml")
	content := `domains:
  - name: redis-slow-commands
    skills: [ve-redis-ops]
    symptoms: [slow-commands, oom-prevention]
    blast_radius: single
    slo_ref: redis-p99-latency
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test envelope: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := RunNConsecutiveIncidentsPath(ctx, 5, envPath, nil)
	if err != nil {
		t.Fatalf("RunNConsecutiveIncidents failed: %v", err)
	}

	if report.TotalIncidents != 5 {
		t.Errorf("expected 5 total incidents, got %d", report.TotalIncidents)
	}
	if report.Passed != 5 {
		t.Errorf("expected 5 passed, got %d", report.Passed)
	}
	if report.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", report.Failed)
	}
	if report.Prompts != 0 {
		t.Errorf("expected 0 prompts (L4 guarantee), got %d", report.Prompts)
	}
	if report.SLOViolations != 0 {
		t.Errorf("expected 0 SLO violations, got %d", report.SLOViolations)
	}
}

func TestRunNConsecutiveIncidents_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test-envelope.yaml")
	content := `domains:
  - name: test
    skills: [test-skill]
    symptoms: [test-symptom]
    blast_radius: single
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test envelope: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := RunNConsecutiveIncidentsPath(ctx, 5, envPath, nil)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestRunNConsecutiveIncidents_InvalidEnvelope(t *testing.T) {
	ctx := context.Background()
	_, err := RunNConsecutiveIncidentsPath(ctx, 5, "/nonexistent/path.yaml", nil)
	if err == nil {
		t.Error("expected error for invalid envelope path")
	}
}
