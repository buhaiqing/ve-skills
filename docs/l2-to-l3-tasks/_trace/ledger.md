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

## T02 2026-07-13 — done

- 交付：`incident-loop-agent/assets/execution-risk.schema.json`（≤3KB, draft 2020-12）
- 9-cell 决策矩阵 + Safety=0→REFUSE（if/then 硬规则）+ 缺元数据→ASK（fail-safe）
- 与 `execution-risk.md` prose 对齐；jsonschema 校验 9 cell 全 PASS、Safety=0 必 REFUSE、缺字段必 ASK
- 提交 dd0ee22（feature/t02-execution-risk-schema → main）
- T05 可解锁（schema 已就绪，T06 scorer 可消费）

## T03 2026-07-13 — done

- 交付：`incident-loop-agent/references/policies/domain-allowlist.md`（48 行）
- 8 协调 skill 1:1 对应 `coordinates`；每 skill 症状白名单；显式排除 destructive op；扩展策略（≥10 traces + 0 incident + 30d）
- 提交 150a477（feature/t03-domain-allowlist → main）
- T05 可解锁

## T04 2026-07-13 — done

- 交付：`ve-skill-generator/references/leaf-op-metadata-spec.md` + 8 leaf SKILL.md 操作表加 `safety_class`/`blast_radius` 两列
- 8 skill 操作表头均含两列，操作行数与加标行数一致（cms 12 / ecs 24 / rds 18 / redis 16 / vpc 11 / iam 23 / kms 17 / billing 25）
- 保守分类：destructive/read-only 按关键词，其余 state-changing；blast_radius 默认 single
- `vet check frontmatter/aiops/assessment` 全绿
- 提交 02a7290（feature/t04-leaf-op-metadata → main）
- T05 可解锁


## T05 2026-07-13 — done

- 改造 Step 5 为 policy 决策门（移除 "Silent default = REFUSE" 硬门）
- 新增 `{{policy.decision}}` / `{{policy.reason}}` 变量声明（Variable Convention）
- 新增 Step 5a — Policy evaluation 小节（描述 T06 scorer 计算逻辑）
- :184 修订为 "No autopilot for non-AUTO class"（非硬门措辞）
- What This Skill Does 新增 auto-exec 一句
- version bump: 0.1.0 → 0.3.1，Changelog 追加 0.3.1 条目
- References 已存在 `execution-risk.md` / `domain-allowlist.md`（T01/T03 已加，无需重复）
- vet check frontmatter (29/29) + gcl (30/30) 全绿
- T06 / T08 可消费（policy 现已在 loop 内）

## T08 2026-07-13 — done

- eval_queries.json 扩至 16 case（9 覆盖 3 类 AUTO/ASK/REFUSE × 场景），每 case 含 expected_decision
- cmd/vet/internal/check/policyguard/ 新建（含 policyguard.go + policyguard_test.go + testdata/ + README.md）
- policyguard.Check/CheckPlan/CheckPlanWithReport/ComputeDecision 4 函数；3 条不变量：safety=0→REFUSE、destructive→not AUTO、missing meta→not AUTO
- cmd/vet/check.go 注册 policyguard 子命令
- go build/vet/test 全绿；vet check policyguard 全绿
- .github/workflows/validate.yml 不存在，跳过 CI 相关 DoD

## T06 2026-07-13 — done

- incident-loop-agent/SKILL.md + rubric.md 的 scripts/gcl_runner.py 全部替换为 vet gcl run ✅（2026-07-17: scripts/ 已删除）
- cmd/vet/internal/gcl/run/run.go 新增 allowedSkills + OpDecision + scoreDecision（含 9 格矩阵）
- scoreDecision 单测：TestScoreDecision_9Cell / TestScoreDecision_DestructiveNeverAuto / TestScoreDecision_SafetyZeroRefuse 全 PASS
- go build/vet/test 全绿；vet check frontmatter/gcl/eval/policyguard 全绿
- .github/workflows/ 不存在，跳过 CI 相关 DoD

## T07 2026-07-13 — done

- incident-loop-agent/assets/trace.schema.json 新建（必填 ticket_id / started_at / finished_at / policy_decision / redaction_pass + iterations[].ve_calls[].request_id）
- cmd/vet/internal/check/trace/trace.go 新建（含 Check / CheckDir，含 3 个 test fixture）
- cmd/vet/check.go 注册 trace 子命令；traceCheck 只检查 incident-trace-*.json（跳过 legacy gcl-trace-*.json）
- go build/vet/test 全绿；vet check trace 全绿（6 条 incident-trace 通过）
- gcl-spec.md §9 / §12 / workflow 无需修改（schema 由 T06 DoD #9 覆盖）
