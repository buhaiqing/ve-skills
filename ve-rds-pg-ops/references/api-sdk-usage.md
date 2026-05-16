# RDS PostgreSQL API & SDK Usage

## OpenAPI

- **Version:** `2022-01-01`
- **Service:** `rds_postgresql`
- **Endpoint:** `rds-postgresql.{region}.volcengineapi.com`
- **Doc:** https://www.volcengine.com/docs/6438

## SDK Operations Map

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| Create Instance | `CreateDBInstance` | DbEngineVersion, NodeSpec, PrimaryZoneId, SecondaryZoneId, StorageSpace, SubnetId, InstanceName, ChargeInfo |
| Describe Instance | `DescribeDBInstanceDetail` | InstanceId |
| List Instances | `DescribeDBInstances` | Region (optional filters) |
| Delete Instance | `DeleteDBInstance` | InstanceId |
| Modify Spec | `ModifyDBInstanceSpec` | InstanceId, NodeSpec or StorageSpace |
| Describe Parameters | `DescribeDBInstanceParameters` | InstanceId |
| Modify Parameter | `ModifyDBInstanceParameter` | InstanceId, Parameters[] |
| Describe Regions | `DescribeRegions` | None |
| Describe AZs | `DescribeAvailabilityZones` | RegionId |
| Create Read-Only | `CreateReadonlyInstance` | SrcInstanceInstanceId, NodeSpec, ZoneId |
| Describe Accounts | `DescribeAccounts` | InstanceId |
| Create Account | `CreateAccount` | InstanceId, AccountName, AccountPassword |
| Describe Backups | `DescribeBackups` | InstanceId |
| Create Backup | `CreateBackup` | InstanceId, BackupName |
| Restore Instance | `RestoreToNewInstance` | BackupId or RestoreTime, InstanceName, NodeSpec, PrimaryZoneId |
| Rebuild Instance | `RebuildDBInstance` | InstanceId |
| Describe Schemas | `DescribeSchemas` | InstanceId, DbName |
| Create Schema | `CreateSchema` | InstanceId, DbName, SchemaName, Owner |
| Create Database | `CreateDatabase` | InstanceId, DbName, CType, Collate |

## Response JSON Paths

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| CreateDBInstance | `$.Result.InstanceId` | InstanceId |
| DescribeDBInstances | `$.Result.Instances[]` | InstanceId, InstanceName, InstanceStatus, DbEngineVersion, NodeSpec, StorageSpace |
| DescribeDBInstanceDetail | `$.Result.*` | Endpoints[], Nodes[], Parameters[] |
| DescribeDBInstanceParameters | `$.Result.Parameters[]` | Name, Value, IsModifiable, DefaultValue |
| DeleteDBInstance | `$.Metadata.RequestId` | RequestId |

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/rds_postgresql"

instance := rds_postgresql.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
resp, err := instance.Client.Request("rds_postgresql", "CreateDBInstance", params)
```
