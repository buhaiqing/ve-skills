# L400 Wave A — 生产地基 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 L400 Capable 的 Wave A（真探针可 AUTO、最小 CI、在线 eval 周报、KB 持久化接线），使 ≥2 个 heal plan 可 `AllowProductionAuto`，且 Confirm 对非 Stub plan 不再因 stub 门挡 AUTO。

**Architecture:** 在既有 `heal` ProbeRunner / Stub 门与 `agent.Confirm` 之上，用 `Promote` 将 `cpu_high`、`redis_slow_query` 升级为带 `ProbeArgv` 的非 Stub 计划；新增 `agent/evalreport` 聚合 triage/GCL/误 REFUSE；`cmd/vet/Makefile` 固化门禁；`ProposeFix` 加载磁盘 KB。不做 Wave B/C（看板、自主域、L500）——见 ADR-0001/0005。

**Tech Stack:** Go 1.26、`cmd/vet` 模块、`go test` / `go vet`、现有 `heal`/`agent`/`strategy`/`gcl` 包；禁止 `sh -c`；无新第三方依赖。

**Upstream:**
- Roadmap: `docs/superpowers/plans/2026-08-01-l400-capable-roadmap.md` §2 Wave A
- ADRs: `docs/adr/0001`–`0002`（及 0005 排除项）；glossary: `docs/adr/glossary-l400.md`
- P0 已落地: Stub 门、`ValueMetrics`、`Plan()` getter

## Global Constraints

- ADR-0001: 只做 Wave A；B1 看板不在本计划
- ADR-0002: Stub ⇒ 禁止生产 AUTO；`AllowStub` 仅测试
- ADR-0005: 不做预测式 / multi-agent / 全量 promote
- AGENTS: `go build` + `go vet` 干净；Probe 必须 argv 直传；中文注释可少、代码英文
- 探针步骤只做 **read_only Describe\***，不在 Wave A 接真实扩缩容 Action
- ADRs 仍为 Proposed：实现以 Decision 正文为准；grilling Accept 后改 Status 不阻塞本计划

## File map

| File | Responsibility |
|------|----------------|
| `cmd/vet/internal/heal/promote.go` | `Promote(name, probes)`：写非 Stub + ProbeArgv |
| `cmd/vet/internal/heal/orchestrator.go` | `defaultPlans` 调用内置 promote 数据 **或** 启动时 Promote 两个 plan |
| `cmd/vet/internal/heal/promote_test.go` | Promote / AllowProductionAuto 测试 |
| `cmd/vet/internal/agent/confirm.go` | 无逻辑大改；补非 Stub → 可 AUTO 测试 |
| `cmd/vet/internal/agent/confirm_heal_test.go` | stub vs promoted Confirm |
| `cmd/vet/internal/agent/evalreport.go` | 在线 eval 聚合纯函数 + 写 JSON |
| `cmd/vet/internal/agent/evalreport_test.go` | 聚合测试 |
| `cmd/vet/agent.go` | 注册 `vet agent eval-report` |
| `cmd/vet/internal/strategy/kbpath.go` | 默认 KB 路径常量 |
| `cmd/vet/internal/agent/propose.go` | ProposeFix 启动时 Load KB |
| `cmd/vet/Makefile` | `test` / `vet` / `check-ci` targets |
| `docs/adr/0001-*.md` … | Task 收尾可选：Status → Accepted（若用户已 Accept） |

**Out of this plan (separate plans later):** Wave B (B1–B5), Wave C (C1–C6), ADR-0006 CLI 硬门（可挂 Wave B）。

---

### Task 1: heal.Promote — 升级命名 plan 为非 Stub

**Files:**
- Create: `cmd/vet/internal/heal/promote.go`
- Create: `cmd/vet/internal/heal/promote_test.go`
- Modify: `cmd/vet/internal/heal/orchestrator.go` (`defaultPlans` 或 `NewOrchestrator`)

**Interfaces:**
- Consumes: `Orchestrator.plans`, `RecoveryStep`, `IsStub`, `AllowProductionAuto`
- Produces:
  - `func (o *Orchestrator) Promote(name string, probes [][]string) error`
  - `func BuiltInPromotions() map[string][][]string` — 返回 `cpu_high` / `redis_slow_query` 的 argv 列表

- [ ] **Step 1: Write the failing test**

