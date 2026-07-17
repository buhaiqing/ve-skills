# Phase 1 — Agent Runtime (SDD Spec)

日期：2026-07-17
对应架构：`docs/vet-agent-evolution-architecture.md`

---

## 1. 功能描述

### 1.1 核心目标

把 `incident-loop-agent/SKILL.md` 的 7 步编排逻辑从 prose 描述变为可执行 Go 代码，使 `vet` 能够以 Agent 方式接收 incident payload 并自动完成全流程。

### 1.2 7 步状态机

```
INGEST → TRIAGE → DIAGNOSE → PROPOSE → CONFIRM → EXECUTE → REFLEXION
```

| Step | 输入 | 输出 | 说明 |
|------|------|------|------|
| INGEST | JSON payload / JIRA ticket | `IncidentPayload` 结构体 | 解析并标准化事件源 |
| TRIAGE | `IncidentPayload` + routing graph | `TriageResult{primary, secondary[], confidence}` | 路由匹配 |
| DIAGNOSE | `TriageResult` | `DiagnosisEvidence` | 并行调用 ve-*-ops 的 Describe* |
| PROPOSE | `DiagnosisEvidence` | `DispatchPlan` | 构建操作计划 + GCL loop |
| CONFIRM | `DispatchPlan` + policy | `ConfirmResult{AUTO/ASK/REFUSE}` | 策略评估 |
| EXECUTE | `DispatchPlan` (AUTO) | `ExecuteResult` | 委托给 ve-*-ops skill |
| REFLEXION | `ExecuteResult` | 写回 failure-patterns.json | 自动学习 |

### 1.3 状态持久化

每个 run 的状态保存在 `.runtime/agent/runs/<run_id>/state.json`，支持：
- 断点续跑（进程重启后从上次 checkpoint 继续）
- 人机交互暂停（ASK 类操作暂停，等待确认后继续）
- 并发安全（每个 run 独立目录）

### 1.4 CLI 接口

```bash
# 从 JSON payload 执行
vet agent run --payload '{"product":"ecs","symptom":"cpu>90%"}'

# 从 JIRA ticket 执行
vet agent run --ticket DOPS-12345

# 恢复暂停的 run
vet agent resume --run-id a1b2c3d4 --confirm

# 查看 run 状态
vet agent status --run-id a1b2c3d4
```

---

## 2. 异常边界

| 场景 | 处理 |
|------|------|
| `product_hint` 缺失 | FAIL_FAST，询问用户 |
| 路由图无匹配 | 降级到 `ve-cms-ops`（Rule 5） |
| DIAGNOSE 超时 | 返回已收集的部分证据，标记 `partial=true` |
| CONFIRM = REFUSE | 跳过 EXECUTE，直接进入 REFLEXION |
| CONFIRM = ASK 无 `--confirm` | 暂停，持久化状态，等待 `vet agent resume` |
| EXECUTE 失败 | GCL heal 重试（max 3），仍失败则 rollback + REFLEXION |

---

## 3. 验收标准

- [ ] `cmd/vet/internal/agent/engine.go` 实现 7 步状态机
- [ ] `vet agent run --payload "..."` 端到端可执行
- [ ] 状态持久化 + 断点续跑
- [ ] 复用现有 GCL/heal/policy/memory 全部基础设施
- [ ] `go build/vet/test` 全绿
- [ ] 集成测试覆盖完整 7 步流程
