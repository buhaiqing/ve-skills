# Troubleshooting MongoDB

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `InvalidParameter.InstanceName` | Name format invalid | Use valid name (2-128 chars, alphanumeric, hyphens) |
| `InvalidParameter.NodeSpec` | Spec invalid | Check available specs via DescribeDBInstanceSpecs |
| `InvalidParameter.StorageSpace` | Storage out of range [20-3000]GB | Provide valid storage size |
| `InvalidParameter.MongoVersion` | Invalid MongoDB version | Valid: 4.0, 4.2, 4.4, 5.0, 6.0 |
| `InvalidParameter.NetworkConfig` | VPC/subnet invalid | Verify network exists in region |
| `InvalidParameter.AccountName` | Account name invalid | Use valid format (1-63 chars) |
| `InvalidParameter.AccountPassword` | Password invalid | Must be 8-32 chars |
| `InvalidParameter.DBName` | Database name invalid | Use valid MongoDB identifier |
| `InvalidParameter.CollectionName` | Collection name invalid | Use valid MongoDB identifier |
| `InvalidParameter.IPList` | IP format invalid | Use valid CIDR notation |
| `InvalidParameter.Parameter` | Parameter value invalid | Check parameter constraints |
| `ResourceNotFound.Instance` | Instance doesn't exist | Verify InstanceId |
| `ResourceNotFound.Account` | Account doesn't exist | Verify AccountName |
| `ResourceNotFound.Database` | Database doesn't exist | Verify DBName |
| `ResourceNotFound.Collection` | Collection doesn't exist | Verify CollectionName |
| `ResourceNotFound.Backup` | Backup doesn't exist | Verify BackupId |
| `ResourceNotFound.Vpc` | VPC doesn't exist | Verify VpcId |
| `ResourceNotFound.Subnet` | Subnet doesn't exist | Verify SubnetId |
| `OperationDenied.InstanceStatus` | Invalid state for operation | Wait for operation to complete |
| `OperationDenied.AccountStatus` | Account in invalid state | Wait or check account |
| `QuotaExceeded.InstanceCount` | Max instances reached | Delete unused or request quota |
| `QuotaExceeded.AccountCount` | Max accounts reached | Delete unused accounts |
| `QuotaExceeded.DatabaseCount` | Max databases reached | Delete unused databases |
| `QuotaExceeded.CollectionCount` | Max collections reached | Delete unused collections |
| `InsufficientBalance` | Account balance insufficient | Recharge account |
| `InternalError` | Server-side error | Retry with backoff; HALT after 3 with RequestId |
| `Throttling` | Rate limit hit | Exponential backoff |
| `ResourceInUse` | Resource in use by another operation | Wait for operation to complete |
| `ResourceAlreadyExists` | Resource already exists | Use different name or reuse existing |
| `Forbidden.RAM` | Insufficient permissions | Add RAM policy for MongoDB |

## Diagnostic Order

### Instance Issues

1. Check instance status
   ```bash
   ve mongodb DescribeDBInstanceDetail --InstanceId <id>
   ```

2. Verify VPC/subnet exists in target region
   ```bash
   ve vpc DescribeVpcs --Region <region>
   ve vpc DescribeSubnets --VpcId <vpc-id>
   ```

3. Check storage compatibility with node spec
   ```bash
   ve mongodb DescribeDBInstanceSpecs --Region <region> --MongoVersion <version>
   ```

4. Verify MongoDB version availability
   ```bash
   ve mongodb DescribeDBInstanceSpecs --Region <region>
   ```

### Connection Issues

1. Check instance status is `RUNNING`
2. Verify connection string and port
3. Check IP whitelist includes client IP
   ```bash
   ve mongodb DescribeDBInstanceIPList --InstanceId <id>
   ```
4. Verify security group allows port 27017
5. Test connectivity
   ```bash
   mongosh "mongodb://<user>:<pass>@<host>:<port>/<db>" --eval "db.adminCommand('ping')"
   ```

### Performance Issues

1. Check current connections
   ```javascript
   // Connect via mongosh and run:
   db.serverStatus().connections
   ```

2. Check slow queries
   ```javascript
   db.system.profile.find().sort({ ts: -1 }).limit(10)
   ```

3. Check instance resource usage in console

4. Consider spec upgrade if CPU/memory constrained

### Backup/Restore Issues

1. Verify backup exists and is completed
   ```bash
   ve mongodb DescribeBackups --InstanceId <id>
   ```

2. Check storage space for restore operation
3. Verify target instance is in `RUNNING` state
4. For large restores, expect longer completion time (up to 1800s)

## Common Patterns

### Instance Creation Stuck in CREATING

**Causes:**
- VPC/subnet misconfiguration
- Storage provisioning delay
- Quota limitation

**Fix:**
1. Wait up to 600s for normal creation
2. Check VPC and subnet configuration
3. Verify quota limits

### Connection Timeout

**Causes:**
- IP whitelist not configured
- Security group blocking port 27017
- Instance not in RUNNING state
- Network ACL blocking traffic

**Fix:**
1. Add client IP to IP whitelist
   ```bash
   ve mongodb ModifyDBInstanceIPList --InstanceId <id> --IPList '["<client-ip>/32"]'
   ```
2. Verify security group allows port 27017
3. Check instance status
4. Test with telnet: `telnet <host> 27017`

### Authentication Failed

**Causes:**
- Wrong username or password
- User doesn't exist
- Insufficient privileges
- Authentication database not specified

**Fix:**
1. Verify account exists
   ```bash
   ve mongodb DescribeDBAccounts --InstanceId <id>
   ```
2. Check password
3. Ensure using correct authentication database (usually `admin`)
4. Verify user privileges for target database

### Slow Spec Modification

**Causes:**
- Storage resize is asynchronous
- Node spec change requires data migration
- High load on instance

**Fix:**
1. Monitor InstanceStatus until RUNNING
2. May take up to 900s for spec changes
3. Consider maintenance window for high-load instances

### Backup Creation Failed

**Causes:**
- Instance in invalid state
- Insufficient storage for backup
- Concurrent backup limit reached

**Fix:**
1. Ensure instance is RUNNING
2. Check backup quota
3. Retry after previous backup completes

### Restore Failed

**Causes:**
- Backup corrupted or incomplete
- Insufficient storage on target
- Version incompatibility

**Fix:**
1. Verify backup integrity
2. Ensure target has sufficient storage
3. Check MongoDB version compatibility

## Recovery Procedures

### Emergency Instance Recovery

If instance is stuck in ERROR state:

1. Contact support with RequestId
2. Consider restore from backup to new instance
3. Document data loss if any

### Password Reset

If admin password is lost:

1. Use `ModifyDBAccountPrivilege` to create new admin user
2. Or restore from backup to new instance
3. Update application connection strings

### Data Recovery

If data is accidentally deleted:

1. Stop writes immediately
2. Restore from latest backup to new instance
3. Export affected collections
4. Import to original instance

## Monitoring Checkpoints

| Checkpoint | Command/API | Healthy Value |
|------------|-------------|---------------|
| Instance status | `DescribeDBInstanceDetail` | `RUNNING` |
| Connection count | `db.serverStatus().connections` | < maxConnections * 0.8 |
| Replication lag | `rs.printSecondaryReplicationInfo()` | < 10 seconds |
| Storage usage | `db.stats()` | < 80% |
| Memory usage | `db.serverStatus().mem` | < 90% |
