// Package links validates local Markdown links and explicit repository path
// references.
//
// It is a faithful Go port of scripts/check_markdown_links.py. The checker
// walks the navigation-root markdown docs (AGENTS.md, README.md, README_CN.md,
// and top-level docs/*.md), extracts internal links and backtick-quoted repo
// paths, and verifies the referenced targets exist on disk.
package links

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Finding describes a single broken reference inside a file.
type Finding struct {
	File   string
	Line   int
	Target string
	Reason string
}

var ignoredDirParts = map[string]bool{
	".git":          true,
	".github":       true,
	".omc":          true,
	".omo":          true,
	".codebuddy":    true,
	"audit-results": true,
}

var ignoredMarkdownPrefixes = []string{
	"docs/superpowers/plans/",
}

var pathPrefixes = []string{
	"AGENTS.md",
	"README.md",
	"README_CN.md",
	"LICENSE",
	"docs/",
	"scripts/",
	"ve-",
	".github/",
}

// binaryNameRe matches release artifacts that look like repo paths but are not.
var binaryNameRe = regexp.MustCompile(`^ve-(darwin|linux|windows)-(amd64|arm64|x86_64)(\.exe)?$`)

// linkRe matches [text](target). Go's RE2 has no lookbehind, so image embeds
// ![](...) are filtered out separately by checking for a leading '!'.
var linkRe = regexp.MustCompile(`\[[^\]]+\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

// backtickRe matches `quoted` spans.
var backtickRe = regexp.MustCompile("`([^`]+)`")

// iterMarkdownFiles returns the navigation-root markdown files, sorted by
// repo-relative path.
func iterMarkdownFiles(root string) []string {
	candidates := []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "README.md"),
		filepath.Join(root, "README_CN.md"),
	}
	docsDir := filepath.Join(root, "docs")
	if entries, err := os.ReadDir(docsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			candidates = append(candidates, filepath.Join(docsDir, e.Name()))
		}
	}

	var files []string
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			continue
		}
		rel := fileRel(root, path)
		for _, part := range strings.Split(rel, "/") {
			if ignoredDirParts[part] {
				goto skip
			}
		}
		for _, p := range ignoredMarkdownPrefixes {
			if strings.HasPrefix(rel, p) {
				goto skip
			}
		}
		files = append(files, path)
	skip:
	}
	sort.Strings(files)
	return files
}

// normalizeTarget strips a link target down to a checkable filesystem path.
// Returns "" when the target should be ignored (URLs, anchors, mailto, etc.).
func normalizeTarget(raw string) string {
	target := strings.TrimSpace(raw)
	if target == "" {
		return ""
	}
	if strings.HasPrefix(target, "http://") ||
		strings.HasPrefix(target, "https://") ||
		strings.HasPrefix(target, "mailto:") ||
		strings.HasPrefix(target, "#") {
		return ""
	}
	if strings.HasPrefix(target, "<") && strings.HasSuffix(target, ">") {
		return ""
	}
	if i := strings.Index(target, "#"); i >= 0 {
		target = target[:i]
	}
	if i := strings.Index(target, "?"); i >= 0 {
		target = target[:i]
	}
	return target
}

// looksLikeRepoPath reports whether a backtick-quoted token is an explicit
// repository path reference worth validating.
func looksLikeRepoPath(text string) bool {
	if strings.ContainsAny(text, " \t\n") {
		return false
	}
	if strings.HasPrefix(text, "http://") ||
		strings.HasPrefix(text, "https://") ||
		strings.HasPrefix(text, "mailto:") ||
		strings.HasPrefix(text, "#") ||
		strings.HasPrefix(text, "{{") ||
		strings.HasPrefix(text, "<") {
		return false
	}
	if strings.ContainsAny(text, "<>") {
		return false
	}
	for _, sym := range []string{"*", "|", "--", "=", "[", "]", "{", "}"} {
		if strings.Contains(text, sym) {
			return false
		}
	}
	if binaryNameRe.MatchString(text) {
		return false
	}
	for _, p := range pathPrefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

// resolveTarget maps a normalized target to an absolute filesystem path.
func resolveTarget(root, source, target string) string {
	candidate := target
	if filepath.IsAbs(candidate) {
		return candidate
	}
	for _, p := range pathPrefixes {
		if strings.HasPrefix(target, p) {
			return filepath.Join(root, candidate)
		}
	}
	return filepath.Join(filepath.Dir(source), candidate)
}

// targetExists reports whether the resolved target is present on disk.
func targetExists(root, source, target string) bool {
	path := resolveTarget(root, source, target)
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		if part == "*" || part == "..." {
			return true
		}
	}
	_, err := os.Stat(path)
	return err == nil
}

// checkFile scans a single markdown file and returns its findings.
func checkFile(root, path string) []Finding {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var findings []Finding
	for lineNo, line := range strings.Split(string(data), "\n") {
		for _, m := range linkRe.FindAllStringSubmatchIndex(line, -1) {
			// Skip image embeds: ! immediately before the '['.
			matchStart := m[0]
			if matchStart > 0 && line[matchStart-1] == '!' {
				continue
			}
			target := normalizeTarget(line[m[2]:m[3]])
			if target != "" && !targetExists(root, path, target) {
				findings = append(findings, Finding{
					File:   path,
					Line:   lineNo + 1,
					Target: target,
					Reason: "missing markdown link target",
				})
			}
		}
		for _, m := range backtickRe.FindAllStringSubmatch(line, -1) {
			raw := strings.TrimSpace(m[1])
			target := normalizeTarget(raw)
			if target == "" || !looksLikeRepoPath(target) {
				continue
			}
			if !targetExists(root, path, target) {
				findings = append(findings, Finding{
					File:   path,
					Line:   lineNo + 1,
					Target: target,
					Reason: "missing backtick path target",
				})
			}
		}
	}
	return findings
}

// CheckDir validates every navigation-root markdown file under root and
// returns a per-file error map (only files with errors are present) plus the
// sorted file list.
func CheckDir(root string) (map[string][]string, []string) {
	files := iterMarkdownFiles(root)
	results := make(map[string][]string)
	var filingList []string
	for _, f := range files {
		filingList = append(filingList, f)
		findings := checkFile(root, f)
		if len(findings) == 0 {
			continue
		}
		var errs []string
		for _, fd := range findings {
			errs = append(errs, fd.Reason+": "+fd.Target)
		}
		results[f] = errs
	}
	return results, filingList
}

// fileRel returns the repo-relative slash-separated path of p.
func fileRel(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(rel)
}
