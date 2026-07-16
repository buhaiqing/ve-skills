# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

> **Content hierarchy**: This file is the entry point. See [docs/README.md](docs/README.md) for full documentation index organized by role (使用者 / 开发者 / 审查者).

---

## Code Quality Rule (applies to `cmd/vet/**` Go code)

All Go code under `cmd/vet/` MUST uphold:

- **Reusability** — extract shared logic into `internal/` packages; do not duplicate helpers across commands.
- **Simplicity** — keep it minimal; no premature abstraction, no dead code, no over-engineering.
- **Readability** — clear names that state intent; comments explain WHY not WHAT.
- **Testability** — pure, deterministic functions; cover happy path + error path with `_test.go`.
- **Security** — never log or commit real credentials; mask secrets in all trace/GCL output; keep the credential-masking path covered by tests.

---

## MANDATORY: All Tools MUST Be Written in Go (compile + publishable)

> **严格规则，无例外。** 本仓库内**任何可执行工具/工具型 CLI/校验器/自动化脚本都必须用 Go 实现**，且在任何变更"完成"前 MUST 满足以下两道门禁：

| 门禁 | 要求 | 验证命令 |
|------|------|----------|
| **Compile** | `go build ./...`（或该工具的模块）零错误通过 | 在工具所属 `go.mod` 模块内执行 `go build ./...` |
| **Publishable** | 工具具备发布路径（如 `Makefile` target / 带版本号的构建），且 `go vet ./...` 干净 | `go vet ./...`、`make build` / 发布 target |

执行细则：
- **禁止用 Python / Bash / shell 编写独立工具。** 所有可执行工具必须是 Go 代码（见 MANDATORY 规则）；新增能力 → 在 `cmd/` 下新增 Go 代码（或扩展既有 Go 模块）。
- **Markdown skill 规格不是"工具"**，不受本规则约束——它们仍是 prose。本规则只管可执行工具、CLI、校验器与自动化。
- 任何新增/修改工具的 PR/变更 MUST 证明 `go build` + `go vet` 干净；CI/发布必须能产出二进制。
- 例外需每笔变更显式获得用户批准；无批准 = MUST 为 Go。

---

## Repository Type — Read This First

This repo contains **only Markdown skill specifications** plus **Go tooling** (`cmd/vet/` and any future `cmd/*` Go modules). There is:

