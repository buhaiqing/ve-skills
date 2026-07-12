# Plan: 将仓库 Python 验证脚本重写为统一 Go CLI 工具（vet）+ CodeGraph 集成

> **状态**：v2 — 基于 CodeGraph 对 Python 的分析重构（待用户确认后进入编码）
> **目标**：
> 1. 把 `scripts/*.py` 验证/CI 工具用 Go 重写为单一二进制 CLI `vet`，支持多平台发布（GitHub Release），面向**无 Python 环境的普通项目工程师**与 **AI Agent**。
> 2. 把 **CodeGraph 集成**固化为仓库规则：任何代码/脚本变动后第一时间 `codegraph sync`。
> **原则**：节奏受管控；子任务按"AI 可独立高质量执行"的粒度拆分；每个里程碑独立 commit + 独立验证；全部确认后再开编码。

---

## 0. 背景与决策（已确认）

- 工具位置：`cmd/vet/` 独立 Go module。二进制名：`vet`。
- 旧 `.py`：先保留标 deprecated，过渡后删。测试：用 `scripts/fixtures/` 对拍保证等价。
- 发布触发：tag `vet/v*` → goreleaser。
- **CodeGraph 已 init 本仓库**（287 nodes / 528 edges），`.codegraph/` 已入 `.gitignore`。

### 0.1 使用入口事实（已澄清）
- **最终用户通过「Agent Skill + Prompt」入口使用能力**，不直接敲 `codegraph` / `vet` 命令。
- `vet`（Go 二进制）和 `codegraph` 都是 **Skill 运行时背后的底层引擎**，对用户透明：Skill 的部署/运行时负责把它们放到 PATH，用户只在 Prompt 里触发 Skill。
- 这强化了「用 Go 而非 Python」的必要性：Skill 运行时环境干净、无 Python 依赖，Go 单二进制最稳。

### 0.2 CodeGraph 集成规则（分层，写入 AGENTS.md）
CodeGraph sync 在两层的触发方式不同：

1. **Skill 运行时层（对用户透明）**：GCL 运行 / 代码改动经 Skill 执行且变更落盘后，Skill 运行时自动 `codegraph sync --quiet`（挂在 GCL 的 `### Trace` 写出之后 / Skill `recover` 步骤末尾）。用户无感知。
   - 已确认：Agent Skill 运行时支持执行后自动 sync。依赖 Skill 运行时 PATH 含 `codegraph` 二进制。
2. **贡献者层（改 skill 规格的人）**：AGENTS.md 规则要求——任何代码/脚本/规格变动后**第一时间** `codegraph sync`（仓库未 init 时先 `codegraph init`）。
   - 目的双轨：① 维持仓库知识图谱最新，供 AI 检索调用链/影响面；② 为 Python→Go 翻译提供语义基底——翻译前用 `codegraph callees` / `codegraph impact <symbol>` 查依赖，避免漏迁被调方。
   - sync 须覆盖 `cmd/vet/` 子目录（Go 工具同样入图）。
   - 可选：安装 git post-commit hook 自动 `codegraph sync --quiet`（本地，不提交）。

> **已确认**：Agent Skill 运行时支持执行后自动 `codegraph sync --quiet`；第 1 层采用运行时 hook，AGENTS.md 仅额外规范贡献者侧手动 sync。

---

## 1. CodeGraph 分析结果（Python 现状，迁移依据）

### 1.1 依赖拓扑（决定迁移顺序）
- **10 个脚本零互依赖**（仅 `gcl_runner_test.py` import `gcl_runner`、`gcl_trace_aggregate_test.py` import `gta`）。→ 各脚本可独立迁，无"先底层后上层"强约束；顺序按"复杂度/风险"排。
- 复杂核心：`gcl_runner.py`（34 符号）> `check_eval_regression.py`（27）> `gcl_ci_gate.py`（18）/ `check_markdown_links.py`（21）。
- 简单叶子：`validate_skills_frontmatter.py`（16）/ `check_aiops_coverage.py`（9）/ `validate_local.py`（17）/ `validate_product_assessment.py`（19）/ `check_gcl_conformance.py`（15）/ `gcl_trace_aggregate.py`（17）/ `gcl_critic_stub.py`（8）。

### 1.2 各脚本内部结构（Go 包方法映射基础）

