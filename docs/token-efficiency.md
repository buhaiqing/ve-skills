# Token Efficiency Requirements (P0 — 强制)

> 本文档是 `AGENTS.md` §Token Efficiency Requirements 的完整规则详解。
> AGENTS.md 只保留摘要表，详细做法和代码示例在此维护。

---

## TE-1: 用 API 查询替代硬编码静态数据

```markdown
# ❌ BAD: 硬编码引擎版本/配额表（50+ 行）
| Engine | Port |
|--------|------|
| MySQL | 3306 |
| PostgreSQL | 5432 |

# ✅ GOOD: 用 API 可查 + 精简表
## Supported Engines (Use API for latest)
ve [product] Describe[EngineVersions]
| Engine | Port |
|--------|------|
| MySQL | 3306 |
| PostgreSQL | 5432 |
```

**预计算约**: ~200-500 Token/文件

**适用范围**: references/core-concepts.md, references/cli-usage.md 中的版本号、端口、配额、区域列表等静态数据。

---

## TE-2: 省略不必要的文档说明

Volcengine 使用 Go SDK，用 `#` 注释代替函数级 docstring：

```go
// ❌ BAD
func createResource(client *ecs.Client) (*ecs.CreateInstanceResponse, error) {
    /* Create a new ECS instance with specified parameters.
     * Args: client - ECS client
     * Returns: response - instance details
     */
}

// ✅ GOOD: inline comment only
func createResource(client *ecs.Client) (*ecs.CreateInstanceResponse, error) {
    // Create instance with required params from OpenAPI
}
```

**预计算约**: ~100-200 Token/函数

**适用范围**: references/api-sdk-usage.md, references/integration.md 中的 Go SDK 代码块。

---

## TE-3: 错误表 → 紧凑格式

```markdown
# ❌ BAD: 每个错误 8-15 行描述
#### InvalidParameter
Cause: ...
Resolution: ...

# ✅ GOOD: 紧凑表格，每行 1 个错误
| Error Code | Agent Action |
|------------|-------------|
| InvalidParameter | FIX — align with OpenAPI |
| QuotaExceeded | HALT — request increase |
```

**预计算约**: ~300-500 Token/文件

**适用范围**: references/troubleshooting.md 中的错误码表。

**推荐格式**:

| 列名 | 必填 | 内容 |
|------|------|------|
| Error Code | ✅ | OpenAPI 错误码 |
| Agent Action | ✅ | FIX / HALT / RETRY |
| Recovery (可选) | ❌ | 一行说明恢复步骤 |
| Related Service | ❌ | 关联服务名称（仅跨服务错误需要） |

---

## TE-4: JSON paths 集中声明（不重复）

```markdown
# ❌ BAD: 每个操作后单独列 JSON paths
## Create
JSON path: $.ResourceId
## Describe
JSON path: $.Status

# ✅ GOOD: 文件顶部集中声明
# Common JSON Paths:
# Create: $.ResourceId
# Describe: $.{Status,ResourceId,RegionId}
```

**预计算约**: ~50-100 Token/文件

**适用范围**: references/api-sdk-usage.md 和 references/cli-usage.md 中频繁出现的 JSON 路径。

**集中声明位置**: 放在文件顶部的 `---` frontmatter 之后、正文之前。

---

## TE-5: YAML anchors 消除重复字段

```yaml
# ❌ BAD: Dev/Prod 各写 15 行重复字段
# ✅ GOOD: 共享 anchors
x-prod: &prod
  region: "cn-beijing"
  zone_id: "cn-beijing-a"
  internet_charge_type: "PayByTraffic"

instance:
  <<: *prod
  instance_name: "prod-web-01"
```

**预计算约**: ~200-400 Token/文件

**适用范围**: `assets/example-config.yaml` 中有多个相似环境配置时。

---

## TE-6: 消除跨文件重复流程

- SKILL.md 已有完整 Pre-flight → Execute → Validate → Recover
- `assets/example-config.yaml` 中的完整流程示例和 SDK 文件中的 Complete Setup 函数是重复内容 → 删除

