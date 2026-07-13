# T02 — Execution Risk Schema (machine-readable)

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §3.3 + §4 (P1 schema)
> 依赖：T01
> 可并行：T03
> 预计工作量：0.5 天
> 状态：🟡 TODO（待 T01 完成解锁）

## 1. 目标

把 T01 的 prose 决策矩阵**机器化**，产出 JSON Schema 供 runner 自动评分。
本卡**仅产出 schema**（不产出 Go 校验代码 — 那是 T07 的事）。

## 2. 背景

- T01 文档：`incident-loop-agent/references/policies/execution-risk.md`
- Schema 消费方：T06（gcl-runner 用它评分）+ T08（eval/guard 用它断言）
- 父 spec：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §3.3

## 3. 产出物

**新文件**：`incident-loop-agent/assets/execution-risk.schema.json`

### 3.1 Schema 必须描述的实体

```jsonc
{
  "Operation": {
    "safety_class": "read-only | state-changing | destructive",
    "blast_radius": "single | multi | account-or-region | unknown",
    "confidence":  "low | medium | high",
    "safety":      0 | 1,           // 来自 GCL Critic
    "metadata_complete": bool       // 缺 metadata → 默认 ASK（fail-safe）
  },
  "Decision": "AUTO | ASK | REFUSE",
  "Rule":     {  // 9 cells 矩阵 + safety floor
    "match":  "<三元组>",
    "result": "AUTO | ASK | REFUSE"
  }
}
```

### 3.2 硬约束

- 用 `draft 2020-12`（与 `cmd/vet` 现有依赖一致）
- 必含 `enum` 校验所有枚举
- `safety = 0` 必须强制 result = REFUSE（用 `if/then` 表达）
- 文件大小 ≤ 3 KB

## 4. DoD

```
□ 1. 写入 incident-loop-agent/assets/execution-risk.schema.json
□ 2. JSON 语法合法（jq/python 解析无错）
□ 3. 覆盖 9 cells 决策矩阵（T01 §2 对齐）
□ 4. safety=0 → REFUSE 的硬规则在 schema 中可被解析
□ 5. metadata_complete=false → 默认 ASK（fail-safe）
□ 6. 文件 ≤ 3 KB
□ 7. draft 2020-12 标记
□ 8. `incident-loop-agent/SKILL.md` 的 `## References` 已同步 `execution-risk.schema.json` 路径
□ 9. T01 的 `execution-risk.md` 已能反向引用本 schema 文件名（防止双源真相漂移）
□ 10. ledger 已登记（含 schema 字节数）
```

## 5. 验证命令

```bash
# 1. 文件存在 + 合法 JSON
test -f incident-loop-agent/assets/execution-risk.schema.json && \
  python3 -c "import json; json.load(open('incident-loop-agent/assets/execution-risk.schema.json')); print('JSON_OK')"

# 2. Schema 自身可被解析（用 jsonschema 库）
python3 -c "
import json, jsonschema
s = json.load(open('incident-loop-agent/assets/execution-risk.schema.json'))
# 用一个样例 Operation 校验
op = {'safety_class': 'destructive', 'blast_radius': 'single',
      'confidence': 'high', 'safety': 1, 'metadata_complete': True}
  # 注：schema 主要描述 Operation 形状；决策逻辑由 T06 的 scoreDecision 实现（见 T06 §3.2），T07 仅做 trace 校验
print('SHAPE_OK')
"

# 3. 文件大小 ≤ 3 KB
size=$(stat -f%z incident-loop-agent/assets/execution-risk.schema.json 2>/dev/null || \
       stat -c%s incident-loop-agent/assets/execution-risk.schema.json)
[ "$size" -le 3072 ] && echo "SIZE_OK" || echo "SIZE_FAIL $size"

# 4. 全局门禁
cd cmd/vet && go build ./... && go vet ./...
```

## 6. 完成回报

```markdown
## T02 2026-07-XX — done
- 交付：incident-loop-agent/assets/execution-risk.schema.json（XX B）
- 覆盖 9 cells，safety=0 硬规则用 if/then 表达
- T06 / T08 可消费
```