| 脚本 | 符号数 | 关键函数（→ Go 方法） | I/O 边界 / 注意点 |
|---|---|---|---|
| `validate_skills_frontmatter.py` | 16 | `extract_frontmatter`, `has_key`, `nested_metadata_field`, `top_level_field`, `validate_skill` | 读 `ve-*-ops/SKILL.md` 的 YAML frontmatter |
| `check_aiops_coverage.py` | 9 | `check_skill` | 读 skill 目录，查 `references/advanced/aiops.md` 存在性 |
| `validate_product_assessment.py` | 19 | `extract_example_jsons`, `validate_finding`, `validate_assessment` | 解析 markdown 中嵌入的 JSON |
| `check_gcl_conformance.py` | 15 | `_count_numbered_sections`, `check_skill`, `check_all`, `_format_human`, `cmd_check` | 校验 SKILL.md 的 GCL 结构 |
| `check_markdown_links.py` | 21 | `Finding`(类), `iter_markdown_files`, `normalize_target`, `resolve_target`, `target_exists`, `check_file` | **遍历所有 .md 并验证链接目标存在** — 文件系统 I/O 重 |
| `gcl_trace_aggregate.py` | 17 | `parse_trace`, `last_scores`, `_build_failure_patterns_table`, `aggregate`, `collect_paths`, `persist_summary`, `update_failure_patterns_file` | 读 trace JSON，聚合写回 |
| `gcl_critic_stub.py` | 8 | `main` | 输出 stub critic JSON — 可内联进 `vet gcl run --critic-stdin` |
| `validate_local.py` | 17 | `Step`(类), `_check_file_integrity`, `_check_required_sections`, `_check_error_taxonomy`, `_check_te1_hardcodes`, `build_steps`, `_inline_script`, `run_step` | **动态生成并执行内联 Python 脚本**（调 `sys.executable`）— Go 版保留"执行外部命令"语义 |
| `check_eval_regression.py` | 27 | `_tokenize`, `_score_matches`, `_validate_eval_schema`, `_get_changed_skills_by_git`, `_check_delta`, `_check_skill` | **调 `git` 命令** + 读 eval_queries.json，做语义漂移检测 — 最复杂校验类 |
| `gcl_runner.py` | 34 | `mask_secrets`, `has_credential_leak`, `run_command`, `structural_critic`, `load_critic`, `run_isolated_critic`, `validate_critic_payload`, `decide`, `extract_failure_pattern`, `load_known_failure_patterns`, `_writeback_failure_pattern`, `detect_credential_fields`, `derive_operation_intent`, `persist_trace`, `cmd_run` | **GCL 核心**：执行生成器命令、隔离批评器、打分裁决、写 trace、回填 failure pattern。含凭据遮蔽逻辑（安全关键） |
| `gcl_ci_gate.py` | 18 | `_runner_path`, `smoke_skill`, `smoke_all`, `_format_human`, `cmd_check` | CI 用：对 skill 跑 structural smoke 测试；概念与 `gcl_runner` 重叠，Go 应抽公共包 |

### 1.3 共享逻辑（Go 应抽公共包，避免重复）
- **凭据遮蔽**：`gcl_runner.mask_secrets` / `has_credential_leak` / `detect_credential_fields` — `vet gcl` 全程复用。
- **critic 评分/裁决**：`structural_critic` / `validate_critic_payload` / `decide` — `gcl_runner` 与 `gcl_ci_gate` 共用。
- **人类可读格式化**：各 `cmd_*` 的 `_format_human` 模式 — 抽 `format` 公共函数。
- **trace 读写**：`gcl_runner.persist_trace` 与 `gcl_trace_aggregate.parse_trace`/`collect_paths` 共用 JSON 结构 — 抽 `trace` 包定义 schema。

---

## 2. 子命令映射（脚本 → CLI）

| Python 脚本 | `vet` 子命令 | Go 包（建议） |
|---|---|---|
| `validate_skills_frontmatter.py` | `vet check frontmatter` | `check/frontmatter` |
| `check_aiops_coverage.py` | `vet check aiops` | `check/aiops` |
| `validate_product_assessment.py` | `vet check assessment` | `check/assessment` |
| `check_gcl_conformance.py` | `vet check gcl` | `check/gcl` |
| `check_markdown_links.py` | `vet check links` | `check/links` |
| `check_eval_regression.py` | `vet check eval` | `check/eval` |
| `validate_local.py` | `vet validate` | `validate` |
| `gcl_runner.py` | `vet gcl run` | `gcl/run` + `gcl/internal`（凭据/评分/trace 公共） |
| `gcl_ci_gate.py` | `vet gcl gate` | `gcl/gate`（复用 `gcl/internal`） |
| `gcl_trace_aggregate.py` | `vet gcl trace` | `gcl/trace`（复用 `gcl/internal` trace schema） |
| `gcl_critic_stub.py` | （内联进 `vet gcl run --critic-stdin`） | — |
| `*_test.py` | 转 Go `_test.go` | 各对应包 |

> 公共包：`gcl/internal/{secret,critic,trace,format}`。

---

## 3. 子任务拆分（AI 友好粒度 + 并行性标注）

> 每个子任务 = 1 个 Python 文件 → 1 个 Go 包 + 1 个 `_test.go`，含**等价判据**。
> 子 spec 独立成文于 `docs/superpowers/plans/vet-*.md`，AI 仅读自己那份。
> **并行标注**：`🔀并行组` = 同组内子任务互不依赖、可同时开发；`🔒锁` = 必须先完成才能解锁后续并行组。

### M0 — Bootstrap（本仓库，已完成大部分）
- [x] `git push` 3 个 commit（已完成）
- [ ] 同步过期计划文档 61 复选框（待你授权）
- [x] `codegraph init` + `.gitignore`（已完成）
- [ ] AGENTS.md 加 `## CodeGraph Integration` 节（待做）

