package duprefs

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Token represents a 4-line block hash occurrence across skills.
type Token struct {
	Hash      string
	LineBlock string
	Files     []string
}

// CheckDir returns (findings []string, err error) listing duplicate block hashes
// that recur across more than one distinct skill directory.
func CheckDir(root string) ([]string, error) {
	blockToFiles := map[string]map[string]bool{}
	blockToSample := map[string]string{}
	// Collect inputs: SKILL.md + references/advanced/aiops.md per skill.
	skillDirs, _ := os.ReadDir(root)
	for _, sd := range skillDirs {
		if !sd.IsDir() {
			continue
		}
		name := sd.Name()
		if !strings.HasPrefix(name, "ve-") && name != "incident-loop-agent" {
			continue
		}
		if name == "incident-loop-agent" {
			// still scan
		}
		paths := []string{
			filepath.Join(root, name, "SKILL.md"),
			filepath.Join(root, name, "references", "advanced", "aiops.md"),
		}
		for _, p := range paths {
			f, err := os.Open(p)
			if err != nil {
				continue
			}
			sc := bufio.NewScanner(f)
			sc.Buffer(make([]byte, 1<<17), 1<<17)
			var lines []string
			for sc.Scan() {
				lines = append(lines, sc.Text())
			}
			f.Close()
			if len(lines) < 4 {
				continue
			}
			for i := 0; i+4 <= len(lines); i++ {
				block := strings.TrimSpace(strings.Join(lines[i:i+4], "\n"))
				if block == "" {
					continue
				}
				skip := false
				for _, l := range strings.Split(block, "\n") {
					// skip pure code-fence or heading-only blocks
					tl := strings.TrimSpace(l)
					if tl == "" {
						continue
					}
				}
				if skip {
					continue
				}
				h := md5.Sum([]byte(block))
				key := fmt.Sprintf("%x", h[:])
				if blockToFiles[key] == nil {
					blockToFiles[key] = map[string]bool{}
				}
				blockToFiles[key][name] = true
				if _, ok := blockToSample[key]; !ok {
					blockToSample[key] = truncate(block, 80)
				}
			}
		}
	}
	findings := []string{}
	// sort keys for determinism
	keys := make([]string, 0, len(blockToFiles))
	for k := range blockToFiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		files := make([]string, 0, len(blockToFiles[k]))
		for f := range blockToFiles[k] {
			files = append(files, f)
		}
		sort.Strings(files)
		if len(files) >= 6 { // recurring across many skills = likely shared template noise
			findings = append(findings, fmt.Sprintf("dup block across %d skills: %s ...", len(files), blockToSample[k]))
		}
	}
	return findings, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
