# Skill Harness 视角专家评审报告

> 评审日期：2026-07-05
> 评审范围：ve-skills 仓库全部 28 个 `ve-*-ops` Skill + `ve-skill-generator` Meta-Skill
> 评审视角：Harness 运行时兼容性 + Skill 质量合规（Two-Round Self-Review 标准）

---

## 1. 评审方法

### 数据来源

| 来源 | 说明 |
|------|------|
| 自动化验证脚本 | `validate_skills_frontmatter.py`, `check_gcl_conformance.py`, `check_markdown_links.py`, `check_aiops_coverage.py` |
| 背景探索 Agent × 3 | Harness frontmatter audit, Skill template compliance, TE and dedup audit |
| 直接采样 | 6 个代表性 Skill（ecs/vpc/redis/cdn/sls/billing）逐行对比模板 |
| Git log | 最近 20 次提交分析演进趋势 |

### 评分标准

- **P0 (Critical)**：Harness 运行时阻断，Skill 无法被正确路由或执行
- **P1 (High)**：明确违反 AGENTS.md 或 TE 规则的违规
- **P2 (Medium)**：模板合规偏差，非功能性缺失
- **P3 (Low)**：长期优化项，不会影响当前执行

---

## 2. 第一维度：Harness 运行时兼容性

### 2.1 汇总

| 检查项 | 通过率 | 详情 |
|--------|--------|------|
| 名称 & 描述 | 28/28 ✅ | 全部符合 OpenSpec 格式，≤600 字符 |
| cli_applicability | 28/28 ✅ | 全部 `dual-path` |
| metadata.runtime | 28/28 ✅ | 全部注明兼容运行时 |
| ## Quality Gate (GCL) | 28/28 ✅ | 全部存在 |
| eval_queries.json | 28/28 ✅ | 全部 ≥5 trigger + ≥2 non-trigger |
| **## Trigger & Scope** | **27/28 🔴** | **ve-nas-ops 缺失** |

### 2.2 🔴 P0 — ve-nas-ops 缺少 `## Trigger & Scope`

**文件**：`ve-nas-ops/SKILL.md`

**问题**：该 Skill 缺少整个 `## Trigger & Scope` 章节，即没有 `### SHOULD Use This Skill When` 和 `### SHOULD NOT Use This Skill When` 子节。

**影响**：在 OpenCode / Claude Code 等 Harness 中，意图分类器无法：
- 将 NAS 相关任务（文件系统、挂载点、存储）确定性路由至此 Skill
- 提供负向边界防止误触发（如 "存储" 类意图可能错误路由到 EBS/TOS Skill）
- 该 Skill 只能通过显式名称调用，对 Harness 自动路由完全不可见

**根因**：`ve-nas-ops` 是最早一批添加的 Skill（commit `6bd0e10`），当时模板还未强制要求 `## Trigger & Scope` 章节。后续新建的 Skill 均包含该章节。

**自动化盲区**：现有 5 个验证脚本均不会检测 `## Trigger & Scope` 的存在性。

### 2.3 🟢 值得肯定

- `description` 字段格式高度统一（"Use when..." + 产品名 + 场景 + 负向排除），有利于 Harness 做模式匹配
- 全部 Skill 标注了明确的运行时兼容性（Harness AI Agent, Claude Code, Cursor）
- `cli_applicability` 一致性 100%，Harness 可以统一假设 CLI 优先的执行路径

---

## 3. 第二维度：Skill 质量合规

### 3.1 自动化门禁结果

| 验证工具 | 检查内容 | 结果 |
|----------|---------|------|
| `validate_skills_frontmatter.py` | frontmatter 必填字段 | 29/29 ✅ |
| `check_gcl_conformance.py` | GCL artifact 完整性 | 29/29 ✅ |
| `check_markdown_links.py` | 内部链接完整性 | ✅ |
| `check_aiops_coverage.py` | AIOps/FinOps 覆盖 | 23/23 required+recommended ✅ |

### 3.2 🔴 P1 — TE-1 违规（硬编码版本号，3 个 Skill 共 11 处）

| Skill | 文件 | 行 | 违规内容 |
|-------|------|----|---------|
| ve-redis-ops | `references/cli-usage.md` | 27, 43, 64, 153 | `"EngineVersion": "7.0"` / `"6.0"` |
| ve-mongodb-ops | `references/cli-usage.md` | 21, 35, 64 | `--MongoVersion 5.0` |
| ve-elasticsearch-ops | `references/cli-usage.md` | 53, 91 | `--Version "7.16"` / `--TargetVersion "8.5"` |
| ve-ecs-ops | `references/api-sdk-usage.md` | 5, 60, 86 | `API Version: 2020-04-01` |
| ve-kafka-ops | `SKILL.md` | 249, 282 | `--Version "2.6"` |

**修复建议**：使用 `ve <service> Describe*Versions` 查询命令替代硬编码值。

