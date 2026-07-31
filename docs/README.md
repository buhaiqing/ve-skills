# ve-skills 文档索引

> 本文件是 ve-skills 仓库文档的中央索引，按使用者角色分组。
> AGENTS.md 是 Agent 工作流入口规则，本文件是详细参考导航。

---

## 快速入口

| 文件 | 说明 |
|------|------|
| [AGENTS.md](../AGENTS.md) | Agent 工作流入口规则（含所有角色通用规则） |
| [README.md](../README.md) | 项目概述 |
| [adr/README.md](./adr/README.md) | L400 Capable 架构决策（ADR）+ [glossary](./adr/glossary-l400.md) |

---

## Skill 使用者

> 使用已有 skill 执行云资源操作的用户。

| 文件 | 说明 |
|------|------|
| `ve-*-ops/SKILL.md`（各产品） | 各云产品的运维 skill，如 [ve-ecs-ops/SKILL.md](../ve-ecs-ops/SKILL.md) |
| [skill-routing-graph.md](./skill-routing-graph.md) | 跨产品告警路由图 |
| [execution-strategy.md](./execution-strategy.md) | 执行策略（E1–E4 决策规则） |

---

## Skill 开发者

> 使用 `ve-skill-generator` 创建或维护 skill 的开发者。

### 入门
| 文件 | 说明 |
|------|------|
| [ve-skill-generator/SKILL.md](../ve-skill-generator/SKILL.md) | Meta-skill 生成器 — 完整工作流 + P0/P1 checklist |
| [ve-skill-generator/references/ve-skill-template.md](../ve-skill-generator/references/ve-skill-template.md) | 规范 SKILL.md 模板（含 GCL block） |
| [ve-skill-generator/references/cli-behavior.md](../ve-skill-generator/references/cli-behavior.md) | 验证过的 `ve` CLI 约定 |
| [ve-skill-generator/references/execution-environment.md](../ve-skill-generator/references/execution-environment.md) | CLI + Go SDK 环境配置 |
| [ve-skill-generator/references/user-experience-spec.md](../ve-skill-generator/references/user-experience-spec.md) | UX 要求（每个生成的 skill 必须满足） |

### 质量与安全
| 文件 | 说明 |
|------|------|
| [token-efficiency.md](./token-efficiency.md) | TE-1~TE-9 Token 效率规则（含代码示例） |
| [skill-harness-review-checklist.md](./skill-harness-review-checklist.md) | Skill Harness 评审清单（P0–P3） |
| [self-review-checklist.md](./self-review-checklist.md) | 两轮自检清单（Round 1 C1–C17 + Round 2 F1–F13） |
| [ve-skill-generator/references/governance-and-adversarial-review.md](../ve-skill-generator/references/governance-and-adversarial-review.md) | 治理与对抗评审（R1–R4 场景） |
| [document-integrity.md](./document-integrity.md) | 文档完整性检查（链接验证 + 去重） |

### 专项指南
| 文件 | 说明 |
|------|------|
| [skill-routing-graph.md](./skill-routing-graph.md) | 跨 skill 告警路由规则 |
| [inline-script-pattern.md](./inline-script-pattern.md) | `_inline_script()` 实现约束（验证脚本） |
| [codegraph-integration.md](./codegraph-integration.md) | CodeGraph 两层同步规则（知识图谱） |
| [agent-runtime-patterns.md](./agent-runtime-patterns.md) | Agent Runtime Patterns（P1–P9 高阶模式 + 代码示例） |

---

## GCL 开发者

> 维护或扩展 GCL（Generator-Critic-Loop）系统的开发者。

| 文件 | 说明 |
|------|------|
| [gcl-spec.md](./gcl-spec.md) | **GCL 完整规范** — 目的、角色、循环流程、trace、prompt 模板、per-skill 默认值、changelog |
| [reflexion-memory.md](./reflexion-memory.md) | Reflexion 规则 — 跨会话 failure-pattern 内存治理 |
| [failure-patterns.md](./failure-patterns.md) | Reflexion 记忆库 — 结构化失败模式（跨会话学习） |

---

## 审查者

> 评审 skill 质量、安全、合规性的审查者。

| 文件 | 说明 |
|------|------|
| [skill-harness-review-checklist.md](./skill-harness-review-checklist.md) | **Skill Harness 评审框架** — P0–P3 检查清单（R/F/TE/I/G/E 全量条目） |
| [self-review-checklist.md](./self-review-checklist.md) | **两轮自检** — Round 1（C1–C17 结构合规）+ Round 2（F1–F13 准确性） |
| [skill-harness-review-2026-07-05.md](./skill-harness-review-2026-07-05.md) | 2026-07-05 全量 skill 评审记录 |

