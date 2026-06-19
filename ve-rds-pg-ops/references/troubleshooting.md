# Troubleshooting RDS PostgreSQL

## Common Error Codes

| Error Code | Agent Action | Recovery |
|-----------|--------------|----------|
| `InvalidParameter.InstanceName` | HALT | 1-128 chars, no leading number/dash |
| `InvalidParameter.NodeSpec` | HALT | Check via DescribeInstanceSpecs |
| `InvalidParameter.StorageSpace` | HALT | Storage [20-3000]GB, step 10GB |
| `InvalidParameter.ZoneConfig` | HALT | Verify primary/secondary zones exist |
| `InvalidParameter.NetworkConfig` | HALT | Verify VPC/subnet in region |
| `InvalidParameter.Parameter` | HALT | Check parameter constraints |
| `ResourceNotFound.Instance` | HALT | Verify InstanceId |
| `ResourceNotFound.Account` | HALT | Verify AccountName |
| `OperationDenied.InstanceStatus` | HALT | Wait for current operation |
| `QuotaExceeded.InstanceCount` | HALT | Delete unused or raise quota |
| `InsufficientBalance` | HALT | Recharge account |
| `InternalError` | Retry (3x) | HALT with RequestId if persists |
| `Throttling` | Retry (exponential) | Back off and retry |
| `ResourceInUse` | HALT | Wait for concurrent operation |
| `Forbidden.RAM` | HALT | Add RAM policy |

## Diagnostic Order

1. Check instance: `ve rds_postgresql DescribeDBInstanceDetail --InstanceId <id>`
2. Verify zones available in region
3. Check storage type (LocalSSD only for PG)
4. Review parameter values
5. Check replication status for read-only nodes

## Common Patterns

### WAL Disk Pressure
- Cause: WAL accumulation, replication lag
- Fix: Check `wal_keep_size`; monitor backup status; increase storage

### Connection Exhaustion
- Cause: `max_connections` limit reached
- Fix: Increase parameter; check for connection leaks; use connection pooling

### Read-Only Node Not Syncing
- Cause: Network issues, heavy write load on primary
- Fix: Check replication lag; restart read-only node if needed

### Parameter Change Ignored
- Cause: `ForceRestart=true` but not restarted
- Fix: Restart instance after modifying restart-required parameters
