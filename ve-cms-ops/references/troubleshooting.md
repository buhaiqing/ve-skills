# Troubleshooting — CMS

## Common Error Codes

| Error Code | Action | Recovery |
|-----------|--------|----------|
| `InvalidParameter` | Invalid param | Verify API docs |
| `NoSuchMetric` | Metric not found | HALT; check namespace/metric |
| `NoSuchResource` | Resource missing | HALT; verify Dimensions ID |
| `Throttling` | Rate exceeded | Exponential backoff + reduced freq |
| `Unauthorized` | No permission | HALT; attach VMSReadOnlyAccess policy |
| `InvalidDimensions` | Bad JSON format | Fix → `[{"key":"value"}]` |
| `MetricDataInsufficient` | Not enough data | HALT; wait for accumulation |
| `RuleAlreadyExists` | Name conflict | HALT; use unique name |
| `ContactGroupNotFound` | Group missing | HALT; create group first |
| `QuotaExceeded.AlarmRule` | Limit reached | HALT; delete unused or request increase |
| `InternalError` | Server error | Retry ×3 with backoff → HALT |
| `ResourceNotFound` | Rule/template missing | Verify ID |
| `InvalidTimestamp` | Time range bad | Verify Unix ms format |
| `Forbidden.NoData` | Quota exceeded | HALT; wait for reset/upgrade |
| `TemplateAlreadyApplied` | Already applied | HALT; remove or use different template |

## Diagnostic Order

1. ✅ Credentials:
   ```bash
   test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "OK"
   ```

2. ✅ Service available:
   ```bash
   ve version
   ```

3. ✅ Metric metadata:
   ```bash
   ve metrics ListMetrics --Namespace Volcengine_ECS
   ```

4. ✅ Alarm rules:
   ```bash
   ve metrics DescribeMetricRuleList --Namespace Volcengine_ECS
   ```

## Common Scenarios

### 1: No Data Returned

**Recovery:**
- Wait 5-15 min after resource creation before querying
- Verify resource ID is correct & active
- Check monitoring agent installed (OS-level metrics)

### 2: Alarm Rule Never Fires

**Recovery:**
- Query historical metrics → establish baseline
- Lower threshold gradually
- Verify `EnableState` is `true`
- Check notification group configured

### 3: API Throttling

**Recovery:**
- Client-side rate limiting
- Use longer `Period` to reduce data points
- Cache query results
