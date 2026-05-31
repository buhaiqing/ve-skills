# Elasticsearch Troubleshooting Guide

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| `InvalidParameter` / 400 | Request parameter invalid | Align parameters with API spec |
| `InvalidInstance.NotFound` / 404 | Instance does not exist | Verify instance ID; may be deleted |
| `InvalidVpc.NotFound` / 400 | VPC does not exist | Create VPC via `ve-vpc-ops` first |
| `InvalidSubnet.NotFound` / 400 | Subnet does not exist | Create subnet first |
| `IndexAlreadyExists` / 400 | Index already exists | Use different name or delete first |
| `IndexNotFound` / 404 | Index does not exist | Verify index name |
| `PluginAlreadyExists` / 400 | Plugin already installed | Skip or uninstall first |
| `PluginNotFound` / 404 | Plugin not available | Verify plugin name and ES version |
| `PluginIncompatible` / 400 | Plugin incompatible with ES | Select compatible plugin version |
| `QuotaExceeded` / 400 | Resource quota exceeded | Request quota increase |
| `QuotaExceeded.Index` / 400 | Index quota exceeded | Delete unused indices |
| `QuotaExceeded.Snapshot` / 400 | Snapshot quota exceeded | Delete old snapshots |
| `InsufficientBalance` / 400 | Account balance insufficient | Recharge account |
| `Unauthorized` / 403 | IAM permission denied | Check IAM policies |
| `Forbidden.RAM` / 403 | RAM policy denies access | Add ES permissions to IAM policy |
| `InvalidInstanceStatus` / 400 | Instance status invalid | Wait for stable state |
| `IncompatibleVersion` / 400 | Version upgrade not supported | Check version compatibility matrix |
| `ClusterHealthNotGreen` / 400 | Cluster health not Green/Yellow | Fix cluster health first |
| `SnapshotInProgress` / 400 | Snapshot already running | Wait for current snapshot to complete |
| `SnapshotNotFound` / 404 | Snapshot does not exist | Verify snapshot name |
| `RestoreInProgress` / 400 | Restore already in progress | Wait for current restore |
| `InternalError` / 500 | Server-side error | Retry with backoff; escalate if persists |
| `Throttling` / 429 | Rate limit exceeded | Back off and retry |
| `ServiceUnavailable` / 503 | Service temporarily unavailable | Retry after delay |

## Diagnostic Order

### 1. Instance Issues

**Symptom: Instance creation fails**

```bash
# Check VPC exists
ve vpc DescribeVpcs --Region cn-beijing

# Check subnet exists
ve vpc DescribeSubnets --Region cn-beijing --SubnetIds '["subnet-xxx"]'

# Check quota
ve elasticsearch ListInstances --Region cn-beijing | jq '.Result.TotalCount'

# Check account balance
ve billing DescribeBalance
```

**Symptom: Instance stuck in Creating/Restarting state**

```bash
# Check instance status
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Status'

# Wait and poll
for i in {1..60}; do
  STATUS=$(ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done
```

### 2. Index Issues

**Symptom: Cannot create index**

```bash
# Check instance status
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Status'

# Check existing indices
ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx | jq '.Result.Indices[].IndexName'

# Check index quota
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0] | {Storage: .StorageSpaceGb}'

# Check for duplicate name
echo "Checking if index name is available..."
ve elasticsearch DescribeIndices --Region cn-beijing --InstanceId es-xxx --IndexName "target-index" 2>&1 | grep -q "IndexNotFound" && echo "Name available" || echo "Index exists"
```

**Symptom: Index health is Red (shards unassigned)**

```bash
# Check which shards are unassigned
# This information is typically visible via Kibana dev tools or ES REST API

# Possible causes:
# - Insufficient disk space on data nodes
# - Data node failure or restart
# - Network partition

# Resolution steps:
# 1. Check node status: ve elasticsearch DescribeInstances --InstanceId es-xxx
# 2. Check storage usage
# 3. Restart instance if needed: ve elasticsearch RestartInstance --InstanceId es-xxx
```

### 3. Snapshot Issues

**Symptom: Snapshot creation fails**

```bash
# Check existing snapshots
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx

# Check snapshot repository status
# The repository is managed automatically; ensure TOS bucket is accessible

# Check if too many concurrent snapshots
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx | jq '.Result.Snapshots | length'
```

**Symptom: Snapshot deletion fails**

```bash
# Verify snapshot exists
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx --SnapshotName "target-snapshot"

# Check snapshot status — only SUCCESS or PARTIAL can be deleted
ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx --SnapshotName "target-snapshot" | jq '.Result.Snapshots[0].Status'
```

### 4. Plugin Issues

**Symptom: Plugin installation fails**

