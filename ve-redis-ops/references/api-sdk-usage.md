# Redis API & SDK Usage

## OpenAPI

- **Version:** `2020-12-07`
- **Service:** `Redis`
- **Endpoint:** `redis.{region}.volcengineapi.com`
- **Doc:** https://www.volcengine.com/docs/6293/65741

## SDK Operations Map

| Goal | API Action | Required Fields |
|------|-----------|----------------|
| Create Instance | `CreateDBInstance` | Region, InstanceName, EngineVersion, NodeNumber, ShardCapacity, VpcId, SubnetId, ChargeType, Password |
| Describe Instance | `DescribeDBInstanceDetail` | InstanceId |
| List Instances | `DescribeDBInstances` | Region (optional filters: InstanceId, InstanceName, ZoneId, VpcId, ChargeType, ServiceType) |
| Delete Instance | `DeleteDBInstance` | InstanceId |
| Modify Spec | `ModifyDBInstanceSpec` | InstanceId, ShardCapacity or NodeNumber |
| Restart | `RestartDBInstance` | InstanceId |
| Describe Parameters | `DescribeDBInstanceParameters` | InstanceId |
| Modify Parameters | `ModifyDBInstanceParameters` | InstanceId, Parameters[] |
| Create AllowList | `CreateAllowList` | Region, Name, IPList[] |
| Describe AllowLists | `DescribeAllowLists` | Region |
| Modify AllowList | `ModifyAllowList` | AllowListId, IPList[] |
| Delete AllowList | `DeleteAllowList` | AllowListId |
| Describe Accounts | `DescribeAccounts` | InstanceId |
| Create Account | `CreateAccount` | InstanceId, AccountName, AccountPassword, AccountRole |
| Describe Backups | `DescribeBackups` | InstanceId |
| Create Backup | `CreateBackup` | InstanceId, BackupName |

## Response JSON Paths (from actual API)

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| CreateDBInstance | `$.Result.InstanceId` | InstanceId |
| DescribeDBInstances | `$.Result.Instances[]` | InstanceId, InstanceName, Status, EngineVersion, Capacity, ChargeType |
| DescribeDBInstances | `$.Result.TotalInstancesNum` | int | Total count |
| DescribeDBInstanceDetail | `$.Result.*` | PrivateAddress, PrivatePort, VpcId, ZoneIds, Status |
| DeleteDBInstance | `$.Metadata.RequestId` | RequestId |

## Pagination

DescribeDBInstances supports:
- `PageNumber`: Starting from 1
- `PageSize`: Default 10
- Response: `$.Result.TotalInstancesNum` for total

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/redis"

instance := redis.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
resp, err := instance.Client.Request("redis", "CreateDBInstance", params)
```
