// Package eval ports scripts/check_eval_regression.py — intent-classification
// regression checking for all ve-* skills.
//
// It reads each skill's assets/eval_queries.json, validates its schema, parses
// the skill's "## Trigger & Scope" section, and verifies that queries labeled
// should_trigger:true match the SHOULD scope while should_trigger:false queries
// do not. It also supports git-diff semantic-drift detection by shelling out to
// git (os/exec), mirroring the Python subprocess calls.
package eval

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Minimum required eval query counts (mirrors the Python constants).
const (
	minTotal        = 8
	minTrigger      = 4
	minNonTrigger   = 3
	matchThreshold  = 0.15
	antiThreshold   = 0.75
	antiAntiMax     = 0.3
	minTokensFlag   = 2
	deltaLossThresh = 0.50
)

// exempt skills from eval regression (orchestration / meta skills).
var exempt = map[string]bool{
	"ve-skill-generator": true,
}

// defaultBaseRev is the git ref compared against when git-diff mode is enabled.
// Mirrors the Python --git-diff default ("HEAD~1"); for the bare CheckDir we
// compare against the merge-base with origin/main (falling back to main).
const defaultBaseRev = "origin/main"

// scoring stop-words excluded from significant-token coverage.
var stopWords = map[string]bool{
	"the": true, "this": true, "that": true, "with": true, "from": true,
	"when": true, "what": true, "how": true, "for": true, "not": true,
	"use": true, "can": true, "all": true, "are": true, "its": true,
	"has": true, "had": true, "but": true, "was": true, "were": true,
	"been": true, "being": true, "have": true, "does": true, "do": true,
	"will": true, "would": true, "should": true, "could": true, "may": true,
	"might": true, "shall": true, "also": true, "about": true, "into": true,
	"over": true, "both": true, "than": true, "then": true, "each": true,
	"other": true, "more": true, "most": true, "some": true, "any": true,
	"need": true,
}

// known Volcengine product acronyms — used for cross-lingual Chinese↔English matching.
var productAcronyms = []string{
	"ecs", "vpc", "rds", "clb", "alb", "iam", "kms", "nat", "vpn",
	"vke", "tos", "cdn", "dns", "sls", "nas", "fg", "ark", "cms",
	"redis", "kafka", "mongodb", "elasticsearch", "polar", "eip",
}

// common Chinese operation/domain keywords kept as tokens for matching.
var cnKeywords = []string{
	"创建", "删除", "查询", "修改", "配置", "启动", "停止", "重启", "扩容",
	"缩容", "备份", "恢复", "迁移", "监控", "告警", "日志", "权限", "策略",
	"安全", "网络", "存储", "实例", "集群", "节点", "账户", "账单", "费用",
	"巡检", "优化", "分析", "诊断", "评估", "禁用", "部署", "释放", "开启",
	"关闭", "申请", "查看", "导出", "设置", "添加", "加入", "授权", "克隆",
	"批量", "取消", "转换", "加速", "轮换", "同步", "暂停", "刷新", "下载",
	"上传", "挂载",
}

var (
	wordRe     = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9\-_]{2,}`)
	triggerSec = regexp.MustCompile(`(?m)^## Trigger & Scope(?: \(Agent-Readable\))?\s*$`)
	diffFileRe = regexp.MustCompile(`ve-([^/]+)-ops/`)
)

// findTriggerScopeSection extracts the "## Trigger & Scope" section body.
// RE2 lacks lookahead, so we locate the heading then slice until the next
// top-level "## " heading (or EOF). Faithful to the Python DOTALL lookahead.
func findTriggerScopeSection(text string) string {
	m := triggerSec.FindStringIndex(text)
	if m == nil {
		return ""
	}
	body := text[m[1]:]
	// slice at the next "\n## " top-level heading
	if idx := strings.Index(body, "\n## "); idx >= 0 {
		body = body[:idx]
	}
	return strings.TrimSpace(body)
}

// acronyms and Chinese keywords found in text. Faithful port of _tokenize.
func Tokenize(text string) map[string]bool {
	tokens := make(map[string]bool)
	for _, m := range wordRe.FindAllString(text, -1) {
		tokens[strings.ToLower(m)] = true
	}
	for _, acro := range productAcronyms {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(acro) + `\b`)
		if re.MatchString(text) {
			tokens[acro] = true
		}
	}
	for _, kw := range cnKeywords {
		if strings.Contains(text, kw) {
			tokens[kw] = true
		}
	}
	return tokens
}

