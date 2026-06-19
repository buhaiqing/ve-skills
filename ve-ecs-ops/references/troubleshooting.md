# Troubleshooting — ECS

## Common Error Codes

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `InvalidRegion.NotFound` | HALT; query valid regions | `ve ecs DescribeRegions` |
| `InvalidImageId.NotFound` | HALT; verify image ID | `ve ecs DescribeImages` |
| `InvalidInstanceType.ValueNotSupported` | HALT; list available types | `ve ecs DescribeInstanceTypes` |
| `InvalidSubnetId.NotFound` | HALT; verify subnet ID & region | `ve vpc DescribeSubnets` |
| `InvalidVpcId.NotFound` | HALT; verify VPC ID | `ve vpc DescribeVpcs` |
| `InvalidSecurityGroupId.NotFound` | HALT; verify SG ID | `ve ecs DescribeSecurityGroups` |
| `InvalidInstanceId.NotFound` | HALT; verify instance ID | `ve ecs DescribeInstances` |
| `InvalidPasswordFormat` | HALT; fix password | 8-30 chars, 3 of: upper, lower, digit, special |
| `QuotaExceeded.Instance` | HALT; request increase | `ve ecs DescribeResourceQuota` |
| `QuotaExceeded.SecurityGroup` | HALT; delete unused or increase | Check SG usage |
| `InsufficientAvailableStock` | Retry different type/zone | `ve ecs DescribeInstanceTypes --InstanceTypeFamilyIds '["g3i"]'` |
| `IncorrectInstanceStatus` | HALT; check status | Perform prerequisite action first |
| `InstanceExpired` | HALT; renew instance | Renew in console or API |
| `Unauthorized` | HALT; attach IAM policy | Attach `ECSFullAccess` |
| `InternalError` | Retry 3x with backoff | HALT after 3 retries |
| `Throttling` | Backoff 2s, 4s, 8s | Reduce request rate |
| `ExpiredOrder` | Retry operation | Re-submit request |
| `InvalidParameter` | HALT; check API docs | Fix parameter value |
| `LimitExceeded.MaxResults` | Reduce to ≤ 100 | Retry once |
| `ResourceNotEnough` | Try different zone/type | `ve ecs DescribeZones --Region "{{user.region}}"` |

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
