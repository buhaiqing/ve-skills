# MongoDB Knowledge Base — Fault Patterns

## Connection Issues

### Pattern: Connection Refused
**Symptoms:** `Connection refused`, app timeouts
**Diagnosis:** Instance not `RUNNING`? IP whitelist missing client? Security group blocking 27017?
**Fix:**
```bash
ve mongodb DescribeDBInstanceDetail --InstanceId <id> | jq '.Result.InstanceStatus'
ve mongodb DescribeDBInstanceIPList --InstanceId <id>
ve mongodb ModifyDBInstanceIPList --InstanceId <id> --IPList '["<client-ip>/32"]'
```

### Pattern: Auth Failed
**Symptoms:** `Authentication failed`, SCRAM failure
**Diagnosis:** Account exists? Password correct? Auth DB correct?
**Fix:**
```bash
ve mongodb DescribeDBAccounts --InstanceId <id>
```

### Pattern: Too Many Connections
**Symptoms:** `connection pool exhausted`, `too many open files`
**Diagnosis:** `db.serverStatus().connections` → current near max?
**Fix:** Increase maxConnections param, add connection pooling, scale up spec

---

## Performance Degradation

### Pattern: Slow Queries
**Symptoms:** Latency > 100ms, high CPU, increasing response times
**Diagnosis:**
```javascript
db.setProfilingLevel(1, { slowms: 100 })
db.system.profile.find().sort({ ts: -1 }).limit(10)
db.collection.find(query).explain("executionStats")
```
**Fix:** Add indexes, avoid large collection scans, optimize queries

### Pattern: High CPU
**Symptoms:** `CPUUtilization > 80%`, slow response, timeouts
**Causes:** CPU-intensive aggregations, large sorts w/o indexes, high write throughput
**Fix:** Optimize queries, add indexes, distribute reads to secondaries, upgrade spec

### Pattern: High Memory
**Symptoms:** `MemoryUtilization > 85%`, page faults, OOM
**Diagnosis:** `db.serverStatus().mem`, `db.serverStatus().wiredTiger.cache`
**Fix:** Increase memory, reduce WiredTiger cache, add indexes, archive old data

---

## Replication Problems

### Pattern: Replication Lag
**Symptoms:** `ReplicationLag > 30s`, stale reads, secondaries out of sync
**Diagnosis:** `rs.printSecondaryReplicationInfo()`, `rs.printReplicationInfo()`
**Causes:** High writes, slow network, under-provisioned secondary, large batch ops
**Fix:** Increase oplog size, upgrade secondary spec, distribute reads, smaller batches

### Pattern: Secondary Not Syncing
**Symptoms:** State `RECOVERING`, never catches up
**Diagnosis:** `rs.status()`
**Fix:** Restart secondary, resync from primary, check network

---

## Storage Issues

### Pattern: Storage Full
**Symptoms:** `StorageUtilization > 85%`, writes fail, `disk full`
**Diagnosis:** `db.collection.stats().size`, `db.collection.stats().storageSize`
**Fix:**
```bash
ve mongodb ModifyDBInstanceSpec --InstanceId <id> --StorageSpaceGB <new-size>
```
Or delete old data, compact collections

### Pattern: Rapid Storage Growth
**Symptoms:** Storage growing faster than expected, high fragmentation
**Diagnosis:** `db.stats()`, `db.collection.stats().fragmentation`
**Fix:** Reduce oplog size, compact fragmented collections, TTL indexes for auto cleanup

---

## Memory Issues

### Pattern: OOM
**Symptoms:** Instance restarts unexpectedly, `out of memory`, memory consistently high
**Diagnosis:** `db.serverStatus().wiredTiger.cache`
**Fix:** Increase memory, tune WiredTiger cache (~50% of RAM), limit connections, optimize aggregations

### Pattern: Cache Miss Rate High
**Symptoms:** `CacheHitRate < 90%`, slow reads, high disk I/O
**Diagnosis:** Check "pages read into cache" vs "pages requested from cache"
**Fix:** Increase memory, add indexes for locality, pre-warm cache after restart, read replicas

---

## Alert Correlations

| Alert A | Alert B | Correlation | Action |
|---------|---------|-------------|--------|
| High CPU | Slow Queries | Queries consuming CPU | Add indexes |
| High Memory | Cache Miss | Insufficient cache | Scale up memory |
| Replication Lag | High Write | Oplog pressure | Increase oplog size |
| Storage Full | Write Errors | No space | Expand storage |
| Connection Exhaustion | App Errors | Connection leak | Fix pooling |

## Recovery Quick Reference

| Scenario | Action |
|----------|--------|
| ERROR state | Collect info → attempt restart → restore from backup → escalate |
| Data deleted | Stop writes → restore to new instance → export → import |
| Performance degraded | ID slow queries → add indexes → kill long-running ops → scale