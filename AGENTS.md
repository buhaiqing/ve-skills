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
- **编辑既有文件前必须先 Read 确认内容，禁止凭 Glob 缺失就 `Write` 覆盖**：子 Agent 在 worktree 内曾因 Glob 未命中既有 `run_test.go`，误判文件不存在并用 `Write` 整体覆盖，误删 249 行既有测试。规则：改任何既有文件前，**先 Read 该文件**（或用 `grep`/`git show` 确认内容），确认存在且了解现状后再 Edit；绝不用 `Write` 创建式覆盖去"改"一个可能已存在的文件。改完用 `grep "^func Test"` 之类的手段核验既有测试/符号未被静默删除。

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

## MANDATORY: Spec + Plan First, Then Code (Superpowers Three-Way Consistency)

> **铁律 — 不可打破。** 任何功能性开发（新增功能 / 新模块 / 行为变更 / 破坏性修复）**必须先有 Superpowers SDD 规格（spec）与执行计划（plan），再写代码**。spec 与 plan 是开发的先决条件，不是事后补录。

**Why**: 用户明确要求——code review 时必须对「spec / plan / code」三者做一致性匹配校验，确保功能需求不漂移。没有上游 spec/plan 的代码等于无锚点的实现，review 无从校验意图，需求易偏离。

**How to apply:**

1. **开工前必须有 spec + plan**（放在 `docs/superpowers/`）：
   - 规格：`docs/superpowers/specs/<date>-<feature>-design.md`（SDD：功能描述、状态机/契约、异常边界、验收标准）。
   - 计划：`docs/superpowers/plans/<date>-<feature>.md`（里程碑拆分、依赖图、DoD、验证命令）。
   - 格式对齐既有样例：`docs/superpowers/specs/2026-05-26-high-issues-fix-design.md`、`docs/superpowers/plans/golang-migration/2026-07-12-vet-m3-gcl.md`。
   - 既有的 `docs/l2-to-l3-tasks/*` / `docs/l3-to-l4-tasks/*` 任务卡**不替代** superpowers spec/plan——任务卡是执行分解，spec/plan 是意图与契约的单一事实源。

2. **代码必须 trace 回 spec/plan**：每个功能点都能在 spec/plan 中找到对应条目；函数名/接口契约与 plan 声明的签名一致。

3. **Code Review 强制做三者一致性校验**（不可跳过）：
   - **Spec ↔ Plan**：plan 的 DoD / 接口契约与 spec 的验收标准对齐，无矛盾。
   - **Plan ↔ Code**：plan 声明的函数 / 文件 / 行为，代码中真实存在且签名一致；不存在的须显式标注「已合并/inline/未实现」并说明原因，禁止静默缺失。
   - **Spec ↔ Code**：代码行为（尤其安全门、决策矩阵、错误码）与 spec 描述一致；发现偏离必须作为 `[BLOCKER]`/`[MAJOR]` 报出。
   - 校验结论写进 review 报告；发现三者漂移即阻断合并，先 reconcile spec/plan 再合代码。

4. **例外（须显式记录）**：< 5 行的 typo/注释/格式化改动可不走此流程，但功能型改动一律无例外。

5. **Spec 引用函数签名/字段前必须 grep 核对当前实现**：写 spec 时若声称「某函数已接收某参数」「某字段已存在」，须先 `grep` 确认磁盘上的真实签名/字段，不得凭记忆或上游任务印象填写。曾发生 spec 误称 `Aggregate(root, traces)` 已接收 `sinceHours`（实际只传到 `CollectPaths`），导致 plan 与代码签名不一致、开发期才返工。这条与全局 fact-check 门禁（命令/路径须有证据）同源，扩展到"spec 内的代码事实声明"。

## MANDATORY: 决策建议须带倾向与理由

> **铁律 — 不可打破。** 每当需要用户从**多条候选方案**中做选择（如 code review 给出 A/B/C 优化建议、修复路径取舍、设计选型），**必须同时给出我自己的推荐方案 + 推荐理由**，供用户参考，不得只罗列选项让用户自行判断。

**Why**: 用户明确要求——列一堆方案却不给立场，等于把判断成本转嫁给用户；Agent 应基于已掌握的事实主动给出专业倾向，用户再据此拍板或反驳。

**How to apply:**

