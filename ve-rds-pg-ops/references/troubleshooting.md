# Troubleshooting RDS PostgreSQL

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `InvalidParameter.InstanceName` | Name invalid | 1-128 chars, no leading number/dash |
| `InvalidParameter.NodeSpec` | Spec invalid | Check via DescribeInstanceSpecs |
| `InvalidParameter.StorageSpace` | Storage [20-3000]GB, step 10GB | Provide valid size |
| `InvalidParameter.ZoneConfig` | Zone invalid | Verify primary/secondary zones exist |
| `InvalidParameter.NetworkConfig` | VPC/subnet invalid | Verify network in region |
| `InvalidParameter.Parameter` | Value outside allowed range | Check parameter constraints |
| `ResourceNotFound.Instance` | Instance not found | Verify InstanceId |
| `ResourceNotFound.Account` | Account not found | Verify AccountName |
| `OperationDenied.InstanceStatus` | Invalid state | Wait for current operation |
| `QuotaExceeded.InstanceCount` | Max instances | Delete unused or raise quota |
| `InsufficientBalance` | No funds | Recharge |
| `InternalError` | Server error | Retry 3x; HALT with RequestId |
| `Throttling` | Rate limit | Exponential backoff |
| `ResourceInUse` | In use | Wait for operation |
| `Forbidden.RAM` | Insufficient permissions | Add RAM policy |

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