```go
// cmd/vet/internal/heal/promote_test.go
package heal

import "testing"

func TestPromoteMakesAllowProductionAuto(t *testing.T) {
	o := NewOrchestrator()
	if AllowProductionAuto(o.Plan("cpu_high")) {
		t.Fatal("precondition: cpu_high should start stub")
	}
	probes := [][]string{
		{"ve", "ecs", "DescribeInstances"},
		{"ve", "ecs", "DescribeInstances"},
	}
	if err := o.Promote("cpu_high", probes); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	p := o.Plan("cpu_high")
	if p.IsStub() {
		t.Fatal("expected non-stub after Promote")
	}
	if !AllowProductionAuto(p) {
		t.Fatal("expected AllowProductionAuto")
	}
	for i, s := range p.Steps {
		if s.Stub || len(s.ProbeArgv) == 0 {
			t.Fatalf("step %d still stub or empty ProbeArgv: %+v", i, s)
		}
	}
}

func TestPromoteUnknownPlan(t *testing.T) {
	o := NewOrchestrator()
	err := o.Promote("no_such", [][]string{{"ve", "ecs", "DescribeInstances"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPromoteProbeCountMismatch(t *testing.T) {
	o := NewOrchestrator()
	err := o.Promote("cpu_high", [][]string{{"ve", "ecs", "DescribeInstances"}})
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cmd/vet && go test ./internal/heal/ -run TestPromote -count=1`

Expected: FAIL — `Promote` undefined

- [ ] **Step 3: Write minimal implementation**

```go
// cmd/vet/internal/heal/promote.go
package heal

import "fmt"

// BuiltInPromotions maps incident types to per-step ProbeArgv (read-only).
func BuiltInPromotions() map[string][][]string {
	return map[string][][]string{
		"cpu_high": {
			{"ve", "ecs", "DescribeInstances"},
			{"ve", "ecs", "DescribeInstances"},
		},
		"redis_slow_query": {
			{"ve", "redis", "DescribeDBInstances"},
			{"ve", "redis", "DescribeDBInstances"},
			{"ve", "redis", "DescribeDBInstances"},
		},
	}
}

// Promote marks plan steps non-stub and attaches ProbeArgv.
// probes length MUST equal len(plan.Steps); each argv non-empty.
func (o *Orchestrator) Promote(name string, probes [][]string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	plan, ok := o.plans[name]
	if !ok {
		return fmt.Errorf("heal: unknown plan %q", name)
	}
	if len(probes) != len(plan.Steps) {
		return fmt.Errorf("heal: promote %q: need %d probes, got %d", name, len(plan.Steps), len(probes))
	}
	for i := range plan.Steps {
		if len(probes[i]) == 0 {
			return fmt.Errorf("heal: promote %q: empty probe at step %d", name, i)
		}
		plan.Steps[i].Stub = false
		plan.Steps[i].ProbeArgv = append([]string(nil), probes[i]...)
		plan.Steps[i].CheckFn = nil
	}
	return nil
}

// ApplyBuiltInPromotions promotes cpu_high and redis_slow_query.
func (o *Orchestrator) ApplyBuiltInPromotions() error {
	for name, probes := range BuiltInPromotions() {
		if err := o.Promote(name, probes); err != nil {
			return err
		}
	}
	return nil
}
```

在 `NewOrchestrator` 末尾（创建后）调用：

```go
func NewOrchestrator() *Orchestrator {
	o := &Orchestrator{
		plans:   defaultPlans(),
		circuit: &CircuitBreaker{threshold: 5, timeout: 30 * time.Second},
	}
	_ = o.ApplyBuiltInPromotions() // Wave A: cpu_high + redis_slow_query
	return o
}
```

注意：`Promote` 自身加锁；`ApplyBuiltInPromotions` 调 `Promote` 会死锁。改为：

