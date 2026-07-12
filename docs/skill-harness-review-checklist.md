# Skill Harness 评审检查表

> **用途**：ve-*-ops skill 评审时的逐项核对清单。
> 沉淀自 2026-07-05 全量 29 skill 评审经验。
> 配合 `AGENTS.md §Skill Harness Review` 使用。

---

## D0 — 评审前准备

- [ ] 运行 `vet validate --root .`，确认自动化套件全绿
- [ ] 确认评审范围：单个 skill / 全部 skill / 按产品线
- [ ] 抓取被审 skill 的 `SKILL.md` + `references/*.md` + `assets/eval_queries.json`

---

## D1 — Harness Runtime（路由与执行）

| # | 检查项 | 检查方法 | 失败级别 |
|---|--------|---------|---------|
| R1 | `## Trigger & Scope` 存在 | grep heading | P0 |
| R2 | SHOULD Use / SHOULD NOT Use 子节存在 | grep subheading | P0 |
| R3 | `## Steps` 存在，每个步骤有输入输出定义 | 检查 steps 结构 | P0 |
| R4 | `## Quality Gate (GCL)` 存在（required/recommended 技能） | grep heading | P0 |
| R5 | `{{output.*}}` 占位符格式正确 | 检查所有 output capture | P1 |
| R6 | `{{env.*}}` 用于 secrets / region，无明文凭证 | 检查 `VOLCENGINE_SECRET_KEY` 等 | P0 |
| R7 | `{{user.*}}` 用于需要用户输入的参数 | 检查所有交互式参数 | P1 |
| R8 | 跨产品操作正确委派（IAM→ve-iam-ops, KMS→ve-kms-ops 等） | 按 routing graph 核对 | P1 |

---

## D2 — Quality Compliance（模板与规范）

### D2a — Frontmatter

| # | 检查项 | 检查方法 | 失败级别 |
|---|--------|---------|---------|
| F1 | name、description、license、compatibility 完整 | frontmatter 字段检查 | P1 |
| F2 | `cli_applicability` 已设置 | frontmatter 字段检查 | P2 |
| F3 | `dual-path` 时 CLI 和 SDK 步骤都存在 | 条件检查 | P2 |

### D2b — Token Efficiency（TE-1 ~ TE-9）

| # | 规则 | 检查方法 | 失败级别 |
|---|------|---------|---------|
| TE1 | 用户可选版本/配额不硬编码（`--Version "5.0"`） | grep hardcoded digits → 区分用户参数 vs API 规范 | P2 |
| TE2 | Go SDK 代码块用 `#` 注释而非函数级 docstring | grep docstring 风格 | P2 |
| TE3 | 错误表 ≤3 列，每行 1 个错误码 | 检查错误表 | P2 |
| TE4 | JSON path 集中声明在文件顶部 | 检查重复 JSON path | P2 |
| TE5 | YAML anchor 消除 `example-config.yaml` 重复 | 检查重复 YAML 字段 | P2 |
| TE6 | SKILL.md 与 references/ 无内容重复 | 跨文件查重（按 AGENTS.md §Layer 3 模板手动检） | P2 |
| TE7 | AIOps/FinOps 放 `references/advanced/`；SQL 标注 Security-Sensitive | grep 目录结构 | P2 |
| TE8 | 使用 `→` `⇒` `✅` `❌` 等符号替代冗余自然语言 | 扫描冗余描述 | P3 |
| TE9 | 压缩级别匹配文档用途（Minimal/Efficient/Compressed） | 按 Token 用量判断 | P3 |

### D2c — 文件完整性

| # | 检查项 | 检查方法 | 失败级别 |
|---|--------|---------|---------|
| I1 | SKILL.md 无 null bytes 损坏 | `python3 -c 'b"\\0" in data'` | P0 |
| I2 | 所有 code fence 闭合 | grep 成对 ``` | P1 |
| I3 | 文件无截断（末尾完整，无突然中断） | 肉眼检查 + diff | P1 |
| I4 | references/ 目录文件齐全（6 个标准文件） | ls | P2 |

### D2d — GCL 质量门

| # | 检查项 | 检查方法 | 失败级别 |
|---|--------|---------|---------|
| G1 | rubric 5 维度（Correctness/Safety/Idempotency/Traceability/Spec Compliance）完整 | 对照 `docs/gcl-spec.md` | P1 |
| G2 | max_iter 明确设定 | grep SKILL.md GCL section | P1 |
| G3 | prompt templates 使用 `{{env.*}}`/`{{user.*}}`/`{{output.*}}` 占位符 | grep template 文件 | P1 |
| G4 | 破坏性操作（delete/stop/release）的 Safety 阈值 = 1.0 | 检查 rubric threshold | P1 |
| G5 | 错误码 ≥ 10 个，含 HALT vs retry 分类 | grep HALT/retry | P1 |

### D2e — 评估数据

| # | 检查项 | 检查方法 | 失败级别 |
|---|--------|---------|---------|
| E1 | `assets/eval_queries.json` 存在 | file exists | P2 |
| E2 | ≥ 5 trigger case + ≥ 2 non-trigger case | 统计 JSON 条目 | P2 |

---

## D3 — 背景 Agent 使用提示

### 搜索策略

- **heading 搜索**：使用 substring（`Trigger & Scope`）而非 exact match（`## Trigger & Scope`）
  - 已有实际案例：`## Trigger & Scope (Agent-Readable)` 后缀导致 exact match 漏报
- **多角度搜索**：同一个信息从 heading、content、comment 三个角度分别搜索
- **交叉验证**：背景 agent 找到的结果，用直接工具快速抽样验证

### 触发时机

| 场景 | 搜索内容 | 搜索范围 |
|------|---------|---------|
| 评审前收集 | 所有 SKILL.md 的 frontmatter 字段值 | `ve-*/SKILL.md` |
| TE-1 检查 | 硬编码版本号 `--Version`、`EngineVersion`、`MongoVersion` | `ve-*/**` |
| 跨产品委派 | IAM/ECS/VPC 等跨产品调用 | `ve-*/SKILL.md` |
| 重复内容 | 跨文件的表格、代码块 | `ve-*/references/*.md` |

---

## D4 — 输出格式

评审结论按以下格式交付：

```
## Skill: ve-<product>-ops

### P0 (blocking)
- [ ] 问题描述 → 修复方案

### P1 (must fix)
- [ ] ...

### P2 (should fix)
- [ ] ...

### P3 (nice to have)
- [ ] ...

### 汇总
- P0: N 项
- P1: N 项
- P2: N 项
- P3: N 项
- 总计: N 项
```

---

## 参考链接

- [AGENTS.md §Skill Harness Review](../AGENTS.md#skill-harness-review--评审方法论)
- [docs/gcl-spec.md](gcl-spec.md) — GCL 规范全文
- [docs/token-efficiency.md](token-efficiency.md) — TE 规则详解（含代码示例）
- [docs/skill-routing-graph.md](skill-routing-graph.md) — 跨产品委派路由
- [ve-skill-generator/references/ve-skill-template.md](../ve-skill-generator/references/ve-skill-template.md) — skill 模板
- [ve-skill-generator/SKILL.md](../ve-skill-generator/SKILL.md) — P0/P1 checklist
- [skill-harness-review-2026-07-05.md](skill-harness-review-2026-07-05.md) — 2026-07-05 全量 29 skill 评审原始报告（已提炼为本文档）
