## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateInstance — Create an Elasticsearch Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` is valid | Region supported | Suggest valid region |
| VPC/Subnet | `ve vpc DescribeVpcs` | VPC exists | HALT; create VPC first |
| Name unique | Instance name not in use | No conflict | Use different name |
| Quota | `ve elasticsearch ListInstances` | Under quota | HALT; request quota increase |
| Version valid | Check supported versions | Version supported | Suggest valid version |

#### Execution — CLI (`ve`)

```bash
# Create an Elasticsearch instance (PostPaid)
ve elasticsearch CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --Version "{{user.instance_version}}" \
  --NodeSpec "{{user.node_spec}}" \
  --NodeNumber {{user.node_count}} \
  --StorageSpaceGb {{user.storage_gb}} \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid"

# Create with specific configuration
ve elasticsearch CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --Version "{{user.es_version}}" \
  --NodeSpec "es.x2.medium" \
  --NodeNumber 3 \
  --StorageSpaceGb 100 \
  --StorageType "ESSD" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/elasticsearch"
)

func main() {
    instance := elasticsearch.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":       os.Getenv("VOLCENGINE_REGION"),
        "InstanceName": os.Getenv("INSTANCE_NAME"),
        "Version":      os.Getenv("ES_VERSION"),
        "NodeSpec":     os.Getenv("NODE_SPEC"),
        "NodeNumber":   3,
        "StorageSpaceGb": 100,
        "VpcId":        os.Getenv("VPC_ID"),
        "SubnetId":     os.Getenv("SUBNET_ID"),
        "ChargeType":   "PostPaid",
    }

    resp, err := instance.Client.Request("elasticsearch", "CreateInstance", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

```bash
# Poll until instance is Running
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done

# Verify instance details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq '.Result.Instances[0] | {Name: .InstanceName, Version: .Version, Status: .Status, NodeSpec: .NodeSpec}'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidVpc.NotFound` | HALT; VPC does not exist — create VPC first via `ve-vpc-ops` |
| `InvalidSubnet.NotFound` | HALT; Subnet does not exist — create subnet first |
| `InvalidParameter` | HALT; fix parameter values per error message |
| `QuotaExceeded` | HALT; instance quota exceeded — request quota increase |
| `InsufficientBalance` | HALT; account balance insufficient — recharge |
| `InvalidVersion` | HALT; ES version not supported — select from valid versions |
| `InvalidNodeSpec` | HALT; node spec not available — select valid spec |

---

### Operation: DeleteInstance — Delete an Elasticsearch Instance

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of instance `{{user.instance_id}}` ({{user.instance_name}})
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all indices, data, snapshots, and configurations will be permanently deleted
- **MUST** verify instance status is `Running` or `Stopped`

```bash
# Verify instance exists and get details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"

# List indices to warn about data loss
echo "WARNING: The following indices and all their data will be permanently deleted:"
ve elasticsearch ListIndices --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Indices[].IndexName'
```

#### Execution

```bash
# Delete the instance
ve elasticsearch DeleteInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"
```

#### Validation

```bash
# Poll until instance is deleted (DescribeInstances returns empty or error)
for i in $(seq 1 60); do
  RESULT=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" 2>&1)
  if echo "$RESULT" | grep -q "InvalidInstance.NotFound" || echo "$RESULT" | grep -q '"Instances": \[\]'; then
    echo "Instance deleted successfully"
    break
  fi
  echo "Waiting for deletion... (attempt $i/60)"
  sleep 30
done
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | Instance already deleted; skip |
| `InvalidInstanceStatus` | HALT; instance is being modified — wait and retry |
| `DependencyViolation` | HALT; resources still attached — remove dependencies first |
| `Unauthorized` | HALT; check IAM permissions |

---

### Operation: CreateIndex — Create an Index

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Index name unique | `ve elasticsearch ListIndices` | Name not in list | Use different name |
| Shard count valid | Check limits | Within range | Adjust shard count |
| Storage available | Instance storage check | Sufficient space | Scale instance storage |

#### Execution

```bash
# Create an index
ve elasticsearch CreateIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" \
  --ShardCount {{user.shard_count}} \
  --ReplicaCount {{user.replica_count}}

