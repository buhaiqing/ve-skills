# Troubleshooting — ECS

## Common Error Codes

| Error Code | HTTP Status | Meaning | Agent Action |
|-----------|-------------|---------|-------------|
| `InvalidRegion.NotFound` | 404 | Region does not exist | List valid regions via `DescribeRegions`; HALT |
| `InvalidImageId.NotFound` | 400 | Image does not exist | Verify image ID; list images via `DescribeImages` |
| `InvalidInstanceType.ValueNotSupported` | 400 | Instance type not supported | List available types via `DescribeInstanceTypes` |
| `InvalidSubnetId.NotFound` | 400 | Subnet does not exist | Verify subnet ID; check region match |
| `InvalidVpcId.NotFound` | 400 | VPC does not exist | Verify VPC ID |
| `InvalidSecurityGroupId.NotFound` | 400 | Security group not found | Verify security group ID |
| `InvalidInstanceId.NotFound` | 400 | Instance does not exist | Verify instance ID |
| `InvalidPasswordFormat` | 400 | Password doesn't meet requirements | Use 8-30 chars with 3 of: uppercase, lowercase, digit, special |
| `QuotaExceeded.Instance` | 400 | Instance quota exceeded | HALT; request quota increase |
| `QuotaExceeded.SecurityGroup` | 400 | Security group quota exceeded | HALT; delete unused or request increase |
| `InsufficientAvailableStock` | 400 | Resource unavailable in zone | Try different instance type or zone |
| `IncorrectInstanceStatus` | 400 | Instance in wrong state for operation | Check instance status; perform prerequisite action |
| `InstanceExpired` | 400 | Instance has expired | Renew instance |
| `Unauthorized` | 403 | Insufficient permissions | Attach required IAM policy (e.g., `ECSFullAccess`) |
| `InternalError` | 500 | Server-side error | Retry with backoff; HALT after 3 retries |
| `Throttling` | 429 | Rate limit exceeded | Exponential backoff; reduce request rate |
| `ExpiredOrder` | 400 | Request has expired | Retry the operation |
| `InvalidParameter` | 400 | Invalid request parameter | Check parameter against API docs |
| `LimitExceeded.MaxResults` | 400 | MaxResults exceeds limit | Reduce to ≤ 100 |
| `ResourceNotEnough` | 400 | Insufficient resources | Try different zone or instance type |

## Diagnostic Order

1. **Describe instance** — Check current status and configuration
   ```bash
   ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]'
   ```

2. **Check region and zone** — Verify resources exist in the specified region
   ```bash
   ve ecs DescribeRegions
   ve ecs DescribeZones --Region "{{user.region}}"
   ```

3. **Check CLI metadata** — Verify the operation is available via CLI
   ```bash
   ve ecs --help
   ve ecs <Action> --help
   ```

4. **Verify credentials** — Ensure env vars are set
   ```bash
   test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "OK"
   ```

5. **Check quota** — Verify resource limits
   ```bash
   ve ecs DescribeResourceQuota --Region "{{user.region}}"
   ```

6. **Check instance system log** — For boot issues
   ```bash
   ve ecs GetInstanceConsoleOutput --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"
   ```

## Common Scenarios

### Scenario 1: Instance Stuck in CREATING

**Diagnosis:**
```bash
ve ecs DescribeInstances --Region "{{user.region}}" --InstanceIds '["{{user.instance_id}}"]'
```

**Recovery:**
- If stuck > 30 min, it may have failed. Check event notifications
- If status transitions to `ERROR`, delete and recreate

### Scenario 2: Cannot Stop Instance

**Diagnosis:**
- Instance is already stopping/stopped
- Instance has local disk and requires `ForceStop`

**Recovery:**
```bash
ve ecs StopInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}" --ForceStop true
```

### Scenario 3: Cannot Delete Instance

**Diagnosis:**
- Instance is running (must be stopped first)
- Deletion protection is enabled

**Recovery:**
```bash
ve ecs StopInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"
ve ecs DeleteInstance --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"
```

### Scenario 4: No Available Instance Types

**Diagnosis:**
- Instance type not available in current zone
- Zone has insufficient stock

**Recovery:**
```bash
ve ecs DescribeInstanceTypes --InstanceTypeFamilyIds '["g3i"]'
# Try different zone or instance type
```

## Multi-Round Diagnosis Review

Before finalizing any diagnosis:

1. **Fact Check:** Are the ECS metrics and status current?
2. **Causal Analysis:** Is the identified cause the true root cause?
3. **Solution Validation:** Will the fix actually resolve the issue? Could it cause side effects?
