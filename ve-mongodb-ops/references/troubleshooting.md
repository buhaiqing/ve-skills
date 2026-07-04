# Troubleshooting MongoDB

> **Version:** 1.1.0 | **Last Updated:** 2026-07-04

## Error Taxonomy

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `InvalidParameter.*` (Name/Spec/Storage/Version/Network/Account/Password/DB/Collection/IP) | **HALT** — param val failed | Check value against API constraints or query `DescribeDBInstanceSpecs` |
| `ResourceNotFound.*` (Instance/Account/Database/Collection/Backup/Vpc/Subnet) | **HALT** — resource not found | Verify ID/name via describe command |
| `OperationDenied.Instance/AccountStatus` | **HALT** — wrong state | Wait for transition or cancel current op |
| `QuotaExceeded.*` (Instance/Account/Database/Collection) | **HALT** — quota exhausted | Delete unused or request increase |
| `InsufficientBalance` | **HALT** — balance insufficient | Recharge via billing |
| `Forbidden.RAM` | **HALT** — insufficient perms | Verify IAM policy |
| `ResourceInUse` | **HALT** — resource busy | Wait for concurrent op to finish |
| `ResourceAlreadyExists` | **HALT** — name taken | Use unique name or reuse |
| `Throttling` | RETRY w/ exponential backoff | Max 3 retries |
| `InternalError` | RETRY w/ backoff 2s,4s,8s | Max 3 retries; capture `RequestId` |

## Diagnostic Order

### Instance Issues

1. Check status: `ve mongodb DescribeDBInstanceDetail --InstanceId <id>`
2. Verify VPC/subnet: `ve vpc DescribeVpcs --Region <region>`
3. Check storage/spec compatibility: `ve mongodb DescribeDBInstanceSpecs --Region <region> --MongoVersion <version>`

### Connection Issues

1. Instance `RUNNING`?
2. IP whitelist includes client? `ve mongodb DescribeDBInstanceIPList --InstanceId <id>`
3. Security group allows 27017?
4. Test: `mongosh "mongodb://<user>:<pass>@<host>:<port>/<db>" --eval "db.adminCommand('ping')"`

### Performance Issues

1. Check connections: `db.serverStatus().connections`
2. Slow queries: `db.system.profile.find().sort({ ts: -1 }).limit(10)`
3. Consider spec upgrade if CPU/memory constrained

### Backup/Restore Issues

1. Verify backup complete: `ve mongodb DescribeBackups --InstanceId <id>`
2. Target instance `RUNNING`?
3. Large restores → up to 1800s

## Common Patterns

### Instance Stuck CREATING
**Causes:** VPC/subnet misconfig, storage delay, quota limit.
**Fix:** Wait up to 600s; verify VPC; check quotas.

### Connection Timeout
**Causes:** IP whitelist missing, security group blocking 27017, instance not RUNNING.
**Fix:**
```bash
ve mongodb ModifyDBInstanceIPList --InstanceId <id> --IPList '["<client-ip>/32"]'
```

### Authentication Failed
**Causes:** Wrong user/pass, user doesn't exist, auth DB not specified.
**Fix:**
```bash
ve mongodb DescribeDBAccounts --InstanceId <id>
```

### Slow Spec Modification
**Causes:** Async storage resize, data migration, high load.
**Fix:** Monitor until RUNNING; up to 900s; schedule maintenance window.

### Backup Creation Failed
**Causes:** Instance invalid state, insufficient storage, concurrent backup.
**Fix:** Ensure RUNNING; check quota; retry after prior backup completes.

## Recovery Procedures

| Scenario | Action |
|----------|--------|
| ERROR state | Contact support w/ RequestId; restore from backup to new instance |
| Lost admin password | `ModifyDBAccountPrivilege` to create new admin, or restore from backup |
| Accidental data delete | Stop writes → restore from backup to new instance → export → import back |