# Prompts Handbook — ve-billing-ops

> **Purpose:** Common prompt templates and response patterns for AI agent interactions. Based on AIOps best practices for billing operations.
> **Version:** 1.3.0>
> **Last Updated:** 2026-06-02

---

## Table of Contents

1. [Query Prompts](#1-query-prompts)
2. [Analysis Prompts](#2-analysis-prompts)
3. [Action Prompts](#3-action-prompts)
4. [Report Prompts](#4-report-prompts)
5. [Troubleshooting Prompts](#5-troubleshooting-prompts)
6. [Response Templates](#6-response-templates)

---

## 1. Query Prompts

### 1.1 Monthly Bill Query

**User Prompt:**
```
查询本月账单
```

**Agent Response Pattern:**
```markdown
## 月度账单查询结果

**账单周期:** {{output.billing_period}}
**总费用:** ¥{{output.total_cost}}

### 费用构成
| 产品 | 费用 | 占比 |
|------|------|------|
{{#each output.product_breakdown}}
| {{product}} | ¥{{cost}} | {{percent}}% |
{{/each}}

### Top 5 成本来源
| 排名 | 资源ID | 产品 | 费用 |
|------|--------|------|------|
{{#each output.top_cost_drivers}}
| {{rank}} | {{resource_id}} | {{product_type}} | ¥{{cost}} |
{{/each}}

[Suggest: 分析成本趋势 / 生成优化建议]
```

---

### 1.2 Balance Check

**User Prompt:**
```
检查账户余额
```

**Agent Response Pattern:**
```markdown
## 账户余额状态

**当前余额:** ¥{{output.balance}}
**币种:** {{output.currency}}
**信用额度:** ¥{{output.credit_limit}} ({{#if output.has_credit}}有{{else}}无{{/if}})

### 消耗预估
| 指标 | 数值 |
|------|------|
| 日均消耗 | ¥{{output.daily_burn_rate}} |
| 预计剩余天数 | {{output.days_remaining}} 天 |
| 月度预测 | ¥{{output.projected_monthly}} |

{{#if output.low_balance}}
⚠️ **余额预警:** 余额可能在 {{output.days_remaining}} 天后耗尽，建议及时充值。
{{/if}}
```

---

### 1.3 Budget Status

**User Prompt:**
```
查看预算状态
```

**Agent Response Pattern:**
```markdown
## 预算状态概览

| 预算名称 | 限额 | 已用 | 剩余 | 使用率 |
|----------|------|------|------|--------|
{{#each output.budgets}}
| {{name}} | ¥{{limit}} | ¥{{spent}} | ¥{{remaining}} | {{utilization}}% |
{{/each}}

### 详细状态
{{#each output.budgets}}
#### {{name}}
- 📊 当前: ¥{{spent}} / ¥{{limit}} ({{utilization}}%)
- 📈 日均消耗: ¥{{daily_burn}}
- 🎯 预计月末: ¥{{projected}}
{{/each}}

{{#if output.any_alert}}
🚨 **告警状态:**
{{#each output.triggered_alerts}}
- {{budget_name}}: 已触发 {{threshold}}% 阈值
{{/each}}
{{/if}}
```

---

### 1.4 RI Utilization

**User Prompt:**
```
查看预留实例使用情况
```

**Agent Response Pattern:**
```markdown
## 预留实例 (RI) 使用分析

### 总体指标
| 指标 | 数值 |
|------|------|
| RI 总数 | {{output.total_ri_count}} |
| 总单元数 | {{output.total_units}} |
| 已使用单元 | {{output.used_units}} |
| 整体利用率 | {{output.overall_utilization}}% |
| 预计节省 | ¥{{output.savings}} |

### 按类型分布
| 实例类型 | 总量 | 已用 | 利用率 | 状态 |
|----------|------|------|--------|------|
{{#each output.ri_by_type}}
| {{instance_type}} | {{total}} | {{used}} | {{utilization}}% | {{status}} |
{{/each}}

### 即将过期
{{#if output.expiring_soon}}
| 实例类型 | 过期时间 | 剩余天数 | 建议 |
|----------|----------|----------|------|
{{#each output.expiring_soon}}
| {{instance_type}} | {{expire_time}} | {{days_left}} 天 | {{recommendation}} |
{{/each}}
{{else}}
✅ 无即将过期的预留实例
{{/if}}
```

---

## 2. Analysis Prompts

### 2.1 Cost Trend Analysis

**User Prompt:**
```
分析本月的成本趋势
```

**Agent Response Pattern:**
```markdown
## 成本趋势分析报告

### 周期对比
| 月份 | 总费用 | MoM 变化 | 变化率 |
|------|--------|----------|--------|
{{#each output.monthly_comparison}}
| {{month}} | ¥{{cost}} | {{#if change}}({{change_sign}}{{change}}){{else}}—{{/if}} | {{change_rate}}% |
{{/each}}

### 趋势指标
| 指标 | 数值 | 状态 |
|------|------|------|
| 环比增长率 | {{output.mom_growth}}% | {{output.mom_status}} |
| 日均消耗 | ¥{{output.daily_avg}} | {{output.daily_status}} |
| 月度预测 | ¥{{output.projection}} | {{output.projection_status}} |
| vs 预算 | {{output.vs_budget}}% | {{output.budget_status}} |

### 趋势图表
```
{{output.trend_chart}}
```

### 关键变化点
{{#each output.key_changes}}
- **{{month}}:** {{description}} ({{impact}} 影响)
{{/each}}

### 建议
{{#each output.recommendations}}
- **{{priority}}:** {{suggestion}} (预计节省 ¥{{est_savings}}/月)
{{/each}}
```

---

### 2.2 Cost Driver Analysis

**User Prompt:**
```
找出成本最高的资源
```

**Agent Response Pattern:**
```markdown
## 成本驱动因素分析

### Top 10 成本资源
| 排名 | 资源ID | 产品类型 | 实例规格 | 费用 | 费用占比 | 趋势 |
|------|--------|----------|----------|------|----------|------|
{{#each output.top_cost_drivers}}
| {{rank}} | `{{resource_id}}` | {{product_type}} | {{spec}} | ¥{{cost}} | {{percent}}% | {{trend}} |
{{/each}}

### 成本分布
```
{{output.distribution_chart}}
```

### 产品维度
| 产品 | 费用 | 占比 | 环比变化 |
|------|------|------|----------|
{{#each output.by_product}}
| {{product}} | ¥{{cost}} | {{percent}}% | {{mom_change}}% |
{{/each}}

### 环境维度
| 环境 | 费用 | 占比 | 标签覆盖率 |
|--------|------|------|-----------|
{{#each output.by_environment}}
| {{environment}} | ¥{{cost}} | {{percent}}% | {{tag_coverage}}% |
{{/each}}

### 成本中心维度
| 成本中心 | 费用 | 占比 | 负责人 |
|----------|------|------|--------|
{{#each output.by_cost_center}}
| {{cost_center}} | ¥{{cost}} | {{percent}}% | {{owner}} |
{{/each}}
```

---

### 2.3 Anomaly Detection

**User Prompt:**
```
检测账单异常
```

**Agent Response Pattern:**
```markdown
## 成本异常检测报告

### 检测结果
| 类型 | 严重程度 | 描述 |
|------|----------|------|
{{#each output.anomalies}}
| {{type}} | {{severity}} | {{description}} |
{{/each}}

### 详细分析

{{#each output.anomalies}}
#### 异常 {{anomaly_id}}: {{type}}
- **检测时间:** {{detected_at}}
- **当前值:** ¥{{current_value}}
- **基线值:** ¥{{baseline_value}}
- **偏差:** {{deviation}}% (阈值: {{threshold}}%)
- **可信度:** {{confidence}}%

**可能原因:**
{{#each possible_causes}}
- {{.}}
{{/each}}

**建议操作:**
{{#each recommended_actions}}
- {{.}}
{{/each}}
{{/each}}

### 历史对比
```
{{output.historical_comparison}}
```
```

---

## 3. Action Prompts

### 3.1 Create Budget

**User Prompt:**
```
创建一个月度预算，限额10万元，告警阈值80%、90%、100%
```

**Agent Response Pattern:**
```markdown
## 创建预算

### 输入确认
| 参数 | 值 |
|------|-----|
| 预算名称 | {{user.budget_name}} |
| 预算类型 | MONTHLY |
| 限额 | ¥{{user.budget_amount}} |
| 告警阈值 | 80%, 90%, 100% |

### 执行中...
```bash
ve billing CreateBudget \
  --BudgetName "{{user.budget_name}}" \
  --BudgetAmount {{user.budget_amount}} \
  --BudgetType MONTHLY \
  --AlertThresholds '[80, 90, 100]' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

### 执行结果
✅ **预算创建成功**

| 字段 | 值 |
|------|-----|
| Budget ID | `{{output.budget_id}}` |
| 名称 | {{user.budget_name}} |
| 类型 | MONTHLY |
| 限额 | ¥{{user.budget_amount}} |
| 阈值 | 80%, 90%, 100% |

### 下一步
- [查看预算状态](DescribeBudgets)
- [修改告警阈值](ModifyBudget)
- [配置告警通知](observability.md#webhook)
```

---

### 3.2 Cost Optimization Request

**User Prompt:**
```
帮我优化成本，有什么可以节省的地方？
```

**Agent Response Pattern:**
```markdown
## 成本优化建议报告

### 执行摘要
| 类别 | 数量 | 预计总节省 |
|------|------|------------|
| Right-sizing | {{output.right_sizing_count}} | ¥{{output.right_sizing_savings}} |
| 闲置资源清理 | {{output.idle_count}} | ¥{{output.idle_savings}} |
| 计费模式优化 | {{output.billing_mode_count}} | ¥{{output.billing_mode_savings}} |
| **总计** | — | **¥{{output.total_savings}}** |

### Right-Sizing 建议
| 资源ID | 当前规格 | 利用率 (CPU/Mem) | 建议规格 | 预计节省 |
|--------|----------|------------------|----------|----------|
{{#each output.right_sizing_recs}}
| `{{resource_id}}` | {{current_type}} | {{avg_util}}% | → {{target_type}} | ¥{{savings}}/月 |
{{/each}}

### 闲置资源清理
| 资源ID | 类型 | 闲置时长 | 月费用 | 建议操作 |
|--------|------|----------|--------|----------|
{{#each output.idle_resources}}
| `{{resource_id}}` | {{type}} | {{idle_days}} 天 | ¥{{cost}} | {{action}} |
{{/each}}

### 计费模式优化
| 资源 | 当前模式 | 建议模式 | 预计节省 | 风险 |
|------|----------|----------|----------|------|
{{#each output.billing_mode_recs}}
| `{{resource_id}}` | {{current_mode}} | → {{target_mode}} | ¥{{savings}}/月 | {{risk}} |
{{/each}}

### 优化优先级
| 优先级 | 行动 | 预计节省 | 复杂度 | 风险 |
|--------|------|----------|--------|------|
{{#each output.prioritized_actions}}
| {{priority}} | {{action}} | ¥{{savings}} | {{complexity}} | {{risk}} |
{{/each}}

### 待批准操作
{{#if output.requires_approval}}
以下操作需要人工确认:
{{#each output.requires_approval}}
- **{{action}}**: {{resource_id}} ({{reason}})
{{/each}}

请回复 "确认" 继续执行，或指定要执行的操作。
{{/if}}
```

---

## 4. Report Prompts

### 4.1 Daily Report

**User Prompt:**
```
生成每日账单报告
```

**Agent Response Pattern:**
```markdown
# 📊 每日账单报告 — {{output.date}}

**生成时间:** {{output.generated_at}}
**报告周期:** {{output.period}}

---

## 状态概览

| 指标 | 状态 | 详情 |
|------|------|------|
| 预算使用 | {{output.budget_status}} | {{output.budget_utilization}}% |
| 账户余额 | {{output.balance_status}} | ¥{{output.balance}} ({{output.days_remaining}} 天) |
| 异常告警 | {{output.alert_status}} | {{output.alert_count}} 条 |
| RI 覆盖率 | {{output.ri_status}} | {{output.ri_coverage}}% |

## 今日成本
| 指标 | 数值 |
|------|------|
| 今日费用 | ¥{{output.today_cost}} |
| 昨日费用 | ¥{{output.yesterday_cost}} |
| 环比变化 | {{output.day_over_day_change}}% |

## Top 5 今日成本资源
| 资源 | 费用 | 产品 |
|------|------|------|
{{#each output.top_today}}
| {{resource_id}} | ¥{{cost}} | {{product}} |
{{/each}}

## 告警列表
{{#if output.alerts}}
| 告警类型 | 严重程度 | 详情 | 时间 |
|----------|----------|------|------|
{{#each output.alerts}}
| {{type}} | {{severity}} | {{details}} | {{time}} |
{{/each}}
{{else}}
✅ 暂无告警
{{/if}}

## 建议行动
{{#each output.action_items}}
- **{{priority}}:** {{action}} ({{owner}})
{{/each}}

---
*此报告由 ve-billing-ops 自动生成*
```

---

### 4.2 Weekly Report

**User Prompt:**
```
生成周报告
```

**Agent Response Pattern:**
```markdown
# 📈 周账单报告 — 第 {{output.week_number}} 周

**报告周期:** {{output.start_date}} ~ {{output.end_date}}
**生成时间:** {{output.generated_at}}

---

## 执行摘要

### 成本概览
| 指标 | 本周 | 上周 | 环比 |
|------|------|------|------|
| 总费用 | ¥{{output.week_total}} | ¥{{output.last_week_total}} | {{output.mom_change}}% |
| 日均费用 | ¥{{output.daily_avg}} | ¥{{output.last_daily_avg}} | {{output.daily_change}}% |
| 峰值日 | {{output.peak_day}} | ¥{{output.peak_cost}} | — |
| 谷值日 | {{output.trough_day}} | ¥{{output.trough_cost}} | — |

### 预算进度
| 预算 | 限额 | 已用 | 使用率 | 预计月末 |
|------|------|------|--------|----------|
| {{budget_name}} | ¥{{budget_limit}} | ¥{{budget_spent}} | {{utilization}}% | ¥{{projected}} |

### RI 覆盖率
| 类型 | 覆盖率 | 状态 |
|------|--------|------|
| 总体 | {{output.overall_ri_coverage}}% | {{output.ri_status}} |
{{#each output.by_ri_type}}
| {{instance_type}} | {{coverage}}% | {{status}} |
{{/each}}

## 成本趋势
```
{{output.weekly_chart}}
```

### 每日明细
| 日期 | 费用 | 累计 | vs 预算日均 |
|------|------|------|-------------|
{{#each output.daily_breakdown}}
| {{date}} | ¥{{cost}} | ¥{{cumulative}} | {{vs_daily_budget}}% |
{{/each}}

## 产品成本分布
| 产品 | 本周费用 | 占比 | 环比变化 |
|------|----------|------|----------|
{{#each output.by_product}}
| {{product}} | ¥{{cost}} | {{percent}}% | {{change}}% |
{{/each}}

## Top 5 成本资源
| 排名 | 资源 | 费用 | 环比变化 | 趋势 |
|------|------|------|----------|------|
{{#each output.top_cost_resources}}
| {{rank}} | {{resource_id}} | ¥{{cost}} | {{change}}% | {{trend}} |
{{/each}}

## 异常检测
{{#if output.anomalies}}
| 异常类型 | 描述 | 严重程度 | 建议 |
|----------|------|----------|------|
{{#each output.anomalies}}
| {{type}} | {{description}} | {{severity}} | {{recommendation}} |
{{/each}}
{{else}}
✅ 本周无异常
{{/if}}

## 优化机会
| 类型 | 数量 | 预计节省 | 可执行 |
|----------|------|----------|--------|
| Right-sizing | {{output.right_sizing_count}} | ¥{{output.right_sizing_savings}} | {{output.right_sizing_executable}} |
| 闲置清理 | {{output.idle_count}} | ¥{{output.idle_savings}} | {{output.idle_executable}} |
| RI 优化 | {{output.ri_count}} | ¥{{output.ri_savings}} | {{output.ri_executable}} |

## 行动项
| 优先级 | 行动 | 负责人 | 截止日期 |
|--------|------|--------|----------|
{{#each output.action_items}}
| {{priority}} | {{action}} | {{owner}} | {{due_date}} |
{{/each}}

---

## 下周预测
| 指标 | 预测值 | 置信度 |
|------|--------|--------|
| 预计总费用 | ¥{{output.next_week_forecast}} | {{output.confidence}}% |
| 预计峰值 | ¥{{output.next_week_peak}} | — |
| 预算达标率 | {{output.budget_achievability}}% | — |

---
*此报告由 ve-billing-ops 自动生成*
```

---

### 4.3 Monthly Report

**User Prompt:**
```
生成月度报告
```

**Agent Response Pattern:**
```markdown
# 📅 月度账单报告 — {{output.month}}

**报告周期:** {{output.period}}
**生成时间:** {{output.generated_at}}

---

## 执行摘要

### 关键指标
| 指标 | 本月 | 上月 | 环比 | vs 预算 |
|------|------|------|------|---------|
| 总费用 | ¥{{output.total_cost}} | ¥{{output.last_month}} | {{output.mom_change}}% | {{output.vs_budget}}% |
| 日均费用 | ¥{{output.daily_avg}} | ¥{{output.last_daily_avg}} | — | — |
| 峰值日 | {{output.peak_day}} | — | — | — |
| 预算使用 | {{output.budget_utilization}}% | — | — | — |

### 财务状况
| 状态 | 数值 |
|------|------|
| 期末余额 | ¥{{output.ending_balance}} |
| 本月支出 | ¥{{output.total_cost}} |
| 充值 | ¥{{output.recharges}} |
| 退款 | ¥{{output.refunds}} |

## 6 个月趋势
| 月份 | 费用 | MoM | vs 预算 | 累计偏差 |
|------|------|-----|---------|----------|
{{#each output.six_month_trend}}
| {{month}} | ¥{{cost}} | {{mom}}% | {{vs_budget}}% | {{cumulative_deviation}}% |
{{/each}}

## 成本分布

### 按产品
| 产品 | 费用 | 占比 | 环比变化 |
|------|------|------|----------|
{{#each output.by_product}}
| {{product}} | ¥{{cost}} | {{percent}}% | {{change}}% |
{{/each}}

### 按环境
| 环境 | 费用 | 占比 | 标签覆盖率 |
|--------|------|------|-----------|
{{#each output.by_environment}}
| {{environment}} | ¥{{cost}} | {{percent}}% | {{tag_coverage}}% |
{{/each}}

### 按成本中心
| 成本中心 | 费用 | 占比 | 预算执行率 |
|----------|------|------|------------|
{{#each output.by_cost_center}}
| {{cost_center}} | ¥{{cost}} | {{percent}}% | {{execution_rate}}% |
{{/each}}

## RI 分析
| 指标 | 数值 |
|------|------|
| RI 总投资 | ¥{{output.ri_investment}} |
| 覆盖资源成本 | ¥{{output.ri_covered_cost}} |
| RI 覆盖率 | {{output.ri_coverage}}% |
| 实际节省 | ¥{{output.actual_savings}} |
| ROI | {{output.ri_roi}}% |

### RI 续费建议
{{#each output.ri_renewal_suggestions}}
- **{{instance_type}}:** {{recommendation}} (到期: {{expire_date}})
{{/each}}

## 优化成果

### 本月已执行
| 优化类型 | 数量 | 实际节省 |
|----------|------|----------|
| Right-sizing | {{output.right_sizing_done}} | ¥{{output.right_sizing_savings}} |
| 资源清理 | {{output.cleanup_done}} | ¥{{output.cleanup_savings}} |
| 计费模式 | {{output.billing_mode_done}} | ¥{{output.billing_mode_savings}} |
| **总计** | — | **¥{{output.total_savings_achieved}}** |

### 未执行优化机会
| 优化类型 | 数量 | 潜在节省 | 优先级 |
|----------|------|----------|--------|
| Right-sizing | {{output.right_sizing_pending}} | ¥{{output.right_sizing_pending_savings}} | P1 |
| 闲置清理 | {{output.idle_pending}} | ¥{{output.idle_pending_savings}} | P2 |
| 计费模式 | {{output.billing_mode_pending}} | ¥{{output.billing_mode_pending_savings}} | P2 |

## 成本异常事件
{{#if output.anomaly_events}}
| 事件 | 日期 | 影响 | 原因 | 处理 |
|------|------|------|------|------|
{{#each output.anomaly_events}}
| {{event}} | {{date}} | ¥{{impact}} | {{cause}} | {{resolution}} |
{{/each}}
{{else}}
✅ 本月无重大异常事件
{{/if}}

## 下月展望
| 指标 | 预测值 | 置信度 |
|------|--------|--------|
| 预计费用 | ¥{{output.next_month_forecast}} | {{output.confidence}}% |
| 建议预算 | ¥{{output.recommended_budget}} | — |
| 预计节省 | ¥{{output.next_month_potential_savings}} | — |

## 行动项
| 优先级 | 行动 | 负责人 | 截止日期 |
|--------|------|--------|----------|
{{#each output.action_items}}
| {{priority}} | {{action}} | {{owner}} | {{due_date}} |
{{/each}}

## 签批
| 角色 | 姓名 | 日期 | 签名 |
|------|------|------|------|
| FinOps 工程师 | | | |
| 财务经理 | | | |
| DevOps 负责人 | | | |

---
*此报告由 ve-billing-ops 自动生成*
```

---

## 5. Troubleshooting Prompts

### 5.1 Cost Spike Investigation

**User Prompt:**
```
为什么这个月费用突然增加了50%？
```

**Agent Response Pattern:**
```markdown
## 费用异常增长调查

### 基础信息
| 指标 | 数值 |
|------|------|
| 本月费用 | ¥{{output.current_cost}} |
| 上月费用 | ¥{{output.previous_cost}} |
| 增长金额 | ¥{{output.increase}} |
| 增长率 | {{output.growth_rate}}% |

### 时间线分析
| 阶段 | 日期范围 | 费用 | 特征 |
|------|----------|------|------|
{{#each output.timeline}}
| {{phase}} | {{date_range}} | ¥{{cost}} | {{characteristic}} |
{{/each}}

### 产品维度分析
| 产品 | 上月 | 本月 | 增长 | 贡献度 |
|------|------|------|------|--------|
{{#each output.by_product}}
| {{product}} | ¥{{last_month}} | ¥{{current}} | {{growth}}% | {{contribution}}% |
{{/each}}

### Top 5 增长来源
| 排名 | 资源ID | 产品 | 增长费用 | 增长原因 |
|------|--------|------|----------|----------|
{{#each output.top_increases}}
| {{rank}} | `{{resource_id}}` | {{product}} | +¥{{increase}} | {{reason}} |
{{/each}}

### 可能原因
{{#each output.possible_causes}}
1. **{{.}}** {{#if evidence}}(证据: {{evidence}}){{/if}}
{{/each}}

### 根本原因分析
{{output.root_cause_analysis}}

### 建议措施
| 优先级 | 措施 | 预计效果 | 实施难度 |
|--------|------|----------|----------|
{{#each output.recommended_actions}}
| {{priority}} | {{action}} | {{effect}} | {{difficulty}} |
{{/each}}

### 预防措施
- {{#each output.preventive_measures}}
- {{.}}
{{/each}}
```

---

### 5.2 Budget Exceeded Resolution

**User Prompt:**
```
预算超了怎么办？
```

**Agent Response Pattern:**
```markdown
## 预算超支处理方案

### 当前状态
| 指标 | 数值 |
|------|------|
| 预算名称 | {{output.budget_name}} |
| 预算限额 | ¥{{output.budget_limit}} |
| 当前使用 | ¥{{output.current_spent}} |
| 超支金额 | ¥{{output.overage}} |
| 超支比例 | {{output.overage_percent}}% |
| 预计月末 | ¥{{output.projected_monthly}} |

### 立即行动 (Day 1)
| 优先级 | 行动 | 预计节省 | 执行人 |
|--------|------|----------|--------|
{{#each output.immediate_actions}}
| {{priority}} | {{action}} | ¥{{savings}} | {{owner}} |
{{/each}}

### 短期措施 (Week 1)
| 行动 | 预计节省 | 影响 |
|------|----------|------|
{{#each output.short_term_actions}}
| {{action}} | ¥{{savings}} | {{impact}} |
{{/each}}

### 预算调整建议
| 选项 | 说明 | 优点 | 缺点 |
|------|------|------|------|
{{#each output.budget_options}}
| {{option}} | {{description}} | {{pros}} | {{cons}} |
{{/each}}

### 建议方案
{{output.recommended_solution}}

### 后续跟进
| 时间点 | 行动 | 负责人 |
|--------|------|--------|
| 今天 | 冻结非关键资源采购 | @{{today_owner}} |
| 明天 | 完成资源清理 | @{{tomorrow_owner}} |
| 本周 | 评估预算调整需求 | @{{week_owner}} |
| 下周 | 回顾执行效果 | @{{next_week_owner}} |
```

---

## 6. Response Templates

### 6.1 Success Response

```markdown
✅ **操作成功**

{{output.success_message}}

| 结果 | 详情 |
|------|------|
{{#each output.result_fields}}
| {{key}} | {{value}} |
{{/each}}

### 下一步操作
- [查看详情](#)
- [执行相关操作](#)
- [返回主菜单](#)
```

### 6.2 Error Response

```markdown
❌ **操作失败**

**错误代码:** `{{error.code}}`
**错误信息:** {{error.message}}

### 问题描述
{{error.description}}

### 解决方法
{{#each error.resolution_steps}}
1. {{.}}
{{/each}}

### 建议下一步
{{error.next_steps}}
```

### 6.3 Confirmation Response

```markdown
⚠️ **操作确认**

您即将执行以下操作：
- **操作:** {{operation_name}}
- **资源:** {{resource_id}}
- **影响:** {{impact_description}}

此操作{{is_reversible}}.

请输入 `confirm` 确认执行，或 `cancel` 取消：
```

### 6.4 Partial Success Response

```markdown
⚠️ **部分成功**

成功: {{success_count}} 项
失败: {{failure_count}} 项

### 成功项目
{{#each output.success_items}}
- ✅ {{.}}
{{/each}}

### 失败项目
{{#each output.failed_items}}
- ❌ **{{.}}** — {{error_reason}}
{{/each}}

### 失败项目重试
是否重试失败的项目？(yes/no)
```