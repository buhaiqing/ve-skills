# T03 — Domain Allow-list (AUTO-eligible scope)

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §3.3 + §7
> 依赖：T01
> 可并行：T02
> 预计工作量：0.25 天
> 状态：🟡 TODO（待 T01 完成解锁）

## 1. 目标

明确**哪些产品/症状**有资格进入 AUTO 决策类。
L3 起步要"窄"（防 safety regression），先放 8 个 `incident-loop-agent` 直接协调的 skill。

## 2. 背景

- 协调清单：`incident-loop-agent/SKILL.md:32-40`（metadata.coordinates）
- 安全原则：L3 起步要**严格窄**（plan §7 风险表"Domain allow-list starts narrow"）
- 扩展路径：`count ≥ 10` patterns → 评估扩域（plan §3.3）

## 3. 产出物

**新文件**：`incident-loop-agent/references/policies/domain-allowlist.md`

### 3.1 必须包含的章节

| § | 标题 | 内容 |
|---|------|------|
| 0 | Purpose | 一句话：AUTO 决策仅对**显式列入**的 (skill, symptom) 生效 |
| 1 | Eligible skills | 列出 8 个协调 skill（`SKILL.md:32-40`） |
| 2 | Eligible symptoms | 每个 skill 的允许症状子集（如 ECS → CPU > 90%, idle 资源；不含"删除实例"） |
| 3 | Explicit exclusions | 列出**绝对不**允许 AUTO 的操作（destructive ops 全列） |
| 4 | Expansion policy | 扩域条件：`count ≥ 10` 成功 trace + safety incident = 0 + ≥ 30 天窗口 |
| 5 | Review cadence | 月度审查，owner / 流程 |

### 3.2 格式约束

- 表格 ≤ 3 列（TE-3）
- 用 `✅` `❌` 符号（TE-8）
- 行数 ≤ 80

## 4. DoD

```
□ 1. 写入 incident-loop-agent/references/policies/domain-allowlist.md
□ 2. §1 列出 8 个 skill（与 SKILL.md:32-40 完全一致，不多不少）
□ 3. §2 至少给出 3 个 skill 的症状白名单
□ 4. §3 explicit exclusions 显式列出"destructive ops 全列"原则
□ 5. §4 expansion policy 给出 count/时间/事故三维条件
□ 6. 行数 ≤ 80
```

## 5. 验证命令

```bash
# 1. 文件存在
test -f incident-loop-agent/references/policies/domain-allowlist.md

# 2. 8 个 skill 全部出现
for s in ve-cms-ops ve-ecs-ops ve-rds-mysql-ops ve-redis-ops ve-vpc-ops \
         ve-iam-ops ve-kms-ops ve-billing-ops; do
  grep -q "$s" incident-loop-agent/references/policies/domain-allowlist.md \
    || { echo "MISS $s"; exit 1; }
done
echo "8_SKILLS_OK"

# 3. 行数 ≤ 80
awk 'END{ if (NR<=80) print "LENGTH_OK"; else print "LENGTH_FAIL "NR }' \
  incident-loop-agent/references/policies/domain-allowlist.md

# 4. 全局门禁
cd cmd/vet && go build ./... && go vet ./...
```

## 6. 完成回报

```markdown
## T03 2026-07-XX — done
- 交付：incident-loop-agent/references/policies/domain-allowlist.md（XX 行）
- 8 协调 skill 全部纳入 eligible 范围
- destructive ops 显式 excluded
- T05 可消费
```
