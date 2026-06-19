# Troubleshooting — CMS

## Common Error Codes

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `InvalidParameter` | Invalid request parameter | Verify all parameters against API docs |
| `NoSuchMetric` | Metric not found in namespace | HALT; check namespace and metric name |
| `NoSuchResource` | Resource does not exist | HALT; verify resource ID in Dimensions |
| `Throttling` | Rate limit exceeded | Exponential backoff; reduce query frequency |
| `Unauthorized` | Insufficient permissions | HALT; attach VMSReadOnlyAccess IAM policy |
| `InvalidDimensions` | Malformed Dimensions JSON | Fix JSON format; use `[{"key":"value"}]` |
| `MetricDataInsufficient` | Not enough data for evaluation | HALT; wait for data accumulation |
| `RuleAlreadyExists` | Alarm rule name conflicts | HALT; use unique rule name |
| `ContactGroupNotFound` | Notification group not found | HALT; create group first |
| `QuotaExceeded.AlarmRule` | Alarm rule limit reached | HALT; delete unused rules or request increase |
| `InternalError` | Server-side error | Retry with backoff; HALT after 3 retries |
| `ResourceNotFound` | Alarm template or rule not found | Verify resource ID |
| `InvalidTimestamp` | StartTime > EndTime or invalid format | Verify timestamps are valid Unix ms |
| `Forbidden.NoData` | API call quota exceeded | HALT; wait for quota reset or upgrade plan |
| `TemplateAlreadyApplied` | Template already applied to resource | HALT; remove first or use different template |

## Diagnostic Order

1. **Check credentials:**
   ```bash
   test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "OK"
   ```

2. **Verify service is available:**
   ```bash
   ve version
   ```

3. **Query metric metadata:** (Check if namespace and metric exist)
   ```bash
   ve metrics ListMetrics --Namespace Volcengine_ECS
   ```

4. **Query alarm rules:**
   ```bash
   ve metrics DescribeMetricRuleList --Namespace Volcengine_ECS
   ```

## Common Scenarios

### Scenario 1: No Data Returned

**Diagnosis:** Metrics take time to report after resource creation.

**Recovery:**
- Wait 5-15 minutes after creating a resource before querying metrics
- Verify the resource ID is correct and active
- Check that monitoring agent is installed (for OS-level metrics)

### Scenario 2: Alarm Rule Never Fires

**Diagnosis:** Threshold too high, or metric data insufficient.

**Recovery:**
- Query historical metrics to establish baseline
- Lower threshold gradually
- Verify alarm rule `EnableState` is `true`
- Check notification group is properly configured

### Scenario 3: API Throttling

**Diagnosis:** Too many queries in short time window.

**Recovery:**
- Implement client-side rate limiting
- Use longer `Period` values to reduce data points
- Cache query results when possible