1. 任何"请选择 A/B/C"式的提问，必须附带一行：`推荐：<方案 X> — <理由>`。
2. 理由须基于**可验证事实**（测试结果、spec 条款、风险权衡），而非"感觉更好"。
3. 若确实无法倾向（信息不足），须显式说"信息不足，无法推荐，需你补充：<缺什么>"，而不是假装中立。
4. 此规则覆盖所有交互场景（code review、方案选型、修复建议、配置决策），无例外。

## MANDATORY: 审计链前置（破坏性/ASK 授权须留痕）

> **铁律 — 不可打破。** 任何会执行破坏性操作或升级为 ASK 类的授权，**必须先有显式人工确认，并把"谁、何时授权"写入可追溯记录（trace / 日志）**，禁止无 provenance 的裸授权开关。

**Why**: 用户确认保持 `destructive → ASK`（非 REFUSE）的设计，但要求对"谁授权了这次破坏性操作"留审计链——避免 `--confirmed` 类开关被无差别滥用而无从追责。

**How to apply:**

1. 破坏性 / ASK 类操作的执行授权，必须由上游人工确认闸门（如 `incident-loop-agent` Step 5 收集 `{{user.confirm}}`）显式产生，不得由 `--confirmed` 等开关凭空放行。
2. 授权信号须携带 provenance（`ticket_id` / `human_handle`），并持久化到 trace（如 `Iteration.ConfirmedBy`）；无 provenance 的裸授权视为审计违规。
3. `Safety = 0` 的 REFUSE 是硬地板，任何授权开关均不得绕过。
4. 此规则与 "Spec + Plan First" 铁律协同：授权/审计行为也须在 spec/plan 中声明，code review 时一并校验三方一致。

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
| scripts/ (repo root) | **已删除** — 已迁移到 `cmd/vet/` | ~~Deprecated~~ |

**Owner skill rule**: the skill that defines and consumes a contract owns it. Create `ve-<owner>-ops/assets/<name>.schema.json`; secondary skills link via relative path. Anti-pattern: `assets/` at repo root — contracts belong to owning skill, not shared at root.

---

## Generator-Critic-Loop (GCL) — Adversarial Quality Gate

> Full spec: [docs/gcl-spec.md](docs/gcl-spec.md)（roles, rubric, loop flow, termination, trace, prompt templates, per-skill defaults, changelog）。

| Role | Job | Forbidden |
|------|-----|-----------|
| **Generator (G)** | Execute the cloud operation | modifying the rubric; self-scoring |
| **Critic (C)** | Independently audit G's output | calling `ve` / SDK / mutating anything |
| **Orchestrator (O)** | Loop control, termination | executing or scoring on its own |

**Hard constraint:** G and C MUST live in **isolated prompt contexts**. Shared context is banned.

**Rubric**: 5 dimensions (Correctness/Safety/Idempotency/Traceability/Spec Compliance). SAFETY_FAIL always ABORTs; Safety must equal 1. Per-skill `max_iter` override in `SKILL.md` (`## Quality Gate (GCL)`).

