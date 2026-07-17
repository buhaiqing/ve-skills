# L2 → L3 主题需求目录

本目录收纳 L2（受控自动执行）→ L3（条件自动执行）演进相关的 Superpowers 规格（spec）与执行计划（plan），按 `specs/` 与 `plans/` 子目录组织。

## 内容索引

| 主题 | Spec | Plan |
|------|------|------|
| incident-loop-agent L3 升级 | [specs/2026-07-17-incident-loop-agent-l3-design.md](specs/2026-07-17-incident-loop-agent-l3-design.md) | [plans/2026-07-17-incident-loop-agent-l3.md](plans/2026-07-17-incident-loop-agent-l3.md) |
| L3 退出闸门（P5/P7 CI 强制） | [specs/2026-07-17-l3-exit-gate-design.md](specs/2026-07-17-l3-exit-gate-design.md) | [plans/2026-07-17-l3-exit-gate.md](plans/2026-07-17-l3-exit-gate.md) |
| T01 执行风险策略（prose 锚定） | [specs/2026-07-17-l2-l3-t01-execution-risk-policy-design.md](specs/2026-07-17-l2-l3-t01-execution-risk-policy-design.md) | [plans/2026-07-17-l2-l3-t01-execution-risk-policy.md](plans/2026-07-17-l2-l3-t01-execution-risk-policy.md) |
| T02 执行风险 Schema（机器化） | [specs/2026-07-17-l2-l3-t02-execution-risk-schema-design.md](specs/2026-07-17-l2-l3-t02-execution-risk-schema-design.md) | [plans/2026-07-17-l2-l3-t02-execution-risk-schema.md](plans/2026-07-17-l2-l3-t02-execution-risk-schema.md) |
| T03 域允许列表（AUTO 资格） | [specs/2026-07-17-l2-l3-t03-domain-allowlist-design.md](specs/2026-07-17-l2-l3-t03-domain-allowlist-design.md) | [plans/2026-07-17-l2-l3-t03-domain-allowlist.md](plans/2026-07-17-l2-l3-t03-domain-allowlist.md) |
| T04 Leaf 操作元数据标注 | [specs/2026-07-17-l2-l3-t04-leaf-op-metadata-design.md](specs/2026-07-17-l2-l3-t04-leaf-op-metadata-design.md) | [plans/2026-07-17-l2-l3-t04-leaf-op-metadata.md](plans/2026-07-17-l2-l3-t04-leaf-op-metadata.md) |
| T05 Incident Loop Skill 改造（Step 5 决策门） | [specs/2026-07-17-l2-l3-t05-incident-loop-wiring-design.md](specs/2026-07-17-l2-l3-t05-incident-loop-wiring-design.md) | [plans/2026-07-17-l2-l3-t05-incident-loop-wiring.md](plans/2026-07-17-l2-l3-t05-incident-loop-wiring.md) |
| T06 GCL Runner 生产化运行时 | [specs/2026-07-17-l2-l3-t06-gcl-runner-runtime-design.md](specs/2026-07-17-l2-l3-t06-gcl-runner-runtime-design.md) | [plans/2026-07-17-l2-l3-t06-gcl-runner-runtime.md](plans/2026-07-17-l2-l3-t06-gcl-runner-runtime.md) |
| T07 Trace Schema + 校验器（双包正交） | [specs/2026-07-17-l2-l3-t07-trace-schema-validator-design.md](specs/2026-07-17-l2-l3-t07-trace-schema-validator-design.md) | [plans/2026-07-17-l2-l3-t07-trace-schema-validator.md](plans/2026-07-17-l2-l3-t07-trace-schema-validator.md) |
| T08 Eval 覆盖 + Safety 不变量 Guard | [specs/2026-07-17-l2-l3-t08-eval-safety-guard-design.md](specs/2026-07-17-l2-l3-t08-eval-safety-guard-design.md) | [plans/2026-07-17-l2-l3-t08-eval-safety-guard.md](plans/2026-07-17-l2-l3-t08-eval-safety-guard.md) |

> **T07 说明**：原 T07 仅覆盖 incident-trace（`check/trace` 包）；2026-07-17 扩展新增 gcl-trace runtime 校验（`gcl/trace` 包），两包正交。
> **T08 说明**：2026-07-17 补完 P7 CI e2e 接线（`validate.yml` 含 `vet check policyguard` + `vet check trace`）。
| T02 执行风险 schema（机器化） | [specs/2026-07-17-l2-l3-t02-execution-risk-schema-design.md](specs/2026-07-17-l2-l3-t02-execution-risk-schema-design.md) | [plans/2026-07-17-l2-l3-t02-execution-risk-schema.md](plans/2026-07-17-l2-l3-t02-execution-risk-schema.md) |
| T03 域允许列表（AUTO 范围） | [specs/2026-07-17-l2-l3-t03-domain-allowlist-design.md](specs/2026-07-17-l2-l3-t03-domain-allowlist-design.md) | [plans/2026-07-17-l2-l3-t03-domain-allowlist.md](plans/2026-07-17-l2-l3-t03-domain-allowlist.md) |
| T04 Leaf Op 元数据标注 | [specs/2026-07-17-l2-l3-t04-leaf-op-metadata-design.md](specs/2026-07-17-l2-l3-t04-leaf-op-metadata-design.md) | [plans/2026-07-17-l2-l3-t04-leaf-op-metadata.md](plans/2026-07-17-l2-l3-t04-leaf-op-metadata.md) |
| T05 Incident Loop 策略门改造 | [specs/2026-07-17-l2-l3-t05-incident-loop-wiring-design.md](specs/2026-07-17-l2-l3-t05-incident-loop-wiring-design.md) | [plans/2026-07-17-l2-l3-t05-incident-loop-wiring.md](plans/2026-07-17-l2-l3-t05-incident-loop-wiring.md) |
| T06 GCL Runner 生产化 | [specs/2026-07-17-l2-l3-t06-gcl-runner-runtime-design.md](specs/2026-07-17-l2-l3-t06-gcl-runner-runtime-design.md) | [plans/2026-07-17-l2-l3-t06-gcl-runner-runtime.md](plans/2026-07-17-l2-l3-t06-gcl-runner-runtime.md) |
| T07 Trace Schema + Validator | [specs/2026-07-17-l2-l3-t07-trace-schema-validator-design.md](specs/2026-07-17-l2-l3-t07-trace-schema-validator-design.md) | [plans/2026-07-17-l2-l3-t07-trace-schema-validator.md](plans/2026-07-17-l2-l3-t07-trace-schema-validator.md) |
| T08 Eval 覆盖 + Safety 不变量 | [specs/2026-07-17-l2-l3-t08-eval-safety-guard-design.md](specs/2026-07-17-l2-l3-t08-eval-safety-guard-design.md) | [plans/2026-07-17-l2-l3-t08-eval-safety-guard.md](plans/2026-07-17-l2-l3-t08-eval-safety-guard.md) |

## 关联

- 任务卡（plan 形态）：`docs/l2-to-l3-tasks/`
- 上游总览：`docs/l2-to-l3-plan.md`
