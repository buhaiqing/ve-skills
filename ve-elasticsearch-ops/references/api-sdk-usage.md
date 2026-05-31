# Elasticsearch API & SDK Usage

## OpenAPI Reference

- **API Documentation**: https://www.volcengine.com/docs/6337
- **Base Path**: `https://es.{region}.volces.com`
- **API Version**: 2022-01-01

## SDK Operations Map

| Goal | API Operation | CLI Command | SDK Method |
|------|---------------|-------------|------------|
| Create Instance | CreateInstance | `ve elasticsearch CreateInstance` | `CreateInstance` |
| Describe Instances | DescribeInstances | `ve elasticsearch DescribeInstances` | `DescribeInstances` |
| Modify Instance Spec | ModifyInstanceSpec | `ve elasticsearch ModifyInstanceSpec` | `ModifyInstanceSpec` |
| Delete Instance | DeleteInstance | `ve elasticsearch DeleteInstance` | `DeleteInstance` |
| List Instances | ListInstances | `ve elasticsearch ListInstances` | `ListInstances` |
| Restart Instance | RestartInstance | `ve elasticsearch RestartInstance` | `RestartInstance` |
| Upgrade Version | UpgradeVersion | `ve elasticsearch UpgradeVersion` | `UpgradeVersion` |
| Modify Node Count | ModifyNodeCount | `ve elasticsearch ModifyNodeCount` | `ModifyNodeCount` |
| Modify Charge Type | ModifyInstanceChargeType | `ve elasticsearch ModifyInstanceChargeType` | `ModifyInstanceChargeType` |
| Create Index | CreateIndex | `ve elasticsearch CreateIndex` | `CreateIndex` |
| Describe Indices | DescribeIndices | `ve elasticsearch DescribeIndices` | `DescribeIndices` |
| Delete Index | DeleteIndex | `ve elasticsearch DeleteIndex` | `DeleteIndex` |
| List Indices | ListIndices | `ve elasticsearch ListIndices` | `ListIndices` |
| Create Snapshot | CreateSnapshot | `ve elasticsearch CreateSnapshot` | `CreateSnapshot` |
| Describe Snapshots | DescribeSnapshots | `ve elasticsearch DescribeSnapshots` | `DescribeSnapshots` |
| Delete Snapshot | DeleteSnapshot | `ve elasticsearch DeleteSnapshot` | `DeleteSnapshot` |
| Install Plugin | InstallPlugin | `ve elasticsearch InstallPlugin` | `InstallPlugin` |
| Describe Plugins | DescribePlugins | `ve elasticsearch DescribePlugins` | `DescribePlugins` |
| Uninstall Plugin | UninstallPlugin | `ve elasticsearch UninstallPlugin` | `UninstallPlugin` |
| Enable Kibana | EnableKibana | `ve elasticsearch EnableKibana` | `EnableKibana` |
| Describe Kibana | DescribeKibana | `ve elasticsearch DescribeKibana` | `DescribeKibana` |
| Disable Kibana | DisableKibana | `ve elasticsearch DisableKibana` | `DisableKibana` |

## Request / Response Examples

### CreateInstance

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceName": "prod-search",
  "Version": "7.16",
  "NodeSpec": "es.x4.medium",
  "NodeNumber": 3,
  "StorageSpaceGb": 100,
  "StorageType": "ESSD",
  "VpcId": "vpc-xxx",
  "SubnetId": "subnet-xxx",
  "ChargeType": "PostPaid"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123456",
    "Action": "CreateInstance",
    "Version": "2022-01-01",
    "Service": "elasticsearch",
    "Region": "cn-beijing"
  },
  "Result": {
    "InstanceId": "es-xxx"
  }
}
```

### DescribeInstances

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123457",
    "Action": "DescribeInstances"
  },
  "Result": {
    "Instances": [{
      "InstanceId": "es-xxx",
      "InstanceName": "prod-search",
      "Version": "7.16",
      "Status": "Running",
      "NodeSpec": "es.x4.medium",
      "NodeNumber": 3,
      "StorageSpaceGb": 100,
      "StorageType": "ESSD",
      "VpcId": "vpc-xxx",
      "SubnetId": "subnet-xxx",
      "CreateTime": "2024-05-20T10:00:00+08:00",
      "ExpiredTime": "2025-05-20T10:00:00+08:00",
      "ChargeType": "PostPaid"
    }]
  }
}
```