### M1 — 脚手架 + 发布管线（串行，有构建依赖）
- 🔒 M1.1：建 `cmd/vet/` module（`go.mod`/`main.go` 路由骨架/`README.md`）→ 后续依赖此产物
- M1.2：goreleaser + GitHub Actions release 流程（占位 `vet version` 跑通 5 平台）｜依赖 M1.1
- M1.3：CodeGraph 收录 `cmd/vet/`（`codegraph sync` + `query vet`）｜最后

### M2 — 校验类（check 组，🔀 全 7 个可并行）
> 依据 CodeGraph：10 脚本零互依赖，故 M2.1–M2.7 之间无耦合，可同时开 7 个 worktree 并行开发。
> 每个子任务独立 worktree + 独立 commit + 独立 Critic 评审（GCL §16.1）。
- 🔀 M2.1 `vet check frontmatter` ← `validate_skills_frontmatter.py`（最简单，先练手）
- 🔀 M2.2 `vet check aiops` ← `check_aiops_coverage.py`
- 🔀 M2.3 `vet check assessment` ← `validate_product_assessment.py`
- 🔀 M2.4 `vet check gcl` ← `check_gcl_conformance.py`
- 🔀 M2.5 `vet check links` ← `check_markdown_links.py`（文件系统 I/O 重）
- 🔀 M2.6 `vet check eval` ← `check_eval_regression.py`（最复杂，调 git + 语义漂移；独占 1 Agent）
- 🔀 M2.7 `vet validate` ← `validate_local.py`（动态内联 Python 执行语义）

### M3 — GCL 运行类（gcl 组，🔒 公共包先锁，🔀 上层并行）
- 🔒 M3.0 `gcl/internal` 公共包：`secret`/`critic`/`trace`/`format` — **先完成并冻结接口**，解锁 M3.1–M3.3
- 🔀 M3.1 `vet gcl run` ← `gcl_runner.py`（含吸收 `gcl_critic_stub` 的 `--critic-stdin`）｜依赖 M3.0
- 🔀 M3.2 `vet gcl gate` ← `gcl_ci_gate.py`（复用 M3.0）｜依赖 M3.0
- 🔀 M3.3 `vet gcl trace` ← `gcl_trace_aggregate.py`（复用 M3.0 trace schema）｜依赖 M3.0
- M3.4 原 `*_test.py` 转 Go `_test.go`｜在 M3.0–M3.3 合并后做

### M4 — 切换 + 清理（串行，依赖 M2/M3 全完成）
- M4.1 CI（`.github/workflows/validate.yml`）改调 `vet`
- M4.2 旧 `scripts/*.py` 标 deprecated（或删，待你拍板）
- M4.3 AGENTS.md / SKILL.md 指引改 `vet`；CodeGraph 节补翻译辅助指引

### M5 — 首次发布（串行）
- M5.1 打 tag `vet/v0.1.0` → goreleaser 出 5 平台 Release
- M5.2 `cmd/vet/README.md` 写安装/用法/CodeGraph 辅助翻译示例

### 并行执行约束
- 每个并行子任务在**独立 git worktree** 开发（git-worktree.md），互不踩文件。
- 公共包（M3.0）接口冻结后，并行任务只能依赖其**已发布签名**，不得改内部。
- 复杂任务（M2.6 / M3.1）独占 1 个 Agent，不与其他塞同批（E1：每 Agent ≤ 3 文件）。
- 并行组内任一子任务失败 ≥2 轮 → 主 Agent 接管，不阻塞其他并行任务。

---

## 4. 等价判据模板（每个子任务 spec 必须含）

每个子 spec 写明：
- **输入**：同 Python 版（如 `vet check frontmatter --root .`）
- **输出契约**：exit code 语义、stdout 格式（human/json）、错误码
- **对拍方法**：用 `scripts/fixtures/` 样例，Python 输出 vs Go 输出逐字节/逐字段比对
- **安全约束**：涉及凭据的（`gcl` 组）必须保留 `mask_secrets` 行为，trace 中不出现明文密钥

---

## 5. 独立子任务文档索引（AI 执行时按里程碑读取对应文档）

| 文档 | 范围 | 并行性 |
|------|------|--------|
| `2026-07-12-vet-m1-scaffold.md` | M1：module + goreleaser + CodeGraph 收录 | 串行 |
| `2026-07-12-vet-m2-check.md` | M2：校验类 7 子任务 | 🔀 7 并行 |
| `2026-07-12-vet-m3-gcl.md` | M3：GCL 公共包 + run/gate/trace | 🔒 公共包锁，🔀 3 并行 |
| `2026-07-12-vet-m4-cutover.md` | M4：CI/文档切换 | 串行 |
| `2026-07-12-vet-m5-release.md` | M5：首次发布 | 串行 |

> 每个子文档含：交付物、等价判据、测试点、退出标准。AI 执行某里程碑时仅读对应文档，不被其他任务噪音干扰。

---

## 6. 状态

- **M0**：✅ 完成（push / codegraph init / 计划定稿 / AGENTS.md CodeGraph 节）
- **M1–M5**：⏳ 待执行（独立文档已就绪）

> 开发过程按里程碑推进，每里程碑结束汇报结果，不跨里程碑自动推进。
