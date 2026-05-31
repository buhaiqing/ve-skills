# API & SDK — MongoDB

## OpenAPI

- **Base URL:** `https://mongodb.{region}.volcengineapi.com`
- **API Version:** `2022-01-01`
- **Service ID:** `mongodb`
- **Protocol:** HTTPS, JSON
- **Authentication:** Access Key + Secret Key (HMAC-SHA256)

## SDK Operations Map

| Goal | API Operation | CLI Command | Go SDK Method |
|------|---------------|-------------|---------------|
| **Instance Lifecycle** ||||
| Create instance | CreateDBInstance | `ve mongodb CreateDBInstance` | `Request("mongodb", "CreateDBInstance", params)` |
| Describe instance | DescribeDBInstanceDetail | `ve mongodb DescribeDBInstanceDetail` | `Request("mongodb", "DescribeDBInstanceDetail", params)` |
| List instances | DescribeDBInstances | `ve mongodb DescribeDBInstances` | `Request("mongodb", "DescribeDBInstances", params)` |
| Delete instance | DeleteDBInstance | `ve mongodb DeleteDBInstance` | `Request("mongodb", "DeleteDBInstance", params)` |
| Modify spec | ModifyDBInstanceSpec | `ve mongodb ModifyDBInstanceSpec` | `Request("mongodb", "ModifyDBInstanceSpec", params)` |
| Restart instance | RestartDBInstance | `ve mongodb RestartDBInstance` | `Request("mongodb", "RestartDBInstance", params)` |
| **User Management** ||||
| Create account | CreateDBAccount | `ve mongodb CreateDBAccount` | `Request("mongodb", "CreateDBAccount", params)` |
| Delete account | DeleteDBAccount | `ve mongodb DeleteDBAccount` | `Request("mongodb", "DeleteDBAccount", params)` |
| List accounts | DescribeDBAccounts | `ve mongodb DescribeDBAccounts` | `Request("mongodb", "DescribeDBAccounts", params)` |
| Modify privilege | ModifyDBAccountPrivilege | `ve mongodb ModifyDBAccountPrivilege` | `Request("mongodb", "ModifyDBAccountPrivilege", params)` |
| **Database Management** ||||
| Create database | CreateDatabase | `ve mongodb CreateDatabase` | `Request("mongodb", "CreateDatabase", params)` |
| Delete database | DeleteDatabase | `ve mongodb DeleteDatabase` | `Request("mongodb", "DeleteDatabase", params)` |
| List databases | DescribeDatabases | `ve mongodb DescribeDatabases` | `Request("mongodb", "DescribeDatabases", params)` |
| **Collection Management** ||||
| Create collection | CreateCollection | `ve mongodb CreateCollection` | `Request("mongodb", "CreateCollection", params)` |
| Delete collection | DeleteCollection | `ve mongodb DeleteCollection` | `Request("mongodb", "DeleteCollection", params)` |
| List collections | DescribeCollections | `ve mongodb DescribeCollections` | `Request("mongodb", "DescribeCollections", params)` |
| **Backup Management** ||||
| Create backup | CreateBackup | `ve mongodb CreateBackup` | `Request("mongodb", "CreateBackup", params)` |
| List backups | DescribeBackups | `ve mongodb DescribeBackups` | `Request("mongodb", "DescribeBackups", params)` |
| Restore instance | RestoreDBInstance | `ve mongodb RestoreDBInstance` | `Request("mongodb", "RestoreDBInstance", params)` |
| **Network & Security** ||||
| Get IP list | DescribeDBInstanceIPList | `ve mongodb DescribeDBInstanceIPList` | `Request("mongodb", "DescribeDBInstanceIPList", params)` |
| Modify IP list | ModifyDBInstanceIPList | `ve mongodb ModifyDBInstanceIPList` | `Request("mongodb", "ModifyDBInstanceIPList", params)` |
| **Parameter Management** ||||
| List parameters | DescribeDBInstanceParameters | `ve mongodb DescribeDBInstanceParameters` | `Request("mongodb", "DescribeDBInstanceParameters", params)` |
| Modify parameters | ModifyDBInstanceParameters | `ve mongodb ModifyDBInstanceParameters` | `Request("mongodb", "ModifyDBInstanceParameters", params)` |
| **Metadata** ||||
| List regions | DescribeRegions | `ve mongodb DescribeRegions` | `Request("mongodb", "DescribeRegions", params)` |
| List zones | DescribeAvailabilityZones | `ve mongodb DescribeAvailabilityZones` | `Request("mongodb", "DescribeAvailabilityZones", params)` |
| List specs | DescribeDBInstanceSpecs | `ve mongodb DescribeDBInstanceSpecs` | `Request("mongodb", "DescribeDBInstanceSpecs", params)` |