### CreateIndex

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "IndexName": "logs-2024.05.27",
  "ShardCount": 5,
  "ReplicaCount": 1
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123458",
    "Action": "CreateIndex"
  },
  "Result": {}
}
```

### DescribeIndices

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "IndexName": "logs-2024.05.27"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123459",
    "Action": "DescribeIndices"
  },
  "Result": {
    "IndexName": "logs-2024.05.27",
    "Health": "Green",
    "ShardCount": 5,
    "ReplicaCount": 1,
    "DocCount": 1234567,
    "StorageSizeBytes": 21474836480,
    "CreateTime": "2024-05-27T10:00:00+08:00"
  }
}
```

### CreateSnapshot

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "SnapshotName": "daily-backup-20240527",
  "Indices": "*"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123460",
    "Action": "CreateSnapshot"
  },
  "Result": {
    "SnapshotName": "daily-backup-20240527"
  }
}
```

### DescribeSnapshots

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123461"
  },
  "Result": {
    "Snapshots": [{
      "SnapshotName": "daily-backup-20240527",
      "Status": "SUCCESS",
      "Indices": "logs-2024.05.27,app-errors",
      "StartTime": "2024-05-27T02:00:00+08:00",
      "EndTime": "2024-05-27T02:05:30+08:00",
      "StorageSizeBytes": 5368709120
    }]
  }
}
```

### EnableKibana

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "KibanaUser": "admin",
  "KibanaPassword": "SecurePass123!"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123462"
  },
  "Result": {
    "KibanaEndpoint": "https://es-xxx.cn-beijing.es.volces.com:5601"
  }
}
```

### InstallPlugin

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "PluginName": "analysis-ik",
  "ForceRestart": true
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123463"
  },
  "Result": {}
}
```

### UpgradeVersion

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "TargetVersion": "8.5"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123464"
  },
  "Result": {}
}
```

### ListIndices

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "PageNumber": 1,
  "PageSize": 20
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123465"
  },
  "Result": {
    "TotalCount": 15,
    "Indices": [
      {"IndexName": "logs-2024.05.27", "Health": "Green", "DocCount": 1234567, "StorageSizeBytes": 21474836480},
      {"IndexName": "app-errors", "Health": "Yellow", "DocCount": 50000, "StorageSizeBytes": 1073741824}
    ]
  }
}
```

## Pagination

List operations support pagination via `PageNumber` and `PageSize` parameters:

```json
{
  "Region": "cn-beijing",
  "InstanceId": "es-xxx",
  "PageNumber": 1,
  "PageSize": 20
}
```

Response includes total count:
```json
{
  "Result": {
    "TotalCount": 100,
    "Indices": [...]
  }
}
```

## Required Fields Summary

| Operation | Required Fields |
|-----------|-----------------|
| CreateInstance | Region, InstanceName, Version, NodeSpec, NodeNumber, StorageSpaceGb, VpcId, SubnetId |
| DescribeInstances | Region (InstanceId optional) |
| DeleteInstance | Region, InstanceId |
| CreateIndex | Region, InstanceId, IndexName, ShardCount, ReplicaCount |
| DeleteIndex | Region, InstanceId, IndexName |
| CreateSnapshot | Region, InstanceId, SnapshotName, Indices |
| DescribeSnapshots | Region, InstanceId (SnapshotName optional) |
| DeleteSnapshot | Region, InstanceId, SnapshotName |
| InstallPlugin | Region, InstanceId, PluginName |
| UninstallPlugin | Region, InstanceId, PluginName |
| DescribePlugins | Region, InstanceId (PluginName optional) |
| EnableKibana | Region, InstanceId, KibanaUser, KibanaPassword |
| DescribeKibana | Region, InstanceId |
| DisableKibana | Region, InstanceId |
| RestartInstance | Region, InstanceId |
| UpgradeVersion | Region, InstanceId, TargetVersion |
| ModifyInstanceSpec | Region, InstanceId (NodeSpec or StorageSpaceGb) |
| ModifyNodeCount | Region, InstanceId, NodeNumber |
| ModifyInstanceChargeType | Region, InstanceId, ChargeType |
