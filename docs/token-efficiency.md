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
| Self-Review 详细清单 | `docs/self-review.md` | AGENTS.md 只留摘要 |

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