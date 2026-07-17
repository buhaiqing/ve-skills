package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatLine(t *testing.T) {
	line := formatLine("a1b2c3d4", INFO, "gcl.run", "start", "skill=ve-ecs-ops", "max_iter=3")
	if !strings.Contains(line, "[a1b2c3d4]") {
		t.Errorf("expected runID in line, got: %s", line)
	}
	if !strings.Contains(line, "INFO") {
		t.Errorf("expected INFO level, got: %s", line)
	}
	if !strings.Contains(line, "gcl.run") {
		t.Errorf("expected component, got: %s", line)
	}
	if !strings.Contains(line, "start") {
		t.Errorf("expected message, got: %s", line)
	}
	if !strings.Contains(line, "skill=ve-ecs-ops") {
		t.Errorf("expected skill KV, got: %s", line)
	}
	if !strings.Contains(line, "max_iter=3") {
		t.Errorf("expected max_iter KV, got: %s", line)
	}
	// Verify format: ISO | [runID] | level | component | message | kvs
	parts := strings.Split(line, " | ")
	if len(parts) < 5 {
		t.Errorf("expected at least 5 pipe-delimited fields, got %d: %s", len(parts), line)
	}
}

func TestAppendAndRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	// Write a log line
	if err := Append(path, "testrun1", INFO, "test", "hello", "key=value"); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify file exists and contains the line
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "[testrun1]") {
		t.Errorf("expected runID in file, got: %s", content)
	}
	if !strings.Contains(content, "hello") {
		t.Errorf("expected message in file, got: %s", content)
	}
	if !strings.Contains(content, "key=value") {
		t.Errorf("expected KV in file, got: %s", content)
	}
}

func TestRotation_MaxSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rotating.log")

	// Use a very small rotation size for testing
	smallRotation := Rotation{MaxSize: 200, MaxBackups: 3, MaxAge: 30}

	// Write enough lines to trigger rotation
	for i := 0; i < 20; i++ {
		if err := appendWithRotation(path, "test", INFO, "test", "line", smallRotation); err != nil {
			t.Fatalf("Append failed at %d: %v", i, err)
		}
	}

	// Verify main file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("main log file missing: %v", err)
	}

	// Verify backups exist (at least one .1 file)
	bp := backupPath(path, 1)
	if _, err := os.Stat(bp); err != nil {
		t.Errorf("backup file %s missing: %v", bp, err)
	}
}

func TestRotation_OldBackupsRemoved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "age.log")

	// Create a backup file with old mod time
	bp := backupPath(path, 5)
	os.WriteFile(bp, []byte("old"), 0o644)
	oldTime := time.Now().AddDate(0, 0, -60) // 60 days ago
	os.Chtimes(bp, oldTime, oldTime)

	r := Rotation{MaxSize: 10 * 1024 * 1024, MaxBackups: 3, MaxAge: 30}
	rotate(path, r)

	// Old backup should be removed
	if _, err := os.Stat(bp); err == nil {
		t.Errorf("old backup %s should have been removed", bp)
	}
}

// appendWithRotation is like Append but uses a custom rotation policy.
func appendWithRotation(path, runID string, level Level, component, message string, r Rotation) error {
	line := formatLine(runID, level, component, message)

	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0o755)

	if needsRotation(path, r.MaxSize) {
		rotate(path, r)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, line) // using fmt here is fine — test helper
	return nil
}

func TestKVHelper(t *testing.T) {
	result := KV("skill", "ve-ecs-ops")
	if result != "skill=ve-ecs-ops" {
		t.Errorf("expected skill=ve-ecs-ops, got %s", result)
	}
}

func TestLevelString(t *testing.T) {
	if DEBUG.String() != "DEBUG" {
		t.Errorf("DEBUG string mismatch")
	}
	if INFO.String() != "INFO" {
		t.Errorf("INFO string mismatch")
	}
	if WARN.String() != "WARN" {
		t.Errorf("WARN string mismatch")
	}
	if ERROR.String() != "ERROR" {
		t.Errorf("ERROR string mismatch")
	}
}
