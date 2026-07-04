# Troubleshooting — ECS

## Common Error Codes

> Error codes are per-operation in SKILL.md §Failure Recovery tables.
> Brief overview below — see `../SKILL.md` per-op recovery for full detail.

| Error Pattern | Action | Recovery (summary) |
|---|--------|--------------------|
| `Invalid*NotFound` | HALT | Verify resource ID → `Describe*` |
| `QuotaExceeded.*` | HALT | `DescribeResourceQuota` → increase or cleanup |
| `InsufficientAvailableStock` | RETRY | Different type/zone → `DescribeInstanceTypes` |
| `IncorrectInstanceStatus` | HALT | Check status → pre-req action first |
| `Unauthorized` | HALT | Attach `ECSFullAccess` IAM policy |
| `InternalError` | RETRY 3x | Backoff → HALT after 3 |
| `Throttling` | RETRY 3x | Backoff 2s, 4s, 8s |
| `InvalidPasswordFormat` | HALT | 8-30 chars, 3 of: upper, lower, digit, special |
| `InvalidParameter` | HALT | Fix param per API docs |
| `LimitExceeded.MaxResults` | FIX | Reduce to ≤ 100 |

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