# Create index with settings
ve elasticsearch CreateIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" \
  --ShardCount 5 \
  --ReplicaCount 1 \
  --Settings '{"index.refresh_interval": "30s", "index.number_of_routing_shards": 10}'
```

#### Validation

```bash
# Verify index exists
ve elasticsearch DescribeIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}"

# Check index health
ve elasticsearch DescribeIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}" | jq '.Result.Health'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `IndexAlreadyExists` | HALT; index exists — use different name or delete first |
| `InvalidParameter.ShardCount` | HALT; shard count exceeds limit — reduce shard count |
| `QuotaExceeded.Index` | HALT; index quota exceeded — delete unused indices |
| `InvalidIndexName` | HALT; index name contains invalid characters |

---

### Operation: DeleteIndex — Delete an Index

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of index `{{user.index_name}}`
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all documents in the index will be permanently deleted
- **MUST** check if any aliases or search applications depend on this index

```bash
# Warn user
echo "WARNING: Deleting index '{{user.index_name}}' will permanently delete all documents."
echo "Ensure no active search queries depend on this index."
```

#### Execution

```bash
# Delete the index
ve elasticsearch DeleteIndex \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --IndexName "{{user.index_name}}"
```

#### Validation

```bash
# Verify index is deleted
if ve elasticsearch DescribeIndices --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --IndexName "{{user.index_name}}" 2>&1 | grep -q "IndexNotFound"; then
  echo "Index deleted successfully"
fi
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `IndexNotFound` | Index already deleted; skip |
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `DeleteIndexPartial` | Some shards failed deletion — investigate cluster health |

---

### Operation: CreateSnapshot — Create a Snapshot

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Snapshot name unique | `ve elasticsearch DescribeSnapshots` | Name not in list | Use different name |
| Repository configured | Snapshot repository exists | Repository ready | HALT; configure repository |

#### Execution

```bash
# Create a snapshot
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" \
  --Indices "*"

# Create snapshot of specific indices
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" \
  --Indices "index1,index2,index3"
```

#### Validation

```bash
# Verify snapshot status
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}" | jq '.Result.Snapshots[0].Status'

# Poll until SUCCESS
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeSnapshots --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --SnapshotName "{{user.snapshot_name}}" | jq -r '.Result.Snapshots[0].Status')
  echo "Snapshot: $STATUS (attempt $i/60)"
  [ "$STATUS" = "SUCCESS" ] && break
  sleep 30
done
```

---

### Operation: InstallPlugin — Install a Plugin

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Plugin not installed | `ve elasticsearch DescribePlugins` | Plugin not in list | HALT; plugin already installed |
| Plugin compatible | Check plugin ES version compatibility | Compatible | Select compatible plugin |

#### Execution

```bash
# Install a plugin
ve elasticsearch InstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"

# Install with forced restart
ve elasticsearch InstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}" \
  --ForceRestart true
```

#### Validation

```bash
# Verify plugin is installed
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"

# List all installed plugins
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq -r '.Result.Plugins[].PluginName'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `PluginAlreadyExists` | HALT; plugin already installed — skip or uninstall first |
| `PluginNotFound` | HALT; plugin not available for this ES version |
| `PluginIncompatible` | HALT; plugin incompatible with current ES version |
| `RestartRequired` | Plugin requires instance restart — proceed with restart |

---

### Operation: UninstallPlugin — Remove a Plugin

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Plugin installed | `ve elasticsearch DescribePlugins` | Plugin found | HALT; plugin not installed |
| No dependency | Check if other plugins depend | No dependency | HALT; remove dependencies first |

#### Execution

```bash
# Uninstall a plugin
ve elasticsearch UninstallPlugin \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"
```

#### Validation

```bash
# Verify plugin is removed
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq -r '.Result.Plugins[].PluginName' | grep -v "{{user.plugin_name}}" || echo "Plugin removed"
```

---

### Operation: EnableKibana — Enable Kibana Access

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Kibana disabled | `ve elasticsearch DescribeKibana` | Not enabled | Already enabled |
| Password valid | Length and complexity | Min 8 chars | Require stronger password |

#### Execution

```bash
# Enable Kibana with admin credentials
ve elasticsearch EnableKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --KibanaUser "{{user.kibana_user}}" \
  --KibanaPassword "{{user.kibana_password}}"
```

> **Security:** Kibana password is masked in all output as `<masked>`.

