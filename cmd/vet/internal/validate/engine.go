package validate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Step is a single validation step, mirroring the Python Step dataclass.
type Step struct {
	name string
	argv []string // command to execute externally (python3 ...)
}

// python3 returns the python interpreter path, preferring "python3".
func python3() string {
	if p, err := exec.LookPath("python3"); err == nil {
		return p
	}
	return "python3"
}

// buildSteps constructs the full validation step list, faithfully mirroring
// build_steps() in scripts/validate_local.py. Inline-python checks are kept as
// pure-Go helpers (see validate.go) but are ALSO emitted as `python3 -c`
// commands below so the external-execution semantics are preserved: when
// runLive is true, runStep executes the argv via os/exec exactly like the
// original. For the four inline checks we additionally expose the generated
// Python source via inlineScriptSource so the behavior is verifiable.
func buildSteps() []Step {
	py := python3()
	steps := []Step{
		{name: "File integrity (null byte check)", argv: []string{py, "-c", inlineScriptSource("file_integrity")}},
		{name: "Frontmatter validation", argv: []string{py, "scripts/validate_skills_frontmatter.py"}},
		{name: "Required sections presence", argv: []string{py, "-c", inlineScriptSource("required_sections")}},
		{name: "Error Taxonomy (≥10 codes, HALT/RETRY)", argv: []string{py, "-c", inlineScriptSource("error_taxonomy")}},
		{name: "TE-1 hardcoded version scan", argv: []string{py, "-c", inlineScriptSource("te1_hardcodes")}},
		{name: "Markdown local links", argv: []string{py, "scripts/check_markdown_links.py"}},
		{
			name: "GCL runner smoke test",
			argv: []string{py, "scripts/gcl_runner.py", "run", "--skill", "ve-skill-generator",
				"--request", "CI smoke test", "--command",
				`echo {"Response":{"RequestId":"ci-smoke"}}`, "--max-iter", "1", "--structural-critic-only"},
		},
		{name: "GCL CI gate (all skills)", argv: []string{py, "scripts/gcl_ci_gate.py", "--skip-incident-loop"}},
		{name: "GCL trace aggregate", argv: []string{py, "scripts/gcl_trace_aggregate.py", "--since-hours", "168"}},
		{name: "Script unit tests", argv: []string{py, "-m", "unittest", "discover", "-s", "scripts", "-p", "*_test.py", "-v"}},
		{name: "GCL Tier-A conformance", argv: []string{py, "scripts/check_gcl_conformance.py"}},
		{name: "Eval regression", argv: []string{py, "scripts/check_eval_regression.py"}},
	}
	return steps
}

// inlineScriptSource returns the generated `python3 -c` source for an inline
// check, mirroring _inline_script in the Python original (function body
// dedented/indented under a main() that calls sys.exit(main())).
func inlineScriptSource(which string) string {
	return "import sys\nfrom pathlib import Path\ndef main():\n    root = Path.cwd()\n" +
		inlineScriptBody(which) +
		"\nsys.exit(main())\n"
}

