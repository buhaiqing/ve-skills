package trace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheck_ValidTrace(t *testing.T) {
	path := filepath.Join("testdata", "trace_valid.json")
	err := Check(path)
	if err != nil {
		t.Fatalf("expected valid trace to pass, got: %v", err)
	}
}

func TestCheck_MissingReqId(t *testing.T) {
	path := filepath.Join("testdata", "trace_missing_reqid.json")
	err := Check(path)
	if err == nil {
		t.Fatal("expected missing request_id to fail, got nil")
	}
	if !contains(err.Error(), "request_id") {
		t.Errorf("expected error about request_id, got: %v", err)
	}
}

func TestCheck_BadRedaction(t *testing.T) {
	path := filepath.Join("testdata", "trace_bad_redaction.json")
	err := Check(path)
	if err == nil {
		t.Fatal("expected redaction_pass=false to fail, got nil")
	}
	if !contains(err.Error(), "redaction_pass") {
		t.Errorf("expected error about redaction_pass, got: %v", err)
	}
}

func TestCheck_ValidFromDisk(t *testing.T) {
	// Verify the valid fixture actually parses correctly
	path := filepath.Join("testdata", "trace_valid.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("valid trace content: %s", string(data))
	if err := Check(path); err != nil {
		t.Fatalf("valid trace should pass: %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
