# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

> **Content hierarchy**: This file is the entry point. Detailed specifications live in `docs/`:
> - `docs/gcl-spec.md` — full GCL specification (purpose, roles, loop flow, trace, prompt templates, changelog)
> - `docs/token-efficiency.md` — full TE rules with code examples (TE-1 through TE-9)
> - `docs/skill-harness-review-checklist.md` — reusable P0-P3 review checklist for all ve-*-ops skills
> - `docs/inline-script-pattern.md` — `_inline_script()` implementation constraints for validation scripts
>
> Keep this file in sync with `docs/` files. When updating either, verify the other stays aligned.

---

## Repository Type — Read This First

This repo contains **only Markdown skill specifications**. There is:

- **No build system, no tests, no lint, no package manager.** Do not run `go build`, `go test`, `npm`, `pip`, `make`, etc. — they will fail or do nothing.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md` (see also **Files that DO NOT exist** below).

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

---

## MANDATORY: Two-Round Self-Review After Every Skill Update

After ANY edit to a `ve-*-ops/SKILL.md`, its `references/*.md`, `assets/*`, or to `ve-skill-generator/**`, you MUST run **two rounds of self-review** before declaring done. Fix every issue surfaced — do not just list them.

### Round 1 — Structural & Specification Compliance

Verify against `ve-skill-generator/SKILL.md` and `ve-skill-generator/references/ve-skill-template.md`:

| # | Check | Detail |
|---|-------|--------|
| C1 | Frontmatter | name, description, license, compatibility, metadata complete |
| C2 | cli_applicability | set; if `dual-path`, both CLI and SDK steps exist |
| C3 | Five Core Standards | boundaries, I/O, steps, failures, single responsibility present & accurate |
| C4 | Placeholder channels | `{{env.*}}` for secrets/region, `{{user.*}}` for interactive, `{{output.*}}` for captured |
| C5 | Credential safety | No credential value logged/printed/echoed — only `<masked>` or existence checks |
| C6 | **Token Efficiency** | **See §Token Efficiency below; verify TE-1~TE-7** |
| C7 | Safety gates | Every destructive operation (delete, stop, release, restore) has explicit confirmation |
| C8 | Error taxonomy | ≥ 10 product-specific codes with HALT vs retry classification |
| C9 | Reference files | All 6 standard files exist when applicable |
| C10 | Cross-product delegation | ECS skill delegates IAM work to IAM skill, etc. |
| C11 | Version bump | `last_updated` and `version` bumped if behavior changed |
| C12 | **Content dedup** | No duplicate operation flows across SKILL.md ↔ references/; TE-6 verified (see §Link & Dedup) |
| C13 | **AIOps coverage** | `references/advanced/aiops.md` and `references/advanced/finops.md` exist for all required+recommended skills (23 total) |
| **C15** | **`### What This Skill Does` (IMPORTANT)** | **MUST** exist with a clear 2-3 sentence description of the skill's purpose and boundary |
| **C16** | **`## Operational Best Practices` (IMPORTANT)** | **MUST** exist with actionable operational guidance (monitoring, backup, security patterns) |
| **C17** | **SecurityOps coverage** | `references/advanced/securityops.md` exists for security-critical skills (security-group-ops, iam-ops, kms-ops, ecs-ops, rds-mysql-ops, redis-ops, mongodb-ops, elasticsearch-ops); recommended for all others |

### Round 2 — Accuracy & Anti-Pattern Sweep

Verify against reality, not memory:

| # | Check | Detail |
|---|-------|--------|
| F1 | API fidelity | Every field/flag grounded in official OpenAPI or `ve <service> --help` |
| F2 | Command prefix | `ve` (not `volcengine-cli`, renamed at v1.0.20) |
| F3 | CLI shape | `ve <service> <Action> --<Param> value` (PascalCase actions and params) |
| F4 | JSON bodies | `--body '{"Key":"Value"}'` form, not invented flags |
| F5 | Originality | No generic prose, speculative claims, or copy-pasted irrelevant boilerplate |
| F6 | Links | Sibling references use correct relative paths |
| F7 | Examples | Minimal, runnable, not pseudocode |
| F8 | Rendering | Tables aligned, code fences closed, headings hierarchical |
| **F9** | **Link integrity** | Cross-references between docs/ and AGENTS.md verified; all internal `[...](...)` resolve to existing files (see §Link & Dedup) |
| **F10** | **Dedup integrity** | No repeated content across multiple files; after every change run dedup check (see §Link & Dedup) |
| **F11** | **Eval data coverage** | `assets/eval_queries.json` exists for all 29 skills with ≥ 5 trigger + ≥ 2 non-trigger cases |
| **F12** | **FinOps coverage** | `references/advanced/finops.md` exists for required+recommended skills |
| **F13** | **SecurityOps coverage** | `references/advanced/securityops.md` exists for security-critical skills (8); recommended for all others |

**Fix every issue found in both rounds before responding "done".** If a finding requires information you don't have, mark it `[blocked: needs OpenAPI verification]`.

### Validation Command Matrix

| Change scope | Required command |
|---|---|
| Full local validation before PR / handoff | Walk P0/P1 checklist in `ve-skill-generator/SKILL.md` (no automated runner in this repo) |
| Any `SKILL.md` frontmatter or metadata change | Re-check P0/P1 checklist items C1–C12 manually |
| Any GCL rubric, prompt template, or `## Quality Gate (GCL)` section change | Verify against `docs/gcl-spec.md` §Rubric + §Prompt Templates |
| Any Markdown spec, README, or path-reference change | Run the link-integrity scan in **§Document Integrity & Link Validation** below |
| Any `docs/*.md` change | Verify cross-reference symmetry (see Layer 2 below) |
| Any skill harness review | Walk [docs/skill-harness-review-checklist.md](docs/skill-harness-review-checklist.md) D0-D4 |
| Any Go SDK example code change | `python3 -m py_compile <file>` if Python; Go examples are illustrative — no build step |

> **Note:** This repo has automated validation scripts in `scripts/` for pre-commit checks (see **Files that DO NOT exist** below). After changes, run `python3 scripts/validate_local.py` if available, then supplement with manual checks against the P0/P1 checklist.

---

## Execution Strategy — 任务执行策略

> 多步骤任务执行时的决策规则，沉淀自多次迭代经验。完整 check items 见 [docs/token-efficiency.md](docs/token-efficiency.md)。

### E1 — 智能并行分解

每个 ≥2 步骤的多目标任务，**必须**在 Phase 0 做分解判断：

| 条件 | 策略 | 示例 |
|------|------|------|
| 2+ 独立模块，无共享状态 | **并行委托** → 每个目标一个 `deep` 或 `unspecified-high` 子代理 | 同时修复 5 个 skill 的 TE 违规 → 5 个并行子代理 |
| 步骤间有依赖链 | **串行执行**，但每个阶段内最大并行化 | TE-4 修复需先加声明再验证 → 声明阶段串行，验证并行 |
| 目标模糊需探索 | **先探索再执行** → `explore`/`librarian` 背景搜索 | "有哪些 TE-6 违规" → 并行扫描 9 个 skill，汇总后再修复 |
| 全量扫描 + 局部修复 | 扫描用 `explore` 并行，修复用 `deep` 分批并行 | TE-6 扫描（3 个 explore）+ TE-6 修复（每 skill 一个 deep） |

**规则**：`run_in_background=true` 是默认偏好，除非任务明确阻塞后续步骤。

### E2 — 运行时自适应

遇到运行环境问题时，按以下顺序调整：

| 问题 | 调整措施 |
|------|---------|
| 子代理超时 / 输出截断 | → 拆分任务到更细粒度，每代理 ≤ 3 个文件或 ≤ 1 个技能 |
| 工具调用失败（file not found 等） | → 先用 `read`/`grep`/`glob` 确认路径再操作，不假设 |
| context 窗口过大 | → 用 `planning-with-files` 或 `note` 将中间结果写入文件，避免全量保留 |
| 外部服务不可达（MCP / web） | → 降级为本地工具 + 静态 fallback 模式，不阻塞流程 |
| 反复修复失败（≥2 轮） | → 停止自动修复 → 记录失败模式 → 通知用户裁决 |

**关键**：不做无谓的重试。每次失败记录根因，不重复同样的错误路径。

### E3 — 定期反思与复盘

完成任务阶段（或每 15 次工具调用）后，执行快速反思：

```
1. ❓ 当前步骤离目标还差多远？（已完成 / 阻塞 / 偏离）
2. 🔍 有没有更简单的方法达到目标？（TE-4 手动 vs 批量）
3. 💡 有没有发现可复用的模式或知识？
```

发现可复用内容 → 立即沉淀（见 E4）。

### E4 — 可复用资产沉淀

运行时发现的通用知识，按以下规则存储：

| 资产类型 | 存放位置 | 示例 |
|---------|---------|------|
| 修复经验（跨 skill 通用） | `docs/failure-patterns.md` | "TE-4 扫描误标 — 需人工核查路径覆盖率" |
| 执行流程改进 | `AGENTS.md` 本文件 | "扫描类任务先用 explore 并行，再基于结果分批修复" |
| 规则解读 / 边界判定 | 对应 `docs/*.md` | "TE 规则边界判定 → 见 docs/token-efficiency.md" |
| 临时工作记录 | `planning-with-files` 产生的 .md 文件 | `findings.md`、`task_plan.md`、`progress.md` |

**规则**：沉淀内容必须去重（查重后再追加），单条 ≤ 200 字，用 `→` `⇒` `✅` `❌` 符号。

### 复盘检查表

每次任务完成后（或标记 `done` 前），执行：

```
□ E1: 是否充分利用了并行？（至少检查一次）
□ E2: 是否有可优化的运行时决策？
□ E3: 是否沉淀了可复用资产到正确位置？
□ E4: 沉淀内容是否已去重且符号化？
```

**NO EVIDENCE = NOT COMPLETE**：沉淀的资产必须在文件中实际存在，不是口头承诺。

---

## Skill Harness Review — 评审方法论

> 评审 ve-*-ops skill 时的两层框架，沉淀自 2026-07-05 全量 skill 评审实践经验。
> 完整检查表见 [docs/skill-harness-review-checklist.md](docs/skill-harness-review-checklist.md)。

### 两层评审框架

每次 skill 评审必须覆盖两个维度：

| 维度 | 关注点 | 检查对象 |
|------|--------|---------|
| **Harness Runtime** | Skill 能否被 Agent Runtime 正确路由、执行、跟踪 | `## Trigger & Scope`、`## Steps`、`## Quality Gate (GCL)`、`{{output.*}}` |
| **Quality Compliance** | Skill 是否符合模板规范、TE 规则、GCL 标准 | frontmatter、section 完整性、TE-1~TE-9、GCL rubric |

### P0-P3 优先级定义

| 级别 | 含义 | 处理方式 | 示例 |
|------|------|---------|------|
| **P0** | blocking — 不可接受 | **必须立即修复** | 缺 Trigger & Scope / GCL / null bytes / 凭证泄露 |
| **P1** | 破坏性操作验证缺失 | **必须修复** | delete/stop 缺少确认门、错误码不足 10 个 |
| **P2** | 质量欠佳但不阻塞 | **建议修复** | 硬编码版本号、跨文件内容重复、缺少 What This Skill Does |
| **P3** | 优化项 | **有空再修** | TE-8 符号替换不彻底、压缩级别不匹配 |

### 评审执行步骤（按顺序）

```
1. 运行自动化检查 → python3 scripts/validate_local.py
2. 检查 frontmatter → name/description/license/compatibility 齐全
3. 检查 Trigger & Scope → SHOULD/SHOULD NOT 子节存在，确认路由边界
4. 检查 Steps → 每个操作有明确的 I/O 定义和失败处理
5. 检查 GCL → rubric/prompt-templates 与 docs/gcl-spec.md 对齐
6. 检查 TE 规则 → TE-1~TE-7 逐条验证（TE-1 见下方边界规则）
7. 检查文件完整性 → 无 null bytes、无截断、code fence 闭合
8. 检查 cross-skill 委派 → 跨产品操作是否正确委托
9. 给出 P0-P3 分级结论
```

### 背景 Agent 搜索指引

搜索 skill 文件时使用 **substring（子串）匹配**而非 exact match，因为：

- `## Trigger & Scope` 可能多 `(Agent-Readable)` 后缀 → 搜索 `Trigger & Scope` 而非 `## Trigger & Scope`
- heading 可能有额外空格或符号 → 搜索核心关键词即可

### TE-1 扫描边界规则

判断一个硬编码值是否算 TE-1 违规：

| 类型 | 示例 | 判定 | 理由 |
|------|------|------|------|
| **用户可选参数** | `--EngineVersion "5.0"`、`"MongoVersion": "4.4"` | ✅ **违规** | 用户可能想选其他版本 |
| **API 规范版本** | `ApiVersion: "2024-01-01"`、API endpoint 中的版本号 | ❌ 不算 | 这是 API 契约的一部分，不由用户选择 |
| **配额/限制** | `MaxResults: 100` | ❌ 不算 | 服务端限制，不取决于用户 |

> 在保持 Agent 可执行性的前提下，最小化每个 Skill 的 Token 消耗。
> 完整规则详解（含代码示例）见 [docs/token-efficiency.md](docs/token-efficiency.md)。

| 规则 | 要点 | 节省 |
|------|------|------|
| **TE-1** API 查询 > 静态表格 | 用 `ve` 命令获取版本/配额，不硬编码 | ~200-500/文件 |
| **TE-2** 省略不必要 docstring | Go SDK 用 `#` 注释代替函数级 docstring | ~100-200/函数 |
| **TE-3** 紧凑错误表 | 每行 1 个错误码，≤3 列 | ~300-500/文件 |
| **TE-4** JSON paths 集中声明 | 文件顶部统一声明，不重复 | ~50-100/文件 |
| **TE-5** YAML anchors | `example-config.yaml` 用 `&anchor` 消除重复 | ~200-400/文件 |
| **TE-6** 消除跨文件重复 | SKILL.md 已有完整流程，references 不重复 | 因 Skill 而异 |
| **TE-7** 专业内容分层 | AIOps/FinOps 放 `references/advanced/`；安全敏感操作单独标注并显式确认 | ~3,000-8,000/文件 |
| **TE-8** 符号与缩写 | 使用 `→` `⇒` `✅` `❌` 等符号和标准缩写替代冗余自然语言 | ~100-300/文件 |
| **TE-9** 压缩级别选择 | 按文档用途选择 Minimal / Efficient / Compressed / Critical / Emergency | ~500-2,000/文件 |

**不可压缩的内容**：Agent 可执行命令本身（参数、JSON paths）、错误恢复逻辑、安全门、Credential 规则、跨技能编排链。

**发现任一违规 → 立即修复 → 重新检查直到全部通过。**

### Symbol System

Standard Token Efficiency symbols for compact skill documentation:

| Category | Symbols | Usage |
|----------|---------|-------|
| **Core Logic** | `→` (leads to), `⇒` (transforms to), `←` (rollback), `⇄` (bidirectional), `&` (and), `\|` (separator), `:` (define), `»` (sequence), `∴` (therefore), `∵` (because), `≡` (equivalent), `≈` (approximately), `≠` (not equal) | Flow mapping, state transitions, causality |
| **Status** | `✅` (completed), `❌` (failed), `⚠️` (warning), `ℹ️` (info), `🔄` (in progress), `⏳` (pending), `🚨` (critical), `🎯` (target), `📊` (metrics), `💡` (insight) | Checkpoint gates, validation results |
| **Domain** | `⚡` (perf), `🔍` (analysis), `🔧` (config), `🛡️` (security), `📦` (deploy), `🎨` (design), `🌐` (network), `📱` (mobile), `🏗️` (architecture), `🧩` (components) | Section markers, scope indicators |

### Abbreviations

Standard abbreviations to reduce token count while preserving meaning:

| Abbr | Meaning | Abbr | Meaning |
|------|---------|------|---------|
| `cfg` | configuration | `impl` | implementation |
| `arch` | architecture | `perf` | performance |
| `ops` | operations | `env` | environment |
| `req` | requirements | `deps` | dependencies |
| `val` | validation | `test` | testing |
| `docs` | documentation | `std` | standards |
| `qual` | quality | `sec` | security |
| `err` | error | `rec` | recovery |
| `sev` | severity | `opt` | optimization |

### Compression Levels

Adaptive compression strategy for token management:

| Level | Range | Use Case |
|-------|-------|----------|
| **Minimal** | 0-40% | Full detail, persona-optimized clarity |
| **Efficient** | 40-70% | Balanced compression with domain awareness |
| **Compressed** | 70-85% | Aggressive optimization, quality gates active |
| **Critical** | 85-95% | Maximum compression, preserve essential context |
| **Emergency** | 95%+ | Ultra-compression, information validation |

### TE 规则速查

| TE 规则 | 检查方法 | 不通过则 |
|---------|---------|---------|
| TE-1 | 检查 references/ 中是否有硬编码的版本号/配额数字 | 替换为 `ve` 查询命令 |
| TE-2 | 检查 Go SDK 代码块是否有函数级 docstring | 删除，改用 `#` 行注释 |
| TE-3 | 检查错误表是否超过 3 列 | 合并列，每行 1 个错误码 |
| TE-4 | 检查 JSON path 是否在文件顶部集中声明 | 移至文件顶部统一声明 |
| TE-5 | 检查 example-config.yaml 是否有重复字段 | 用 YAML anchors 消除 |
| TE-6 | 检查 SKILL.md 与 references/ 是否有内容重复 | 删除 references 中的重复 |
| TE-7 | 检查 AIOps/FinOps 是否在 `references/advanced/`；SQL 执行是否标注 Security-Sensitive | 分层修复 |
| TE-8 | 检查 Markdown 中是否混用了自然语言描述与 TE symbols | 统一使用 `→` `⇒` `✅` `❌` 等符号替代冗余文字 |
| TE-9 | 检查当前压缩级别是否匹配文档目的（Minimal/Efficient/Compressed） | 调整压缩策略 |

---

## Skill Authoring Guardrails

- **Never create a new `ve-*-ops/` directory by hand.** Use the `ve-skill-generator` meta-skill flow.
- **Never invent CLI flags or API parameters.** Only official OpenAPI or `ve <service> <action> --help` verified fields.
- **Single-product rule.** One `ve-*-ops` skill = one product = one primary resource model.
- **Prefer editing existing skills over creating new files.** No tutorial-style `.md` at the repo root.
- **Orchestration / loop-agent skills are exempt from the `ve-*` prefix and `cli_applicability` rule.** Skills whose `metadata.type ∈ {orchestration-skill, loop-agent, meta-skill}` (e.g. `incident-loop-agent`) need not follow the `ve-<product>-ops` naming or carry a mandatory `cli_applicability` field. They MUST still live under their own `<name>/` directory with a `SKILL.md` + `references/`, and MUST be added to the GCL conformance allowlist (`scripts/validate_skills_frontmatter.py` `ORCHESTRATION_TYPES`). The single-product rule does not apply to them (they coordinate, not own, a product).

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
| `scripts/` (repo root) | Shared validation scripts (`validate_local.py`, `check_gcl_conformance.py`, etc.) — pre-commit use | JSON Schema, handoff contracts, example YAML |

**Owner skill rule:** the skill that **defines and primarily consumes** the contract owns the file. Secondary consumers link to the owner via relative path — they do not copy or re-home the schema.

**When adding a new `*.schema.json` or handoff contract:**
1. Pick the owner skill (primary consumer of the JSON contract).
2. Create under `ve-<owner>-ops/assets/<name>.schema.json`.
3. Reference from owner `SKILL.md` / `references/` and owner `example-config.yaml` if config-driven.
4. Secondary skills cite the owner path (e.g. `../ve-<owner>-ops/assets/...`) — never `assets/` at repo root.
5. If a future `scripts/*.py` emits JSON matching the schema, its docstring MUST point at the owner skill path (script ≠ schema owner).

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

### Rubric (5 dimensions)

| Dimension | Scale | Default threshold |
|-----------|-------|-------------------|
| **Correctness** | 0 / 0.5 / 1 | ≥ 0.5 (1.0 for `delete` / `stop` / IAM / KMS / DDL) |
| **Safety** | 0 / 1 | **= 1** (0 → ABORT immediately) |
| **Idempotency** | 0 / 0.5 / 1 | ≥ 0.5 |
| **Traceability** | 0 / 0.5 / 1 | ≥ 0.5 |
| **Spec Compliance** | 0 / 0.5 / 1 | ≥ 0.5 |

### Termination

| Condition | Behavior |
|-----------|----------|
| **PASS** | All dimensions meet threshold → return G's result |
| **MAX_ITER** | Reached max_iter → return best-so-far + unresolved items |
| **SAFETY_FAIL** | Safety=0 → **ABORT**; never partial return |

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

Automated validation scripts are available in `scripts/` for pre-commit checks (see **Files that DO NOT exist** below). After changing any skill, run `python3 scripts/validate_local.py` if available, then manually verify against the P0/P1 checklist in `ve-skill-generator/SKILL.md` and the 2-round self-review above.

### Runtime GCL

`scripts/gcl_runner.py` implements the GCL Orchestrator (Phase 2). Use it for automated GCL loops:
- External Critic via `--critic-json` or stdin (production mode)
- `--structural-critic-only` for CI/local structural smoke tests only (NOT for production mutations)

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

- **`scripts/` directory** at repo root contains validation scripts (`validate_local.py`, `check_gcl_conformance.py`, `check_markdown_links.py`, etc.). These are used for pre-commit validation. There is no `.github/workflows/` directory — CI has not been set up yet.
- **No `package.json`, `Makefile`, `CI configs`, `build scripts`, `typechecker`, or non-stdlib test runner** — except:
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

### 两层触发

1. **Skill 运行时层（对用户透明）**：GCL 运行 / 代码改动经 Skill 执行且变更落盘后，运行时自动 `codegraph sync --quiet`（挂在 GCL `### Trace` 写出之后 / Skill `recover` 步骤末尾）。用户无感知。依赖运行时 PATH 含 `codegraph` 二进制。
2. **贡献者层（改 skill 规格 / 脚本的人）**：任何代码、脚本或规格变动后，**第一时间** `codegraph sync`（仓库未 `init` 时先 `codegraph init`）。sync 须覆盖 `cmd/vet/` 子目录（Go 工具同样入图）。可选：装 git post-commit hook 自动 `codegraph sync --quiet`（本地，不提交）。

### 规则

- 新克隆仓库若未索引：先 `codegraph init`，再把 `.codegraph/` 加入 `.gitignore`（本地索引，禁止提交）。
- 每次变动后 `codegraph status` 确认新增/变更文件已入图；翻译类任务前用 `codegraph query <新符号>` 验证可达。
- `codegraph` 为本仓库**已验证存在的工具**（v1.1.6，`~/.local/bin/codegraph`），非假设。

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