**检查方法**: 在编辑完成后全局搜索是否有相同段落出现在两个文件中。

---

## TE-7: 专业内容分层

| 内容类型 | 放置位置 | 要求 |
|---------|---------|------|
| AIOps / FinOps 深度分析 | `references/advanced/` | 单独文件，SKILL.md 只留入口链接 |
| Well-Architected 评估 | `references/well-architected-assessment.md` | 强制，但放在独立文件 |
| SQL 执行 / DDL 等安全敏感操作 | SKILL.md 中标注 Security-Sensitive | 要求显式确认，不在 `references/` 中绕开安全门 |
| GCL 详细规范 | `docs/gcl-spec.md` | AGENTS.md 只留摘要 |
| Token Efficiency 详解 | `docs/token-efficiency.md` | AGENTS.md 只留摘要表 |
| Self-Review 详细清单 | `AGENTS.md` §2-round self-review | AGENTS.md 只留摘要 |

---

## TE 边界 — 不可压缩的内容

| 可压缩 | 不可压缩 |
|--------|---------|
| DocStrings、静态表格、重复流程 | Agent 可执行命令本身（参数、JSON paths） |
| 长篇架构描述 | 错误恢复逻辑、安全门、Credential 规则 |
| 多个示例变体（保留 1-2 个核心） | 跨技能编排链、AIOps/Well-Architected 场景定义 |

---

## 验证方法（C6 检查项）

逐条验证：

| TE 规则 | 检查方法 | 不通过则 |
|---------|---------|---------|
| TE-1 | 检查 references/ 中是否有硬编码的版本号/配额数字 | 替换为 `ve` 查询命令 |
| TE-2 | 检查 Go SDK 代码块是否有函数级 docstring | 删除 docstring，改用 `#` 行注释 |
| TE-3 | 检查错误表是否超过 3 列 | 合并列，每行 1 个错误码 |
| TE-4 | 检查 JSON path 是否在文件顶部集中声明 | 移至文件顶部统一声明 |
| TE-5 | 检查 example-config.yaml 是否有重复字段 | 用 YAML anchors 消除 |
| TE-6 | 检查 SKILL.md 与 references/ 是否有内容重复 | 删除 references 中的重复 |
| TE-7 | 检查 AIOps/FinOps 是否在 `references/advanced/`；SQL 执行是否标注为 Security-Sensitive | 移至 `advanced/` + 添加 Advanced Analytics 节 + Security-Sensitive 子节 |

---

## TE-8: Symbol System & Abbreviations

使用标准符号和缩写替代冗余自然语言描述，减少 Token 消耗同时保持 Agent 可读性。

### Symbol System

```markdown
# ❌ BAD: 冗余自然语言描述
If the API call fails, then the operation should retry. After that,
if the validation passes, we proceed to the next step. Otherwise
we roll back to the previous state.

# ✅ GOOD: 使用 TE symbols 紧凑表达
API call → fail → RETRY | pass → next step | else ← rollback
```

**Core Logic Symbols**:

| Symbol | Meaning | Skill Example |
|--------|---------|---------------|
| `→` | leads to, implies | `CreateInstance → InstanceRunning` |
| `⇒` | transforms to | `input.json ⇒ parsed config` |
| `←` | rollback, reverse | `migration ← rollback` |
| `⇄` | bidirectional | `sync ⇄ remote` |
| `&` | and, combine | `security & compliance` |
| `\|` | separator, or | `stop\|delete\|restart` |
| `:` | define, specify | `scope: file\|module` |
| `»` | sequence, then | `pre-flight » execute » validate » recover` |
| `∴` | therefore | `state=error ∴ trigger alarm` |
| `∵` | because | `timeout ∵ QuotaExceeded` |
| `≡` | equivalent | `TE-1 ≡ API query` |
| `≈` | approximately | `≈500 tokens saved` |
| `≠` | not equal | `actual ≠ expected` |

