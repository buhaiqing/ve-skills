# API & SDK — MongoDB

## Common JSON Paths

```markdown
# Instance detail: $.Result.{InstanceId,InstanceName,InstanceStatus,MongoVersion,NodeSpec,StorageSpaceGB}
# Instance list: $.Result.Instances[] | $.Result.Instances[].{InstanceId,Status,MongoVersion,NodeSpec}
# Create result: $.Result.InstanceId
# Account: $.Result.Accounts[] | $.Result.Accounts[].{AccountName,AccountPrivilege}
# Database: $.Result.Databases[] | $.Result.Databases[].DBName
# Backup: $.Result.Backups[] | $.Result.Backups[].{BackupId,BackupName,BackupStatus}
# Parameter: $.Result.Parameters[] | $.Result.Parameters[].{ParameterName,ParameterValue}
# IP List: $.Result.IPList[] | $.Result.InstanceId (for ModifyDBInstanceIPList)
# Pagination: $.Result.{Total,PageNumber,PageSize}
# Error: $.ResponseMetadata.RequestId
```

---

## OpenAPI

- **Base URL:** `https://mongodb.{region}.volcengineapi.com`
- **API Version:** `2022-01-01`
- **Service ID:** `mongodb`
- **Protocol:** HTTPS, JSON
- **Auth:** Access Key + Secret Key (HMAC-SHA256)

## Request / Response Notes

> See SKILL.md Execution Flows for the complete operation map. This file covers parameter details not in SKILL.md.

### CreateDBInstance

**Required Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| Region | string | Region ID (e.g., cn-beijing) |
| InstanceName | string | 2-128 chars |
| MongoVersion | string | Query via `DescribeDBInstanceSpecs` |
| NodeSpec | string | Query via `DescribeDBInstanceSpecs` |
| StorageSpaceGB | integer | 20-3000 GB |
| NodeNumber | integer | 3 for replica set |
| VpcId | string | VPC ID |
| SubnetId | string | Subnet ID |
| ChargeType | string | `PostPaid` or `PrePaid` |

### DescribeDBInstances

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| Region | string | Yes | Region ID |
| PageNumber | integer | No | Default 1 |
| PageSize | integer | No | Default 10, max 100 |
| InstanceId | string | No | Filter |
| InstanceName | string | No | Filter |
| InstanceStatus | string | No | Filter |

### CreateDBAccount

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| AccountName | string | Yes | 1-63 chars |
| AccountPassword | string | Yes | 8-32 chars |
| AccountPrivilege | string | Yes | ReadWrite, ReadOnly, dbAdmin, root |

### ModifyDBInstanceIPList

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| IPList | array | Yes | CIDR strings |

### ModifyDBInstanceParameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| Parameters | array | Yes | `[{"ParameterName":"...","ParameterValue":"..."}]` |

## Pagination Pattern

```json
{"PageNumber": 1, "PageSize": 100}
```
→ Response: `$.Result.{Total,PageNumber,PageSize}`

## Common Parameter Types

### InstanceStatus
`CREATING` → `RUNNING` → `RESTARTING` / `MODIFYING` / `BACKING_UP` / `RESTORING` / `CONFIG_CHANGING` → `DELETING` → `DELETED` | `ERROR`

### AccountPrivilege
| Privilege | Scope |
|-----------|-------|
| `Read` | Read-only (single DB) |
| `ReadWrite` | Read + Write (single DB) |
| `dbAdmin` | DB admin |
| `userAdmin` | User mgmt |
| `root` | Full access |
| `ReadAnyDatabase` | Read all |
| `ReadWriteAnyDatabase` | Read/write all |

### ChargeType
`PostPaid` — pay-as-you-go | `PrePaid` — subscription

## Error Response Format

```json
{
  "ResponseMetadata": {
    "RequestId": "xxx",
    "Action": "CreateDBInstance",
    "Version": "2022-01-01",
    "Service": "mongodb",
    "Region": "cn-beijing"
  },
  "Error": {
    "Code": "InvalidParameter.InstanceName",
    "Message": "Invalid instance name format"
  }
}
```