// tokenizeSet wraps Tokenize returning a set-like map.
func tokenizeSet(text string) map[string]bool {
	return Tokenize(text)
}

// scoreMatches returns the fraction of query tokens present in scope tokens
// (excluding stop-words). Faithful port of _score_matches.
func scoreMatches(queryTokens, scopeTokens map[string]bool) float64 {
	if len(queryTokens) == 0 {
		return 0.0
	}
	significant := make(map[string]bool)
	for t := range queryTokens {
		if !stopWords[t] {
			significant[t] = true
		}
	}
	if len(significant) == 0 {
		return 0.0
	}
	matched := 0
	for t := range significant {
		if scopeTokens[t] {
			matched++
		}
	}
	return float64(matched) / float64(len(significant))
}

// hasEnglishTokens reports whether text contains extractable English words.
func hasEnglishTokens(text string) bool {
	return wordRe.MatchString(text)
}

// validateEvalSchema validates the shape of an eval_queries.json file.
func validateEvalSchema(data []map[string]any, skillDir string) []string {
	var errors []string
	requiredKeys := []string{"query", "should_trigger", "skill", "confidence"}
	validConf := map[string]bool{"high": true, "medium": true, "low": true}

	if len(data) == 0 {
		return []string{skillDir + ": eval_queries.json is empty"}
	}

	for i, entry := range data {
		if entry == nil {
			errors = append(errors, skillDir+": entry["+itoa(i)+"] is not a dict")
			continue
		}
		for _, k := range requiredKeys {
			if _, ok := entry[k]; !ok {
				errors = append(errors, skillDir+": entry["+itoa(i)+"] missing keys: {"+k+"}")
			}
		}
		q, ok := entry["query"].(string)
		if !ok || strings.TrimSpace(q) == "" {
			errors = append(errors, skillDir+": entry["+itoa(i)+"] 'query' must be non-empty string")
		}
		if b, ok := entry["should_trigger"].(bool); !ok || (b != true && b != false) {
			errors = append(errors, skillDir+": entry["+itoa(i)+"] 'should_trigger' must be bool")
		}
		s, ok := entry["skill"].(string)
		if !ok || strings.TrimSpace(s) == "" {
			errors = append(errors, skillDir+": entry["+itoa(i)+"] 'skill' must be non-empty string")
		}
		c, _ := entry["confidence"].(string)
		if !validConf[c] {
			errors = append(errors, skillDir+": entry["+itoa(i)+"] 'confidence' must be one of high|medium|low")
		}
	}
	return errors
}

// extractSectionBullets returns bullet items under a specific ### heading.
// RE2 lacks lookahead, so after locating the heading we slice its body up to
// the next "### " heading or EOF. Faithful to the Python section regex.
func extractSectionBullets(scopeText, heading string) []string {
	headingPat := "^### " + regexp.QuoteMeta(heading) + `\s*$`
	re := regexp.MustCompile(`(?m)` + headingPat)
	m := re.FindStringIndex(scopeText)
	if m == nil {
		return nil
	}
	body := scopeText[m[1]:]
	if idx := strings.Index(body, "\n### "); idx >= 0 {
		body = body[:idx]
	}
	var bullets []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			bullets = append(bullets, line[2:])
		}
	}
	return bullets
}

