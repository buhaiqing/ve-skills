# 进度台账（Trace Ledger）

> 每次 Task 完成时，按对应卡片的"完成回报"格式追加一条。
> 不要覆盖旧条目 — 这是审计 trail。
>
> **执行前基线（2026-07-13 评估）**：T01–T08 全部 **未开始**。
> - `01-index.md` 状态列均为 🟡 TODO；
> - 磁盘上尚无任何交付物：`incident-loop-agent/references/policies/`、`execution-risk.schema.json`、`ve-skill-generator/references/leaf-op-metadata-spec.md`、`cmd/vet/internal/check/{trace,policyguard}/` 均不存在。
> - 此前误填的 "done" 占位条目（含 `2026-07-XX` / `XX 行` 占位）已清空，杜绝假完成信号。
>
> 完成一条、登记一条；不要预填。

## T01 2026-07-13 — done

- 交付：`incident-loop-agent/references/policies/execution-risk.md`（68 行）
- 9 cells 决策矩阵（3×3 网格）已对齐 plan §3.2；含 0/0/0 边界 + destructive 3 cells
- Safety=0 hard floor 显式声明且为 §5 首分支
- `SKILL.md` 已加 `## References`（链接 policy）+ `## Changelog` 0.2.0 条目
- GCL 纪律式自审 PASS（Safety=1），trace: `_trace/gcl-trace-T01-20260713-a1f3c9.md`
- T02/T03 可解锁

