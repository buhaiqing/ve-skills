# CLI — PolarDB MySQL (`ve polardb_mysql`)

## Install and Config

- Install: `ve` CLI from https://github.com/volcengine/volcengine-cli/releases
- Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
- Output: **JSON by default**

## Command Map

### Cluster Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List clusters | `ve polardb_mysql DescribeDBClusters --Region cn-beijing` | Paginated |
| Cluster details | `ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx` | Full info |
| Create cluster | `ve polardb_mysql CreateDBCluster --Region cn-beijing --ZoneId cn-beijing-a --VpcId vpc-xxx --SubnetId subnet-xxx --ClusterName my-cluster --DBEngineVersion MySQL_8_0 --NodeClass polar.mysql.x4.large --NodeNumber 2 --StorageSpace 100 --ChargeType PostPaid` | Returns cluster ID |
| Delete cluster | `ve polardb_mysql DeleteDBCluster --ClusterId pc-xxx` | **Irreversible** |
| Restart cluster | `ve polardb_mysql RestartDBCluster --ClusterId pc-xxx` | Restarts all nodes |
| Failover | `ve polardb_mysql FailoverDBCluster --ClusterId pc-xxx` | Manual failover |

### Node Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List nodes | `ve polardb_mysql DescribeDBNodes --ClusterId pc-xxx` | Shows all nodes |
| Add RO node | `ve polardb_mysql CreateDBNode --ClusterId pc-xxx --NodeClass polar.mysql.x4.large --NodeNumber 1` | Add read-only nodes |
| Delete node | `ve polardb_mysql DeleteDBNode --ClusterId pc-xxx --NodeId pi-xxx` | Remove RO node |
| Restart node | `ve polardb_mysql RestartDBNode --NodeId pi-xxx` | Single node restart |
| Scale compute | `ve polardb_mysql ModifyDBNodeClass --ClusterId pc-xxx --NodeClass polar.mysql.x4.2xlarge` | Scale all nodes |

### Storage Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| Scale storage | `ve polardb_mysql ScaleStorage --ClusterId pc-xxx --StorageSpace 500` | Expand storage pool |

### Endpoint Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List endpoints | `ve polardb_mysql DescribeDBClusterEndpoints --ClusterId pc-xxx` | RW and RO endpoints |
| Modify endpoint | `ve polardb_mysql ModifyDBClusterEndpoint --ClusterId pc-xxx --EndpointId ep-xxx --AutoAddNewNodes true` | Auto-add new nodes |

### Backup Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List backups | `ve polardb_mysql DescribeBackups --ClusterId pc-xxx` | Backup history |
| Create backup | `ve polardb_mysql CreateBackup --ClusterId pc-xxx --BackupName manual-backup` | On-demand backup |
| Restore cluster | `ve polardb_mysql RestoreDBCluster --BackupId bk-xxx --ClusterName restored-cluster --VpcId vpc-xxx --SubnetId subnet-xxx` | Creates new cluster |

### Parameter Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List parameters | `ve polardb_mysql DescribeDBClusterParameters --ClusterId pc-xxx` | Cluster parameters |
| Modify parameters | `ve polardb_mysql ModifyDBClusterParameters --ClusterId pc-xxx --Parameters '[{"ParameterName":"max_connections","ParameterValue":"2000"}]'` | JSON array |
| List param groups | `ve polardb_mysql DescribeParameterGroups --Region cn-beijing` | Available templates |
| Create param group | `ve polardb_mysql CreateParameterGroup --Region cn-beijing --ParameterGroupName my-params --DBEngineVersion MySQL_8_0` | New template |
| Delete param group | `ve polardb_mysql DeleteParameterGroup --ParameterGroupId pg-xxx` | Remove template |

### Metadata Operations

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List regions | `ve polardb_mysql DescribeRegions` | Available regions |
| List zones | `ve polardb_mysql DescribeAvailabilityZones --RegionId cn-beijing` | AZs in region |
| List node classes | `ve polardb_mysql DescribeDBNodeClasses --Region cn-beijing --DBEngineVersion MySQL_8_0` | Available specs |

## Parameter Discovery

```bash
ve polardb_mysql --help
ve polardb_mysql CreateDBCluster --help
ve polardb_mysql DescribeDBClusterDetail --help
```

## JSON Body Passing

For complex nested parameters, use `--body` with JSON:

```bash
ve polardb_mysql CreateDBCluster \
  --Region cn-beijing \
  --body '{
    "ZoneId": "cn-beijing-a",
    "VpcId": "vpc-xxx",
    "SubnetId": "subnet-xxx",
    "ClusterName": "my-polardb",
    "DBEngineVersion": "MySQL_8_0",
    "NodeClass": "polar.mysql.x4.large",
    "NodeNumber": 2,
    "StorageSpace": 100,
    "ChargeType": "PostPaid"
  }'
```

## Polling Patterns

### Wait for Cluster Running

```bash
CLUSTER_ID="pc-xxx"
for i in $(seq 1 90); do
  STATUS=$(ve polardb_mysql DescribeDBClusterDetail --ClusterId "$CLUSTER_ID" | jq -r '.Result.ClusterStatus // ""')
  echo "Status: $STATUS (attempt $i/90)"
  [ "$STATUS" = "RUNNING" ] && break
  [ "$STATUS" = "ERROR" ] && echo "Cluster failed" && exit 1
  sleep 10
done
```

### Wait for Node Running

```bash
NODE_ID="pi-xxx"
for i in $(seq 1 60); do
  STATUS=$(ve polardb_mysql DescribeDBNodes --ClusterId "$CLUSTER_ID" | jq -r ".Result.Nodes[] | select(.NodeId == \"$NODE_ID\") | .NodeStatus")
  echo "Node status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "RUNNING" ] && break
  sleep 5
done
```

## Output Parsing Examples

### Get Cluster ID from Create Response

```bash
CLUSTER_ID=$(ve polardb_mysql CreateDBCluster ... | jq -r '.Result.ClusterId')
echo "Created: $CLUSTER_ID"
```

### Get Endpoint Address

```bash
ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq -r '.Result.Endpoints[].Address'
```

### List All Cluster IDs

```bash
ve polardb_mysql DescribeDBClusters --Region cn-beijing | jq -r '.Result.Clusters[].ClusterId'
```

### Get Storage Usage

```bash
ve polardb_mysql DescribeDBClusterDetail --ClusterId pc-xxx | jq -r '{total: .Result.StorageSpace, used: .Result.StorageUsed}'
```
