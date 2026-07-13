# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

> **Content hierarchy**: This file is the entry point. Detailed specifications live in `docs/`:
> - `docs/gcl-spec.md` — full GCL specification (purpose, roles, loop flow, trace, prompt templates, changelog)
> - `docs/token-efficiency.md` — full TE rules with code examples (TE-1 through TE-9)
> - `docs/skill-harness-review-checklist.md` — reusable P0-P3 review checklist for all ve-*-ops skills
> - `docs/self-review-checklist.md` — Round 1 (C1–C17) + Round 2 (F1–F13) checklists + validation matrix
> - `docs/execution-strategy.md` — E1–E4 decision rules, runtime adaptation, retro reflection
> - `docs/l2-to-l3-plan.md` — detailed L2→L3 (conditional autonomy) execution plan (M1 expansion)
> - `docs/codegraph-integration.md` — CodeGraph two-tier sync rules
> - `docs/inline-script-pattern.md` — `_inline_script()` implementation constraints for validation scripts
>
> Keep this file in sync with `docs/` files. When updating either, verify the other stays aligned.

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
- **禁止用 Python / Bash / shell 编写独立工具。** 遗留的 `scripts/*.py` 仅作参考实现，MUST NOT 再扩展新功能；新增能力 → 在 `cmd/` 下新增 Go 代码（或扩展既有 Go 模块）。
- **Markdown skill 规格不是"工具"**，不受本规则约束——它们仍是 prose。本规则只管可执行工具、CLI、校验器与自动化。
- 任何新增/修改工具的 PR/变更 MUST 证明 `go build` + `go vet` 干净；CI/发布必须能产出二进制。
- 例外需每笔变更显式获得用户批准；无批准 = MUST 为 Go。

---

## Repository Type — Read This First

This repo contains **only Markdown skill specifications** plus **Go tooling** (`cmd/vet/` and any future `cmd/*` Go modules). There is:

- **Go tooling has a real build system** — each tool is its own Go module (`go.mod`); `go build`, `go test`, `go vet`, `make` apply *to the Go modules only*. For the Markdown skill specs themselves, do not run build/test commands — they are prose, not code.
- **No non-Go tools.** New operational tooling MUST be Go (see MANDATORY rule above). The legacy Python scripts in `scripts/` are deprecated reference only.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md` (see also **Files that DO NOT exist** below).

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

---

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
- **Orchestration / loop-agent skills are exempt from the `ve-*` prefix and `cli_applicability` rule.** Skills whose `metadata.type ∈ {orchestration-skill, loop-agent, meta-skill}` (e.g. `incident-loop-agent`) need not follow the `ve-<product>-ops` naming or carry a mandatory `cli_applicability` field. They MUST still live under their own `<name>/` directory with a `SKILL.md` + `references/`, and MUST be added to the GCL conformance allowlist (`ORCHESTRATION_TYPES`, defined in `cmd/vet/internal/check/frontmatter/frontmatter.go`). The single-product rule does not apply to them (they coordinate, not own, a product). The legacy `scripts/validate_skills_frontmatter.py` mirrored this allowlist.

## File Layout Anchors (do not relocate without reason)

```
ve-skill-generator/                  meta-skill: how to author new skills
  SKILL.md                           generator workflow + P0/P1 checklist
  references/
    ve-skill-template.md             canonical skill template
    cli-behavior.md                  verified `ve` CLI conventions
    execution-environment.md         CLI + Go SDK setup
    user-experience-spec.md          UX requirements every generated skill must follow
    governance-and-adversarial-review.md
ve-[product]-ops/                    one per Volcengine product
  SKILL.md                           main runbook
  references/                        6 standard reference files
  assets/example-config.yaml
docs/
  gcl-spec.md                        full GCL specification
  token-efficiency.md                detailed TE rules with code examples
  document-integrity.md              link validation & dedup scripts
  failure-patterns.md                Reflexion memory store
  inline-script-pattern.md           _inline_script() constraints
  reflexion-memory.md                cross-session failure-pattern memory
  skill-harness-review-checklist.md  P0-P3 review checklist
  skill-routing-graph.md             cross-skill alarm routing