```go
func (o *Orchestrator) ApplyBuiltInPromotions() error {
	for name, probes := range BuiltInPromotions() {
		if err := o.promoteLocked(name, probes); err != nil {
			return err
		}
	}
	return nil
}

// promoteLocked assumes o.mu not held; Promote wraps it with Lock.
func (o *Orchestrator) Promote(name string, probes [][]string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.promoteUnderLock(name, probes)
}

func (o *Orchestrator) promoteUnderLock(name string, probes [][]string) error { /* body */ }

func (o *Orchestrator) ApplyBuiltInPromotions() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for name, probes := range BuiltInPromotions() {
		if err := o.promoteUnderLock(name, probes); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: Update TestDefaultPlansAreStub**

`defaultPlans` 在 New 后已被 promote：改测试为：

```go
func TestDefaultPlansPromotedSubset(t *testing.T) {
	o := NewOrchestrator()
	if o.Plan("cpu_high").IsStub() {
		t.Fatal("cpu_high should be promoted")
	}
	if o.Plan("redis_slow_query").IsStub() {
		t.Fatal("redis_slow_query should be promoted")
	}
	if !o.Plan("mysql_connection_pool").IsStub() {
		t.Fatal("mysql should remain stub")
	}
	if !o.Plan("vpc_route_table").IsStub() {
		t.Fatal("vpc should remain stub")
	}
}
```

更新所有依赖「NewOrchestrator 后 cpu_high 为 stub」的 RejectStub 测试：改为测 `mysql_connection_pool`，或 `NewOrchestrator` 不用 Apply、单独测 stub fixture。

推荐：保留 `defaultPlans()` 仍全 Stub；仅 `NewOrchestrator` 调用 `ApplyBuiltInPromotions`。`TestExecutePlanRejectsStub` 用 `mysql_connection_pool`。

- [ ] **Step 5: Run tests**

Run: `cd cmd/vet && go test ./internal/heal/ -count=1`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/bohaiqing/opensource/git/ve-skills
git add cmd/vet/internal/heal/promote.go cmd/vet/internal/heal/promote_test.go cmd/vet/internal/heal/orchestrator.go cmd/vet/internal/heal/orchestrator_test.go
git commit -m "$(cat <<'EOF'
feat(heal): promote cpu_high and redis_slow_query with real ProbeArgv

Wave A / ADR-0002: enable AllowProductionAuto for two recovery plans.
EOF
)"
```

---

### Task 2: Confirm — 非 Stub heal 允许 AUTO

**Files:**
- Create: `cmd/vet/internal/agent/confirm_heal_test.go`
- Modify: none required if Task 1 makes `NewOrchestrator().Plan("cpu_high")` non-stub（Confirm 已有 stub 分支）

**Interfaces:**
- Consumes: `heal.NewOrchestrator().Plan`, `Confirm`
- Produces: 回归证明 promoted ⇒ AUTO；stub ⇒ ASK

- [ ] **Step 1: Write the failing/expected tests**

```go
// cmd/vet/internal/agent/confirm_heal_test.go
package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMinimalPolicy(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "incident-loop-agent", "references", "policies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "execution-risk.md"), []byte(`# Execution Risk Policy

## 2. Decision matrix

| risk | blast_radius | decision |
|------|--------------|----------|
| read-only | single | AUTO |
`), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "domain-allowlist.md"), []byte("# Domain\n\n## 1. Eligible skills\n\n`ve-ecs-ops`\n`ve-redis-ops`\n"), 0o644)
}

func autoPlan(healType, skill, cmd string) *DispatchPlan {
	return &DispatchPlan{
		Operations: []DispatchOp{{
			Skill: skill, Command: cmd, SafetyClass: "read_only",
			BlastRadius: "single", Confidence: "high", Safety: 1.0,
		}},
		BlastRadius:      "single",
		HealIncidentType: healType,
	}
}

func TestConfirmPromotedHealAllowsAUTO(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	res := Confirm(root, autoPlan("cpu_high", "ve-ecs-ops", "ve ecs DescribeInstances"))
	if res.Decision != "AUTO" {
		t.Fatalf("got %s (%s), want AUTO", res.Decision, res.Reason)
	}
}

func TestConfirmStubHealStillASK(t *testing.T) {
	root := t.TempDir()
	writeMinimalPolicy(t, root)
	res := Confirm(root, autoPlan("mysql_connection_pool", "ve-ecs-ops", "ve ecs DescribeInstances"))
	if res.Decision != "ASK" || !strings.Contains(strings.ToLower(res.Reason), "stub") {
		t.Fatalf("got %s (%s), want ASK stub", res.Decision, res.Reason)
	}
}
```

- [ ] **Step 2: Run tests**

Run: `cd cmd/vet && go test ./internal/agent/ -run 'TestConfirmPromoted|TestConfirmStubHeal' -count=1`

Expected: PASS（依赖 Task 1）。若 Task 1 未合入则 Promoted 用例 FAIL。

- [ ] **Step 3: Commit**

```bash
git add cmd/vet/internal/agent/confirm_heal_test.go
git commit -m "$(cat <<'EOF'
test(agent): Confirm AUTO for promoted heal, ASK for remaining stubs

EOF
)"
```

---

### Task 3: 在线 eval 周报（纯函数 + CLI）

