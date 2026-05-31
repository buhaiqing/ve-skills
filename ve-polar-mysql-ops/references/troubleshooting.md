# Troubleshooting PolarDB MySQL

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `InvalidParameter.ClusterName` | Name format invalid | Use 1-64 chars, letters/numbers/hyphens/underscores |
| `InvalidParameter.NodeClass` | Node class invalid | Check available classes via DescribeDBNodeClasses |
| `InvalidParameter.NodeNumber` | Node count out of range | Valid range: 2-16 (1 primary + 1 secondary + 0-14 RO) |
| `InvalidParameter.StorageSpace` | Storage out of range [100-100000]GB | Provide valid storage size |
| `InvalidParameter.VpcId` | VPC invalid or not found | Verify VPC exists in region via ve-vpc-ops |
| `InvalidParameter.SubnetId` | Subnet invalid or not in VPC | Verify subnet exists and belongs to specified VPC |
| `InvalidParameter.ZoneId` | Zone invalid | Check available zones via DescribeAvailabilityZones |
| `InvalidParameter.Parameter` | Parameter value invalid | Check against CheckingCode range |
| `InvalidParameter.EndpointId` | Endpoint not found | Verify endpoint ID via DescribeDBClusterEndpoints |
| `ResourceNotFound.Cluster` | Cluster doesn't exist | Verify ClusterId |
| `ResourceNotFound.Node` | Node doesn't exist | Verify NodeId via DescribeDBNodes |
| `ResourceNotFound.Backup` | Backup doesn't exist | Verify BackupId via DescribeBackups |
| `OperationDenied.ClusterStatus` | Invalid cluster state | Wait for current operation to complete |
| `OperationDenied.NodeStatus` | Invalid node state | Wait for node operation to complete |
| `OperationDenied.InsufficientNodes` | Not enough nodes for operation | Minimum 2 nodes (primary + secondary) required |
| `QuotaExceeded.ClusterCount` | Max clusters reached | Delete unused or raise quota |
| `QuotaExceeded.NodeCount` | Max nodes per cluster reached | Max 16 nodes per cluster |
| `InsufficientBalance` | Account balance insufficient | Recharge account |
| `Throttling` | Rate limit exceeded | Exponential backoff |
| `InternalError` | Server-side error | Retry with backoff; HALT after 3 with RequestId |
| `ResourceInUse` | Resource in use by another operation | Wait for operation to complete |
| `Forbidden.RAM` | Insufficient permissions | Add RAM policy for PolarDB |
| `Conflict.InvalidOperation` | Operation conflicts with current state | Check cluster status and retry |
| `ServiceUnavailable` | Service temporarily unavailable | Retry after delay |

## Diagnostic Order

### General Diagnostic Flow

1. **Check cluster status:**
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx
   ```

2. **Check node statuses:**
   ```bash
   ve polardb_mysql DescribeDBNodes --ClusterId pc-xxx
   ```

3. **Verify VPC/subnet exists in target region:**
   ```bash
   ve vpc DescribeVpcs --VpcIds vpc-xxx
   ve vpc DescribeSubnetAttributes --SubnetId subnet-xxx
   ```

4. **Check storage type and space compatibility:**
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq '.Result.StorageSpace, .Result.StorageUsed'
   ```

5. **Verify engine version (MySQL_5_7 / MySQL_8_0):**
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq -r '.Result.DBEngineVersion'
   ```

6. **Review parameter values:**
   ```bash
   ve polardb_mysql DescribeDBClusterParameters --ClusterId pc-xxx
   ```

## Common Patterns

### Cluster Creation Stuck in CREATING

**Symptoms:** Cluster status remains `CREATING` for extended period (> 15 minutes).

**Causes:**
- Storage pool provisioning delay
- VPC/subnet misconfiguration
- Insufficient IP addresses in subnet
- AZ capacity constraints

**Fix:**
1. Verify VPC and subnet configuration
2. Check subnet has available IP addresses
3. Check cluster status for specific error details:
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq '.Result.ClusterStatus, .Result.ErrorMessage'
   ```
4. If stuck > 30 minutes, consider deleting and recreating

### Failover Failed

**Symptoms:** Failover command returns error or cluster enters `ERROR` state.

**Causes:**
- Secondary node not in `RUNNING` state
- Storage layer issues
- Network partition

**Fix:**
1. Check secondary node status:
   ```bash
   ve polardb_mysql DescribeDBNodes --ClusterId pc-xxx | jq '.Result.Nodes[] | select(.NodeRole == "Secondary")'
   ```
2. Verify both primary and secondary are healthy
3. Wait for any ongoing operations to complete
4. Retry failover