- **Go tooling has a real build system** — each tool is its own Go module (`go.mod`); `go build`, `go test`, `go vet`, `make` apply *to the Go modules only*. For the Markdown skill specs themselves, do not run build/test commands — they are prose, not code.
- **No non-Go tools.** All operational tooling MUST be Go (see MANDATORY rule above).
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md` (see also **Files that DO NOT exist** below).

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

---

## Session-Observed Environment Notes (2026-07-13)

> 沉淀自 L2→L3 任务（T01–T04）执行复盘，适用于本仓库所有 Agent 任务。

- **Background 子代理派发在本环境不可用**：`task(subagent_type="executor"/"build", run_in_background=true)` 会静默超时（"Task timed out after 30 minutes of inactivity" / "Task delegation failed"），代理实际从未运行；前台 `task()` 也须用 `category` 参数，不能直接传 `subagent_type="Sisyphus-Junior"`（报 "Cannot use subagent_type directly"）。→ **多文件机械性改动直接在主代理内完成**，勿派发子代理。
- **机械性批量改写**（如给 N 个 SKILL.md 操作表统一加列）：用一次性 `/tmp/*.py` 脚本（不入库、非仓库工具，符合"仓库内不新增 Python 工具"约束）最可靠；改完用 grep 验证每行生效。
- `vet validate --root .` 的链接检查有**预存坏链**（`enhanced-self-healing-framework.md` / `AGENTS.md` / `ve-ecs-ops/SKILL.md` / `reflexion-memory.md` 的 `./` 链接，早于本任务、非本任务引入），与卡片范围无关；卡片 DoD 门禁 `vet check frontmatter/aiops/assessment` 保持绿即可，勿被预存链接失败阻塞。

## MANDATORY: Two-Round Self-Review After Every Skill Update

After ANY edit to a `ve-*-ops/SKILL.md`, its `references/*.md`, `assets/*`, or to `ve-skill-generator/**`, you MUST run **two rounds of self-review** before declaring done. Fix every issue surfaced — do not just list them.

### Round 1 + Round 2 checklists

Full Round 1 (C1–C17) structural/spec-compliance table and Round 2 (F1–F13) accuracy/anti-pattern table, plus the Validation Command Matrix, live in **[docs/self-review-checklist.md](docs/self-review-checklist.md)**.

**Hard rule**: Fix every issue found in both rounds before responding "done". If a finding needs info you lack, mark it `[blocked: needs OpenAPI verification]`.

**Build-time validation**: after changes run `vet validate --root .` (or `vet check frontmatter|links|gcl|eval|aiops|assessment`), then walk the P0/P1 checklist in `ve-skill-generator/SKILL.md`.

---

## Execution Strategy — 任务执行策略

> 多步骤任务执行时的决策规则，沉淀自多次迭代经验。完整 check items 见 [docs/token-efficiency.md](docs/token-efficiency.md)。

### E1–E4 决策规则（摘要）

| 规则 | 一句话 |
|------|--------|
| E1 智能并行分解 | ≥2 步任务先判定并行/串行；`run_in_background=true` 默认 |
| E2 运行时自适应 | 工具失败先确认路径；外部不可达降级本地；≥2 轮失败停止 |
| E3 定期反思 | 每阶段/15 次调用后快速反思（离目标/更简单/可复用） |
| E4 资产沉淀 | 通用知识写入 `docs/failure-patterns.md` 等，去重、≤200 字、符号化 |

完整决策矩阵、E2–E4 明细与复盘检查表见 **[docs/execution-strategy.md](docs/execution-strategy.md)**。

---

## Skill Harness Review — 评审方法论

> 评审 ve-*-ops skill 时的两层框架，沉淀自 2026-07-05 全量 skill 评审实践经验。
> 完整检查表见 [docs/skill-harness-review-checklist.md](docs/skill-harness-review-checklist.md)。

### 评审框架摘要

每次 skill 评审覆盖两个维度：**Harness Runtime**（路由/执行/跟踪）与 **Quality Compliance**（模板/TE/GCL）。完整检查表（R/F/TE/I/G/E 全量条目 + D0–D4、评审步骤、背景搜索指引）见 **[docs/skill-harness-review-checklist.md](docs/skill-harness-review-checklist.md)**。

- **P0** blocking（缺 Trigger&Scope / GCL / null bytes / 凭证泄露）→ 必须立即修复
- **P1** 破坏性验证缺失（delete/stop 缺确认门、错误码 <10）→ 必须修复
- **P2** 质量欠佳不阻塞（硬编码版本号、跨文件重复、缺 What This Skill Does）→ 建议修复
- **P3** 优化项（TE-8 符号、压缩级别）→ 有空再修

### TE 规则（摘要）

完整 TE-1~TE-9 规则、Symbol System、Abbreviations、Compression Levels、TE-1 扫描边界见 **[docs/token-efficiency.md](docs/token-efficiency.md)**。

| TE | 要点 |
|----|------|
| TE-1 | API 查询替代硬编码版本/配额 |
| TE-2 | Go SDK 用 `#` 注释而非函数级 docstring |
| TE-3 | 错误表每行 1 码、≤3 列 |
| TE-4 | JSON path 集中声明于文件顶部 |
| TE-5 | example-config.yaml 用 YAML anchors |
| TE-6 | SKILL.md 与 references 不重复 |
| TE-7 | AIOps/FinOps 入 `references/advanced/`；SQL 标 Security-Sensitive |
| TE-8 | 用 `→` `⇒` `✅` `❌` 等符号替代冗余文字 |
| TE-9 | 按用途选压缩级别（Minimal~Emergency） |

> 不可压缩：可执行命令、错误恢复、安全门、Credential 规则、跨技能编排链。发现任一违规 → 立即修复 → 重检至全过。

---

## Skill Authoring Guardrails

- **Never create a new `ve-*-ops/` directory by hand.** Use the `ve-skill-generator` meta-skill flow.
- **Never invent CLI flags or API parameters.** Only official OpenAPI or `ve <service> <action> --help` verified fields.
- **Single-product rule.** One `ve-*-ops` skill = one product = one primary resource model.
- **Prefer editing existing skills over creating new files.** No tutorial-style `.md` at the repo root.
- **Orchestration / loop-agent skills are exempt from the `ve-*` prefix and `cli_applicability` rule.** Skills whose `metadata.type ∈ {orchestration-skill, loop-agent, meta-skill}` (e.g. `incident-loop-agent`) need not follow the `ve-<product>-ops` naming or carry a mandatory `cli_applicability` field. They MUST still live under their own `<name>/` directory with a `SKILL.md` + `references/`, and MUST be added to the GCL conformance allowlist (`ORCHESTRATION_TYPES`, defined in `cmd/vet/internal/check/frontmatter/frontmatter.go`). The single-product rule does not apply to them (they coordinate, not own, a product). The `ORCHESTRATION_TYPES` allowlist is defined in `cmd/vet/internal/check/frontmatter/frontmatter.go`.

## File Layout Anchors (do not relocate without reason)

完整目录树见 [docs/README.md](docs/README.md)。核心结构：

```
ve-skill-generator/    meta-skill: how to author new skills
ve-[product]-ops/       one per Volcengine product (SKILL.md + references/ + assets/)
docs/                  详细索引见 [docs/README.md](docs/README.md)
incident-loop-agent/    orchestration skill (exempt from ve-* naming rule)
```

`.omc/` and `.omo/` are tool-local state and are gitignored.

### Asset & Schema Placement Rules (mandatory)

| Location | Allowed | Forbidden |
|---|---|---|
| `ve-*-ops/assets/` | `eval_queries.json`, `example-config.yaml`, `*.schema.json` | Cross-skill executables |
| `ve-*-ops/references/` | Runbooks, output contracts in Markdown | JSON schemas (→ `assets/`) |
| `scripts/` (repo root) | **已删除** — 已迁移到 `cmd/vet/` | ~~Deprecated~~ |

**Owner skill rule**: the skill that defines and consumes a contract owns it. Create `ve-<owner>-ops/assets/<name>.schema.json`; secondary skills link via relative path. Anti-pattern: `assets/` at repo root — contracts belong to owning skill, not shared at root.

---

## Generator-Critic-Loop (GCL) — Adversarial Quality Gate

> Full specification: [docs/gcl-spec.md](docs/gcl-spec.md) — purpose, roles, rubric, loop flow,
> termination, trace schema, prompt templates, and changelog.

### Summary

| Role | Job | Forbidden |
|------|-----|-----------|
| **Generator (G)** | Execute the cloud operation | modifying the rubric; self-scoring |
| **Critic (C)** | Independently audit G's output | calling `ve` / SDK / mutating anything |
| **Orchestrator (O)** | Loop control, termination | executing or scoring on its own |

**Hard constraint:** G and C MUST live in **isolated prompt contexts**. Shared context is banned.

### Rubric (5 dimensions) & Termination

Full rubric (Correctness/Safety/Idempotency/Traceability/Spec Compliance) + thresholds and PASS/MAX_ITER/SAFETY_FAIL termination semantics: **[docs/gcl-spec.md](docs/gcl-spec.md)** §Rubric + §Termination. SAFETY_FAIL always ABORTs; Safety must equal 1.

### Per-Skill Defaults

Full tier/default mapping: [docs/gcl-spec.md §8](docs/gcl-spec.md#8-per-skill-defaults).
Each skill may override `max_iter` in its own `SKILL.md` (`## Quality Gate (GCL)`).

### Anti-Patterns (banned)

> Full list: [docs/gcl-spec.md](docs/gcl-spec.md) §9.
> **Key**: Shared G+C context banned · Safety=0 must ABORT · Trace must persist · Credential leakage in trace · No unbounded loops.

### Cross-Skill Delegation

When GCL identifies cross-product gaps, the Orchestrator MUST delegate, not absorb.
Full alarm-pattern → skill routing rules in [docs/skill-routing-graph.md](docs/skill-routing-graph.md).
The Critic itself MUST NOT call any skill — it only emits suggestions.

### GCL Rollout Complete — All 29 Skills Equipped

All 13 `required`-tier, 10 `recommended`-tier, and 6 `optional`-tier skills have GCL rubric + prompt templates.
See [docs/gcl-spec.md](docs/gcl-spec.md) §11 for the full rollout changelog.

> Additionally, 1 **orchestration skill** `incident-loop-agent` (`metadata.type=orchestration-skill`) is equipped with GCL rubric + prompt templates. It is exempt from the `ve-*` prefix and `cli_applicability` mandate per the Skill Authoring Guardrails exception clause.

## Evaluation

`assets/eval_queries.json` per skill holds intent-classification test cases (`should_trigger: true/false`). These are consumed by external evaluation harnesses, not by an in-repo test runner. When adding capability, add eval cases in the same change.

### Build-time regression

The `vet` CLI (from `cmd/vet/`) is the primary validation tool for pre-commit checks. After changing any skill, run `vet validate --root .` (or `vet check frontmatter|links|gcl|eval|aiops|assessment`), then manually verify against the P0/P1 checklist in `ve-skill-generator/SKILL.md` and the 2-round self-review above.

### Runtime GCL

`vet gcl run` implements the GCL Orchestrator (Phase 2). Use it for automated GCL loops:
- External Critic via `--critic-json` or stdin (production mode)
- `--structural-critic-only` for CI/local structural smoke tests only (NOT for production mutations)
- `vet gcl gate` runs the structural CI gate across all skills; `vet gcl trace` aggregates `audit-results/gcl-trace-*.json` into a quality summary.

GCL runs MUST use externally supplied isolated Critic scores in production. Manual execution following `docs/gcl-spec.md` is also acceptable.

---

## Document Integrity & Link Validation (C12 / F9 / F10 — 强制)

> 每次文档变更后，自动对所有受影响的文件执行 3 层完整性检查。
> 完整检查脚本及模板见 [docs/document-integrity.md](docs/document-integrity.md)。

| Layer | 检查项 | 方法 |
|-------|--------|------|
| 1 | 链接完整性（F9） | 扫描所有 `[...](...)` → `test -f` 验证路径存在 |
| 2 | 交叉引用对称性（F9/TE-7） | AGENTS.md → docs/ 的反向引用存在性 |
| 3 | 内容去重（C12/TE-6） | 跨文件代码块/表格完全重复检测 |

**发现任何问题 → 修复并重新验证 → 确认全部通过后方可继续。**

---

## Files that DO NOT exist

<!-- `scripts/` 已删除 — 已迁移到 cmd/vet/ Go 工具 -->
- **No `package.json`, `CI configs`, `build scripts`, `typechecker`, or non-stdlib test runner** at repo root *for the Markdown specs*.
- **Go tool modules (`cmd/*`) legitimately own their own `Makefile`, `go.mod`, `go.sum`, CI, and release targets** — these are required by the MANDATORY "All Tools MUST Be Written in Go" rule above, not banned.
- **No `CLAUDE.md` at repo root** — this file (`AGENTS.md`) is the agent guidance entry point, imported via `@CLAUDE.md`.
- **No `opencode.json`, `.cursorrules`, or similar IDE configs.**
- **No `audit-results/` directory** — GCL traces are stored in tool-local state only.
- `.omc/`, `.omo/` are gitignored cache data — not source.

---

## Runtime Quality Gates: GCL & Reflexion

详细规范见 [docs/gcl-spec.md](docs/gcl-spec.md)（GCL）和 [docs/reflexion-memory.md](docs/reflexion-memory.md)（Reflexion）。

**GCL 硬约束**：G+C 上下文隔离 · Critic 只读 · Safety=0 必须 ABORT · 有界 max_iterations · `{{env.*}}`/`{{user.*}}`/`{{output.*}}` 占位符 · trace 持久化并脱敏凭证。

**Build-time 自检与 runtime GCL 互相独立**：自检通过 ≠ GCL 通过；GCL 通过 ≠ 自检可以敷衍。

---

## CodeGraph Integration — 代码变动即时同步

CodeGraph (`codegraph` CLI) 维护仓库知识图谱，使 AI 能检索符号、调用链与影响面；并为 Python→Go 翻译提供语义基底（翻译前用 `codegraph callees` / `codegraph impact <symbol>` 查依赖，避免漏迁被调方）。

完整的两层触发与同步规则见 **[docs/codegraph-integration.md](docs/codegraph-integration.md)**。

### MANDATORY: CodeGraph MCP 优先于 grep/read

> **强规则：所有代码理解任务（符号查找、调用链追溯、影响面分析）必须先用 CodeGraph MCP (`codegraph_explore`)，再用 grep/read 补充。**

**为什么**：CodeGraph 的 AST + 调用图覆盖了 grep 无法到达的跳转表（接口实现、动态派送、跨包调用）。纯文本搜索会漏掉真实调用者。

**强制执行顺序**：
1. `codegraph_explore <symbol>` → 获取符号定义 + 调用者 + 影响面
2. 仅在 CodeGraph 索引缺失或不确定时用 grep/read 交叉验证
3. 修改代码前必须用 CodeGraph 确认所有调用方都已知

**禁用**：用 grep 替代 CodeGraph 做调用链分析；用 read 替代 codegraph_explore 做符号定位（CodeGraph 更快更准）。

**例外**：纯文本内容搜索（如日志关键字、文档内容）除外。

---

## Communication (Language)

- **默认使用中文回复**。除非用户明确要求使用英文，否则所有回复、总结、报告均使用中文。
- 代码注释、CLI 命令和代码块内的内容保持原文（英文），不做翻译。

---

## Key References

| Document | Description |
|----------|-------------|
| `docs/gcl-spec.md` | **Runtime GCL spec** — purpose, roles, loop flow, trace, prompt templates, per-skill defaults, changelog |
| `docs/token-efficiency.md` | Detailed TE rules with code examples (TE-1 through TE-9) |
| `docs/reflexion-memory.md` | **Reflexion rules** — lightweight cross-session failure-pattern memory governance |
| `docs/failure-patterns.md` | **Reflexion memory store** — bounded structured failure patterns for cross-session learning |
| `docs/skill-harness-review-checklist.md` | **Skill harness review checklist** — reusable P0-P3 review template for all ve-*-ops skills |
| `docs/self-review-checklist.md` | **Two-round self-review checklist** — Round 1 (C1–C17) + Round 2 (F1–F13) + validation matrix |
| `docs/execution-strategy.md` | **Execution strategy** — E1–E4 parallel/adaptive/reflection rules + retro checklist |
| `docs/l2-to-l3-plan.md` | **L2→L3 plan** — conditional-autonomy execution plan (M1): execution-risk policy + P1–P7 tasks + L3 DoD |
| `docs/l2-to-l3-tasks/AGENTS.md` | **L2→L3 task AGENTS.md** — AI entry rules (Karpathy + TDD + GCL + 文档更新 + 12 节工程化规则) |
| `docs/l3-to-l4-tasks/AGENTS.md` | **L3→L4 task AGENTS.md** — AI entry rules (Karpathy + L4 envelope / SLO / 自愈 / Reflexion 4 级 + 13 节工程化规则) |
| `docs/codegraph-integration.md` | **CodeGraph integration** — two-tier sync rules for knowledge-graph maintenance |
| `docs/inline-script-pattern.md` | **Inline script pattern** — `_inline_script()` implementation constraints for validation scripts |
| `docs/superpowers/plans/golang-migration/2026-07-12-python-to-go-cli.md` | **Python→Go (`vet`) + CodeGraph 计划** — 里程碑、子任务拆分、并行策略 |
| `ve-skill-generator/SKILL.md` | Meta-skill generator — full workflow, P0/P1 checklist, Token Efficiency rules |
| `ve-skill-generator/references/ve-skill-template.md` | Canonical SKILL.md template with GCL block |
| `ve-skill-generator/references/governance-and-adversarial-review.md` | Governance & adversarial review — R1-R4 pre-merge security/resilience/UX scenarios |
| `ve-skill-generator/references/cli-behavior.md` | Verified `ve` CLI conventions |
| `ve-skill-generator/references/execution-environment.md` | CLI + Go SDK setup details |
| `ve-skill-generator/references/user-experience-spec.md` | UX requirements every generated skill must follow |
| Each skill's `references/rubric.md` | The rubric instance |
| Each skill's `references/prompt-templates.md` | G/C/O prompt skeletons |