**Files:**
- Create: `cmd/vet/internal/agent/evalreport.go`
- Create: `cmd/vet/internal/agent/evalreport_test.go`
- Modify: `cmd/vet/agent.go` — 增加 `eval-report` 子命令

**Interfaces:**
- Consumes: 文件系统上 `audit-results/gcl-trace-*.json`（可选）、`audit-results/value-metrics.jsonl`、可选 hand-labeled JSON
- Produces:
  - `type EvalReport struct { TriageTop1Accuracy, GCLFirstPassRate, FalseRefuseRate float64; Samples int; ... }`
  - `func BuildEvalReport(in EvalReportInput) EvalReport`
  - `func WriteEvalReport(path string, r EvalReport) error`

设计（YAGNI）：不跑真实 LLM。输入为结构化样本：

```go
type EvalSample struct {
	PredictedSkill string `json:"predicted_skill"`
	LabeledSkill   string `json:"labeled_skill"` // 空则跳过 triage 分母
	GCLFirstPass   bool   `json:"gcl_first_pass"`
	PolicyDecision string `json:"policy_decision"` // AUTO|ASK|REFUSE
	ShouldRefuse   bool   `json:"should_refuse"`   // 标签：本应 REFUSE
}

type EvalReportInput struct {
	Samples []EvalSample
}
```

- [ ] **Step 1: Write the failing test**

```go
func TestBuildEvalReportRates(t *testing.T) {
	r := BuildEvalReport(EvalReportInput{Samples: []EvalSample{
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: true, PolicyDecision: "AUTO", ShouldRefuse: false},
		{PredictedSkill: "ve-redis-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: false, PolicyDecision: "REFUSE", ShouldRefuse: false},
		{PredictedSkill: "ve-ecs-ops", LabeledSkill: "ve-ecs-ops", GCLFirstPass: true, PolicyDecision: "REFUSE", ShouldRefuse: true},
	}})
	// triage: 2/3 labeled correct = 2/3
	if r.TriageTop1Accuracy < 0.66 || r.TriageTop1Accuracy > 0.67 {
		t.Fatalf("triage=%v", r.TriageTop1Accuracy)
	}
	// GCL first pass: 2/3
	if r.GCLFirstPassRate < 0.66 || r.GCLFirstPassRate > 0.67 {
		t.Fatalf("gcl=%v", r.GCLFirstPassRate)
	}
	// false refuse: REFUSE && !ShouldRefuse → 1 / (REFUSE count=2) = 0.5
	if r.FalseRefuseRate != 0.5 {
		t.Fatalf("falseRefuse=%v", r.FalseRefuseRate)
	}
	if r.Samples != 3 {
		t.Fatalf("samples=%d", r.Samples)
	}
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `cd cmd/vet && go test ./internal/agent/ -run TestBuildEvalReportRates -count=1`

- [ ] **Step 3: Implement**

```go
// cmd/vet/internal/agent/evalreport.go
package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EvalSample struct {
	PredictedSkill string `json:"predicted_skill"`
	LabeledSkill   string `json:"labeled_skill"`
	GCLFirstPass   bool   `json:"gcl_first_pass"`
	PolicyDecision string `json:"policy_decision"`
	ShouldRefuse   bool   `json:"should_refuse"`
}

type EvalReportInput struct {
	Samples []EvalSample `json:"samples"`
}

type EvalReport struct {
	TriageTop1Accuracy float64 `json:"triage_top1_accuracy"`
	GCLFirstPassRate   float64 `json:"gcl_first_pass_rate"`
	FalseRefuseRate    float64 `json:"false_refuse_rate"`
	Samples            int     `json:"samples"`
}

func BuildEvalReport(in EvalReportInput) EvalReport {
	r := EvalReport{Samples: len(in.Samples)}
	var triageOK, triageN, gclOK, refuseN, falseRefuse int
	for _, s := range in.Samples {
		if s.LabeledSkill != "" {
			triageN++
			if s.PredictedSkill == s.LabeledSkill {
				triageOK++
			}
		}
		if s.GCLFirstPass {
			gclOK++
		}
		if s.PolicyDecision == "REFUSE" {
			refuseN++
			if !s.ShouldRefuse {
				falseRefuse++
			}
		}
	}
	if triageN > 0 {
		r.TriageTop1Accuracy = float64(triageOK) / float64(triageN)
	}
	if r.Samples > 0 {
		r.GCLFirstPassRate = float64(gclOK) / float64(r.Samples)
	}
	if refuseN > 0 {
		r.FalseRefuseRate = float64(falseRefuse) / float64(refuseN)
	}
	return r
}

