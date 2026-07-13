package links

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// writeFile creates a file (and its parent dirs) with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeTarget(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"docs/gcl-spec.md", "docs/gcl-spec.md"},
		{"  docs/gcl-spec.md  ", "docs/gcl-spec.md"},
		{"#anchor", ""},
		{"https://example.com", ""},
		{"http://example.com", ""},
		{"mailto:a@b.com", ""},
		{"docs/x.md#sec", "docs/x.md"},
		{"docs/x.md?q=1", "docs/x.md"},
		{"<http://x.com>", ""},
		{"", ""},
		// "see file at line N" notations resolve to the file itself.
		{"docs/reflexion-memory.md:63", "docs/reflexion-memory.md"},
		{"docs/reflexion-memory.md:76", "docs/reflexion-memory.md"},
		{"ve-ecs-ops/SKILL.md:769-772", "ve-ecs-ops/SKILL.md"},
		{"ve-ecs-ops/SKILL.md:769", "ve-ecs-ops/SKILL.md"},
		{"ve-skill-generator/references/enhanced-self-healing-framework.md:29",
			"ve-skill-generator/references/enhanced-self-healing-framework.md"},
		// Non-numeric suffix must be left untouched.
		{"docs/x.md:note", "docs/x.md:note"},
	}
	for _, c := range cases {
		if got := normalizeTarget(c.in); got != c.want {
			t.Errorf("normalizeTarget(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLooksLikeRepoPath(t *testing.T) {
	if !looksLikeRepoPath("docs/gcl-spec.md") {
		t.Error("docs/gcl-spec.md should look like a repo path")
	}
	if looksLikeRepoPath("ve-darwin-arm64") {
		t.Error("release binary name must not look like a repo path")
	}
	if looksLikeRepoPath("http://x.com") {
		t.Error("URL must not look like a repo path")
	}
	if looksLikeRepoPath("some random text") {
		t.Error("text with spaces must not look like a repo path")
	}
}

func TestCheckDir(t *testing.T) {
	root := t.TempDir()

	// Valid doc with resolvable links (relative + repo-prefixed).
	valid := "See [spec](docs/gcl-spec.md) and [root](README.md).\n"
	valid += "Backtick path `docs/token-efficiency.md` is valid.\n"
	writeFile(t, filepath.Join(root, "AGENTS.md"), valid)
	writeFile(t, filepath.Join(root, "README.md"), "# readme\n")
	writeFile(t, filepath.Join(root, "docs", "gcl-spec.md"), "# gcl\n")
	writeFile(t, filepath.Join(root, "docs", "token-efficiency.md"), "# te\n")

	// doc with a broken link + broken backtick path.
	broken := "Broken [x](docs/missing.md) and good [y](README.md).\n"
	broken += "Bad backtick `docs/ghost.md`.\n"
	broken += "Ignore http://example.com and #anchor and ve-darwin-arm64.\n"
	writeFile(t, filepath.Join(root, "docs", "broken.md"), broken)

	results, files := CheckDir(root)

	sort.Strings(files)
	if len(files) != 5 {
		t.Fatalf("expected 5 navigation-root files, got %d: %v", len(files), files)
	}

	brokenPath := filepath.Join(root, "docs", "broken.md")
	errs, ok := results[brokenPath]
	if !ok {
		t.Fatalf("expected errors for %s, got none (results=%v)", brokenPath, results)
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 broken refs in broken.md, got %d: %v", len(errs), errs)
	}
	if !containsStr(errs, "missing markdown link target: docs/missing.md") {
		t.Errorf("missing link not detected: %v", errs)
	}
	if !containsStr(errs, "missing backtick path target: docs/ghost.md") {
		t.Errorf("missing backtick path not detected: %v", errs)
	}

	// Valid files must produce no errors.
	for _, f := range files {
		if f == brokenPath {
			continue
		}
		if e, ok := results[f]; ok && len(e) > 0 {
			t.Errorf("unexpected errors for %s: %v", f, e)
		}
	}
}

func containsStr(s []string, sub string) bool {
	for _, v := range s {
		if v == sub {
			return true
		}
	}
	return false
}
