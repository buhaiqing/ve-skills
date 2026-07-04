# RDS MySQL API & SDK Usage

## OpenAPI

- **Version:** `2022-01-01`
- **Service:** `rds_mysql`
- **Endpoint:** `rds-mysql.{region}.volcengineapi.com`
- **Doc:** https://www.volcengine.com/docs/6313/170637

## SDK Operations Map

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| Create Instance | `CreateDBInstance` | DBEngine, DBEngineVersion, NodeInfo, StorageType, StorageSpace, VpcId, SubnetId, ChargeType, InstanceName |
| Describe Instance | `DescribeDBInstanceDetail` | InstanceId |
| List Instances | `DescribeDBInstances` | Region (optional filters) |
| Delete Instance | `DeleteDBInstance` | InstanceId |
| Modify Node Spec | `ModifyDBNodeSpec` | InstanceId, NodeInfo or StorageSpace |
| Describe Parameters | `DescribeDBInstanceParameters` | InstanceId |
| Modify Parameters | `ModifyDBInstanceParameter` | InstanceId, Parameters[] |
| Describe Parameter Log | `DescribeDBInstanceParametersLog` | InstanceId, StartTime, EndTime |
| Describe Regions | `DescribeRegions` | None |
| Describe AZs | `DescribeAvailabilityZones` | RegionId |
| List IP Lists | `ListDBInstanceIPLists` | InstanceId |
| Modify IP List | `ModifyDBInstanceIPList` | InstanceId, IPList, ModifyMode |
| Describe Accounts | `DescribeDBAccounts` | InstanceId |
| Create Account | `CreateDBAccount` | InstanceId, AccountName, AccountPassword |
| Delete Account | `DeleteDBAccount` | InstanceId, AccountName |
| Describe Backups | `DescribeBackups` | InstanceId |
| Create Backup | `CreateBackup` | InstanceId, BackupName |
| Restore Instance | `RestoreToNewInstance` | BackupId, InstanceName, NodeInfo |
| Rebuild Instance | `RebuildDBInstance` | InstanceId |
| Describe Minor Versions | `DescribeDBInstanceEngineMinorVersions` | InstanceIds[] |

## Response JSON Paths

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| CreateDBInstance | `$.Result.InstanceId` | InstanceId |
| DescribeDBInstances | `$.Result.Instances[]` | InstanceId, InstanceName, InstanceStatus, DBEngineVersion, StorageType, StorageSpace |
| DescribeDBInstanceDetail | `$.Result.*` | ConnectionStrings, NodeInfo, Tags, ZoneId |
| DescribeDBInstanceParameters | `$.Result.Parameters[]` | ParameterName, ParameterValue, ForceRestart, CheckingCode |
| DescribeDBInstanceParametersLog | `$.Result.ParameterChangeLogs[]` | ParameterName, Old/New Value, ModifyTime, Status |
| DeleteDBInstance | `$.Metadata.RequestId` | RequestId |
| Pagination | `$.Result.TotalCount` | Total items |

## Pagination

DescribeDBInstances supports:
- `PageNumber`: Starting from 1
- `PageSize`: Default 10
- Response: TotalCount for total items

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/rds_mysql"

instance := rds_mysql.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
resp, err := instance.Client.Request("rds_mysql", "CreateDBInstance", params)
```