// checkSkill runs eval regression against one skill directory.
func checkSkill(skillDir string) (string, []string) {
	name := filepath.Base(skillDir)
	var errors []string

	evalPath := filepath.Join(skillDir, "assets", "eval_queries.json")
	raw, err := os.ReadFile(evalPath)
	if err != nil {
		return name, []string{name + ": missing assets/eval_queries.json"}
	}
	var data []map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return name, []string{name + ": assets/eval_queries.json parse error: " + err.Error()}
	}

	schemaErrors := validateEvalSchema(data, name)
	if len(schemaErrors) > 0 {
		return name, schemaErrors
	}

	var triggers, nonTriggers []map[string]any
	for _, e := range data {
		if b, ok := e["should_trigger"].(bool); ok && b {
			triggers = append(triggers, e)
		} else {
			nonTriggers = append(nonTriggers, e)
		}
	}

	var issues []string
	if len(data) < minTotal {
		issues = append(issues, "only "+itoa(len(data))+" entries (<"+itoa(minTotal)+")")
	}
	if len(triggers) < minTrigger {
		issues = append(issues, "only "+itoa(len(triggers))+" should_trigger=true (<"+itoa(minTrigger)+")")
	}
	if len(nonTriggers) < minNonTrigger {
		issues = append(issues, "only "+itoa(len(nonTriggers))+" should_trigger=false (<"+itoa(minNonTrigger)+")")
	}
	for _, e := range data {
		if s, ok := e["skill"].(string); !ok || s != name {
			issues = append(issues, "entry 'skill' mismatch: expected '"+name+"', got '"+stringOr(e["skill"])+"'")
		}
	}
	if len(issues) > 0 {
		for _, i := range issues {
			errors = append(errors, name+": "+i)
		}
		if len(triggers) == 0 && len(nonTriggers) == 0 {
			return name, errors
		}
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	skillText, err := os.ReadFile(skillPath)
	if err != nil {
		return name, append(errors, name+": missing SKILL.md")
	}
	scopeSection := findTriggerScopeSection(string(skillText))
	if scopeSection == "" {
		return name, append(errors, name+": cannot find ## Trigger & Scope section in SKILL.md")
	}

	shouldBullets := extractSectionBullets(scopeSection, "SHOULD Use This Skill When")
	shouldNotBullets := extractSectionBullets(scopeSection, "SHOULD NOT Use This Skill When")
	shouldTokens := tokenizeSet(strings.Join(shouldBullets, " "))
	shouldNotTokens := tokenizeSet(strings.Join(shouldNotBullets, " "))

	for _, entry := range data {
		query, _ := entry["query"].(string)
		shouldTrigger, _ := entry["should_trigger"].(bool)
		qt := tokenizeSet(query)
		matchScore := scoreMatches(qt, shouldTokens)
		antiScore := 0.0
		if len(shouldNotTokens) > 0 {
			antiScore = scoreMatches(qt, shouldNotTokens)
		}
		sigTokens := make(map[string]bool)
		for t := range qt {
			if !stopWords[t] {
				sigTokens[t] = true
			}
		}
		nTokens := len(sigTokens)

		// Skip content regression for queries with no English tokens (cannot
		// do meaningful cross-lingual matching against English scope).
		if !hasEnglishTokens(query) {
			continue
		}

		if shouldTrigger && matchScore < matchThreshold {
			errors = append(errors, name+": SHOULD trigger but low scope match "+
				"(score="+fmtFloat(matchScore)+", tokens="+itoa(nTokens)+"): "+query)
		} else if !shouldTrigger && nTokens >= minTokensFlag && matchScore > antiThreshold {
			if antiScore < antiAntiMax {
				errors = append(errors, name+": SHOULD NOT trigger but high scope match "+
					"(score="+fmtFloat(matchScore)+", tokens="+itoa(nTokens)+", "+
					"anti="+fmtFloat(antiScore)+"): "+query)
			}
		}
	}
	return name, errors
}

// ── git-diff semantic drift ─────────────────────────────────────────

// getChangedSkillsByGit finds skills whose "## Trigger & Scope" changed vs
// baseRev. Faithful port of _get_changed_skills_by_git.
func getChangedSkillsByGit(root, baseRev string) map[string]bool {
	changed := make(map[string]bool)
	base, err := gitOutput(root, "merge-base", "HEAD", baseRev)
	if err != nil {
		return changed
	}
	base = strings.TrimSpace(base)
	diff, err := gitOutput(root, "diff", base, "--", "ve-*/SKILL.md")
	if err != nil {
		return changed
	}
	var currentSkill string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "+++ b/") {
			if m := diffFileRe.FindStringSubmatch(line); m != nil {
				currentSkill = "ve-" + m[1] + "-ops"
			} else {
				currentSkill = ""
			}
		}
		if strings.HasPrefix(line, "@@") && currentSkill != "" {
			if strings.Contains(line, "Trigger & Scope") {
				changed[currentSkill] = true
			}
		}
	}
	return changed
}

// getOldScope retrieves the old "## Trigger & Scope" from git history.
func getOldScope(root, skillDir, baseRev string) string {
	rel, err := filepath.Rel(root, skillDir)
	if err != nil {
		return ""
	}
	base, err := gitOutput(root, "merge-base", "HEAD", baseRev)
	if err != nil {
		return ""
	}
	base = strings.TrimSpace(base)
	path := rel + "/SKILL.md"
	old, err := gitOutput(root, "show", base+":"+path)
	if err != nil {
		return ""
	}
	return findTriggerScopeSection(old)
}