```bash
# Check current installed plugins
ve elasticsearch DescribePlugins --Region cn-beijing --InstanceId es-xxx | jq '.Result.Plugins[].PluginName'

# Verify plugin is not already installed
ve elasticsearch DescribePlugins --Region cn-beijing --InstanceId es-xxx | grep "target-plugin" || echo "Plugin not installed"

# Check ES version compatibility
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].Version'
```

**Symptom: Plugin causes node instability**

```bash
# Uninstall the problematic plugin immediately
ve elasticsearch UninstallPlugin --Region cn-beijing --InstanceId es-xxx --PluginName "target-plugin"

# Restart instance if needed
ve elasticsearch RestartInstance --Region cn-beijing --InstanceId es-xxx
```

### 5. Kibana Issues

**Symptom: Cannot access Kibana**

```bash
# Check if Kibana is enabled
ve elasticsearch DescribeKibana --Region cn-beijing --InstanceId es-xxx | jq '.Result.KibanaEndpoint'

# If Kibana is disabled, enable it
ve elasticsearch EnableKibana --Region cn-beijing --InstanceId es-xxx --KibanaUser "admin" --KibanaPassword "<masked>"

# Verify security group allows port 5601
# Ensure client IP is in Kibana IP whitelist
```

**Symptom: Kibana login fails**

```bash
# Reset Kibana password
# Note: This may require disabling and re-enabling Kibana

# Step 1: Disable Kibana
ve elasticsearch DisableKibana --Region cn-beijing --InstanceId es-xxx

# Step 2: Re-enable with new password
ve elasticsearch EnableKibana --Region cn-beijing --InstanceId es-xxx --KibanaUser "admin" --KibanaPassword "<masked>"
```

### 6. Performance Issues

**Symptom: Slow search queries**

```bash
# Check instance specs
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].NodeSpec'

# Check node count
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].NodeNumber'

# Check disk usage
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].StorageSpaceGb'

# Resolution: scale up NodeSpec or add more nodes
```

**Symptom: High disk usage**

```bash
# Check current storage
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0].StorageSpaceGb'

# List indices by size
ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Indices[] | "\(.IndexName): \(.StorageSizeBytes / 1073741824) GB"' | sort -t: -k2 -rn | head -10

# Delete old/unused indices or increase storage
# ve elasticsearch ModifyInstanceSpec --StorageSpaceGb 200
```

## Error Recovery Procedures

### Instance Creation Failure

1. **Check VPC/Subnet:**
   ```bash
   ve vpc DescribeVpcs --Region cn-beijing
   ve vpc DescribeSubnets --Region cn-beijing
   ```

2. **Verify quota:**
   ```bash
   CURRENT=$(ve elasticsearch ListInstances --Region cn-beijing | jq -r '.Result.TotalCount // 0')
   echo "Current instances: $CURRENT"
   ```

3. **Check balance:**
   ```bash
   ve billing DescribeBalance
   ```

### Index Creation Failure

1. **Check instance status:**
   ```bash
   ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Instances[0].Status'
   ```

2. **Check existing indices:**
   ```bash
   ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Indices[].IndexName' | grep -i "target-index"
   ```

### Snapshot Failure

1. **Check if snapshot exists:**
   ```bash
   ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx --SnapshotName "target-snapshot"
   ```

2. **Check concurrent snapshot limit:**
   ```bash
   ve elasticsearch DescribeSnapshots --Region cn-beijing --InstanceId es-xxx | jq '.Result.Snapshots | length'
   ```

3. **Wait or retry:**
   ```bash
   sleep 60 && ve elasticsearch CreateSnapshot --Region cn-beijing --InstanceId es-xxx --SnapshotName "target-snapshot" --Indices "*"
   ```

## Version Upgrade Recovery

### Upgrade Failure

1. **Check version compatibility:**
   - 7.10 → 8.5: NOT supported (must go through 7.16)
   - 7.10 → 7.16 → 8.5: Supported path

2. **Restore from pre-upgrade snapshot:**
   ```bash
   # Create new instance from snapshot if upgrade fails catastrophically
   ve elasticsearch CreateInstance --Region cn-beijing --Version "7.16" --SnapshotName "pre-upgrade-snapshot"
   ```

## Prevention Checklist

- [ ] Instance naming follows convention (env-purpose)
- [ ] VPC and subnet verified before instance creation
- [ ] Shard sizing optimized (10-50 GB per shard)
- [ ] Snapshot schedule configured for all production instances
- [ ] Pre-upgrade snapshot taken before version upgrades
- [ ] Plugin compatibility verified before installation
- [ ] Kibana IP whitelist configured for production
- [ ] Monitoring alerts set up for disk usage and cluster health
- [ ] Storage scaling planned before reaching 75% usage
- [ ] Test version upgrades in staging first
