# AIOps 优化 GCL 评审 Rubric

> Task: 6 AIOps 优化实现（observability, triage, slo/budget, heal, strategy, autonomy）
> MAX_ITER: 3

## R1: Correctness — 功能正确性
| ID | 检查项 | 证据 | 权重 |
|----|--------|------|------|
| R1.1 | 6 个新包全部编译通过 | `go build ./...` 输出 | P0 |
| R1.2 | 6 个新包全部测试通过 | `go test ./internal/{obs,triage,slo,heal,strategy,autonomy}/ -count=1` | P0 |
| R1.3 | 集成后全量测试通过 | `go test ./... -count=1` 33 包 | P0 |
| R1.4 | go vet 干净 | `go vet ./...` 零警告 | P0 |
| R1.5 | 语义分诊正确分类 5 个测试场景 | triage_test.go: CPU/Redis/MySQL/VPC/Kafka | P1 |
| R1.6 | 熔断器 30s 超时恢复 | heal: CircuitBreaker 测试 | P1 |
| R1.7 | 错误预算分级正确 | slo: BudgetGrade 5 边界测试 | P1 |
| R1.8 | 策略图谱 generatePlan 返回排序结果 | strategy: TestGeneratePlan | P1 |

## R2: Safety — 安全性
| ID | 检查项 | 证据 | 权重 |
|----|--------|------|------|
| R2.1 | 无明文凭证/密钥 | 代码审查: grep password|secret|token | P0 |
| R2.2 | 熔断器防止级联失败 | heal: CircuitOpen 测试 | P0 |
| R2.3 | 自愈回滚机制存在 | heal: TestOrchestratorRollback | P0 |
| R2.4 | 知识库查询安全（无注入） | strategy: Query 方法审查 | P1 |

## R3: Idempotency — 幂等性
| ID | 检查项 | 证据 | 权重 |
|----|--------|------|------|
| R3.1 | 预算计算确定性 | slo: TestCalculateBudgetZeroTime | P1 |
| R3.2 | 熔断器状态确定 | heal: TestCircuitBreaker | P1 |
| R3.3 | 分诊对相同输入返回相同结果 | triage: TestTriageClassifier | P1 |

## R4: Traceability — 可追溯性
| ID | 检查项 | 证据 | 权重 |
|----|--------|------|------|
| R4.1 | RunID 贯穿日志 | engine.go: logStep 使用 runID | P0 |
| R4.2 | 链路追踪 span 创建 | engine.go: NewRootTrace + StartSpan | P1 |
| R4.3 | Span.End 记录状态 | engine.go: span.End(err) | P1 |

## R5: Spec Compliance — 规格符合
| ID | 检查项 | 证据 | 权重 |
|----|--------|------|------|
| R5.1 | 6 个新 internal 包与 SDD 一致 | 对比 design.md §4 | P0 |
| R5.2 | 类型签名与 spec 一致 | TraceContext, ErrorBudget, CircuitBreaker | P0 |
| R5.3 | 零新外部依赖 | go.mod 无新增 | P1 |
| R5.4 | 上下文传播使用 context | WithTrace/FromContext | P1 |

---

**评审规则:**
- P0 项任一 fail → BLOCKER，必须修复
- P1 项累计 3+ fail → 必须修复
- Critic 只读不改，输出必须带 rubric_id
- 超 MAX_ITER=3 仍 fail → escalate