func WriteEvalReport(path string, r EvalReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadEvalSamples(path string) (EvalReportInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EvalReportInput{}, err
	}
	var in EvalReportInput
	if err := json.Unmarshal(data, &in); err != nil {
		return EvalReportInput{}, err
	}
	return in, nil
}
```

- [ ] **Step 4: Wire CLI**

在 `cmd/vet/agent.go` 的 `runAgent` switch 增加：

```go
case "eval-report":
    runAgentEvalReport(rest)
```

```go
func runAgentEvalReport(args []string) {
	fs := flag.NewFlagSet("agent eval-report", flag.ExitOnError)
	inPath := fs.String("in", "", "path to EvalReportInput JSON")
	outPath := fs.String("out", "audit-results/eval-report.json", "output report path")
	fs.Parse(args)
	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "agent eval-report: --in is required")
		os.Exit(2)
	}
	in, err := agent.LoadEvalSamples(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent eval-report: load: %v\n", err)
		os.Exit(1)
	}
	rep := agent.BuildEvalReport(in)
	if err := agent.WriteEvalReport(*outPath, rep); err != nil {
		fmt.Fprintf(os.Stderr, "agent eval-report: write: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "eval-report written to %s (samples=%d)\n", *outPath, rep.Samples)
}
```

更新 usage 字符串含 `eval-report`。

- [ ] **Step 5: Test + commit**

Run: `cd cmd/vet && go test ./internal/agent/ -run EvalReport -count=1`

```bash
git add cmd/vet/internal/agent/evalreport.go cmd/vet/internal/agent/evalreport_test.go cmd/vet/agent.go
git commit -m "$(cat <<'EOF'
feat(agent): add online eval-report aggregator and CLI

Wave A4: triage/GCL/false-refuse rates from labeled samples.
EOF
)"
```

---

### Task 4: ProposeFix 加载持久化 KB

**Files:**
- Create: `cmd/vet/internal/strategy/kbpath.go`
- Modify: `cmd/vet/internal/agent/propose.go`
- Create: `cmd/vet/internal/agent/propose_kb_test.go`

**Interfaces:**
- Consumes: `strategy.KnowledgeBase.Load/Save`, `Query`
- Produces: `const DefaultKnowledgeBasePath = ".runtime/strategy/kb.json"`；`func LoadKnowledgeBase(root string) *strategy.KnowledgeBase`

- [ ] **Step 1: Failing test**

```go
func TestProposeFixUsesPersistedPattern(t *testing.T) {
	root := t.TempDir()
	kb := strategy.NewKnowledgeBase()
	kb.Learn(strategy.FailurePattern{
		Pattern: "cpu spike special", Skill: "ve-ecs-ops",
		Solution: "ve ecs DescribeInstances --InstanceIds i-from-kb",
		Confidence: 0.95,
	})
	path := filepath.Join(root, ".runtime", "strategy", "kb.json")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	if err := kb.Save(path); err != nil {
		t.Fatal(err)
	}
	ev := &DiagnosisEvidence{Skill: "ve-ecs-ops"}
	plan := ProposeFixWithRoot(root, ev, &IncidentPayload{Symptom: "cpu spike special", ProductHint: "ecs"})
	if len(plan.Operations) != 1 || !strings.Contains(plan.Operations[0].Command, "i-from-kb") {
		t.Fatalf("expected KB solution, got %+v", plan.Operations)
	}
}
```

注意：现有 `ProposeFix(evidence, payload)` 无 root。新增：

```go
func ProposeFix(evidence *DiagnosisEvidence, payload *IncidentPayload) *DispatchPlan {
	return ProposeFixWithRoot(".", evidence, payload)
}

func ProposeFixWithRoot(root string, evidence *DiagnosisEvidence, payload *IncidentPayload) *DispatchPlan {
	kb := strategy.LoadKnowledgeBase(root) // New + Load default path
	// ... rest same, use kb.Query
}
```

`engine.go` 中 `ProposeFix(...)` 改为 `ProposeFixWithRoot(root, ...)`。

- [ ] **Step 2: Implement kbpath + LoadKnowledgeBase**

```go
// cmd/vet/internal/strategy/kbpath.go
package strategy

import "path/filepath"

const RelKnowledgeBasePath = ".runtime/strategy/kb.json"