func inlineScriptBody(which string) string {
	switch which {
	case "file_integrity":
		return "    errors = []\n" +
			"    for f in sorted(root.glob('ve-*/SKILL.md')):\n" +
			"        if b'\\x00' in f.read_bytes():\n" +
			"            errors.append(str(f.relative_to(root)) + ': contains null bytes')\n" +
			"    if errors:\n" +
			"        for e in errors:\n" +
			"            print('  FAIL: ' + e)\n" +
			"        return 1\n" +
			"    print('  OK: all SKILL.md files are clean UTF-8 text')\n" +
			"    return 0"
	case "required_sections":
		return "    HARD = {'## Trigger & Scope', '## Quality Gate (GCL)'}\n" +
			"    errors = []\n" +
			"    warnings = []\n" +
			"    for f in sorted(root.glob('ve-*/SKILL.md')):\n" +
			"        skill = f.parent.name\n" +
			"        if skill == 've-skill-generator':\n" +
			"            continue\n" +
			"        text = f.read_text(encoding='utf-8')\n" +
			"        has_ts = '## Trigger & Scope' in text or '## Trigger & Scope (Agent-Readable)' in text\n" +
			"        has_shall = '### SHOULD Use This Skill When' in text\n" +
			"        has_shall_not = '### SHOULD NOT Use This Skill When' in text\n" +
			"        has_gcl = '## Quality Gate (GCL)' in text\n" +
			"        has_what = '### What This Skill Does' in text\n" +
			"        has_ops = '## Operational Best Practices' in text\n" +
			"        has_next = '### Next Steps' in text\n" +
			"        if not has_ts:\n" +
			"            errors.append(skill + ': missing ## Trigger & Scope')\n" +
			"        elif not has_shall or not has_shall_not:\n" +
			"            errors.append(skill + ': ## Trigger & Scope lacks SHOULD/SHOULD NOT subsections')\n" +
			"        if not has_gcl:\n" +
			"            errors.append(skill + ': missing ## Quality Gate (GCL)')\n" +
			"        if not has_what:\n" +
			"            errors.append(skill + ': missing ### What This Skill Does (IMPORTANT — MUST exist)')\n" +
			"        if not has_ops:\n" +
			"            errors.append(skill + ': missing ## Operational Best Practices (IMPORTANT — MUST exist)')\n" +
			"        if not has_next:\n" +
			"            warnings.append(skill + ': missing ### Next Steps')\n" +
			"    for e in errors:\n" +
			"        print('  FAIL: ' + e)\n" +
			"    for w in warnings:\n" +
			"        print('  WARN: ' + w)\n" +
			"    if errors:\n" +
			"        return 1\n" +
			"    print('  OK: all harness-critical sections present')\n" +
			"    return 0"
	case "error_taxonomy":
		return "    import re\n" +
			"    warnings = []\n" +
			"    for f in sorted(root.glob('ve-*/SKILL.md')):\n" +
			"        skill = f.parent.name\n" +
			"        if skill == 've-skill-generator':\n" +
			"            continue\n" +
			"        text = f.read_text(encoding='utf-8')\n" +
			"        if '## Error Taxonomy' not in text:\n" +
			"            warnings.append(skill + ': missing ## Error Taxonomy')\n" +
			"            continue\n" +
			"        classes = re.findall(r\"^\\|\\s*`[^`]+`\\s*\\|\\s*[^|]+?\\|\\s*[^|]*?\\*\\*(HALT|RETRY)\\*\\*\", text, re.MULTILINE)\n" +
			"        if len(classes) < 10:\n" +
			"            warnings.append(skill + ': ## Error Taxonomy has only ' + str(len(classes)) + ' codes, need ≥10')\n" +
			"        elif 'HALT' not in classes:\n" +
			"            warnings.append(skill + ': ## Error Taxonomy missing HALT classification')\n" +
			"        elif 'RETRY' not in classes:\n" +
			"            warnings.append(skill + ': ## Error Taxonomy missing RETRY classification')\n" +
			"    for w in warnings:\n" +
			"        print('  WARN: ' + w)\n" +
			"    if warnings:\n" +
			"        print('  → ' + str(len(warnings)) + ' error taxonomy issue(s) found (advisory)')\n" +
			"        return 0\n" +
			"    print('  OK: all skills have ## Error Taxonomy with ≥10 codes including HALT/RETRY')\n" +
			"    return 0"
	case "te1_hardcodes":
		return "    import re\n" +
			"    PATTERNS = [\n" +
			"        ('EngineVersion', r'\"EngineVersion\":\\s*\"\\d+\\.\\d+\"'),\n" +
			"        ('MongoVersion', r'\"MongoVersion\":\\s*\"\\d+\\.\\d+\"'),\n" +
			"        ('--MongoVersion', r'--MongoVersion\\s+\\d+\\.\\d+'),\n" +
			"        ('--Version', r'--Version\\s+\"\\d+\\.\\d+\"'),\n" +
			"        ('--TargetVersion', r'--TargetVersion\\s+\"\\d+\\.\\d+\"'),\n" +
			"    ]\n" +
			"    warnings = []\n" +
			"    for glob_pat in ('ve-*/references/cli-usage.md', 've-*/SKILL.md'):\n" +
			"        for f in sorted(root.glob(glob_pat)):\n" +
			"            text = f.read_text(encoding='utf-8')\n" +
			"            rel = f.relative_to(root)\n" +
			"            for label, pattern in PATTERNS:\n" +
			"                for m in re.finditer(pattern, text):\n" +
			"                    warnings.append(str(rel) + ': TE-1 hardcoded ' + label + ' → ' + m.group())\n" +
			"    for w in warnings:\n" +
			"        print('  WARN: ' + w)\n" +
			"    if not warnings:\n" +
			"        print('  OK: no hardcoded version numbers detected')\n" +
			"    else:\n" +
			"        print('  → ' + str(len(warnings)) + ' TE-1 candidate(s) found (advisory)')\n" +
			"    return 0"
	}
	return ""
}

// runStep executes one step via os/exec in cwd=root, mirroring run_step().
func runStep(root string, step Step) int {
	fmt.Fprintf(os.Stderr, "\n==> %s\n", step.name)
	fmt.Fprintf(os.Stderr, "$ %s\n", strings.Join(step.argv, " "))
	cmd := exec.Command(step.argv[0], step.argv[1:]...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return 1
	}
	return 0
}

// Run mirrors main(): executes every step in order, stopping at the first
// non-zero return. Returns (errors, failedStepNames) where errors maps each
// failed step name to its captured stderr lines, and the second value is the
// sorted list of failing step names (matching frontmatter.CheckDir's shape
// conceptually). When listOnly is true, no commands are executed; the step
// commands are printed instead and (nil, nil) is returned.
func Run(root string, listOnly bool) (map[string][]string, []string) {
	steps := buildSteps()
	if listOnly {
		for _, s := range steps {
			fmt.Printf("%s: %s\n", s.name, strings.Join(s.argv, " "))
		}
		return nil, nil
	}

	failErrs := make(map[string][]string)
	var failed []string
	for _, s := range steps {
		if rc := runStep(root, s); rc != 0 {
			fmt.Fprintf(os.Stderr, "\nFAILED: %s exited with %d\n", s.name, rc)
			failErrs[s.name] = []string{fmt.Sprintf("exited with %d", rc)}
			failed = append(failed, s.name)
			break
		}
	}
	sort.Strings(failed)
	if len(failed) == 0 {
		fmt.Fprintln(os.Stderr, "\nOK: local validation suite passed")
	}
	return failErrs, failed
}

// relSkillPath mirrors the Python f.relative_to(root) for a skill SKILL.md.
func relSkillPath(root, skillPath string) string {
	rel, err := filepath.Rel(root, skillPath)
	if err != nil {
		return skillPath
	}
	return rel
}