// checkDelta compares eval query coverage against old vs new scope.
func checkDelta(name string, data []map[string]any, oldScope, newScope string) []string {
	var errors []string
	oldTokens := tokenizeSet(strings.Join(extractSectionBullets(oldScope, "SHOULD Use This Skill When"), " "))
	newTokens := tokenizeSet(strings.Join(extractSectionBullets(newScope, "SHOULD Use This Skill When"), " "))

	for _, entry := range data {
		b, _ := entry["should_trigger"].(bool)
		if !b {
			continue
		}
		query, _ := entry["query"].(string)
		if !hasEnglishTokens(query) {
			continue
		}
		qt := tokenizeSet(query)
		sig := make(map[string]bool)
		for t := range qt {
			if !stopWords[t] {
				sig[t] = true
			}
		}
		if len(sig) == 0 {
			continue
		}
		oldScore := scoreMatches(qt, oldTokens)
		newScore := scoreMatches(qt, newTokens)
		if oldScore < matchThreshold {
			continue
		}
		relativeLoss := 0.0
		if oldScore > 0 {
			relativeLoss = (oldScore - newScore) / oldScore
		}
		if relativeLoss >= deltaLossThresh {
			errors = append(errors, name+": [GIT-DIFF] scope change dropped "+
				itoa(int(relativeLoss*100))+"% coverage (old="+
				fmtFloat(oldScore)+" → new="+fmtFloat(newScore)+"): "+query)
		}
	}
	return errors
}

// gitOutput runs a git subcommand in dir and returns its stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// ── helpers ─────────────────────────────────────────────────────────

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func fmtFloat(f float64) string {
	neg := f < 0
	if neg {
		f = -f
	}
	d := int((f + 0.0000001) * 100)
	if d > 99 {
		d = 99
	}
	s := itoa(d/10) + "." + itoa(d%10)
	if neg {
		s = "-" + s
	}
	return s
}

func stringOr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "<nil>"
}

// ── public API ──────────────────────────────────────────────────────

// CheckDir validates every ve-*/SKILL.md under root and returns a per-file
// error map (only failing skills are present, keyed by the SKILL.md path) plus
// the sorted skill list. The signature mirrors the frontmatter package.
//
// It runs Pass 1 (absolute match) only. Git-diff semantic-drift detection is
// available via CheckDirGitDiff for callers that opt in.
func CheckDir(root string) (map[string][]string, []string) {
	skills := sortedSkills(root)
	results := make(map[string][]string)
	for _, sk := range skills {
		name := filepath.Base(filepath.Dir(sk))
		if exempt[name] {
			continue
		}
		_, errs := checkSkill(filepath.Dir(sk))
		if len(errs) > 0 {
			results[sk] = errs
		}
	}
	return results, skills
}

// CheckDirGitDiff runs Pass 1 plus Pass 2 git-diff semantic drift against
// baseRev (or defaultBaseRev when empty). Errors keyed by SKILL.md path.
func CheckDirGitDiff(root, baseRev string) (map[string][]string, []string) {
	if baseRev == "" {
		baseRev = defaultBaseRev
	}
	results, skills := CheckDir(root)
	changed := getChangedSkillsByGit(root, baseRev)
	if len(changed) == 0 {
		return results, skills
	}
	for _, sk := range skills {
		name := filepath.Base(filepath.Dir(sk))
		if exempt[name] {
			continue
		}
		if !changed[name] {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(filepath.Dir(sk), "assets", "eval_queries.json"))
		if err != nil {
			continue
		}
		var data []map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			continue
		}
		skillText, err := os.ReadFile(sk)
		if err != nil {
			continue
		}
		newScope := findTriggerScopeSection(string(skillText))
		if newScope == "" {
			continue
		}
		oldScope := getOldScope(root, filepath.Dir(sk), baseRev)
		if oldScope == "" {
			continue
		}
		if delta := checkDelta(name, data, oldScope, newScope); len(delta) > 0 {
			results[sk] = append(results[sk], delta...)
		}
	}
	return results, skills
}

func sortedSkills(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "ve-") {
			out = append(out, filepath.Join(root, e.Name(), "SKILL.md"))
		}
	}
	sort.Strings(out)
	return out
}
