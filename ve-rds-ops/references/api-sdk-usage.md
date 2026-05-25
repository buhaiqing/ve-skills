# API & SDK Usage — Volcengine RDS MySQL

> **Purpose:** Detailed API reference for RDS MySQL operations. API V2 (2022-01-01).
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [API Overview](#1-api-overview)
2. [Instance Management](#2-instance-management)
3. [Database Management](#3-database-management)
4. [Account Management](#4-account-management)
5. [Parameter Management](#5-parameter-management)
6. [Backup and Recovery](#6-backup-and-recovery)
7. [Allow List Management](#7-allow-list-management)
8. [Response Parsing](#8-response-parsing)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `rds_mysql` |
| API Version | `2022-01-01` |
| Endpoint | `rds-mysql.volcengineapi.com` or `rds-mysql.{region}.volcengineapi.com` |
| Protocol | HTTPS, JSON body |

### Common Response Structure

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "...", "Service": "rds_mysql" },
  "Result": { ... }
}
```

---

## 2. Instance Management

### 2.1 CreateDBInstance

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `InstanceName` | String | No | Instance name | `prod-mysql` |
| `DBEngineVersion` | String | Yes | MySQL version | `MySQL_8_0` |
| `InstanceType` | String | Yes | Single/HA/MultiNode | `HA` |
| `NodeSpec` | String | Yes | Node specification | `rds.mysql.2c4g` |
| `StorageSpace` | Integer | Yes | Storage GB | `100` |
| `StorageType` | String | Yes | LocalSSD or ESSD | `ESSD` |
| `VpcId` | String | Yes | VPC ID | `vpc-xxx` |
| `SubnetId` | String | Yes | Subnet ID | `subnet-xxx` |
| `ChargeType` | String | No | PostPaid/PrePaid | `PostPaid` |
| `LowerCaseTableNames` | String | No | Table name case | `GlobalCaseSensitive=0` |
| `ProjectName` | String | No | Project name | `default` |
| `NodeInfo` | Array | Yes | Nodes with zones | `[{ZoneId: "cn-beijing-a"}]` |
| `AllowListIds` | Array | No | Allow list IDs | `["acl-xxx"]` |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `InstanceId` | String | RDS instance ID |
| `OrderId` | String | Order number |

### 2.2 DescribeDBInstances

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RegionId` | String | Yes | Region ID |
| `InstanceId` | String | No | Filter by instance ID |
| `InstanceName` | String | No | Filter by name (fuzzy) |
| `InstanceStatus` | String | No | Filter by status |
| `VpcId` | String | No | Filter by VPC |
| `PageNumber` | Integer | No | Page number |
| `PageSize` | Integer | No | Page size (1-100) |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Instances` | Array | Instance list |
| `Instances[].InstanceId` | String | Instance ID |
| `Instances[].InstanceName` | String | Instance name |
| `Instances[].InstanceStatus` | String | `Running`, `Creating`, `Error` |
| `Instances[].DBEngineVersion` | String | MySQL version |
| `Instances[].InstanceType` | String | Single/HA/MultiNode |
| `Instances[].NodeSpec` | String | Node specification |
| `Instances[].PrimaryIp` | String | Primary node IP |
| `Instances[].Port` | Integer | Port (default 3306) |
| `Instances[].VpcId` | String | VPC ID |
| `Instances[].RegionId` | String | Region ID |
| `Instances[].StorageSpace` | Integer | Storage GB |
| `Instances[].StorageType` | String | LocalSSD/ESSD |
| `TotalCount` | Integer | Total matching |

### 2.3 DeleteDBInstance

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `DataKeepPolicy` | String | Yes | `Last`/`All`/`None` |
| `DataKeepDays` | Integer | No | Backup retention days (default 7) |

**Response:** Empty Result.

### 2.4 RestartDBInstance

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |

### 2.5 ModifyDBInstanceSpec

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `NodeSpec` | String | Yes | New node specification |
| `IsDowngrade` | Boolean | No | Whether downgrading |
| `ApplyImmediately` | Boolean | No | Apply now or in maintenance window |

---

## 3. Database Management

### 3.1 CreateDB

| Parameter | Type | Required | Description | Default |
|-----------|------|----------|-------------|---------|
| `InstanceId` | String | Yes | Instance ID | — |
| `DBName` | String | Yes | Database name | — |
| `CharacterSet` | String | No | Character set | `utf8mb4` |
| `Description` | String | No | Description | — |

**Response:** `$.Result.DBName`

### 3.2 DescribeDatabases

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `DBName` | String | No | Filter by name (fuzzy) |

**Response:** `$.Result.Databases[]`

### 3.3 DeleteDB

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `DBName` | String | Yes | Database name |

---

## 4. Account Management

### 4.1 CreateAccount

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `AccountName` | String | Yes | Account name |
| `AccountPassword` | String | Yes | Password |
| `AccountDesc` | String | No | Description |
| `AccountType` | String | No | `Super`/`Normal` |

### 4.2 ListDBAccounts (DescribeAccounts)

**Response:** `$.Result.Accounts[]` with fields: `AccountName`, `AccountStatus`, `AccountType`, `AccountPrivileges`

### 4.3 ResetAccountPassword

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `AccountName` | String | Yes | Account name |
| `AccountPassword` | String | Yes | New password |

### 4.4 GrantAccountPrivileges

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `AccountName` | String | Yes | Account name |
| `Privileges` | Array | Yes | Privileges |

Privilege format:
```json
[{"DBName": "mydb", "AccountPrivilege": "ReadOnly"}]
```

Privilege values: `ReadOnly`, `ReadWrite`, `ReadWriteDDL`

### 4.5 RevokeAccountPrivileges

Same parameters as GrantAccountPrivileges.

---

## 5. Parameter Management

### 5.1 DescribeDBInstanceParameters

**Response:** `$.Result.Parameters[]` with fields: `ParameterName`, `ParameterValue`, `ParameterDefaultValue`, `ForceRestart`, `CheckingCode`, `ParameterDescription`

### 5.2 ModifyDBInstanceParameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `Parameters` | Array | Yes | Parameters to modify |

Parameters format:
```json
[{"ParameterName": "max_connections", "ParameterValue": "1500"}]
```

---

## 6. Backup and Recovery

### 6.1 CreateBackup

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `BackupName` | String | No | Backup name |
| `BackupStrategy` | String | No | `Snapshot` or `Physical` |

**Response:** `$.Result.BackupId`

### 6.2 DescribeBackups

**Response:** `$.Result.Backups[]` with fields: `BackupId`, `BackupName`, `BackupStatus`, `BackupStartTime`, `BackupEndTime`, `BackupSize`, `BackupType`

### 6.3 DeleteBackup

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Instance ID |
| `BackupId` | String | Yes | Backup ID |

### 6.4 RestoreDBInstance

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `InstanceId` | String | Yes | Source instance ID |
| `BackupId` | String | Yes | Backup ID |
| `RestoreType` | String | Yes | `Backup` or `PITR` |
| `NewInstanceName` | String | No | New instance name |

---

## 7. Allow List Management

### 7.1 DescribeAllowLists

**Response:** `$.Result.AllowLists[]` with fields: `AllowListId`, `AllowListName`, `AllowList`, `AllowListType`, `AssociatedInstances`

### 7.2 CreateAllowList

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllowListName` | String | Yes | Allow list name |
| `AllowList` | Array | Yes | IP list |
| `AllowListType` | String | No | `IPv4` |

### 7.3 ModifyAllowList

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllowListId` | String | Yes | Allow list ID |
| `AllowList` | Array | Yes | New IP list |

---

## 8. Response Parsing

```bash
# Extract instance ID
INSTANCE_ID=$(ve rds_mysql CreateDBInstance --Region "$VOLCENGINE_REGION" --body '...' \
  | jq -r '.Result.InstanceId')

# List instances with details
ve rds_mysql DescribeDBInstances --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.Instances[] | "\(.InstanceId)\t\(.InstanceName)\t\(.InstanceStatus)\t\(.PrimaryIp):\(.Port)\t\(.DBEngineVersion)"'

# Poll until Running
for i in {1..60}; do
  STATUS=$(ve rds_mysql DescribeDBInstances --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" \
    | jq -r '.Result.Instances[0].InstanceStatus')
  [ "$STATUS" = "Running" ] && break
  sleep 10
done
```

---

*This reference document is part of the ve-rds-ops skill.*
