# Elasticsearch Knowledge Base

## Fault Pattern Library

### Pattern 1: Cluster Health Turns Red (Shards Unassigned)

**Symptoms:**
- Cluster health status changes to Red
- Some indices become unavailable
- Search queries return partial results or time out

**Root Causes:**
- Data node failure or restart
- Disk space exhausted on data nodes (>95% usage triggers read-only mode)
- Network partition between nodes
- JVM heap pressure causing node to stop responding

**Diagnosis:**
```bash
# Check instance status
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Status'

# Check storage
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].StorageSpaceGb'

# Check node count
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].NodeNumber'
```

**Resolution:**
1. If disk full: increase storage via `ModifyInstanceSpec --StorageSpaceGb`
2. If node failed: restart instance via `RestartInstance`
3. If network issue: check VPC/subnet configuration via `ve-vpc-ops`

**Prevention:**
- Set up disk usage alerts at 75% and 85%
- Configure automatic snapshots before storage fills up
- Use at least 3 data nodes for production

---

### Pattern 2: JVM Heap Pressure / Circuit Breaker Trips

**Symptoms:**
- Bulk rejection errors in application logs
- Slow query responses
- Circuit breaker warning messages
- Nodes become unresponsive intermittently

**Root Causes:**
- Too many shards per node (recommended: < 1000 per GB of heap)
- Memory-intensive aggregations without proper limiting
- High cardinality fields in terms aggregations
- Insufficient JVM heap for workload

**Diagnosis:**
```bash
# Check node specs
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].NodeSpec'

# Count indices and estimate shard count
ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx | jq '.Result.TotalCount'
```

**Resolution:**
1. Upgrade node spec to larger memory (`ModifyInstanceSpec --NodeSpec es.x8.large`)
2. Reduce shard count by reindexing with fewer shards
3. Optimize queries: avoid high-cardinality aggregation, use `composite` aggregation with pagination
4. Increase bulk queue size (if configurable)

**Prevention:**
- Keep shard count per node under 1000 per GB of heap
- Use `_cat/shards` API to monitor shard distribution
- Set up JVM heap usage alerts at 75%

---

### Pattern 3: Snapshot Failures

**Symptoms:**
- Snapshot creation fails or is stuck
- Snapshot status shows `FAILED` instead of `SUCCESS`
- Restore operations fail

**Root Causes:**
- Repository (TOS bucket) connectivity issue
- Concurrent snapshot limit reached (max 5 concurrent)
- Index being actively modified during snapshot
- Insufficient storage in snapshot repository

**Diagnosis:**
```bash
# Check snapshot status
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx --SnapshotName "target-snapshot" | jq '.Result.Snapshots[0].Status'

# Count existing snapshots
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx | jq '.Result.Snapshots | length'

# Try creating a new snapshot to test
ve elasticsearch CreateSnapshot --Region cn-beijing --InstanceId es-xxx --SnapshotName "test-snapshot-$(date +%s)" --Indices "*"
```

**Resolution:**
1. If concurrent limit hit: wait for running snapshots to complete, then retry
2. If repository issue: check TOS bucket access in the same region
3. If an index is problematic: exclude it from the snapshot (`--Indices "*,-problem-index"`)

**Prevention:**
- Schedule snapshots during low-traffic periods
- Monitor snapshot success rate via alerts
- Keep at least 2× snapshot retention for disaster recovery

---

### Pattern 4: Upgrade Version Failure

**Symptoms:**
- Upgrade operation returns error
- Instance stuck in `Upgrading` state
- Incompatible version error

**Root Causes:**
- Direct version jump not supported (e.g., 7.10 → 8.5)
- Cluster health not at least Yellow
- Plugins incompatible with target version
- Indices need reindexing due to breaking changes

**Diagnosis:**
```bash
# Check current version
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Version'

# Check cluster health via ES REST API
curl -u "admin:password" "https://es-xxx.cn-beijing.es.volces.com:9200/_cluster/health" | jq '.status'

# Check for incompatible plugins
ve elasticsearch DescribePlugins --Region cn-beijing --InstanceId es-xxx | jq '.Result.Plugins[].PluginName'
```

**Resolution:**
1. Verify upgrade path is supported (use version compatibility matrix)
2. Take a full snapshot before attempting upgrade
3. Remove incompatible plugins before upgrade
4. For unsupported paths: create new instance with target version and reindex

**Prevention:**
- Always test version upgrades in a staging environment first
- Keep a current snapshot before any upgrade
- Review Elasticsearch breaking changes documentation before upgrading

---

### Pattern 5: Kibana Access Issues