func LoadKnowledgeBase(root string) *KnowledgeBase {
	kb := NewKnowledgeBase()
	_ = kb.Load(filepath.Join(root, RelKnowledgeBasePath)) // missing file = empty
	return kb
}
```

检查 `Learn` / `Query` / `FailurePattern` 是否已存在；若 `Learn` 名不同，用现有 API（grep `func (kb *KnowledgeBase)`）。**实现前必须 `codegraph explore KnowledgeBase`**，按真实方法名写入测试（若无 Learn，用直接改 patterns 的测试辅助或已有 Add 方法）。

- [ ] **Step 3: Wire engine**

```go
plan := ProposeFixWithRoot(root, state.Evidence, payload)
```

- [ ] **Step 4: Test + commit**

Run: `cd cmd/vet && go test ./internal/strategy/ ./internal/agent/ -count=1`

```bash
git add cmd/vet/internal/strategy/kbpath.go cmd/vet/internal/agent/propose.go cmd/vet/internal/agent/engine.go cmd/vet/internal/agent/propose_kb_test.go
git commit -m "$(cat <<'EOF'
feat(agent): load strategy KB from disk in ProposeFix

Wave A5: restart-safe pattern reuse via .runtime/strategy/kb.json.
EOF
)"
```

---

### Task 5: 最小 CI Makefile

**Files:**
- Create: `cmd/vet/Makefile`

**Interfaces:**
- Produces: `make test`, `make vet`, `make check-ci`

- [ ] **Step 1: Write Makefile**

```makefile
.PHONY: test vet build check-ci

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o /dev/null .

check-ci: vet test
	go run . check frontmatter --root ../.. || true
	go run . gcl gate --root ../.. --structural-critic-only || true
```

**说明:** `check frontmatter` / `gcl gate` 若仓库预存坏链导致非零，Wave A 先把 **`go test` + `go vet` 作为硬门**；gate 用 `|| true` 仅作 smoke，或查现有 `gcl gate` 是否支持只跑 agent 相关 skill。

更干净的硬门版本：

```makefile
check-ci: vet test build
```

文档注释写明：full `vet check` / `gcl gate` 在 Wave B 收紧为硬失败。

- [ ] **Step 2: Run**

Run: `cd cmd/vet && make check-ci`

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add cmd/vet/Makefile
git commit -m "$(cat <<'EOF'
build(vet): add Makefile check-ci gate for Wave A ALM

EOF
)"
```

---

### Task 6: 全量回归 + ADR/codegraph 收尾

**Files:** none required（验证）

- [ ] **Step 1: Full verify**

```bash
cd cmd/vet && make check-ci && go test ./internal/heal/ ./internal/agent/ ./internal/strategy/ -count=1
codegraph sync --quiet
```

Expected: all PASS

- [ ] **Step 2: Spec coverage checklist（人工）**

| Wave A ID | Task |
|-----------|------|
| A1 真探针 ≥2 plan | Task 1 |
| A2 Confirm 非 Stub AUTO | Task 2 |
| A3 最小 CI | Task 5 |
| A4 在线 eval | Task 3 |
| A5 KB 持久化接线 | Task 4 |

- [ ] **Step 3: Optional ADR Status**

若用户已 Accept ADR-0001/0002：将对应 ADR `Status` 改为 `Accepted`。未 grilling 确认则保持 `Proposed`。

---

## Self-review

1. **Spec coverage:** Roadmap Wave A A1–A5 均有 Task；ADR-0005 排除项未写入任务。Wave B/C 明确另案。
2. **Placeholders:** 无 TBD；Task 4 要求实现前核对 `KnowledgeBase` 真实 API（Learn 名以 codegraph 为准）。
3. **Types:** `Promote` / `EvalReport` / `ProposeFixWithRoot` 在后续 Task 引用一致。
4. **Deadlock:** Task 1 明确 Promote 与 ApplyBuiltInPromotions 锁分层。

## Follow-on plans（不在本文件展开）

- `docs/superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md` — B1–B5 + ADR-0003/0006
- `docs/superpowers/plans/2026-08-01-l400-wave-c-capable.md` — C1–C6 + ADR-0004

---

**Plan complete and saved to `docs/superpowers/plans/2026-08-01-l400-wave-a-foundation.md`. Two execution options:**

**1. Subagent-Driven (recommended)** — 每 Task 新子代理，Task 间 review  

**2. Inline Execution** — 本会话按 executing-plans 连续做完  

**Which approach?**