### 3.3 🟡 P2 — TE-8 符号使用不充分

`ve-redis-ops/references/api-sdk-usage.md`（263 行）零个 TE-8 符号（`→` `✅` `❌`），仍使用自然语言长句描述接口调用。这是最近 TE 重构后的残留。

### 3.4 🟡 P2 — CLI vs SDK 双路径覆盖不全

`ve-ecs-ops/references/cli-usage.md` "CLI vs API Coverage" 表缺少 SDK 列。虽然 `api-sdk-usage.md` 已覆盖 SDK 路径，但两个文件缺少交叉引用。

### 3.5 模板合规性：6 采样 Skill 对比

| 维度 | ve-ecs-ops | ve-vpc-ops | ve-redis-ops | ve-cdn-ops | ve-sls-ops | ve-billing-ops |
|------|-----------|-----------|-------------|-----------|-----------|---------------|
| Five Core Standards | 7 | 5 | 5 | 7 | 7 | 5 (带验证列) |
| 集中式错误分类表 | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ 16 codes |
| GCL 完整度 | 4/7 | 0/7 | 5/7 | 1/7 | 1/7 | 1/7 |
| 占位符残留 | 0 | 0 | 0 | 0 | 0 | 0 |
| What This Skill Does | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Operational Best Practices | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ |

**关键发现**：
- `ve-billing-ops` 是最接近模板标准的 Skill（唯一有第 4 列验证条件、唯一有集中式错误表）
- `ve-redis-ops` GCL 最完整（5/7），但 TE 重构中去掉了两个必要章节
- `ve-sls-ops` 错误引用最少（仅 4 处），不满足 ≥10 错误码要求
- **没有 Skill 有完整的 GCL 章节**（完整需 7 子节）

---

## 4. 修复优先级

### P0 — 立即修复

- [ ] **ve-nas-ops/SKILL.md**：补充 `## Trigger & Scope` 章节（SHOULD/SHOULD NOT 子节），参考 `ve-tos-ops` 的模式

### P1 — 系统性违规

- [ ] **ve-redis-ops/references/cli-usage.md**：替换 4 处 `EngineVersion` 硬编码为动态查询
- [ ] **ve-mongodb-ops/references/cli-usage.md**：替换 3 处 `MongoVersion` 硬编码
- [ ] **ve-elasticsearch-ops/references/cli-usage.md**：替换 2 处版本明码
- [ ] **ve-ecs-ops/references/api-sdk-usage.md**：替换 3 处 `API Version` 硬编码
- [ ] **ve-kafka-ops/SKILL.md**：替换 2 处 `--Version "2.6"`

### P2 — 模板合规

- [ ] **ve-redis-ops/references/api-sdk-usage.md**：补充 TE-8 符号
- [ ] **ve-redis-ops/SKILL.md**：恢复 `What This Skill Does` 和 `Operational Best Practices`
- [ ] **ve-ecs-ops/references/cli-usage.md**：补全 "CLI vs API" 表的 SDK 列
- [ ] **ve-sls-ops/SKILL.md**：建立集中式错误分类表（当前仅 4 个错误引用）
- [ ] 将 `ve-billing-ops` 的最佳实践（集中式错误表、第 4 列验证条件）推广到其他 Skill

### P3 — 长期优化

- [ ] 扩展所有 Skill 的 GCL 章节至完整 7 个子节
- [ ] 在 `validate_local.py` 中补充 `## Trigger & Scope` 存在性检查
- [ ] 补充 ve-billing-ops/references/advanced/finops.md (as P3 recommendation)

---

## 5. 自动化验证覆盖分析

| 验证点 | 当前覆盖 | 盲区 |
|--------|---------|------|
| Frontmatter 完整性 | ✅ 脚本覆盖 | — |
| GCL 文件存在性 | ✅ 脚本覆盖 | — |
| Markdown 链接 | ✅ 脚本覆盖 | — |
| AIOps/FinOps 覆盖 | ✅ 脚本覆盖 | — |
| **`## Trigger & Scope` 存在性** | **❌ 未覆盖** | **需新增检查** |
| TE 规则合规 | ❌ 未覆盖 | 无自动化检查 |
| 硬编码版本号 | ❌ 未覆盖 | 无自动化检查 |
| 模板章节完整性 | ❌ 未覆盖 | 无自动化检查 |

---

## 6. 总结

该仓库总体质量良好：自动化门禁全部通过，Harness 接口一致性高（96%），GCL 门禁全面部署（100%），AIOps 覆盖全面。

核心问题集中在两方面：
1. **Harness 兼容性**：1 个 Skill 缺少触发边界定义，对自动路由不可见
2. **Token Efficiency 落地不一致**：TE 规则在不同 Skill 中的执行深度不同，ve-redis-ops 重构了 SKILL.md 但 references/ 仍落后