---

## L2→L3 任务（内部执行记录）

> 这些文档是 L2→L3 任务执行过程中产生的中间产物，供历史参考。

### 任务索引
| 文件 | 说明 |
|------|------|
| [l2-to-l3-plan.md](./l2-to-l3-plan.md) | **L2→L3 执行计划** — 条件化自主执行（M1 展开）：P1–P7 任务 + L3 DoD |
| [l2-to-l3-tasks/01-index.md](./l2-to-l3-tasks/01-index.md) | T01–T08 任务索引 |
| [l2-to-l3-tasks/AGENTS.md](./l2-to-l3-tasks/AGENTS.md) | L2→L3 任务 AGENTS.md — AI 入口规则 |

### 各任务详情
| 任务 | 文件 | 说明 |
|------|------|------|
| T01 | [l2-to-l3-tasks/T01-execution-risk-policy.md](./l2-to-l3-tasks/T01-execution-risk-policy.md) | 执行风险策略 |
| T02 | [l2-to-l3-tasks/T02-execution-risk-schema.md](./l2-to-l3-tasks/T02-execution-risk-schema.md) | 执行风险 schema |
| T03 | [l2-to-l3-tasks/T03-domain-allowlist.md](./l2-to-l3-tasks/T03-domain-allowlist.md) | 域名白名单 |
| T04 | [l2-to-l3-tasks/T04-leaf-op-metadata-annotation.md](./l2-to-l3-tasks/T04-leaf-op-metadata-annotation.md) | Leaf 操作元数据标注 |
| T05 | [l2-to-l3-tasks/T05-incident-loop-skill-wiring.md](./l2-to-l3-tasks/T05-incident-loop-skill-wiring.md) | Incident loop skill 接线 |
| T06 | [l2-to-l3-tasks/T06-gcl-runner-runtime.md](./l2-to-l3-tasks/T06-gcl-runner-runtime.md) | GCL runner 运行时 |
| T07 | [l2-to-l3-tasks/T07-trace-schema-and-validator.md](./l2-to-l3-tasks/T07-trace-schema-and-validator.md) | Trace schema + 验证器 |
| T08 | [l2-to-l3-tasks/T08-eval-and-safety-guard.md](./l2-to-l3-tasks/T08-eval-and-safety-guard.md) | Eval + 安全门 |

---

## L3→L4 任务（内部规划）

> L3→L4 自主化演进的任务规划。

| 文件 | 说明 |
|------|------|
| [l3-to-l4-tasks/AGENTS.md](./l3-to-l4-tasks/AGENTS.md) | L3→L4 任务 AGENTS.md — AI 入口规则 |
| [l3-to-l4-tasks/01-index.md](./l3-to-l4-tasks/01-index.md) | T09–T16 任务索引 |

---

## 超级权限计划（历史）

> `docs/superpowers/plans/` 和 `docs/superpowers/specs/` 是历史执行计划存档。

| 文件 | 说明 |
|------|------|
| [superpowers/plans/2026-08-01-l400-capable-roadmap.md](./superpowers/plans/2026-08-01-l400-capable-roadmap.md) | L400 Capable 三波路线图 |
| [superpowers/plans/2026-08-01-l400-wave-a-foundation.md](./superpowers/plans/2026-08-01-l400-wave-a-foundation.md) | Wave A 生产地基（已实现） |
| [superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md](./superpowers/plans/2026-08-01-l400-wave-b-ops-embed.md) | Wave B 嵌入运营执行计划 |
| [superpowers/specs/2026-08-01-l400-wave-b-ops-embed-design.md](./superpowers/specs/2026-08-01-l400-wave-b-ops-embed-design.md) | Wave B 设计规格 |
| [superpowers/plans/2026-05-27-finops-aiops-optimization.md](./superpowers/plans/2026-05-27-finops-aiops-optimization.md) | FinOps + AIOps 优化计划 |
| [superpowers/plans/golang-migration/2026-07-12-python-to-go-cli.md](./superpowers/plans/golang-migration/2026-07-12-python-to-go-cli.md) | Python→Go (vet) + CodeGraph 迁移计划 |
| [superpowers/specs/2026-05-25-ve-p0-cloud-resources-design.md](./superpowers/specs/2026-05-25-ve-p0-cloud-resources-design.md) | P0 云资源设计 |
| [superpowers/specs/2026-05-26-high-issues-fix-design.md](./superpowers/specs/2026-05-26-high-issues-fix-design.md) | 高优先级问题修复设计 |

---

## Key References（来自 AGENTS.md）

> 以下是 AGENTS.md Key References 表格的完整索引。

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
