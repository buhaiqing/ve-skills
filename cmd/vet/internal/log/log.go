// Package log provides structured logging with automatic rotation for
// runtime log files. It implements the pipe-delimited key=value format
// specified in AGENTS.md §结构化日志与诊断能力.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Level represents a log severity level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "UNKNOWN"
}

// Rotation configures log file rotation.
type Rotation struct {
	MaxSize    int64 // max file size in bytes before rotation (default 10MB)
	MaxBackups int   // max old files to keep (default 5)
	MaxAge     int   // max days to keep (default 30)
}

// DefaultRotation returns the standard rotation policy.
func DefaultRotation() Rotation {
	return Rotation{MaxSize: 10 * 1024 * 1024, MaxBackups: 5, MaxAge: 30}
}

// Append writes a structured log line to the given file path.
// Format: <ISO_8601_ts> | [<runID>] | <level> | <component> | <message> | <key=value>...
// If the file exceeds MaxSize, it is rotated before writing.
func Append(path string, runID string, level Level, component, message string, kvs ...string) error {
	line := formatLine(runID, level, component, message, kvs...)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("log: mkdir %s: %w", dir, err)
	}

	// Check rotation before writing
	if needsRotation(path, DefaultRotation().MaxSize) {
		if err := rotate(path, DefaultRotation()); err != nil {
			// Rotation failed — still try to write (best-effort)
			fmt.Fprintf(os.Stderr, "WARN: log rotation failed for %s: %v\n", path, err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("log: open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, line); err != nil {
		return fmt.Errorf("log: write %s: %w", path, err)
	}
	return nil
}

// formatLine builds a pipe-delimited log line.
func formatLine(runID string, level Level, component, message string, kvs ...string) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	parts := []string{ts, "[" + runID + "]", level.String(), component, message}
	if len(kvs) > 0 {
		parts = append(parts, strings.Join(kvs, " "))
	}
	return strings.Join(parts, " | ")
}

// needsRotation returns true if the file exceeds maxSize.
func needsRotation(path string, maxSize int64) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false // file doesn't exist yet
	}
	return info.Size() >= maxSize
}

// rotate renames current file to .1, .1 to .2, etc., and removes files
// exceeding MaxBackups or MaxAge.
func rotate(path string, r Rotation) error {
	// Rotate existing backups: .2 → .3, .1 → .2, current → .1
	for i := r.MaxBackups - 1; i >= 0; i-- {
		oldPath := backupPath(path, i)
		newPath := backupPath(path, i+1)
		if i == 0 {
			oldPath = path
		}
		if _, err := os.Stat(oldPath); err == nil {
			if i+1 > r.MaxBackups {
				os.Remove(oldPath)
			} else {
				os.Rename(oldPath, newPath)
			}
		}
	}

	// Remove backups older than MaxAge (check all backup files in directory)
	if r.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -r.MaxAge)
		dir := filepath.Dir(path)
		prefix := filepath.Base(path) + "."
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) {
				bp := filepath.Join(dir, e.Name())
				if info, err := os.Stat(bp); err == nil && info.ModTime().Before(cutoff) {
					os.Remove(bp)
				}
			}
		}
	}

	return nil
}

// backupPath returns the path for backup number n (e.g., "file.log.1").
func backupPath(base string, n int) string {
	return base + fmt.Sprintf(".%d", n)
}

// Cleanup removes log backups exceeding MaxBackups or MaxAge.
// Should be called periodically (e.g., on startup).
func Cleanup(path string, r Rotation) {
	for i := r.MaxBackups + 1; i <= r.MaxBackups+10; i++ {
		bp := backupPath(path, i)
		os.Remove(bp) // best-effort
	}
	if r.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -r.MaxAge)
		entries, _ := os.ReadDir(filepath.Dir(path))
		prefix := filepath.Base(path) + "."
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) {
				if info, err := e.Info(); err == nil && info.ModTime().Before(cutoff) {
					os.Remove(filepath.Join(filepath.Dir(path), e.Name()))
				}
			}
		}
	}
}

// KVPair is a helper for building key=value pairs.
type KVPair struct {
	K, V string
}

// KV returns a "key=value" string.
func KV(k, v string) string {
	return k + "=" + v
}

// BuildKVs converts pairs into a sorted slice of "key=value" strings.
func BuildKVs(pairs ...KVPair) []string {
	var out []string
	for _, p := range pairs {
		out = append(out, KV(p.K, p.V))
	}
	sort.Strings(out)
	return out
}
