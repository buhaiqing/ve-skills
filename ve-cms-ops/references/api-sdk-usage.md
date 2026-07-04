# API & SDK — CMS

## OpenAPI

- **API Version:** 2018-03-14
- **Endpoint:** `open.volcengineapi.com` (or `monitor.volcengineapi.com`)
- **Documentation:** https://www.volcengine.com/docs/6408/78941

## Go SDK

**Package:** `github.com/volcengine/volc-sdk-golang`
**Minimum Go version:** 1.14

## SDK Operations Map

| Goal | API Action | SDK Method Pattern |
|------|-----------|-------------------|
| Query metric data | `GetMetricData` | `client.Get("metrics_v2", params)` |
| List metrics meta | `ListMetrics` | `client.Get("metrics_v2", params)` |
| Create/update alarm rule | `PutResourceMetricRule` | `client.Post("metrics_v2", params, body)` |
| Delete alarm rules | `DeleteMetricRules` | `client.Post("metrics_v2", params, body)` |
| List alarm rules | `DescribeMetricRuleList` | `client.Get("metrics_v2", params)` |
| Enable alarm rule | `EnableMetricRule` | `client.Post("metrics_v2", params, body)` |
| Disable alarm rule | `DisableMetricRule` | `client.Post("metrics_v2", params, body)` |
| Create alarm template | `PutMetricRuleTemplate` | `client.Post("metrics_v2", params, body)` |
| Describe alarm template | `DescribeMetricRuleTemplate` | `client.Get("metrics_v2", params)` |
| Apply alarm template | `ApplyMetricRuleTemplate` | `client.Post("metrics_v2", params, body)` |
| List contact groups | `DescribeContactGroups` | `client.Get("volc_content_platform", params)` |
| Create contact group | `CreateContactGroup` | `client.Post("volc_content_platform", params, body)` |

## GetMetricData Response

```json
{
  "ResponseMetadata": {
    "RequestId": "request-id-string",
    "Action": "GetMetricData",
    "Version": "2018-03-14",
    "Service": "metrics_v2"
  },
  "Result": {
    "MetricName": "CpuUtilization",
    "Namespace": "Volcengine_ECS",
    "Datapoints": [
      {"Timestamp": 1715000000000, "Value": 85.5},
      {"Timestamp": 1715000300000, "Value": 87.2}
    ],
    "Period": 300
  }
}
```

## Go SDK Examples

```go
// Client boilerplate (reuse in each function):
//   c := base.NewClient(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
//   c.SetHost("open.volcengineapi.com")
```

### GetMetricData

```go
func main() {
    c := base.NewClient(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    c.SetHost("open.volcengineapi.com")
    params := map[string]string{
        "Action": "GetMetricData", "Version": "2018-03-14",
        "Namespace": "Volcengine_ECS", "MetricName": "CpuUtilization",
        "Dimensions": `[{"InstanceId":"i-xxx"}]`,
        "StartTime": "1715000000000", "EndTime": "1715003600000", "Period": "60",
    }
    resp, err := c.Get("metrics_v2", params)
    if err != nil { panic(err) }
    fmt.Println(string(resp))
}
```

### PutResourceMetricRule

```go
func createAlarm(name, ns, metric string, threshold float64) {
    c := base.NewClient(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    c.SetHost("open.volcengineapi.com")
    params := map[string]string{
        "Action": "PutResourceMetricRule", "Version": "2018-03-14",
        "RuleName": name, "Namespace": ns, "MetricName": metric,
        "AlertState": "Critical", "ComparisonOperator": "GreaterThanThreshold",
        "Statistics": "Average", "Threshold": fmt.Sprintf("%.0f", threshold),
        "Times": "3", "Period": "60", "NotifyType": "DefaultGroup",
    }
    resp, err := c.Get("metrics_v2", params)
    if err != nil { panic(err) }
    fmt.Println(string(resp))
}
```

### DescribeMetricRuleList

```go
func listAlarms() {
    c := base.NewClient(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    c.SetHost("open.volcengineapi.com")
    params := map[string]string{
        "Action": "DescribeMetricRuleList", "Version": "2018-03-14",
        "Namespace": "Volcengine_ECS",
    }
    resp, err := c.Get("metrics_v2", params)
    if err != nil { panic(err) }
    fmt.Println(string(resp))
}
```
