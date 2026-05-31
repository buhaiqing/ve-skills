# Elasticsearch CLI Usage

## Installation and Configuration

### Install ve CLI

```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify installation
ve version
```

### Configure Credentials

**Environment Variables (Recommended for Agents):**

```bash
export VOLCENGINE_ACCESS_KEY="your-access-key"
export VOLCENGINE_SECRET_KEY="<masked>"  # Never display in output
export VOLCENGINE_REGION="cn-beijing"
```

**Verify credentials are set:**
```bash
test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "Credentials configured"
```

## CLI Conventions

- **Output is JSON by default**
- **Service prefix**: `ve elasticsearch`
- **Help**: `ve elasticsearch --help` or `ve elasticsearch <action> --help`
- **Region is required** for most operations

## Command Reference

### Instance Management

```bash
# List instances
ve elasticsearch ListInstances --Region cn-beijing

# Describe instances
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx

# Create instance
ve elasticsearch CreateInstance \
  --Region cn-beijing \
  --InstanceName "prod-search" \
  --Version "7.16" \
  --NodeSpec "es.x4.medium" \
  --NodeNumber 3 \
  --StorageSpaceGb 100 \
  --VpcId "vpc-xxx" \
  --SubnetId "subnet-xxx" \
  --ChargeType "PostPaid"

# Modify instance spec
ve elasticsearch ModifyInstanceSpec \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --NodeSpec "es.x4.large"

# Modify node count
ve elasticsearch ModifyNodeCount \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --NodeNumber 5

# Restart instance
ve elasticsearch RestartInstance \
  --Region cn-beijing \
  --InstanceId es-xxx

# Delete instance
ve elasticsearch DeleteInstance \
  --Region cn-beijing \
  --InstanceId es-xxx
```

### Version Upgrade

```bash
# Upgrade ES version
ve elasticsearch UpgradeVersion \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --TargetVersion "8.5"
```

### Index Management

```bash
# List indices
ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx

# Describe index
ve elasticsearch DescribeIndices \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --IndexName "logs-2024.05.27"

# Create index
ve elasticsearch CreateIndex \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --IndexName "logs-2024.05.27" \
  --ShardCount 5 \
  --ReplicaCount 1

# Delete index
ve elasticsearch DeleteIndex \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --IndexName "logs-2024.05.27"
```

### Snapshot Management

```bash
# Create snapshot
ve elasticsearch CreateSnapshot \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --SnapshotName "daily-backup" \
  --Indices "*"

# Describe snapshots
ve elasticsearch DescribeSnapshots \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --SnapshotName "daily-backup"

# Delete snapshot
ve elasticsearch DeleteSnapshot \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --SnapshotName "daily-backup"
```

### Plugin Management

```bash
# List installed plugins
ve elasticsearch DescribePlugins --Region cn-beijing --InstanceId es-xxx

# Install plugin
ve elasticsearch InstallPlugin \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --PluginName "analysis-ik" \
  --ForceRestart true

# Uninstall plugin
ve elasticsearch UninstallPlugin \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --PluginName "analysis-ik"
```

### Kibana Management

```bash
# Enable Kibana
ve elasticsearch EnableKibana \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --KibanaUser "admin" \
  --KibanaPassword "<masked>"

# Describe Kibana
ve elasticsearch DescribeKibana \
  --Region cn-beijing \
  --InstanceId es-xxx

# Disable Kibana
ve elasticsearch DisableKibana \
  --Region cn-beijing \
  --InstanceId es-xxx
```

### Billing

```bash
# Modify charge type
ve elasticsearch ModifyInstanceChargeType \
  --Region cn-beijing \
  --InstanceId es-xxx \
  --ChargeType "PrePaid" \
  --PeriodUnit "Month" \
  --Period 1
```

## CLI vs API Coverage Gap

| Operation | Available via `ve` CLI | Notes |
|-----------|------------------------|-------|
| CreateInstance | Yes | — |
| DescribeInstances | Yes | — |
| ListInstances | Yes | — |
| ModifyInstanceSpec | Yes | — |
| ModifyNodeCount | Yes | — |
| DeleteInstance | Yes | — |
| RestartInstance | Yes | — |
| UpgradeVersion | Yes | — |
| ModifyInstanceChargeType | Yes | — |
| CreateIndex | Yes | — |
| DescribeIndices | Yes | — |
| DeleteIndex | Yes | — |
| ListIndices | Yes | Pagination supported |
| CreateSnapshot | Yes | — |
| DescribeSnapshots | Yes | — |
| DeleteSnapshot | Yes | — |
| InstallPlugin | Yes | — |
| DescribePlugins | Yes | — |
| UninstallPlugin | Yes | — |
| EnableKibana | Yes | — |
| DescribeKibana | Yes | — |
| DisableKibana | Yes | — |

## Common Patterns

### Check Instance Status

```bash
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Instances[0].Status'
```

### Get Index Health

```bash
ve elasticsearch DescribeIndices --Region cn-beijing --InstanceId es-xxx --IndexName "logs-*" | jq -r '.Result.Health'
```

### Monitor Disk Usage

```bash
ve elasticsearch DescribeInstances --Region cn-beijing --InstanceId es-xxx | jq '.Result.Instances[0] | {Name: .InstanceName, Spec: .NodeSpec, Storage: .StorageSpaceGb}'
```

### List All Indices with Sizes

```bash
ve elasticsearch ListIndices --Region cn-beijing --InstanceId es-xxx | jq -r '.Result.Indices[] | "\(.IndexName): \(.DocCount) docs, \(.StorageSizeBytes / 1073741824) GB"'
```

## JSON Path Quick Reference

| Field | JSON Path |
|-------|-----------|
| Instance ID | `.Result.Instances[0].InstanceId` |
| Instance Status | `.Result.Instances[0].Status` |
| Instance Name | `.Result.Instances[0].InstanceName` |
| ES Version | `.Result.Instances[0].Version` |
| Node Spec | `.Result.Instances[0].NodeSpec` |
| Node Count | `.Result.Instances[0].NodeNumber` |
| Storage (GB) | `.Result.Instances[0].StorageSpaceGb` |
| Index Name | `.Result.IndexName` |
| Index Health | `.Result.Health` |
| Shard Count | `.Result.ShardCount` |
| Document Count | `.Result.DocCount` |
| Snapshot Status | `.Result.Snapshots[0].Status` |
| Kibana Endpoint | `.Result.KibanaEndpoint` |
| Request ID | `.ResponseMetadata.RequestId` |
| Error Code | `.Error.Code` |
| Error Message | `.Error.Message` |
