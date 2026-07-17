# 01 — 任务索引与依赖图

> 详细规划见 [`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M2/M3/M4。
> 继续沿用 L2→L3 卡片格式（每张 Txx 单独 `Txx-<name>.md`）。

## 任务一览

| ID | 卡片 | 主题 | 产出物 | 依赖 | 估时 | 状态 |
|----|------|------|-------|------|------|------|
| T09 | [T09-smart-retry.md](./T09-smart-retry.md) | L2 智能重试：错误分类驱动 | 1 Go 包 + 1 错误码映射 | T06 (vet gcl run) | 1d | ✅ DONE |
| T10 | [T10-multi-path-healing.md](./T10-multi-path-healing.md) | L3 多路径自愈：auto-select best | 1 Go 包 + 2+ 路径表 | T09 | 1.5d | ✅ DONE |
| T11 | [T11-self-healing-telemetry.md](./T11-self-healing-telemetry.md) | 自愈遥测 + 日志 | 1 Go 指标包 + 1 日志 schema | T09, T10 | 1d | ✅ DONE |
| T12 | [T12-predictive-trigger.md](./T12-predictive-trigger.md) | 预测式触发源 | 1 Go 触发器 + 1 路由扩展 | T06 | 2d | ✅ DONE |
| T13 | [T13-pattern-to-policy.md](./T13-pattern-to-policy.md) | Pattern→Policy 转译器 | 1 Go 转译器 + 1 护栏 schema | T11, T14 | 1.5d | 🟡 TODO |
| T14 | [T14-reflexion-promotion.md](./T14-reflexion-promotion.md) | Reflexion HINT→Constraint 升级 | 1 Go 升级机制 | T11 | 1d | 🟡 TODO |
| T15 | [T15-versioned-policy-lib.md](./T15-versioned-policy-lib.md) | 版本化 policy library | 1 Go loader + 目录结构 | T13, T14 | 1d | 🟡 TODO |
| T16 | [T16-l4-slo-envelope-dashboard.md](./T16-l4-slo-envelope-dashboard.md) | SLO 目标 + 自治域 + goals dashboard | 1 Go SLO 引擎 + 1 dashboard | T10, T12, T15 | 3d | 🟡 TODO |

## 依赖图

```
                     ┌── T09 ── T10 ──┐
                     │       │         │
T06 (L2→L3) ───────┬┘       ▼         │
                   │      T11 ─────┐   │
                   │              │   │
                   └──▶ T12 ──────┼───┤
                                  ▼   ▼
                              T14  T13 ── T15 ──┐
                                  │              ▼
                                  └──────▶ T16 (L4 出口)
```

**关键路径**：T06 → T09 → T10 → T11 → T13 → T15 → T16（7 卡片）
**最长路径长度**：7 个卡片
**可并行点**：T09 ∥ T12；T10 ∥ T12；T13 ∥ T14（T14 不依赖 T13）

## 调度建议

| 阶段 | 可并行卡片 | 说明 |
|------|------------|------|
| Stage 0 | T09, T12 | 自愈 + 预测同时启动 |
| Stage 1 | T10, T11 | T11 可在 T10 完成后并入 |
| Stage 2 | T13, T14 | Pattern 转译与 Reflexion 升级独立 |
| Stage 3 | T15 | 收口 policy library |
| Stage 4 | T16 | 终点：SLO 目标 + envelope + dashboard |

## 完成度记录

每张卡片完成后：
```bash
# 1. 登记进度（追加到 _trace/ledger.md）
echo "## Txx 2026-07-XX — done" >> docs/l3-to-l4-tasks/_trace/ledger.md
echo "- Txx: <一句话交付物>" >> docs/l3-to-l4-tasks/_trace/ledger.md

# 2. 跑全局门禁
cd cmd/vet && go build ./... && go vet ./... && go test ./...
cd ../.. && vet validate --root .

# 3. 标记状态
# 把本表对应行 "🟡 TODO" 改成 "✅ DONE"
```

## 全局 L4 出口

T09–T16 全部 ✅ 后，按 [`../autonomous-ops-roadmap.md`](../autonomous-ops-roadmap.md) §M4 Exit 跑：
```bash
# 1. L2→L3 L3→L4 全部卡片 done
# 2. L4 envelope 内连续 N 次 incident end-to-end
# 3. 零 per-op prompt（envelope 内）
# 4. 自动回滚验证可用
# 5. SLO 维持达标
# 6. pattern→policy 升级闭环
```

N 次连续成功 + 5 项 L4 DoD 全部勾选 → **L4 已达成**。
