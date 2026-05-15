# CLI Usage — CMS

## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md) for CLI installation.

**CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`.

## Conventions

- Output is **JSON by default**
- Use `--help` for any action: `ve metrics <Action> --help`

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| GetMetricData | Yes | Primary query method; JSON output |
| ListMetrics | Yes | List available metrics per namespace |
| PutResourceMetricRule | Yes | Create/update alarm rule |
| EnableMetricRule | Yes | Enable alarm rule |
| DisableMetricRule | Yes | Disable alarm rule |
| DeleteMetricRules | Yes | Delete alarm rules |
| DescribeMetricRuleList | Yes | List alarm rules |
| PutMetricRuleTemplate | Yes | Create alarm template |
| DescribeMetricRuleTemplate | Yes | Query alarm template |
| ApplyMetricRuleTemplate | Yes | Apply template to resources |
| DescribeContactGroups | Yes | List notification groups |
| CreateContactGroup | Yes | Create notification group |
| DeleteContactGroup | Yes | Delete notification group |

## Command Map

| Goal | Example `ve` Invocation | Notes |
|------|------------------------|-------|
| Query metric data | `ve metrics GetMetricData --Namespace Volcengine_ECS --MetricName CpuUtilization --Dimensions '[{"InstanceId":"i-xxx"}]' --StartTime 1715000000000 --EndTime 1715003600000 --Period 60` | JSON output |
| List available metrics | `ve metrics ListMetrics --Namespace Volcengine_ECS` | |
| Create alarm rule | `ve metrics PutResourceMetricRule --RuleName cpu-alarm --Namespace Volcengine_ECS --MetricName CpuUtilization --AlertState Critical --ComparisonOperator GreaterThanThreshold --Statistics Average --Threshold 90 --Times 3 --Period 60 --NotifyType "DefaultGroup"` | |
| Enable alarm | `ve metrics EnableMetricRule --RuleIds '["rule-xxx"]'` | |
| Disable alarm | `ve metrics DisableMetricRule --RuleIds '["rule-xxx"]'` | |
| Delete alarm | `ve metrics DeleteMetricRules --RuleIds '["rule-xxx"]'` | |
| List alarm rules | `ve metrics DescribeMetricRuleList --Namespace Volcengine_ECS` | |
| Create alarm template | `ve metrics PutMetricRuleTemplate --TemplateName ecs-template --Namespace Volcengine_ECS --Rules '[{"MetricName":"CpuUtilization","Threshold":90}]'` | |
| List notification groups | `ve metrics DescribeContactGroups` | |
| Create notification group | `ve metrics CreateContactGroup --ContactGroupName ops-team --ContactType sms --ContactValue "13800000000"` | |