## Request / Response Notes

### CreateDBInstance

**Required Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| Region | string | Region ID (e.g., cn-beijing) |
| InstanceName | string | Instance name (2-128 chars) |
| MongoVersion | string | 4.0, 4.2, 4.4, 5.0, 6.0 |
| NodeSpec | string | Node specification |
| StorageSpaceGB | integer | 20-3000 GB |
| NodeNumber | integer | 3 for replica set |
| VpcId | string | VPC ID |
| SubnetId | string | Subnet ID |
| ChargeType | string | PostPaid or PrePaid |

**Response:**
| Field | Path | Type |
|-------|------|------|
| InstanceId | `$.Result.InstanceId` | string |

### DescribeDBInstances

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| Region | string | Yes | Region ID |
| PageNumber | integer | No | Default 1 |
| PageSize | integer | No | Default 10, max 100 |
| InstanceId | string | No | Filter by ID |
| InstanceName | string | No | Filter by name |
| InstanceStatus | string | No | Filter by status |

**Response:**
| Field | Path | Type |
|-------|------|------|
| Instances | `$.Result.Instances[]` | array |
| InstanceId | `$.Result.Instances[].InstanceId` | string |
| InstanceName | `$.Result.Instances[].InstanceName` | string |
| InstanceStatus | `$.Result.Instances[].InstanceStatus` | string |
| MongoVersion | `$.Result.Instances[].MongoVersion` | string |
| NodeSpec | `$.Result.Instances[].NodeSpec` | string |
| StorageSpaceGB | `$.Result.Instances[].StorageSpaceGB` | integer |
| VpcId | `$.Result.Instances[].VpcId` | string |
| ZoneId | `$.Result.Instances[].ZoneId` | string |
| ConnectionString | `$.Result.Instances[].ConnectionString` | string |
| Port | `$.Result.Instances[].Port` | integer |

### CreateDBAccount

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| AccountName | string | Yes | Username (1-63 chars) |
| AccountPassword | string | Yes | Password (8-32 chars) |
| AccountPrivilege | string | Yes | ReadWrite, ReadOnly, dbAdmin, root |

### ModifyDBInstanceIPList

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| IPList | array | Yes | Array of CIDR strings |

Example: `{"IPList": ["10.0.0.0/8", "192.168.1.0/24"]}`

### ModifyDBInstanceParameters

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| InstanceId | string | Yes | Instance ID |
| Parameters | array | Yes | Array of parameter objects |

Example:
```json
{
  "Parameters": [
    {"ParameterName": "maxConnections", "ParameterValue": "2000"},
    {"ParameterName": "slowOpThresholdMs", "ParameterValue": "100"}
  ]
}
```

## Pagination Pattern

All list operations use standard pagination:

```json
{
  "PageNumber": 1,
  "PageSize": 100
}
```

Response includes:
| Field | Path | Description |
|-------|------|-------------|
| Total | `$.Result.Total` | Total item count |
| PageNumber | `$.Result.PageNumber` | Current page |
| PageSize | `$.Result.PageSize` | Items per page |

## Common Parameter Types

### MongoVersion
- `4.0`
- `4.2`
- `4.4`
- `5.0`
- `6.0`

### InstanceStatus
- `CREATING`
- `RUNNING`
- `RESTARTING`
- `MODIFYING`
- `BACKING_UP`
- `RESTORING`
- `CONFIG_CHANGING`
- `DELETING`
- `DELETED`
- `ERROR`

### ChargeType
- `PostPaid` - Pay as you go
- `PrePaid` - Subscription

### AccountPrivilege
- `Read` - Read-only access
- `ReadWrite` - Read and write access
- `dbAdmin` - Database administration
- `userAdmin` - User management
- `root` - Full administrative access
- `ReadAnyDatabase` - Read all databases
- `ReadWriteAnyDatabase` - Read/write all databases

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
