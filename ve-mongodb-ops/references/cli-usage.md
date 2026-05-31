# CLI — MongoDB (`ve mongodb`)

## Install and Config

- **Install:** See [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`

## Conventions

- Output is **JSON by default**
- Help: `ve mongodb --help` or `ve mongodb <action> --help`
- CLI invocation: `ve mongodb <action> --parameter value`
- JSON body: `ve mongodb <action> --body '{"key":"value"}'`

## CLI Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateDBInstance | Yes | Full support |
| DescribeDBInstanceDetail | Yes | Full support |
| DescribeDBInstances | Yes | Full support |
| DeleteDBInstance | Yes | Full support |
| ModifyDBInstanceSpec | Yes | Full support |
| RestartDBInstance | Yes | Full support |
| CreateDBAccount | Yes | Full support |
| DeleteDBAccount | Yes | Full support |
| DescribeDBAccounts | Yes | Full support |
| ModifyDBAccountPrivilege | Yes | Full support |
| CreateDatabase | Yes | Full support |
| DeleteDatabase | Yes | Full support |
| DescribeDatabases | Yes | Full support |
| CreateCollection | Yes | Full support |
| DeleteCollection | Yes | Full support |
| DescribeCollections | Yes | Full support |
| CreateBackup | Yes | Full support |
| DescribeBackups | Yes | Full support |
| RestoreDBInstance | Yes | Full support |
| DescribeDBInstanceIPList | Yes | Full support |
| ModifyDBInstanceIPList | Yes | Full support |
| DescribeDBInstanceParameters | Yes | Full support |
| ModifyDBInstanceParameters | Yes | Full support |
| DescribeRegions | Yes | Full support |
| DescribeAvailabilityZones | Yes | Full support |
| DescribeDBInstanceSpecs | Yes | Full support |

## Command Map

### Instance Lifecycle

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Create | `ve mongodb CreateDBInstance --Region cn-beijing --InstanceName mymongo --MongoVersion 5.0 --NodeSpec mongo.2c4g --StorageSpaceGB 100 --NodeNumber 3 --VpcId vpc-xxx --SubnetId subnet-xxx --ChargeType PostPaid` | Returns InstanceId |
| Describe | `ve mongodb DescribeDBInstanceDetail --InstanceId mongo-xxx` | Full details |
| List | `ve mongodb DescribeDBInstances --Region cn-beijing --PageNumber 1 --PageSize 100` | Paginated list |
| Delete | `ve mongodb DeleteDBInstance --InstanceId mongo-xxx` | Irreversible |
| Modify | `ve mongodb ModifyDBInstanceSpec --InstanceId mongo-xxx --NodeSpec mongo.4c8g --StorageSpaceGB 200` | May require restart |
| Restart | `ve mongodb RestartDBInstance --InstanceId mongo-xxx` | Brief downtime |

### User Management

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Create user | `ve mongodb CreateDBAccount --InstanceId mongo-xxx --AccountName myuser --AccountPassword mypass123 --AccountPrivilege ReadWrite` | Password 8-32 chars |
| Delete user | `ve mongodb DeleteDBAccount --InstanceId mongo-xxx --AccountName myuser` | Immediate |
| List users | `ve mongodb DescribeDBAccounts --InstanceId mongo-xxx` | All accounts |
| Modify privilege | `ve mongodb ModifyDBAccountPrivilege --InstanceId mongo-xxx --AccountName myuser --AccountPrivilege ReadOnly` | Change access |

### Database Management

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Create DB | `ve mongodb CreateDatabase --InstanceId mongo-xxx --DBName mydb` | Logical database |
| Delete DB | `ve mongodb DeleteDatabase --InstanceId mongo-xxx --DBName mydb` | Drops all collections |
| List DBs | `ve mongodb DescribeDatabases --InstanceId mongo-xxx` | All databases |

### Collection Management

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Create collection | `ve mongodb CreateCollection --InstanceId mongo-xxx --DBName mydb --CollectionName mycoll` | Empty collection |
| Delete collection | `ve mongodb DeleteCollection --InstanceId mongo-xxx --DBName mydb --CollectionName mycoll` | Irreversible |
| List collections | `ve mongodb DescribeCollections --InstanceId mongo-xxx --DBName mydb` | All collections in DB |

### Backup Management

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Create backup | `ve mongodb CreateBackup --InstanceId mongo-xxx --BackupName mybackup` | Manual backup |
| List backups | `ve mongodb DescribeBackups --InstanceId mongo-xxx` | All backups |
| Restore | `ve mongodb RestoreDBInstance --InstanceId mongo-xxx --BackupId backup-xxx` | From backup |

### Network & Security

| Goal | Example Command | Notes |
|------|-----------------|-------|
| Get IP list | `ve mongodb DescribeDBInstanceIPList --InstanceId mongo-xxx` | Current whitelist |
| Modify IP list | `ve mongodb ModifyDBInstanceIPList --InstanceId mongo-xxx --IPList '["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]'` | JSON array |

### Parameter Management

| Goal | Example Command | Notes |
|------|-----------------|-------|
| List params | `ve mongodb DescribeDBInstanceParameters --InstanceId mongo-xxx` | All parameters |
| Modify params | `ve mongodb ModifyDBInstanceParameters --InstanceId mongo-xxx --Parameters '[{"ParameterName":"maxConnections","ParameterValue":"2000"}]'` | JSON array |

### Metadata Queries

| Goal | Example Command | Notes |
|------|-----------------|-------|
| List regions | `ve mongodb DescribeRegions` | All regions |
| List zones | `ve mongodb DescribeAvailabilityZones --Region cn-beijing` | AZs in region |
| List specs | `ve mongodb DescribeDBInstanceSpecs --Region cn-beijing --MongoVersion 5.0` | Available specs |

## CLI Output Examples

### DescribeDBInstanceDetail

```json
{
  "Result": {
    "InstanceId": "mongo-xxx",
    "InstanceName": "my-mongo",
    "InstanceStatus": "RUNNING",
    "MongoVersion": "5.0",
    "NodeSpec": "mongo.2c4g",
    "NodeNumber": 3,
    "StorageSpaceGB": 100,
    "StorageType": "ESSD_PL1",
    "VpcId": "vpc-xxx",
    "SubnetId": "subnet-xxx",
    "ZoneId": "cn-beijing-a",
    "ConnectionString": "mongo-xxx.mongodb.volces.com",
    "Port": 27017,
    "ChargeType": "PostPaid",
    "CreateTime": "2026-05-27T10:00:00+08:00"
  }
}
```

### DescribeDBInstances

```json
{
  "Result": {
    "Total": 2,
    "PageNumber": 1,
    "PageSize": 10,
    "Instances": [
      {
        "InstanceId": "mongo-xxx",
        "InstanceName": "my-mongo",
        "InstanceStatus": "RUNNING",
        "MongoVersion": "5.0",
        "NodeSpec": "mongo.2c4g",
        "StorageSpaceGB": 100,
        "ConnectionString": "mongo-xxx.mongodb.volces.com",
        "Port": 27017
      }
    ]
  }
}
```

## Tips

1. **Filter lists:** Use `--InstanceName`, `--InstanceStatus` to filter results
2. **Pagination:** Always use `--PageNumber` and `--PageSize` for lists
3. **JSON parsing:** Pipe output to `jq` for easier parsing
   ```bash
   ve mongodb DescribeDBInstances --Region cn-beijing | jq '.Result.Instances[].InstanceId'
   ```
4. **Help on demand:** Add `--help` to any command for parameter reference
