# ADR-0002: Stub heal 禁止生产 AUTO

## Status

Proposed

## Context

P0 已引入 `RecoveryStep.Stub` / `ProbeArgv` / `AllowProductionAuto` / `ExecutePlan` 拒 stub。Confirm 对绑定 stub `HealIncidentType` 的 plan 强制 ASK。defaultPlans 仍全部 Stub。

L400 Capable 要求 agents **可靠地**参与运营动作；恒 true / 无探针的 Check 是成熟度反模式（治理看起来通过、实际未验证）。

## Decision

1. **生产路径**：`plan.IsStub() == true` ⇒ 不得 `AllowProductionAuto`；Confirm 最高 `ASK`（不得 AUTO）。  
2. **升级路径**：仅当步骤具备真实 `ProbeArgv`（argv 直传 `ve`/`metric`，禁止 `sh -c`）且 `Stub=false` 时，该 plan 可参与 AUTO。  
3. **首批范围（Wave A）**：至少 **2** 个 incident type（推荐 `cpu_high`、`redis_slow_query`）完成非 Stub 升级；其余保持 Stub+ASK。  
4. **测试**：`AllowStub` 仅测试用；生产 `ExecuteOpts{}` 永不设 AllowStub。

## Consequences

- 正面：AUTO 与真实只读验证绑定；与 AGENTS P1 Shell 安全一致。  
- 正面：可渐进 promote，避免一次改全部 plan。  
- 负面：短期大量 symptom 仍 ASK，人工闸门负担不变（可接受）。  
- 后续：A1/A2；Promote API / 配置化 ProbeArgv 可在 Wave A spec 细化。

## Alternatives considered

| 方案 | 为何不选 |
|------|----------|
| Stub 也可 AUTO（标 shadow） | 与 P0 目标矛盾，成熟度倒退 |
| 一次升级全部 defaultPlans | 范围过大；探针依赖真实 `ve` 环境 |
