# 01 — 任务索引与依赖图

> 详细规划见 [`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4。
> 每张卡片可独立领走，但**执行顺序由依赖图决定**（见下）。

## 任务一览

| ID | 卡片 | 主题 | 预计影响面 | 依赖 | 估时 | 状态 |
|----|------|------|-----------|------|------|------|
| T01 | [T01-execution-risk-policy.md](./T01-execution-risk-policy.md) | 编写 `execution-risk.md` 决策规范（prose） | 1 新文件 | — | 0.5d | ✅ DONE |
| T02 | [T02-execution-risk-schema.md](./T02-execution-risk-schema.md) | 编写 `execution-risk.schema.json` JSON Schema | 1 新文件 | T01 | 0.5d | 🟡 TODO |
| T03 | [T03-domain-allowlist.md](./T03-domain-allowlist.md) | 编写 `domain-allowlist.md`（8 协调 skill 种子列表） | 1 新文件 | T01 | 0.25d | 🟡 TODO |
| T04 | [T04-leaf-op-metadata-annotation.md](./T04-leaf-op-metadata-annotation.md) | 在 8 个 leaf skill 加上 `safety_class` / `blast_radius` 元数据 | 8 个 SKILL.md 改 1 行 | — | 1d | 🟡 TODO |
| T05 | [T05-incident-loop-skill-wiring.md](./T05-incident-loop-skill-wiring.md) | 把策略接入 loop（Step 5 改造 + 变量表） | 1 个 SKILL.md | T01, T02, T03, T04 | 1d | 🟡 TODO |
| T06 | [T06-gcl-runner-runtime.md](./T06-gcl-runner-runtime.md) | 把 v0.1.0 skeleton 提升为生产 runtime（`vet gcl run` 替换 Python 脚本） | 1 个 .go + 1 个 Go 工具 | T05 | 2d | 🟡 TODO |
| T07 | [T07-trace-schema-and-validator.md](./T07-trace-schema-and-validator.md) | 全链路 trace schema + `vet` 校验（RequestId 必填） | 1 个 JSON Schema + 1 个 Go 校验 | T06 | 1d | 🟡 TODO |
| T08 | [T08-eval-and-safety-guard.md](./T08-eval-and-safety-guard.md) | eval 覆盖 AUTO/ASK/REFUSE + Safety-invariant guard | 1 个 JSON + 1 个 Go 断言 | T01, T05 | 1d | 🟡 TODO |

> **⚠️ 执行须知**
> 1. **行号漂移**：所有卡片中 `incident-loop-agent/SKILL.md:N` 行号请在执行时**重新 grep**（2026-07-13 评估已确认漂移）：8 个协调 skill 实际在 `:33-40`（非 `:32-39`），硬门在 `:152-153`/`:184`。行号仅作锚点，以 grep 实际结果为准。
> 2. **TDD 铁律**：所有 Go 实现卡片（T06/T07/T08）必须先写**失败测试**再写实现；单测是 DoD 硬证据，禁止"先实现后补测试"。
> 3. **GCL 规范**：运行时改动须遵循 [`../gcl-spec.md`](../gcl-spec.md)（隔离 G/C 上下文、Safety=0 必 ABORT、trace 持久化）。
> 4. **进度登记**：完成一条才在 [`_trace/ledger.md`](./_trace/ledger.md) 追加一条，禁止预填。

## 依赖图（严格串行在门禁层）

```
T01 ─┬─▶ T02 ─┐
     │        ├─▶ T05 ─┬─▶ T06 ─▶ T07
     ├─▶ T03 ─┘        │
     │                 └─▶ T08
T04 ─────────────────────┘
                          (T08 也依赖 T01)
```

**关键路径**：T01 → T05 → T06 → T07
**最长路径长度**：4 个卡片（T01→T05→T06→T07）
**可并行点**：T02 / T03 / T04 可与 T01 并行启动；T08 必须等 T05。

## 调度建议（人/AI 协同时）

| 阶段 | 可并行卡片 | 说明 |
|------|------------|------|
| Stage 0 | T01, T04 | 策略 + leaf 元数据可同时启动 |
| Stage 1 | T02, T03, T05 | T05 必须在 T01+T04 完成后开始 |
| Stage 2 | T06, T08 | 各自独立，可同时启动 |
| Stage 3 | T07 | 收尾，接入 trace 校验 |

## 完成度记录

每张卡片完成后，执行：
```bash
# 1. 登记进度
echo "## Txx 2026-07-13 — done" >> docs/l2-to-l3-tasks/_trace/ledger.md
echo "- Txx: <一句话交付物>" >> docs/l2-to-l3-tasks/_trace/ledger.md

# 2. 跑全局门禁
cd cmd/vet && go build ./... && go vet ./... && go test ./...
cd ../.. && vet validate --root .

# 3. 标记状态
# 把 README 表格中对应行 "🟡 TODO" 改成 "✅ DONE"
```

## 全局 L3 出口

T01–T08 全部 ✅ 后，按 [`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §6 跑：
```bash
vet check frontmatter --root .
vet check gcl --root .
vet gcl gate --root . --skip-incident-loop
vet check eval --root .
```
8 项 DoD 全部勾选 → L3 已达成。
