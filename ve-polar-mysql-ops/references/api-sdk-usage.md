# PolarDB MySQL API & SDK Usage

## OpenAPI

- **Version:** `2022-01-01`
- **Service:** `polardb_mysql`
- **Endpoint:** `polardb-mysql.{region}.volcengineapi.com`
- **Doc:** https://www.volcengine.com/docs/6498

## SDK Operations Map

### Cluster Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| Create Cluster | `CreateDBCluster` | Region, ZoneId, VpcId, SubnetId, ClusterName, DBEngineVersion, NodeClass, NodeNumber, StorageSpace, ChargeType |
| Describe Cluster | `DescribeDBClusterDetail` | ClusterId |
| List Clusters | `DescribeDBClusters` | Region (optional filters) |
| Delete Cluster | `DeleteDBCluster` | ClusterId |
| Restart Cluster | `RestartDBCluster` | ClusterId |
| Failover | `FailoverDBCluster` | ClusterId |

### Node Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| List Nodes | `DescribeDBNodes` | ClusterId |
| Add Read-Only Node | `CreateDBNode` | ClusterId, NodeClass, NodeNumber |
| Delete Node | `DeleteDBNode` | ClusterId, NodeId |
| Restart Node | `RestartDBNode` | NodeId |
| Modify Node Class | `ModifyDBNodeClass` | ClusterId, NodeClass |

### Storage Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| Scale Storage | `ScaleStorage` | ClusterId, StorageSpace |

### Endpoint Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| List Endpoints | `DescribeDBClusterEndpoints` | ClusterId |
| Modify Endpoint | `ModifyDBClusterEndpoint` | ClusterId, EndpointId |

### Backup Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| List Backups | `DescribeBackups` | ClusterId |
| Create Backup | `CreateBackup` | ClusterId, BackupName |
| Restore Cluster | `RestoreDBCluster` | BackupId, ClusterName, VpcId, SubnetId |

### Parameter Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| List Parameters | `DescribeDBClusterParameters` | ClusterId |
| Modify Parameters | `ModifyDBClusterParameters` | ClusterId, Parameters[] |
| List Param Groups | `DescribeParameterGroups` | Region |
| Create Param Group | `CreateParameterGroup` | Region, ParameterGroupName, DBEngineVersion |
| Delete Param Group | `DeleteParameterGroup` | ParameterGroupId |

### Metadata Operations

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| List Regions | `DescribeRegions` | None |
| List Zones | `DescribeAvailabilityZones` | RegionId |
| List Node Classes | `DescribeDBNodeClasses` | Region, DBEngineVersion |

## Response JSON Paths

### Cluster Operations

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| CreateDBCluster | `$.Result.ClusterId` | ClusterId |
| DescribeDBClusters | `$.Result.Clusters[]` | ClusterId, ClusterName, ClusterStatus, DBEngineVersion |
| DescribeDBClusterDetail | `$.Result.*` | Full cluster details |
| DeleteDBCluster | `$.Metadata.RequestId` | RequestId |

### Node Operations

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| DescribeDBNodes | `$.Result.Nodes[]` | NodeId, NodeClass, NodeRole, NodeStatus |
| CreateDBNode | `$.Result.NodeIds[]` | Created node IDs |
| DeleteDBNode | `$.Metadata.RequestId` | RequestId |
| ModifyDBNodeClass | `$.Metadata.RequestId` | RequestId |

### Storage Operations

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| ScaleStorage | `$.Metadata.RequestId` | RequestId |

### Backup Operations

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| DescribeBackups | `$.Result.Backups[]` | BackupId, BackupName, BackupStatus, BackupSize |
| CreateBackup | `$.Result.BackupId` | BackupId |
| RestoreDBCluster | `$.Result.ClusterId` | New cluster ID |
| Pagination | `$.Result.TotalCount` | Total items |

## Pagination

DescribeDBClusters supports:
- `PageNumber`: Starting from 1
- `PageSize`: Default 10, max 100
- Response: TotalCount for total items

DescribeBackups supports:
- `PageNumber`: Starting from 1
- `PageSize`: Default 10
- Response: TotalCount

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"

instance := polardb_mysql.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
resp, err := instance.Client.Request("polardb_mysql", "CreateDBCluster", params)
```

## Common Request Patterns

### Creating Cluster with Multiple Nodes

```go
params := map[string]interface{}{
    "Region":          "cn-beijing",
    "ZoneId":          "cn-beijing-a",
    "VpcId":           "vpc-xxx",
    "SubnetId":        "subnet-xxx",
    "ClusterName":     "my-polardb",
    "DBEngineVersion": "MySQL_8_0",
    "NodeClass":       "polar.mysql.x4.large",
    "NodeNumber":      2,  // Primary + Secondary
    "StorageSpace":    100,
    "ChargeType":      "PostPaid",
}
```

### Adding Read-Only Nodes

```go
params := map[string]interface{}{
    "ClusterId":  "pc-xxx",
    "NodeClass":  "polar.mysql.x4.large",
    "NodeNumber": 2,  // Add 2 RO nodes
}
resp, err := instance.Client.Request("polardb_mysql", "CreateDBNode", params)
```

### Modifying Parameters

```go
params := map[string]interface{}{
    "ClusterId": "pc-xxx",
    "Parameters": []map[string]interface{}{
        {"ParameterName": "max_connections", "ParameterValue": "2000"},
        {"ParameterName": "innodb_buffer_pool_size", "ParameterValue": "8589934592"},
    },
}
resp, err := instance.Client.Request("polardb_mysql", "ModifyDBClusterParameters", params)
```

### Scaling Storage

```go
params := map[string]interface{}{
    "ClusterId":    "pc-xxx",
    "StorageSpace": 500,  // New total size in GB
}
resp, err := instance.Client.Request("polardb_mysql", "ScaleStorage", params)
```
