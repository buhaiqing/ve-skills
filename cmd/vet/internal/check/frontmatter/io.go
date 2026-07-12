package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
)

func readFile(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}

func readDir(root string) ([]os.DirEntry, error) {
	return os.ReadDir(root)
}

func joinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// skillName extracts the "ve-xxx" directory name from a path like
// /repo/ve-ecs-ops/SKILL.md.
func skillName(skillPath string) string {
	dir := filepath.Dir(skillPath)
	return strings.TrimSuffix(filepath.Base(dir), "")
}