### Node Scaling Stuck

**Symptoms:** Node class modification remains in `SCALING` state.

**Causes:**
- Compute resource shortage in target AZ
- Storage layer synchronization

**Fix:**
1. Monitor status:
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq -r '.Result.ClusterStatus'
   ```
2. Check for error messages
3. May take up to 15 minutes for large instances
4. If stuck > 30 minutes, contact support with RequestId

### Storage Scaling Failed

**Symptoms:** Storage scaling returns error or cluster enters `ERROR` state.

**Causes:**
- Requested size smaller than current
- Storage quota exceeded
- Storage layer issues

**Fix:**
1. Verify new size > current size:
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq '.Result.StorageSpace'
   ```
2. Check storage is within limits (100-100000 GB)
3. Verify account has sufficient quota

### Connection Timeout

**Symptoms:** Cannot connect to cluster endpoint.

**Causes:**
- Security group blocking port 3306
- IP whitelist not configured
- Endpoint not active
- Cluster not in `RUNNING` state

**Fix:**
1. Check cluster status is `RUNNING`
2. Verify security group allows port 3306 from client IP
3. Check endpoint addresses:
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq '.Result.Endpoints[].Address'
   ```
4. Test connectivity from within VPC

### Read Replica Lag

**Symptoms:** Read-only nodes showing stale data.

**Causes:**
- Read-only node uses shared storage (PolarDB has no lag normally)
- If using separate read-only endpoint, check node health

**Note:** PolarDB read-only nodes share storage with primary, so there is normally **zero replication lag**. If seeing lag:

**Fix:**
1. Check node status:
   ```bash
   ve polardb_mysql DescribeDBNodes --ClusterId pc-xxx | jq '.Result.Nodes[] | {id: .NodeId, role: .NodeRole, status: .NodeStatus}'
   ```
2. Restart problematic read-only node if needed

### Parameter Change Not Applied

**Symptoms:** Modified parameter value not taking effect.

**Causes:**
- Parameter requires restart but cluster not restarted
- Parameter value outside valid range
- Parameter is read-only

**Fix:**
1. Check if parameter requires restart:
   ```bash
   ve polardb_mysql DescribeDBClusterParameters --ClusterId pc-xxx | jq '.Result.Parameters[] | select(.ParameterName == "max_connections") | .ForceRestart'
   ```
2. If `ForceRestart` is true, restart cluster:
   ```bash
   ve polardb_mysql RestartDBCluster --ClusterId pc-xxx
   ```
3. Verify parameter value against `CheckingCode`

### High Storage Usage Alert

**Symptoms:** Storage usage approaching 100%.

**Causes:**
- Data growth
- Binary logs not purged
- Temporary tables

**Fix:**
1. Check current usage:
   ```bash
   ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq '{total: .Result.StorageSpace, used: .Result.StorageUsed}'
   ```
2. Scale storage if needed:
   ```bash
   ve polardb_mysql ScaleStorage --ClusterId pc-xxx --StorageSpace 200
   ```
3. Review and optimize data retention policies
4. Check for unnecessary large tables or indexes

### Backup Creation Failed

**Symptoms:** Backup creation returns error or backup status shows `FAILED`.

**Causes:**
- Cluster not in `RUNNING` state
- Insufficient storage quota for backup
- Concurrent backup limit reached

**Fix:**
1. Verify cluster is `RUNNING`
2. Check backup quota
3. Retry after any ongoing backups complete
4. Check for error details:
   ```bash
   ve polardb_mysql DescribeBackups --ClusterId pc-xxx | jq '.Result.Backups[] | select(.BackupStatus == "FAILED")'
   ```

## Error Recovery Quick Reference

| Scenario | Command | Expected Result |
|----------|---------|-----------------|
| Check cluster health | `DescribeDBClusterDetail` | Status `RUNNING` |
| Check node health | `DescribeDBNodes` | All nodes `RUNNING` |
| Restart stuck cluster | `RestartDBCluster` | Status → `RUNNING` |
| Force failover | `FailoverDBCluster` | Secondary promoted |
| Scale storage urgently | `ScaleStorage` | Storage increased |
| Get error details | `DescribeDBClusterDetail` | Check `ErrorMessage` |

## Escalation Criteria

Escalate to Volcengine support when:

1. Cluster stuck in `ERROR` state after retry
2. Storage scaling fails repeatedly
3. Failover fails with `InternalError`
4. Data loss suspected
5. Issue persists > 2 hours despite troubleshooting

Include in support ticket:
- Cluster ID
- RequestId from error response
- Timeline of events
- Steps already attempted
