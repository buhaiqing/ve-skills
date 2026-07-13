# T04 — Leaf Operation Metadata Annotation

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4 (P3) + §4 "On P3"
> 依赖：无（可与 T01 并行启动）
> 可并行：T01, T02, T03
> 预计工作量：1 天
> 状态：🟡 TODO

## 1. 目标（关键卡）

为 8 个协调 leaf skill 的 SKILL.md 中**每个操作**加上机器可读的
`safety_class` / `blast_radius` 元数据。

> **这是 L2→L3 的真实卡点（plan §4 "On P3"）**：
> 今天所有 leaf skill 只以散文形式分类操作，runner 无机器可读输入。
> 没有本卡，T02 的 schema 无法被消费，T05 的策略无法评分。

## 2. 背景

- 散文现状样例：`ve-ecs-ops/SKILL.md:769-772` "Destructive / Read-only" 表
- grep 验证：`grep -E 'blast_radius|safety_class' ve-*-ops/SKILL.md` = 0 hits
- 协调清单（8 个）：`incident-loop-agent/SKILL.md:32-40`

## 3. 产出物

8 个 leaf skill 的 SKILL.md 操作分类表**每行**新增 2 列：

| 原列 | 新增列 1 | 新增列 2 |
|------|---------|---------|
| Operation / Action | `safety_class` | `blast_radius` |
| DescribeXxx | `read-only` | `single` |
| StopXxx | `state-changing` | `single` |
| DeleteXxx | `destructive` | `single` 或 `multi` |

### 3.1 8 个目标 skill（按 `SKILL.md:32-40`）

| Skill | 当前是否有 ops 表 | 操作表行号（示例） |
|-------|------------------|-------------------|
| ve-cms-ops | TBD | grep 定位 |
| ve-ecs-ops | ✅ | `:769-772` |
| ve-rds-mysql-ops | TBD | grep 定位 |
| ve-redis-ops | TBD | grep 定位 |
| ve-vpc-ops | TBD | grep 定位 |
| ve-iam-ops | TBD | grep 定位 |
| ve-kms-ops | TBD | grep 定位 |
| ve-billing-ops | TBD | grep 定位 |

### 3.2 必含的"标注规范"文档

> 这是产出之一：让未来 21 个 skill 也能复用。

**新文件**：`ve-skill-generator/references/leaf-op-metadata-spec.md`

| § | 标题 | 内容 |
|---|------|------|
| 0 | Purpose | 为什么需要机器可读 metadata |
| 1 | Two columns | `safety_class` / `blast_radius` 的枚举和定义 |
| 2 | Placement rule | 必须放在操作表的**最右两列**，不新增表 |
| 3 | Default if missing | 缺列 → 默认 ASK（fail-safe） |
| 4 | Update rule | 新增操作时必须填这两列（C18 检查项） |

> **产出位置决策（已敲定，2026-07-13）**：标注规范以**独立新文件** `ve-skill-generator/references/leaf-op-metadata-spec.md` 交付，**不并入** `ve-skill-template.md`（覆盖 plan §4 P3 旧措辞 "into ve-skill-template.md"）。理由：独立文件更内聚、可单独 review，且不污染生成器模板。

## 4. DoD

```
□ 1. 写入 ve-skill-generator/references/leaf-op-metadata-spec.md（标注规范）
□ 2. 8 个目标 skill 的 SKILL.md 操作表全部加 safety_class + blast_radius 两列
□ 3. 每个有 destructive 操作的 skill，destructive 行 blast_radius 明确（single/multi/account-or-region）
□ 4. 标注规范文档第 §3 显式声明"缺列 → ASK"
□ 5. 标注规范文档第 §4 把 metadata 列为 C18（新增 DoD 项）
□ 6. 全局门禁：go build + go vet 干净
□ 7. vet check frontmatter --root . 干净（验证 SKILL.md 未破 frontmatter）
□ 8. vet check aiops --root . / vet check assessment --root . 仍干净（防误改高级章节）
□ 9. 每个被修改的 8 个 SKILL.md 的 `Changelog` 已追加版本条目（版本/日期/变更说明）
□ 10. `ve-skill-generator/references/leaf-op-metadata-spec.md` 的 `last_updated` 字段已刷新
□ 11. `incident-loop-agent/references/policies/domain-allowlist.md`（T03）已能引用本规范作为输入
□ 12. ledger 已登记（格式：`## T04 YYYY-MM-DD — done` + 8 skill 与规范 doc 一句话清单）
```

## 5. 验证命令

```bash
# 1. 8 个 skill 全部含 safety_class / blast_radius
for s in ve-cms-ops ve-ecs-ops ve-rds-mysql-ops ve-redis-ops \
         ve-vpc-ops ve-iam-ops ve-kms-ops ve-billing-ops; do
  if grep -qE 'safety_class' "$s/SKILL.md" && \
     grep -qE 'blast_radius' "$s/SKILL.md"; then
    echo "OK  $s"
  else
    echo "MISS $s"
    exit 1
  fi
done

# 2. 标注规范存在
test -f ve-skill-generator/references/leaf-op-metadata-spec.md

# 3. 全局门禁
cd cmd/vet && go build ./... && go vet ./...

# 4. vet frontmatter 仍干净（确认没破坏 SKILL.md frontmatter）
cd ../..
go build -o /tmp/vet cmd/vet/  # 或 make build
/tmp/vet check frontmatter --root .

# 5. vet aiops / assessment 仍干净
/tmp/vet check aiops --root .
/tmp/vet check assessment --root .
```

## 6. 完成回报

```markdown
## T04 2026-07-XX — done
- 8 个 leaf skill 全部标注 safety_class + blast_radius
- 标注规范：ve-skill-generator/references/leaf-op-metadata-spec.md
- vet frontmatter / aiops / assessment 仍干净
- T05 可消费（policy 现在有真实数据可评分）
```

## 7. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 标注错（destructive 漏标） | 缺列 → ASK 的 fail-safe 已保证；标注需 review |
| 8 个 skill 改太多行触发大 PR | 每个 skill 一次 commit，4 个 batch × 2 skill |
| vet 检查变红 | 操作表**只新增列**，不破坏 frontmatter；改完跑 vet 必检 |
| 回滚 | 单个 SKILL.md `git checkout` 即回滚；标注规范文档独立回滚 |