**Banned anti-patterns**: Shared G+C context · Safety=0 must ABORT · Trace must persist · Credential leakage in trace · No unbounded loops. Full list: [docs/gcl-spec.md §9](docs/gcl-spec.md#9-anti-patterns-banned).

**Cross-skill delegation**: Orchestrator MUST delegate, not absorb. Routing rules: [docs/skill-routing-graph.md](docs/skill-routing-graph.md). Critic MUST NOT call any skill — only emits suggestions.

All 29 skills (13 required + 10 recommended + 6 optional) have GCL rubric + prompt templates. 1 orchestration skill (`incident-loop-agent`) also equipped; exempt from `ve-*` prefix per Skill Authoring Guardrails.

## Evaluation

`assets/eval_queries.json` per skill holds intent-classification test cases (`should_trigger: true/false`). When adding capability, add eval cases in the same change.

**Build-time**: `vet validate --root .` (or `vet check frontmatter|links|gcl|eval|aiops|assessment`) + P0/P1 checklist + 2-round self-review. **Runtime**: `vet gcl run` — External Critic via `--critic-json`/stdin; `--structural-critic-only` for CI smoke tests; `vet gcl gate` for structural CI gate; `vet gcl trace` aggregates traces. GCL runs MUST use externally supplied isolated Critic scores in production.

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

<!-- scripts/ 已删除 — 已迁移到 cmd/vet/ Go 工具 -->
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

CodeGraph (`codegraph` CLI, `~/.local/bin/codegraph`) 维护仓库知识图谱。本仓库已配置 MCP Server（`.mcp.json`），Agent 启动时自动获得 `codegraph_explore` 工具。

### MANDATORY: 每次代码变更后必须 sync

> **铁律 — 不可打破。** 每次 Go 代码变更提交前，**必须**执行 `codegraph sync --quiet` 确保索引最新。

**Why**: 过期索引导致调用链分析和影响面判断错误。不 sync = 索引不可信 = 后续所有查询无效。

### MANDATORY: CodeGraph MCP 优先于 grep/read

> **强规则：所有代码理解任务必须先用 `codegraph explore <symbol>`，再用 grep/read 补充。**

CodeGraph 的 AST + 调用图覆盖了 grep 无法到达的跳转表（接口实现、动态派送、跨包调用）。

**执行顺序**：`codegraph explore <symbol>` → grep/read 交叉验证 → 修改代码前确认所有调用方已知

**例外**：纯文本内容搜索（日志关键字、文档内容）除外。

### 优先级

| 优先级 | 工具 | 适用场景 | Fallback |
|---|---|---|---|
| **Primary** | `codegraph explore` | 理解符号定义/调用链/影响面 | — |
| **Secondary** | `grep` / `read` | 纯文本搜索、配置文件 | `grep` 等原生工具 |
| **Tertiary** | LSP (`lsp`) | 符号重命名、引用查找、类型推断 | `grep` |

### Fallback 规则

当 CodeGraph 不可用或索引过期时，按以下顺序降级：
1. `grep` — 纯文本模式/关键字搜索
2. `glob` + `read` — 文件存在性 + 内容确认
3. `lsp` — 符号级操作（definition/references/rename）

### 触发条件

以下场景**必须**使用 CodeGraph：
- 理解某个函数/方法的实现逻辑
- 查找某个符号的所有调用方（影响面分析）
- 追踪跨包的接口实现或动态派送
- 修改导出符号前的调用方确认
- 代码重构前的结构分析

### 禁止事项

- ❌ 在 CodeGraph 可用时直接用 `grep` 做代码理解（跳过语义层）
- ❌ 修改前不确认 blast radius 就直接编辑
- ❌ 修改 Go 代码后不执行 `codegraph sync --quiet`

> 完整配置、MCP 设置、工作流集成见 [docs/codegraph-integration.md](docs/codegraph-integration.md)。

---

## MANDATORY: 结构化日志与诊断能力（Iron Law）

> **铁律 — 不可打破。** 所有 Go 工具（`cmd/vet/` 下的任何 CLI 和内部包）**必须**遵循以下日志规范，确保运行时可追溯、可诊断、可快速根因定位。

**Why**: 当前系统日志是 ad-hoc `fmt.Fprintf`，无 run ID、无时间戳、无结构化格式，多进程并发运行时无法区分日志来源，故障时只能靠 trace JSON 文件事后分析——根因定位效率极低。规范化后，`grep <run_id> *.log` 即可重建完整执行过程。

### 核心规则

| 规则 | 要求 |
|------|------|
| **Run ID** | 每个 GCL 执行 / CLI 子命令启动时生成 UUID 作为 run ID，贯穿所有日志输出 |
| **日志格式** | `<ISO_8601_ts> \| [<run_id>] \| <level> \| <component> \| <message> \| <key=value>...` |
| **关键日志点** | 启动、每轮迭代开始/结束、自愈触发/结果、Reflexion writeback、终止、JSON store 损坏 |
| **Trace 结构体** | `Trace` 必须包含 `RunID string`；`Iteration` 必须包含 `Timestamp string` 和 `DurationMs int64` |
| **日志滚动** | `MaxSize=10MB`, `MaxBackups=5`, `MaxAge=30d` |
| **审计要求** | ERROR/FATAL 日志须包含足够上下文；禁止明文凭据；原子追加写入 |

> 完整代码示例、日志点表格见 [docs/agent-runtime-patterns.md](docs/agent-runtime-patterns.md)。

---

## Communication (Language)

- **默认使用中文回复**。除非用户明确要求使用英文，否则所有回复、总结、报告均使用中文。
- 代码注释、CLI 命令和代码块内的内容保持原文（英文），不做翻译。

---

## Agent Runtime Patterns（Phase 1 Code Review 沉淀）

> 从 Phase 1 Agent Runtime 开发 + Code Review 中提炼的高阶模式，适用于所有 Go Agent 开发。完整规则与代码示例见 [docs/agent-runtime-patterns.md](docs/agent-runtime-patterns.md)。

### 规则摘要

| 规则 | 一句话 |
|------|--------|
| P1 Shell 安全 | 禁止 `sh -c`，直接传参无 shell 解析 |
| P2 Checkpoint/Resume | 有状态引擎必须支持断点续跑，用 `<=` 判断跳过已完成步骤 |
| P3 DRY Dry-Run | Dry-run 和正常执行必须走同一条引擎路径 |
| P4 提取即使用 | 从输入提取的数据必须在下游使用 |
| P5 委托超时 | 调用外部 runner 必须传 timeout |
| P6 RunID 唯一性 | 完整纳秒时间戳，不截断 |
| P7 配置化 | 硬编码值必须通过参数/环境变量暴露 |
| P8 FlagSet 解析 | 先提取子命令，再 parse 剩余 args（flag.Parse 遇非 flag 停止） |
| P9 Range 未使用 | `for i, step := range` step 未使用 → 用 `for i := range` |
| P10 CodeGraph First | 理解代码先用 `codegraph explore`，再用 `grep`/`read` 补充；修改前确认所有调用方 |

## CodeGraph 优先级原则

> 完整规则见上方 [CodeGraph Integration](#codegraph-integration--代码变动即时同步) 章节（含优先级表、Fallback 规则、触发条件、禁止事项）。

本节仅保留 Agent Runtime Patterns 中的 P10 规则摘要（见上表）。


## Knowledge Distillation Principle（知识萃取原则）

> **铁律 — 不可打破。** 所有知识产物（spec/plan/ADR/runbook）必须为 AI 协作优化，而非人类阅读优化。

### 核心原则

| 原则 | 含义 |
|------|------|
| **结构化 > 叙事化** | 状态机、接口契约、决策表、代码片段优先。禁止时间线/故事叙述。 |
| **压缩但完整** | 保留契约、不变量、决策理由。删除背景、动机、会议记录。 |
| **单一事实源** | 一个概念只在一处定义。禁止跨文件重复。 |
| **可查询** | 清晰的章节标题，便于按章节号定位。 |
| **可执行** | 包含验证命令（`make check-ci`）、测试路径、状态检查方法。 |
| **英文优先** | 代码术语已是英文；中文叙述对 Agent 是 token 噪声。 |

### 实践规则

| 规则 | 要求 | 反例 |
|------|------|------|
| **R1** | 每个 spec/plan 必须有「当前状态」章节 | ❌ 只有「背景」「目标」 |
| **R2** | ADR 必须包含「决策」「理由」「后果」三要素 | ❌ 只写「我们决定用 X」 |
| **R3** | 接口契约必须包含 Go 代码签名 | ❌ 只用自然语言描述 |
| **R4** | 状态机必须包含状态转移图（mermaid 或表格） | ❌ 只用文字描述状态 |
| **R5** | 决策表必须包含条件、动作、理由 | ❌ 只列「如果 X 则 Y」 |
| **R6** | 禁止「我们」「他们」等模糊主语 | ❌ 「我们认为应该…」 |
| **R7** | 禁止「可能」「也许」「大概」等模糊限定词 | ❌ 「这可能会导致…」 |

### 示例：AI-First vs 人类视角

**❌ 人类视角（禁止）**:
> "在 2026-08-01 的会议中，团队讨论了 Wave A 的实现方案。经过激烈讨论，我们决定先实现 heal probe，因为没有真实的探测就无法安全地自动执行。这个决策基于之前的经验教训…"

**✅ AI-First（推荐）**:
```markdown
### ADR-0001: Wave A First (heal probe + CI + eval + KB)

**Decision**: Implement heal probe before any AUTO execution.

**Rationale**: No real probe → no safe AUTO. Stub heal plans (`CheckFn: func() bool { return true }`) are fake execution.

**Consequence**: Wave A blocked until `heal/probe.go` + `heal/promote.go` complete and tested.

**Verification**: `go test ./internal/heal/ -count=1` must pass.
```

### 触发条件（MUST apply）

- 编写或修改 spec/plan/ADR/runbook 时
- 创建新的知识产物（failure pattern、decision record、architecture doc）时
- 评审（review）任何文档时，检查是否符合上述原则

### 禁止事项

- ❌ 时间线/故事型文档结构
- ❌ 背景说明和动机叙述
- ❌ 跨文件重复上下文
- ❌ 知识产物中混用中英文
- ❌ 没有明确契约/接口的模糊「建议」

### 自动验证（MUST run before commit）

```bash
# 1. 检查模糊限定词（R6/R7）
grep -nE '(我们|他们|可能|也许|大概)' <file> && echo "❌ VIOLATION: vague language"

# 2. 检查缺少契约的代码块（R3）
grep -B2 '```go' <file> | grep -v 'func\|type\|interface' && echo "⚠️ WARNING: code without signature"

# 3. 检查 ADR 三要素完整性（R2）
for f in docs/adr/*.md; do
  grep -q 'Decision' "$f" || echo "❌ $f: missing Decision"
  grep -q 'Rationale' "$f" || echo "❌ $f: missing Rationale"
  grep -q 'Consequence' "$f" || echo "❌ $f: missing Consequence"
done

# 4. 检查单一事实源（TE-6）
find . -name '*.md' -exec grep -l 'CodeGraph.*优先级' {} \; | wc -l
# 期望输出：1（仅 AGENTS.md）
```

CI 集成：在 `vet validate` 中添加 `--check-distillation` flag（TODO）。

---

## Knowledge Compound Loop（知识复利循环）

> **铁律 — 不可打破。** 每次工作都必须让系统变得更好。禁止"只消费知识、不沉淀知识"的行为。

### 复利触发器

| Trigger | Action | Compound Effect |
|---------|--------|-----------------|
| Go interface 变更 | `codegraph impact` → 更新 ARCHITECTURE.md §3 | 文档始终反映当前代码 |
| Runtime failure | Reflexion → `docs/failure-patterns.md` → spec 更新 | 失败不再重复 |
| ADR 决策 | ADR rationale → 代码注释 `// ADR-0001: ...` | 决策理由可追溯 |
| Wave C 完成 | Post-mortem → ARCHITECTURE.md §5 更新 | 经验沉淀 |
| Bug fix | 根因 → `failure-patterns.json` → 测试用例 | 回归免疫 |
| Code review | 发现 → `AGENTS.md` 规则补充 | 规则进化 |

### 复利度量（Compound Interest Metrics）

| Metric | Target | Measurement |
|--------|--------|-------------|
| Knowledge reuse rate | ≥80% | Agent sessions 加载 ARCHITECTURE.md vs 读 raw specs 的比例 |
| Failure recurrence | ≤5% | 匹配 `failure-patterns.md` 已有模式的失败占比 |
| Spec drift | 0 | ARCHITECTURE.md 契约 vs 实际 Go 签名的差异数 |
| Decision traceability | 100% | ADR 有对应代码注释的比例 |
| 文档维护成本 | ↓ 每季 | 每季度文档更新所需的 agent token 数 |

### 反模式（禁止）

- ❌ **知识消费不沉淀**：完成任务后不更新 failure-patterns / ARCHITECTURE.md
- ❌ **重复失败**：同一类 bug 出现第二次（说明 Reflexion 闭环断裂）
- ❌ **文档漂移**：代码改了但 ARCHITECTURE.md 没更新（spec drift > 0）
- ❌ **决策失忆**：ADR 写了但代码里没有注释引用（traceability < 100%）

### 复利检查清单（每次 PR 前）

```bash
# 1. 检查 spec drift
codegraph diff --against ARCHITECTURE.md  # TODO: 待实现

# 2. 检查 failure-patterns 是否更新
git diff --name-only HEAD~1 | grep -q 'failure-patterns' || \
  echo "⚠️ 本次修复是否产生了新的 failure pattern?"

# 3. 检查 ADR traceability
grep -r 'ADR-' cmd/ internal/ | wc -l  # 期望：每个 ADR 至少一处引用
```

## Key References

完整索引见 [docs/README.md](docs/README.md)（按角色分组：使用者 / 开发者 / 审查者）。核心文档：
