# MongoDB Knowledge Base

> Fault pattern library for AIOps diagnostic skills.
> Version: 1.0.0

## Pattern Categories

- [Connection Issues](#connection-issues)
- [Performance Degradation](#performance-degradation)
- [Replication Problems](#replication-problems)
- [Storage Issues](#storage-issues)
- [Memory Issues](#memory-issues)

---

## Connection Issues

### Pattern: Connection Refused

**Symptoms:**
- Error: `Connection refused`
- Cannot connect to MongoDB instance
- Application connection timeouts

**Diagnosis:**
1. Check instance status: `InstanceStatus != RUNNING`
2. Check IP whitelist: client IP not in whitelist
3. Check security group: port 27017 not open
4. Check network ACL: blocking traffic

**Resolution:**
```bash
# 1. Verify instance status
ve mongodb DescribeDBInstanceDetail --InstanceId <id> | jq '.Result.InstanceStatus'

# 2. Check IP whitelist
ve mongodb DescribeDBInstanceIPList --InstanceId <id>

# 3. Add client IP if missing
ve mongodb ModifyDBInstanceIPList --InstanceId <id> --IPList '["<client-ip>/32"]'
```

### Pattern: Authentication Failed

**Symptoms:**
- Error: `Authentication failed`
- Invalid credentials error
- SCRAM authentication failure

**Diagnosis:**
1. Verify account exists
2. Check password correctness
3. Check authentication database
4. Verify user privileges

**Resolution:**
```bash
# List accounts
ve mongodb DescribeDBAccounts --InstanceId <id>

# Reset password if needed
ve mongodb CreateDBAccount --InstanceId <id> --AccountName <user> --AccountPassword <newpass> --AccountPrivilege ReadWrite
```

### Pattern: Too Many Connections

**Symptoms:**
- Error: `connection pool exhausted`
- Error: `too many open files`
- Connection refused during peak load

**Diagnosis:**
```javascript
// Check current connections
db.serverStatus().connections
// { current: 1500, available: 500, totalCreated: 50000 }
```

**Resolution:**
1. Increase max connections parameter
2. Implement connection pooling in application
3. Close unused connections
4. Consider scaling up instance spec

---

## Performance Degradation

### Pattern: Slow Queries

**Symptoms:**
- Query latency > 100ms
- High CPU utilization
- Increasing response times

**Diagnosis:**
```javascript
// Enable profiling
db.setProfilingLevel(1, { slowms: 100 })

// View slow queries
db.system.profile.find().sort({ ts: -1 }).limit(10)

// Check for missing indexes
db.collection.find(query).explain("executionStats")
```

**Common Causes:**
1. Missing indexes
2. Large collection scans
3. Inefficient queries
4. Memory pressure

**Resolution:**
```javascript
// Create index
db.collection.createIndex({ field: 1 })

// Compound index
db.collection.createIndex({ field1: 1, field2: -1 })

// Check index usage
db.collection.aggregate([{ $indexStats: {} }])
```

### Pattern: High CPU Usage

**Symptoms:**
- `CPUUtilization > 80%`
- Slow response times
- Connection timeouts

**Diagnosis:**
```bash
# Check CloudMonitor metrics
ve cms DescribeMetricData --Namespace Volcengine_MongoDB --MetricName CPUUtilization --Dimensions '[{"InstanceId":"<id>"}]'
```

**Common Causes:**
1. CPU-intensive aggregations
2. Large sorts without indexes
3. High write throughput
4. Insufficient instance spec

**Resolution:**
1. Optimize slow queries
2. Add appropriate indexes
3. Distribute read load to secondaries
4. Upgrade to higher spec

### Pattern: High Memory Usage

**Symptoms:**
- `MemoryUtilization > 85%`
- Frequent page faults
- OOM errors

**Diagnosis:**
```javascript
db.serverStatus().mem
// { resident: 4096, virtual: 8192, supported: true }

db.serverStatus().wiredTiger.cache
// Check "bytes currently in the cache" vs "maximum bytes configured"
```

**Resolution:**
1. Increase instance memory
2. Reduce WiredTiger cache size if too large
3. Add indexes to reduce working set
4. Archive old data

---

## Replication Problems

### Pattern: Replication Lag

**Symptoms:**
- `ReplicationLag > 30 seconds`
- Stale reads from secondaries
- Secondary nodes out of sync

**Diagnosis:**
```javascript
// Check replication status
rs.printSecondaryReplicationInfo()

// Check oplog window
rs.printReplicationInfo()
```

**Common Causes:**
1. High write throughput
2. Slow network between nodes
3. Secondary under-provisioned
4. Large batch operations

**Resolution:**
1. Increase oplog size
2. Upgrade secondary specs
3. Distribute read load
4. Break large operations into smaller batches

### Pattern: Secondary Not Syncing

**Symptoms:**
- Secondary state: `RECOVERING`
- Secondary never catches up
- Replication error in logs

**Diagnosis:**
```javascript
rs.status()
// Check member states and optime differences
```

**Resolution:**
1. Restart secondary node
2. Resync from primary if data corruption
3. Check network connectivity between nodes
4. Verify firewall rules

---

## Storage Issues

### Pattern: Storage Full

**Symptoms:**
- `StorageUtilization > 85%`
- Write operations fail
- Error: `disk full`

**Diagnosis:**
```bash
# Check storage metrics
ve mongodb DescribeDBInstanceDetail --InstanceId <id> | jq '.Result.StorageSpaceGB'

# Check collection sizes
db.collection.stats().size
db.collection.stats().storageSize
```

**Resolution:**
```bash
# Expand storage
ve mongodb ModifyDBInstanceSpec --InstanceId <id> --StorageSpaceGB <new-size>

# Or delete old data
db.collection.deleteMany({ created_at: { $lt: new Date(Date.now() - 90*24*60*60*1000) } })

# Compact collections
db.collection.compact()
```

### Pattern: Rapid Storage Growth

**Symptoms:**
- Storage growing faster than expected
- Oplog consuming excessive space
- Fragmentation high

**Diagnosis:**
```javascript
// Check database sizes
db.stats()

// Check collection fragmentation
db.collection.stats().paddingFactor
db.collection.stats().fragmentation
```

**Resolution:**
1. Reduce oplog size if too large
2. Compact fragmented collections
3. Implement TTL indexes for automatic cleanup
4. Archive cold data

---

## Memory Issues

### Pattern: OOM (Out of Memory)

**Symptoms:**
- Instance restarts unexpectedly
- Error: `out of memory`
- Memory usage consistently high

**Diagnosis:**
```javascript
db.serverStatus().mem
db.serverStatus().wiredTiger.cache

// Check for memory leaks in application connections
```

**Resolution:**
1. Increase instance memory
2. Tune WiredTiger cache: ~50% of RAM
3. Limit concurrent connections
4. Optimize aggregation pipelines

### Pattern: Cache Miss Rate High

**Symptoms:**
- `CacheHitRate < 90%`
- Slow read performance
- High disk I/O

**Diagnosis:**
```javascript
db.serverStatus().wiredTiger.cache
// Check "pages read into cache" vs "pages requested from cache"
```

**Resolution:**
1. Increase instance memory
2. Add indexes to improve locality
3. Pre-warm cache after restart
4. Consider read replicas for distribution

---

## Diagnostic Workflows

### Complete Health Check

```bash
#!/bin/bash
INSTANCE_ID="mongo-xxx"

# 1. Instance status
ve mongodb DescribeDBInstanceDetail --InstanceId $INSTANCE_ID | jq '.Result.InstanceStatus'

# 2. Connection status
ve mongodb DescribeDBAccounts --InstanceId $INSTANCE_ID

# 3. Recent backups
ve mongodb DescribeBackups --InstanceId $INSTANCE_ID | jq '.Result.Backups[:3]'

# 4. IP whitelist
ve mongodb DescribeDBInstanceIPList --InstanceId $INSTANCE_ID
```

### Performance Baseline

```javascript
// Record baseline metrics
db.serverStatus().connections
db.serverStatus().opcounters
db.serverStatus().mem
db.serverStatus().wiredTiger.cache

// Document current indexes
db.getCollectionNames().forEach(c => {
  print(c + ":");
  db[c].getIndexes().forEach(i => printjson(i));
});
```

---

## Alert Correlations

| Alert A | Alert B | Correlation | Action |
|---------|---------|-------------|--------|
| High CPU | Slow Queries | Queries consuming CPU | Add indexes |
| High Memory | Cache Miss | Insufficient cache | Scale up memory |
| Replication Lag | High Write | Oplog pressure | Increase oplog size |
| Storage Full | Write Errors | No space left | Expand storage |
| Connection Exhaustion | App Errors | Connection leak | Fix pooling |

---

## Recovery Procedures

### Instance Recovery (ERROR state)

1. Collect diagnostic info
2. Attempt restart
3. If fails, restore from backup
4. Escalate to support with logs

### Data Recovery (Accidental Delete)

1. Stop writes immediately
2. Restore from backup to new instance
3. Export affected data
4. Import to original instance

### Performance Recovery

1. Identify slow queries
2. Add missing indexes
3. Kill long-running operations
4. Scale instance if needed