```

`.omc/` and `.omo/` are tool-local state and are gitignored — do not commit anything inside them.

### Asset & Schema Placement Rules (mandatory)

Skill-owned artifacts MUST NOT be placed at repo root. Use this split:

| Location | Allowed contents | Forbidden |
|---|---|---|
| `ve-*-ops/assets/` | `eval_queries.json`, `example-config.yaml`, `*.schema.json`, skill-specific templates | Cross-skill executables |
| `ve-*-ops/references/` | Runbooks, output contracts in Markdown, delegation stubs | Duplicate JSON schemas that belong in `assets/` |
| `scripts/` (repo root) | **Deprecated** Python validation scripts (`validate_local.py`, `check_gcl_conformance.py`, etc.) — superseded by `cmd/vet/`; kept only as reference | JSON Schema, handoff contracts, example YAML |

**Owner skill rule:** the skill that **defines and primarily consumes** the contract owns the file. Secondary consumers link to the owner via relative path — they do not copy or re-home the schema.

**When adding a new `*.schema.json` or handoff contract:**
1. Pick the owner skill (primary consumer of the JSON contract).
2. Create under `ve-<owner>-ops/assets/<name>.schema.json`.
3. Reference from owner `SKILL.md` / `references/` and owner `example-config.yaml` if config-driven.
4. Secondary skills cite the owner path (e.g. `../ve-<owner>-ops/assets/...`) — never `assets/` at repo root.
5. If a future `scripts/*.py` (or `cmd/vet/` check) emits JSON matching the schema, its docstring MUST point at the owner skill path (script ≠ schema owner).

**Anti-pattern (banned):** creating `assets/` at repo root because a script is shared — shared **code** lives in `scripts/`; shared **contracts** still belong to an owning skill.

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

The `vet` CLI (from `cmd/vet/`) is the primary validation tool for pre-commit checks, replacing the legacy Python scripts in `scripts/`. After changing any skill, run `vet validate --root .` (or `vet check frontmatter|links|gcl|eval|aiops|assessment`), then manually verify against the P0/P1 checklist in `ve-skill-generator/SKILL.md` and the 2-round self-review above. The Python scripts remain as the original reference implementation.

### Runtime GCL

`vet gcl run` (Go port of the deprecated `scripts/gcl_runner.py`) implements the GCL Orchestrator (Phase 2). Use it for automated GCL loops:
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

- **`scripts/` directory** at repo root contains **deprecated** Python validation scripts (`validate_local.py`, `check_gcl_conformance.py`, `check_markdown_links.py`, etc.) — superseded by `cmd/vet/`. They are kept only as reference implementations.
- **No `package.json`, `CI configs`, `build scripts`, `typechecker`, or non-stdlib test runner** at repo root *for the Markdown specs*.
- **Go tool modules (`cmd/*`) legitimately own their own `Makefile`, `go.mod`, `go.sum`, CI, and release targets** — these are required by the MANDATORY "All Tools MUST Be Written in Go" rule above, not banned.
- **No `CLAUDE.md` at repo root** — this file (`AGENTS.md`) is the agent guidance entry point, imported via `@CLAUDE.md`.
- **No `opencode.json`, `.cursorrules`, or similar IDE configs.**
- **No `audit-results/` directory** — GCL traces are stored in tool-local state only.
- `.omc/`, `.omo/` are gitignored cache data — not source.

---

## Runtime Quality Gates: GCL & Reflexion

Detailed specs in [docs/gcl-spec.md](docs/gcl-spec.md) and [docs/reflexion-memory.md](docs/reflexion-memory.md).

| Spec | Read before modifying |
|---|---|
| `docs/gcl-spec.md` | any `## Quality Gate (GCL)` section, `references/rubric.md`, `references/prompt-templates.md`, or GCL-related runner code |
| `docs/reflexion-memory.md` | `docs/failure-patterns.md`, trace `failure_pattern` extraction, Reflexion retrieval/persistence logic, or failure-memory governance |

### Hard Constraints Summary

**GCL:** Isolated G+C contexts · Critic read-only · Safety=0 must ABORT · max_iterations bounded · `{{env.*}}`/`{{user.*}}`/`{{output.*}}` placeholders · Trace persisted with masked credentials.

**Reflexion:** Optional hint · `docs/failure-patterns.md` ≤ 200 lines · Dedup by skill+command+error · Patterns from GCL trace only.

### Relationship to build-time self-review

Build-time 2-round self-review and runtime GCL are independent gates. A clean self-review does not exempt runtime scoring; a passing GCL rubric does not exempt sloppy skill updates.

---

## CodeGraph Integration — 代码变动即时同步

CodeGraph (`codegraph` CLI) 维护仓库知识图谱，使 AI 能检索符号、调用链与影响面；并为 Python→Go 翻译提供语义基底（翻译前用 `codegraph callees` / `codegraph impact <symbol>` 查依赖，避免漏迁被调方）。

完整的两层触发与同步规则见 **[docs/codegraph-integration.md](docs/codegraph-integration.md)**。

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