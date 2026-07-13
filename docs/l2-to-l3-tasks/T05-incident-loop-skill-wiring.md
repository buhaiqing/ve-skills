# T05 — Incident Loop Skill Wiring (Step 5 改造)

> 任务来源：[`../l2-to-l3-plan.md`](../l2-to-l3-plan.md) §4 (P2)
> 依赖：T01, T02, T03, T04
> 预计工作量：1 天
> 状态：🟡 TODO

## 1. 目标

把"强制 `{{user.confirm}}` 硬门"换成"**策略决策门**"。
ASK 类别仍要求 `{{user.confirm}}`（保持 L2 安全底线），
但 AUTO 类别不再 prompt 用户。

## 2. 背景

- 当前硬门：`incident-loop-agent/SKILL.md:150-154,184`
- 输入数据：T01 策略 + T02 schema + T03 allow-list + T04 leaf metadata
- 输出：`{{policy.decision}} ∈ {AUTO, ASK, REFUSE}`（由 T06 `RunLoop`→`scoreDecision` 在运行时计算并回填，见 T06 §3.2/§3.3）

## 3. 产出物

**修改文件**：`incident-loop-agent/SKILL.md`

### 3.1 必须改的 4 处

| § | 当前行 | 改动 |
|---|-------|------|
| Variable Convention (`:101-117`) | `{{user.confirm}}` 描述 | 新增 `{{policy.decision}}` / `{{policy.reason}}` 变量说明 |
| Step 5 — Confirm (`:150-154`) | "Collect `{{user.confirm}}` for every destructive" | 改为"读 `{{policy.decision}}`；仅 ASK 类收集 `{{user.confirm}}`；AUTO 直接进入 Step 6；REFUSE 退出并记录" |
| Operational Best Practices (`:184`) | "No autopilot for destructive ops" | 改为"No autopilot for non-AUTO class — AUTO 仅适用于 read-only / single-resource state-changing" |
| What This Skill Does (`:198-204`) | "Runs a bounded GCL loop with Safety = 1 floor" | 新增一句"Auto-executes policy-AUTO ops without human prompt" |

### 3.2 必加的 2 处

- Step 5 后**新增** `### Step 5a — Policy evaluation`：
  - 输入：每个 operation 的 (safety_class, blast_radius, confidence, safety)
  - 输出：每 operation 的 decision (AUTO/ASK/REFUSE)
  - 引用：`incident-loop-agent/references/policies/execution-risk.md`
- Reference（`## References` 或新小节）补链：
  - `incident-loop-agent/references/policies/execution-risk.md`
  - `incident-loop-agent/references/policies/domain-allowlist.md`

## 4. DoD

```
□ 1. incident-loop-agent/SKILL.md:150-154 不再含 "Silent default = REFUSE" 硬门语义
□ 2. 新增 Step 5a 描述 policy 评估
□ 3. Variable Convention 新增 {{policy.decision}} / {{policy.reason}}
□ 4. :184 修订为"No autopilot for non-AUTO class"
□ 5. What This Skill Does 加 auto-exec 一句
□ 6. 至少 2 个新的 references 链接存在
□ 7. vet check frontmatter / gcl 仍干净
```

## 5. 验证命令

```bash
# 1. 硬门文字已消失
! grep -q "Silent default = REFUSE" incident-loop-agent/SKILL.md && \
  echo "HARD_GATE_REMOVED_OK"

# 2. 新文字已存在
grep -q "{{policy.decision}}" incident-loop-agent/SKILL.md && \
  echo "POLICY_VAR_OK"
grep -q "No autopilot for non-AUTO" incident-loop-agent/SKILL.md && \
  echo "BEST_PRACTICE_OK"

# 3. Step 5a 小节存在
grep -q "Step 5a" incident-loop-agent/SKILL.md && \
  echo "STEP5A_OK"

# 4. vet 仍干净
cd cmd/vet && go build ./... && go vet ./...
go build -o /tmp/vet . && /tmp/vet check frontmatter --root . && \
  /tmp/vet check gcl --root .
```

## 6. 完成回报

```markdown
## T05 2026-07-XX — done
- 改造 Step 5 为 policy 决策门
- 新增 {{policy.decision}} 变量
- :184 修订为非硬门措辞
- T06 / T08 可消费（policy 现已在 loop 内）
```

## 7. 风险与回滚

- **最大风险**：改错让破坏性操作被 AUTO
  - 缓解：AUTO 仅对 read-only + single-resource state-changing 开放（plan §3.2 表格硬约束）
  - 缓解：destructive 全列 ASK（plan §3.2 表格硬约束）
  - 缓解：T08 safety-guard CI 阻断
- **回滚**：`git checkout incident-loop-agent/SKILL.md`（单文件回滚）