**Status Symbols**:

| Symbol | Meaning | Example |
|--------|---------|---------|
| `✅` | completed, passed | `✅ SnapshotCreated` |
| `❌` | failed, error | `❌ QuotaExceeded → HALT` |
| `⚠️` | warning | `⚠️ Disk >80%` |
| `ℹ️` | information | `ℹ️ Using default VPC` |
| `🔄` | in progress | `🔄 Creating instance...` |
| `⏳` | pending | `⏳ Awaiting approval` |
| `🚨` | critical, urgent | `🚨 Security group exposed` |
| `🎯` | target, goal | `🎯 TE target: ≤5K tokens` |
| `📊` | metrics, data | `📊 CPU: 45%, Mem: 62%` |
| `💡` | insight, learning | `💡 Use pagination for >100` |

**Domain Symbols**:

| Symbol | Domain | Usage |
|--------|--------|-------|
| `⚡` | Performance | Speed, optimization, latency |
| `🔍` | Analysis | Search, investigation, debugging |
| `🔧` | Configuration | Setup, tools, env vars |
| `🛡️` | Security | Protection, compliance, auth |
| `📦` | Deployment | Package, release, rollout |
| `🎨` | Design | UI, components, layout |
| `🌐` | Network | Connectivity, DNS, CDN |
| `📱` | Mobile | App, responsive |
| `🏗️` | Architecture | System structure, patterns |
| `🧩` | Components | Modules, microservices |

### Abbreviations

```markdown
# ❌ BAD: 完整单词 + 冗余上下文
The configuration for the implementation of the performance optimization
requires environment-specific validation and testing.

# ✅ GOOD: 标准缩写 + 紧凑表达
cfg for impl of perf opt requires env-specific val & test.
```

**Standard Abbreviations**:

| Abbr | Meaning | Skill Context |
|------|---------|---------------|
| `cfg` | configuration | `Apollo cfg`, `CLI cfg` |
| `impl` | implementation | `impl steps`, `Go SDK impl` |
| `arch` | architecture | `system arch`, `deployment arch` |
| `perf` | performance | `perf tuning`, `perf budget` |
| `ops` | operations | `daily ops`, `ops runbook` |
| `env` | environment | `prod env`, `env vars` |
| `req` | requirements | `min req`, `prereq` |
| `deps` | dependencies | `pkg deps`, `module deps` |
| `val` | validation | `pre-flight val`, `post-op val` |
| `test` | testing | `unit test`, `e2e test` |
| `docs` | documentation | `ref docs`, `API docs` |
| `std` | standards | `naming std`, `coding std` |
| `qual` | quality | `qual gate`, `code qual` |
| `sec` | security | `sec review`, `sec group` |
| `err` | error | `err code`, `err recovery` |
| `rec` | recovery | `rec strategy`, `auto rec` |
| `sev` | severity | `sev 1 incident`, `sev: critical` |
| `opt` | optimization | `query opt`, `cost opt` |

### 应用场景

| 场景 | 建议 | 示例 |
|------|------|------|
| 状态机/流程 | 使用 `→` `⇒` `←` | `pre-flight → exec → val → rec` |
| 分支判断 | 使用 `\|` `∴` `∵` | `err ≠ nil ∴ retry \| else → next` |
| 检查清单 | 使用 `✅` `❌` `⚠️` | `✅ C1 pass \| ❌ C2 fail` |
| 领域标记 | 使用 domain symbols | `⚡ Query perf: <100ms` |
| 缩写替代 | 使用 std abbreviations | `sec review: check IAM cfg` |

### 注意事项

- **符号不替代 Agent 可执行命令**：参数、JSON paths、CLI 命令本身保持完整
- **缩写需首次出现时定义**：如果读者可能不熟悉，首次使用用括号注明 `perf (performance)`
- **避免过度符号化**：每段不超过 3 个符号，否则影响可读性
- **错误恢复逻辑保持清晰**：错误码、恢复步骤不使用符号替代

