# API & SDK Usage — Volcengine Redis

> **Purpose:** Detailed API reference for Redis operations. API 2020-12-07.
> **Version:** 1.1.0
> **Last Updated:** 2026-07-04

---

# Common JSON Paths
# ---
# Instance list: $.Result.Instances[] | $.Result.Instances[].{InstanceId,InstanceName,Status,EngineVersion}
# Instance detail: $.Result.{Status,Capacity.{Total,Used},PrivateAddress,PrivatePort,VpcId}
# Create result: $.Result.{InstanceId,OrderNo}
# Account: $.Result.Accounts[] | $.Result.Accounts[].{AccountName,Role,Status}
# Backup: $.Result.Backups[] | $.Result.Backups[].{BackupId,BackupName,BackupStatus,StartTime,BackupType}
# Parameter: $.Result.Parameters[] | $.Result.Parameters[].{Name,Value,Description,ReadOnly,Range}
# AllowList: $.Result.AllowLists[] (via DescribeAllowLists)

## Table of Contents

1. [API Overview](#1-api-overview)
2. [Instance Management](#2-instance-management)
3. [Account Management](#3-account-management)
4. [Parameter Management](#4-parameter-management)
5. [Backup & Recovery](#5-backup--recovery)
6. [Allow List Management](#6-allow-list-management)
7. [Connection Management](#7-connection-management)
8. [Performance Analysis](#8-performance-analysis)
9. [Response Parsing](#9-response-parsing)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `Redis` |
| API Version | `2020-12-07` |
| Endpoint | `redis.volcengineapi.com` or `redis.{region}.volcengineapi.com` |
| Protocol | HTTPS, JSON body |

### Common Response Structure

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "...", "Service": "Redis" },
  "Result": { ... }
}
```

> **Note:** Redis CLI uses `--body` JSON format for most API calls.

---

## 2. Instance Management

### 2.1 CreateDBInstance

**Request Parameters (JSON body):**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `InstanceName` | String | No | Instance name | `prod-redis` |
| `EngineVersion` | String | Yes | Redis version | `6.0`, `7.0` |
| `InstanceClass` | String | Yes | Standalone/PrimarySecondary/ShardedCluster | `PrimarySecondary` |
| `NodeNumber` | Integer | Yes | Number of nodes | `2` |
| `ShardNumber` | Integer | No | Number of shards (cluster only) | `2` |
| `ShardCapacity` | Integer | Yes | Capacity per shard (MB) | `2048` |
| `Password` | String | Yes | Instance password | — |
| `VpcId` | String | Yes | VPC ID | `vpc-xxx` |
| `SubnetId` | String | Yes | Subnet ID | `subnet-xxx` |
| `ChargeType` | String | No | `PostPaid`/`PrePaid` | `PostPaid` |
| `MultiAZ` | String | No | `enabled`/`disabled` | `enabled` |
| `ConfigureNodes` | Array | No | Node zone placement | `[{AZ: "cn-beijing-a"}, {AZ: "cn-beijing-b"}]` |
| `Tags` | Array | No | Tags | `[{Key: "env", Value: "prod"}]` |

**Response:** `$.Result.InstanceId`, `$.Result.OrderNo`

### 2.2 DescribeDBInstances

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Instances` | Array | Instance list |
| `Instances[].InstanceId` | String | Instance ID |
| `Instances[].InstanceName` | String | Instance name |
| `Instances[].Status` | String | `Running`, `Creating`, `Error` |
| `Instances[].EngineVersion` | String | Redis version |
| `Instances[].InstanceClass` | String | Instance class |
| `Instances[].ShardNumber` | Integer | Number of shards |
| `Instances[].Capacity.Total` | Integer | Total capacity MB |
| `Instances[].Capacity.Used` | Integer | Used capacity MB |
| `Instances[].PrivateAddress` | String | Private connection address |
| `Instances[].PrivatePort` | String | Port (default 6379) |
| `Instances[].VpcId` | String | VPC ID |
| `TotalInstancesNum` | Integer | Total matching |

### 2.3 DeleteDBInstance

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `BackupPointName` | String | No | Final backup name (PrimarySecondary only) |

### 2.4 RestartDBInstance

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |

### 2.5 Spec Change Operations

| Operation | Description |
|-----------|-------------|
| `IncreaseDBInstanceNodeNumber` | Add read replicas or shards |
| `DecreaseDBInstanceNodeNumber` | Remove replicas |
| `ModifyDBInstanceShardCapacity` | Change per-shard memory |
| `ModifyDBInstanceShardNumber` | Change shard count |
| `EnableShardedCluster` | Enable sharding on non-cluster |

---

## 3. Account Management

### 3.1 CreateDBAccount

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `AccountName` | String | Yes | Account name |
| `AccountPassword` | String | Yes | Password (8-32 chars, mixed case+digits) |
| `Role` | String | No | `ReadWrite`/`ReadOnly` |

### 3.2 ListDBAccount

**Response:** `$.Result.Accounts[]` with `AccountName`, `Role`, `Status`

### 3.3 DeleteDBAccount

### 3.4 ModifyDBAccount

---

## 4. Parameter Management

### 4.1 DescribeDBInstanceParams

**Response:** `$.Result.Parameters[]` with `Name`, `Value`, `Description`, `ReadOnly`, `Range`

### 4.2 ModifyDBInstanceParams

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `Parameters` | Array | Yes | `[{Name: "maxmemory-policy", Value: "allkeys-lru"}]` |

---

## 5. Backup & Recovery

### 5.1 CreateBackup

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `BackupName` | String | No | Backup name |

### 5.2 DescribeBackups

**Response:** `$.Result.Backups[]` with `BackupId`, `BackupName`, `BackupStatus`, `StartTime`, `BackupType`

### 5.3 ModifyBackupPlan

### 5.4 RestoreDBInstance

---

## 6. Allow List Management

### 6.1 CreateAllowList

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllowListName` | String | Yes | Allow list name |
| `AllowList` | Array | Yes | IP/CIDR list |
| `AllowListDesc` | String | No | Description |

### 6.2 DescribeAllowLists

### 6.3 ModifyAllowList

### 6.4 AssociateAllowList / DisassociateAllowList

---

## 7. Connection Management

### 7.1 CreateDBEndpointPublicAddress

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `NetworkType` | String | No | `SingleLine` or `BGP` |

### 7.2 DeleteDBEndpointPublicAddress

### 7.3 DescribeDBInstanceBandwidthPerShard

---

## 8. Performance Analysis

### 8.1 DescribeBigKeys

Query top big keys (memory-intensive operation).

### 8.2 DescribeHotKeys

Query top hot keys (frequency-based).

### 8.3 CreateKeyScanJob

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `ScanKeyNumPerSecond` | Integer | No | Keys per second (2000–100000) | Default: 2000 |
| `TimeoutMinutes` | Integer | No | Max scan time (30–1440 min) | Default: 30 |

---

## 9. Response Parsing

```bash
# Extract instance ID
INSTANCE_ID=$(ve redis CreateDBInstance --body '{
  "InstanceName": "demo",
  "EngineVersion": "6.0",
  "InstanceClass": "PrimarySecondary",
  "NodeNumber": 2,
  "ShardCapacity": 2048,
  "Password": "SecureP@ss123",
  "VpcId": "vpc-xxx",
  "SubnetId": "subnet-xxx",
  "ChargeType": "PostPaid"
}' | jq -r '.Result.InstanceId')

# List instances with capacity
ve redis DescribeDBInstances --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.Instances[] | "\(.InstanceId)\t\(.InstanceName)\t\(.Status)\t\(.Capacity.Total)MB\t\(.PrivateAddress):\(.PrivatePort)"'

# Poll until Running
for i in {1..60}; do
  STATUS=$(ve redis DescribeDBInstances --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" \
    | jq -r '.Result.Instances[0].Status')
  [ "$STATUS" = "Running" ] && break
  sleep 5
done
```

---

*This reference document is part of the ve-redis-ops skill.*