**Symptoms:**
- Cannot access Kibana endpoint
- Kibana login page shows but authentication fails
- Kibana shows connection error to ES cluster

**Root Causes:**
- Kibana not enabled on the instance
- Incorrect username/password
- Network security group blocking port 5601
- Kibana session expired

**Diagnosis:**
```bash
# Check if Kibana is enabled
ve elasticsearch DescribeKibana --Region cn-beijing --InstanceId es-xxx

# Enable Kibana if disabled
ve elasticsearch EnableKibana --Region cn-beijing --InstanceId es-xxx --KibanaUser "admin" --KibanaPassword "<masked>"
```

**Resolution:**
1. If Kibana disabled: EnableKibana with new credentials
2. If authentication fails: DisableKibana → EnableKibana with new password
3. If network issue: check security group allows port 5601 from client IP

**Prevention:**
- Store Kibana credentials in a secure password manager
- Configure IP whitelist for Kibana access
- Use IAM roles instead of static credentials where possible

---

### Pattern 6: High Indexing Latency / Slow Bulk Indexing

**Symptoms:**
- Bulk indexing requests time out
- Indexing latency above 500ms
- Node CPU usage consistently high
- Bulk rejection errors

**Root Causes:**
- Inadequate node specs for indexing volume
- Too many shards (small shards cause high overhead)
- Refresh interval too frequent
- Insufficient number of nodes

**Diagnosis:**
```bash
# Check node specs and count
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0] | {Spec: .NodeSpec, Count: .NodeNumber}'

# Check disk I/O — ESSD PL level determines IOPS
# PL0: 10K, PL1: 50K, PL2: 100K, PL3: 1M IOPS
```

**Resolution:**
1. Scale up node specs: `ModifyInstanceSpec --NodeSpec es.x8.large`
2. Add more nodes: `ModifyNodeCount --NodeNumber 5`
3. Increase refresh interval: use `index.refresh_interval: 30s` for bulk loading
4. Reduce shard count: reindex with fewer primary shards
5. Use bulk API with optimal batch size (5-15 MB per batch)

**Prevention:**
- Provision for peak indexing throughput with 20% headroom
- Use ESSD PL2 or higher for write-heavy workloads
- Monitor bulk rejection rate and thread pool queues

---

### Pattern 7: Index Read-Only (Disk Watermark)

**Symptoms:**
- Write operations fail with `cluster_block_exception`
- Index metadata shows `index.blocks.read_only_allow_delete: true`
- Cannot index new documents

**Root Causes:**
- Disk usage exceeded the flood-stage watermark (95% by default)
- ES automatically puts indices in read-only mode to prevent data loss

**Diagnosis:**
```bash
# Check storage
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].StorageSpaceGb'
```

**Resolution:**
1. Free up disk space: delete old indices (`DeleteIndex`) or snapshots
2. Increase storage: `ModifyInstanceSpec --StorageSpaceGb 200`
3. ES will automatically remove read-only block when disk usage drops below flood stage

**Prevention:**
- Set up disk usage alerts at 75% and 85%
- Implement index lifecycle management (ILM) rollover policies
- Plan storage growth with 3-6 month horizon

---

### Pattern 8: Plugin Compatibility Issues

**Symptoms:**
- Plugin installation fails with `PluginIncompatible` error
- Node fails to start after plugin installation
- Elasticsearch service logs show plugin loading errors

**Root Causes:**
- Plugin not compatible with current ES version
- Plugin requires specific Elasticsearch features not available in managed service
- Conflicting plugins installed

**Diagnosis:**
```bash
# Check ES version
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Version'

# List all plugins
ve elasticsearch DescribePlugins --Region cn-beijing --InstanceId es-xxx | jq '.Result.Plugins[].PluginName'
```

**Resolution:**
1. Check plugin compatibility matrix for the specific ES version
2. Remove conflicting plugins before installing new ones
3. If plugin is critical, consider upgrading ES to a version that supports it

**Prevention:**
- Always verify plugin compatibility before installation
- Test plugins in a non-production instance first
- Keep a list of approved and verified plugins

---

## Runbook Index

| Pattern | Severity | Typical Recovery Time | Requires Snapshot? |
|---------|----------|----------------------|--------------------|
| Cluster Health Red | Critical | 15-60 min | No |
| JVM Heap Pressure | High | 30-120 min | No |
| Snapshot Failures | Medium | 5-30 min | N/A |
| Upgrade Failure | High | 30-120 min | Yes |
| Kibana Access Issues | Medium | 5-15 min | No |
| High Indexing Latency | Medium | 30-60 min | No |
| Index Read-Only | Critical | 5-30 min | No |
| Plugin Compatibility | Medium | 15-45 min | Recommended |
