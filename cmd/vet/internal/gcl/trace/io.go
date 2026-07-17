package trace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// stderr is a shared writer for diagnostic messages.
var stderr = os.Stderr

func readFileIfExists(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func mkdirAll(dir string) error { return os.MkdirAll(dir, 0o755) }

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeFile(path, string(b))
}

func joinLines(lines []string) string { return strings.Join(lines, "\n") }

func contains(s, sub string) bool { return strings.Contains(s, sub) }

// replaceBlock replaces the "## Extracted from GCL Traces ..." section up to
// the next "## " heading or EOF with newContent.
func replaceBlock(existing, marker, newContent string) string {
	idx := strings.Index(existing, marker)
	if idx < 0 {
		return existing + newContent + "\n"
	}
	// find end of block: next "\n## " after the marker
	rest := existing[idx:]
	next := strings.Index(rest[1:], "\n## ")
	var before, after string
	if next < 0 {
		before = existing[:idx]
		after = ""
	} else {
		end := idx + 1 + next
		before = existing[:idx]
		after = existing[end:]
	}
	// Self-heal accumulated stray "---" separators: trim trailing blank lines
	// and orphan "---" blocks left by prior runs, then emit exactly one
	// separator so re-running never accumulates empty dividers.
	before = strings.TrimRight(before, "\n")
	for strings.HasSuffix(before, "---") {
		before = strings.TrimRight(before[:len(before)-3], "\n ")
	}
	return before + "\n\n---\n\n" + strings.TrimPrefix(newContent, "\n") + after
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		return ""
	}
}