#### Validation

```bash
# Get Kibana endpoint
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Extract the Kibana URL
KIBANA_URL=$(ve elasticsearch DescribeKibana --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.KibanaEndpoint')
echo "Kibana is available at: $KIBANA_URL"
```

---

### Operation: UpgradeVersion — Upgrade ES Version

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: upgrade ES instance `{{user.instance_id}}` from current version to target version
- **MUST** warn about potential incompatibilities: deprecated features, breaking API changes
- **MUST** take a snapshot before upgrade for rollback capability
- **MUST** verify the upgrade path is supported (no skip versions)

```bash
# Take pre-upgrade snapshot
ve elasticsearch CreateSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "pre-upgrade-$(date +%Y%m%d%H%M%S)" \
  --Indices "*"
```

#### Execution

```bash
# Upgrade ES version
ve elasticsearch UpgradeVersion \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TargetVersion "{{user.target_version}}"
```

#### Validation

```bash
# Poll until Running on new version
for i in $(seq 1 120); do
  VERSION=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Version')
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Version: $VERSION, Status: $STATUS (attempt $i/120)"
  [ "$STATUS" = "Running" ] && [ "$VERSION" = "{{user.target_version}}" ] && break
  sleep 30
done
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `IncompatibleVersion` | HALT; upgrade path not supported — check version compatibility matrix |
| `VersionNotFound` | HALT; target version not available — check supported versions |
| `UpgradeInProgress` | Another upgrade in progress — wait and retry |
| `SnapshotRequired` | HALT; take snapshot before upgrade |
| `ClusterHealthNotGreen` | HALT; cluster health must be at least Yellow |

---

### Operation: ModifyInstanceSpec — Modify Spec / Scale

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| Valid spec | Target spec is available | Spec found | Select valid spec |
| No ongoing operation | Instance status | Not modifying | Wait for completion |

#### Execution

```bash
# Modify node specification
ve elasticsearch ModifyInstanceSpec \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeSpec "{{user.new_node_spec}}"

# Modify node count (horizontal scaling)
ve elasticsearch ModifyNodeCount \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeNumber {{user.new_node_count}}

# Modify storage per node
ve elasticsearch ModifyInstanceSpec \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --StorageSpaceGb {{user.new_storage_gb}}
```

#### Validation

```bash
# Poll until spec change completes
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done

# Verify new spec
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq '.Result.Instances[0] | {Spec: .NodeSpec, Count: .NodeNumber, Storage: .StorageSpaceGb}'
```

---

### Operation: RestartInstance — Restart an Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| No pending operations | Instance status | Stable | Wait for completion |

#### Execution

```bash
# Restart the instance
ve elasticsearch RestartInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Restart with force (ignore cluster health)
ve elasticsearch RestartInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --Force true
```

#### Validation

```bash
# Poll until Running again
for i in $(seq 1 20); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  echo "Status: $STATUS (attempt $i/20)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done
```

---

### Operation: DescribeSnapshots — List Snapshots

#### Execution

```bash
# List all snapshots
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Get specific snapshot details
ve elasticsearch DescribeSnapshots \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}"
```

---

### Operation: ModifyChargeType — Change Billing Method

#### Pre-flight (Safety Gate)

- **MUST** confirm billing change: switching from PostPaid to PrePaid or vice versa
- **MUST** warn about potential cost impact

#### Execution

```bash
# Switch to PrePaid (subscription)
ve elasticsearch ModifyInstanceChargeType \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ChargeType "PrePaid" \
  --PeriodUnit "Month" \
  --Period 1

# Switch to PostPaid (pay-as-you-go)
ve elasticsearch ModifyInstanceChargeType \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ChargeType "PostPaid"
```

---

### Operation: ListInstances — List Elasticsearch Instances

#### Execution

```bash
# List all instances
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}"

# Filter by name
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceName "{{user.instance_name}}"

# Filter by status
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --Status "Running"