---

## TE-9: Compression Levels

自适应压缩策略，根据文档目的和读者选择合适级别。

### Compression Levels

| Level | Range | 适用场景 | 示例 |
|-------|-------|---------|------|
| **Minimal** | 0-40% | 新技能创建、详细参考文档、安全敏感操作 | 完整 Pre-flight → Execute → Validate → Recover |
| **Efficient** | 40-70% | 日常操作指南、成熟技能的更新、内部团队使用 | 省略冗余描述，使用缩写，保留核心步骤 |
| **Compressed** | 70-85% | 快速参考卡片、错误恢复速查、CI/CD 集成 | 仅保留命令 + JSON paths + 错误码 |
| **Critical** | 85-95% | 紧急故障处理、实时 Agent 执行、Token 预算紧张 | 仅命令 + 最小验证 |
| **Emergency** | 95%+ | 极端 Token 限制（如大模型上下文窗口压力） | 单行命令 + 占位符 |

### Level Selection Matrix

| 文档类型 | 推荐 Level | Reason |
|---------|-----------|--------|
| `SKILL.md` 主 runbook | **Efficient** (40-70%) | 平衡 Agent 可执行性与 Token 消耗 |
| `references/core-concepts.md` | **Minimal** (0-40%) | 概念理解需要完整上下文 |
| `references/troubleshooting.md` | **Compressed** (70-85%) | 快速查找错误码，冗余信息有害 |
| `references/api-sdk-usage.md` | **Efficient** (40-70%) | 保留参数说明，省略详细解释 |
| `references/cli-usage.md` | **Compressed** (70-85%) | 命令速查为主 |
| `assets/example-config.yaml` | **Minimal** (0-40%) | 完整配置示例，不能省略字段 |
| GCL prompt templates | **Efficient** (40-70%) | 保留所有 GCL 必需结构 |
| `docs/gcl-spec.md` | **Minimal** (0-40%) | 规范文档需完整无歧义 |

### Level-Specific Guidelines

**Minimal (0-40%)**:
- 保留完整 docstring、示例、架构描述
- 不省略任何执行步骤
- 适用于新技能、安全敏感操作、外部审计

**Efficient (40-70%)**:
- 使用 TE symbols 替代自然语言描述
- 省略非必需的 docstring（参见 TE-2）
- 使用紧凑错误表（参见 TE-3）
- 适用于大部分日常 skill 操作

**Compressed (70-85%)**:
- 仅保留命令、参数、JSON paths
- 错误表每行仅错误码 + 建议操作
- 省略所有示例变体（保留 1 个核心）
- 适用于快速参考和熟练操作

**Critical (85-95%)**:
- 仅保留必须的命令行
- 使用占位符 `{{env.*}}` `{{user.*}}`
- 省略所有描述性文字
- 适用于实时故障处理

**Emergency (95%+)**:
- 仅单行命令 + 最小占位符
- 安全门缩减到一行显式确认
- 仅适用于极端上下文限制

### Level Transition Rules

```markdown
# 从 Minimal 到 Efficient 的典型压缩流程
1. 删除 docstring → 替换为 # 行注释 (TE-2)
2. 替换自然语言 flow → 使用 → ⇒ ← 符号
3. 展开错误表 → 紧凑格式 (TE-3)
4. 集中声明 JSON paths (TE-4)
5. 使用 YAML anchors 消除重复 (TE-5)

# 从 Efficient 到 Compressed
6. 删除所有示例变体，保留 1 个核心
7. 删除 references 中与 SKILL.md 重复的内容 (TE-6)
8. 将 AIOps/FinOps 移出主文件 (TE-7)
9. 统一使用标准缩写 (TE-8)
```

---

## TE 检查清单（Self-Review C6 扩展）

Self-Review C6 Token Efficiency 检查的完整清单，覆盖 TE-1 到 TE-9。

### TE-1 ~ TE-9 逐条检查

