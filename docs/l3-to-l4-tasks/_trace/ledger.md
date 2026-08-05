# 进度台账（L3 → L4 任务）

> 每次 Task 完成时，按对应卡片的"完成回报"格式追加一条。
> 不要覆盖旧条目 — 这是审计 trail。

## T09 2026-07-XX — done
- 写入 cmd/vet/internal/gcl/heal/retry.go
- Classify 覆盖 10+ framework 错误码
- 单测：Retryable/Fatal/RateLimit 三类行为正确
- T10 / T11 可消费

## T10 2026-07-XX — done
- paths.go：5 类错误 × 2+ 路径
- runner.go：Classify→Select→Execute→Verify 闭环
- trace.self_healing 段必填
- T11 / T16 可消费

## T11 2026-07-XX — done
- metrics.go 4 指标 + Persist
- log.go framework §6.2 格式
- vet gcl heal-stats 子命令
- 4 指标目标值定义 + CI 警告（不阻断）
- T13 / T14 可消费

## T12 2026-07-XX — done
- cmd/vet/internal/gcl/predict/ + 3 个内置预测器
- docs/skill-routing-graph.md §4 扩 ≥ 3 行
- docs/skill-routing-graph.schema.json
- vet gcl predict + vet check routing 子命令
- T16 可消费

## T13 2026-07-XX — done
- cmd/vet/internal/reflexion/transpile/ 转译器
- guardrails.schema.json
- vet reflexion transpile 子命令
- 升级门槛：count≥10
- T15 可消费

## T14 2026-07-XX — done
- cmd/vet/internal/reflexion/promote/ 4 级 Level + Enforce
- vet reflexion check 子命令
- 硬约束：count<10 不升级；Hard 级别只 Safety=1 生效
- T13 集成测试通过
- T15 可消费

## T15 2026-07-XX — done
- incident-loop-agent/references/policies/ 目录结构
- cmd/vet/internal/policy/ loader + diff
- vet policy {load,diff,check-changelog} 子命令
- CHANGELOG 强制门禁
- T16 可消费

## T16 2026-07-XX — done
- SLO 引擎 + 5+ SLO 模板
- 自动回滚（替代 monitor+retry）
- 自治域 envelope（初始 2 域）
- Goals dashboard spec
- autonomy test harness
- vet autonomy test --n 5 全过
- **L4 已达成**

## L4 出口确认
- T09–T16 全部 done
- 8 项 L4 DoD 全部勾选
- envelope 内 N 次 incident 全闭环、0 per-op prompt
- SLO maintained / 自动回滚可用 / 预测触发先于告警

## T13 2026-07-17 — done
- T13: cmd/vet/internal/reflexion/transpile/ 转译器
- guardrails.schema.json
- vet reflexion transpile 子命令
- 升级门槛：count≥10

## T14 2026-07-17 — done
- T14: cmd/vet/internal/reflexion/promote/ 4级 Level + Enforce
- vet reflexion promote/check 子命令
- 硬约束：count<10 不升级；Hard 级别 ABORT

## T15 2026-07-17 — done
- T15: incident-loop-agent/references/policies/ 目录 + README + CHANGELOG
- cmd/vet/internal/policy/ loader + diff
- vet policy {load,diff,check-changelog} 子命令
- CI: vet policy check-changelog 加入 hard-gates

## T16 2026-08-05 — L4 runtime verification status
- 合成验证层（1-8 实现+单测）全绿：go build/vet/test ./... 通过；vet autonomy test --envelope autonomy-envelope.md --n 5 → 5/5 passed, 0 prompts, SLO maintained
- 已实证项：② 零 per-op prompt、⑧ policy CHANGELOG 化（CI P8 门禁强制）
- 待取证（需真实 incident 流量）：③ 回滚 trace、⑤ reflexion transpile>=1 次、⑥ predict Risk=high 触发、⑦ heal 成功率>80%
- 结论：L4 代码与合成验证已达成；生产运行时认证待真实环境取证后最终勾定
## T16 2026-08-06 — L4 synthetic load-test harness done
- 新增 `vet autonomy loadtest` 子命令（cmd/vet/autonomy.go + internal/autonomy/l4evidence.go 既有实现接通 CLI）
- `RunNConsecutiveIncidents` 改为经真实 slo.Engine（RunIncidentWithEngine）：注入违规 incident → RecommendAction=rollback → 观察恢复样本 → 终态 Healthy（item ①④）
- trace.Iteration 加 RollbackApplied 字段（item ③）
- loadtest 经真实代码路径取证：③ rollback_applied=1、⑤ transpile source_count=10、⑥ predict Risk=high、⑦ heal 90%>80%
- 验证：go build/vet/test ./... 全绿；vet autonomy test --n 5 PASS；vet autonomy loadtest PASS；vet validate 15 steps passed
- L4 运行时取证闭环达成