# Get single instance details
ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Instance ID | `.Result.Instances[].InstanceId` | Format `es-xxxxxxxxx` |
| Name | `.Result.Instances[].InstanceName` | User-defined |
| Version | `.Result.Instances[].Version` | ES version |
| Status | `.Result.Instances[].Status` | Running, Creating, Modifying, Deleting |
| Node Spec | `.Result.Instances[].NodeSpec` | Instance type |
| Node Count | `.Result.Instances[].NodeNumber` | Data node count |
| Storage | `.Result.Instances[].StorageSpaceGb` | GB per node |
| ChargeType | `.Result.Instances[].ChargeType` | PostPaid or PrePaid |
| CreateTime | `.Result.Instances[].CreateTime` | ISO 8601 |

---

### Operation: ListIndices — List Indices

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |

#### Execution

```bash
# List all indices
ve elasticsearch ListIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# List indices with details
ve elasticsearch ListIndices \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --Verbose true
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Index Name | `.Result.Indices[].IndexName` | Index identifier |
| Health | `.Result.Indices[].Health` | green, yellow, red |
| Docs Count | `.Result.Indices[].DocsCount` | Document count |
| Storage Size | `.Result.Indices[].StoreSize` | Storage used |

---

### Operation: DeleteSnapshot — Delete a Snapshot

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: delete snapshot `{{user.snapshot_name}}`
- **MUST** warn that deleted snapshots cannot be recovered

#### Execution

```bash
# Delete a snapshot
ve elasticsearch DeleteSnapshot \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --SnapshotName "{{user.snapshot_name}}"
```

#### Validation

```bash
# Verify snapshot is deleted
if ve elasticsearch DescribeSnapshots --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --SnapshotName "{{user.snapshot_name}}" 2>&1 | grep -q "SnapshotNotFound"; then
  echo "Snapshot deleted successfully"
fi
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `SnapshotNotFound` | 0 | Already deleted; skip — `[INFO] Snapshot already deleted.` |
| `SnapshotInProgress` | 3 | 10s wait for completion — `[WARNING] Snapshot in progress. Retrying...` |

---

### Operation: DisableKibana — Disable Kibana Access

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: disable Kibana for instance `{{user.instance_id}}`
- **MUST** warn that Kibana URL will become inaccessible
- **MUST** verify Kibana is currently enabled

```bash
# Verify Kibana is enabled
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq '.Result.KibanaEndpoint'
```

#### Execution

```bash
# Disable Kibana
ve elasticsearch DisableKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"
```

#### Validation

```bash
# Verify Kibana is disabled
ve elasticsearch DescribeKibana \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" 2>&1 | grep -q "KibanaDisabled" || echo "Kibana has been disabled"
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `KibanaAlreadyDisabled` | 0 | Already disabled; skip — `[INFO] Kibana already disabled.` |
| `InvalidInstance.NotFound` | 0 | HALT; instance does not exist — `[ERROR] Instance not found.` |

---

### Operation: DescribePlugins — List Installed Plugins

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |

#### Execution

```bash
# List all installed plugins
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Check specific plugin
ve elasticsearch DescribePlugins \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --PluginName "{{user.plugin_name}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Plugin Name | `.Result.Plugins[].PluginName` | Plugin identifier |
| Version | `.Result.Plugins[].Version` | Plugin version |
| Status | `.Result.Plugins[].Status` | Installed or installing |

---

### Operation: ModifyNodeCount — Scale Nodes Horizontally

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve elasticsearch DescribeInstances` | Instance Running | HALT |
| No ongoing operation | Instance status | Not modifying | Wait for completion |
| Valid node count | Check min/max limits | Within range | Adjust count |

#### Execution

```bash
# Scale data nodes
ve elasticsearch ModifyNodeCount \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --NodeNumber {{user.new_node_count}}
```

#### Validation

```bash
# Poll until spec change completes
for i in $(seq 1 60); do
  STATUS=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].Status')
  COUNT=$(ve elasticsearch DescribeInstances --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Instances[0].NodeNumber')
  echo "Status: $STATUS, NodeCount: $COUNT (attempt $i/60)"
  [ "$STATUS" = "Running" ] && [ "$COUNT" -eq "{{user.new_node_count}}" ] && break
  sleep 30
done
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidInstance.NotFound` | 0 | HALT; instance does not exist — `[ERROR] Instance not found.` |
| `InvalidInstanceStatus` | 3 | 10s wait for stable status — `[WARNING] Instance busy. Retrying...` |
| `QuotaExceeded` | 0 | HALT; node quota exceeded — `[ERROR] Node quota exceeded. Request quota increase.` |

---