| # | 检查项 | 验证方法 | 通过条件 |
|---|--------|---------|---------|
| **TE-1** | API 查询替代硬编码 | grep references/ 中版本号/配额/端口数字 | 无不必要的硬编码静态数据 |
| **TE-2** | 省略不必要 docstring | 检查 Go SDK 代码块 | 无函数级 docstring，用 `#` 行注释 |
| **TE-3** | 紧凑错误表 | 检查 troubleshooting.md | 每行 1 个错误码，≤3 列 |
| **TE-4** | JSON paths 集中声明 | 检查 api-sdk-usage.md / cli-usage.md 顶部 | 文件顶部集中声明，非散落各处 |
| **TE-5** | YAML anchors | 检查 example-config.yaml | 相似环境配置使用了 `&anchor` `<<: *anchor` |
| **TE-6** | 消除跨文件重复 | 搜索 SKILL.md ↔ references/ | 无完全相同的流程段落 |
| **TE-7** | 专业内容分层 | 检查 advanced/ 目录 + Security-Sensitive 标注 | AIOps/FinOps 在 advanced/；SQL 标注 Security-Sensitive |
| **TE-8** | 符号与缩写 | 检查 Markdown 中自然语言描述 | 流程使用 `→` `⇒` `✅` `❌`；缩写首次出现有定义 |
| **TE-9** | 压缩级别匹配 | 对照 TE-9 Level Selection Matrix | 文档压缩级别匹配其用途 |

### 检查清单（可复制使用）

```markdown
## Token Efficiency 检查清单

### TE-1 硬编码检查
- [ ] 无硬编码版本号（用 `ve DescribeXxx` 替代）
- [ ] 无硬编码端口/配额/区域列表
- [ ] 静态表格仅保留不可动态查询的数据

### TE-2 Docstring 检查
- [ ] Go SDK 代码块无函数级 docstring
- [ ] 使用 `#` 行注释而非 `/* */` 块注释
- [ ] 注释仅说明"做什么"，不重复参数类型

### TE-3 错误表检查
- [ ] 每行 1 个错误码
- [ ] 列数 ≤ 3
- [ ] Agent Action 明确（FIX / HALT / RETRY）

### TE-4 JSON Paths 检查
- [ ] JSON paths 在文件顶部集中声明
- [ ] 操作中不重复声明
- [ ] 使用 `{a,b}` 语法合并多路径

### TE-5 YAML Anchors 检查
- [ ] 相似环境配置使用 `&anchor`
- [ ] 实例使用 `<<: *anchor` 引用
- [ ] 无重复字段跨环境

### TE-6 跨文件去重
- [ ] SKILL.md 与 references/ 无重复流程
- [ ] references/ 中无 SKILL.md 已有内容的副本
- [ ] 各文件保持单一责任

### TE-7 内容分层检查
- [ ] AIOps/FinOps 在 `references/advanced/`
- [ ] SQL 执行标注 Security-Sensitive
- [ ] GCL 详细规范在 `docs/gcl-spec.md`
- [ ] Token Efficiency 详解在 `docs/token-efficiency.md`

### TE-8 符号与缩写检查
- [ ] 流程使用 `→` `⇒` `⇄` 等符号
- [ ] 状态使用 `✅` `❌` `⚠️` `🔄` 等
- [ ] 缩写首次出现有定义
- [ ] 不替代 Agent 可执行命令
- [ ] 错误恢复逻辑保持清晰

### TE-9 压缩级别检查
- [ ] 文档压缩级别匹配 TE-9 Level Selection Matrix
- [ ] 未在 Minimal 级别错误使用 Compressed 技巧
- [ ] 未在 Compressed 级别遗漏关键步骤
```

### Self-Review 集成

在 AGENTS.md C6 检查项中：

```markdown
1. 对受影响的文件运行上述 TE-1 ~ TE-9 清单
2. 发现任一违规 → 立即修复
3. 重新检查直到全部通过
4. 修改 `docs/token-efficiency.md` 时需同时检查 AGENTS.md 的 TE 摘要是否对齐
```