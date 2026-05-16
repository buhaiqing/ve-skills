# CLI — RDS MySQL (`ve rds_mysql`)

## Install and Config

- Install: `ve` CLI from https://github.com/volcengine/volcengine-cli/releases
- Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
- Output: **JSON by default**

## Command Map

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List instances | `ve rds_mysql DescribeDBInstances --Region cn-beijing` | Paginated |
| Instance details | `ve rds_mysql DescribeDBInstanceDetail --InstanceId mysql-xxx` | Full info |
| Create instance | `ve rds_mysql CreateDBInstance --Region cn-beijing --DBEngine Mysql --DBEngineVersion MySQL_8_0 --InstanceName my-db --StorageType LocalSSD --StorageSpace 100 --VpcId vpc-xxx --SubnetId subnet-xxx --ChargeType PostPaid` | JSON output |
| Delete instance | `ve rds_mysql DeleteDBInstance --InstanceId mysql-xxx` | **Irreversible** |
| Modify spec | `ve rds_mysql ModifyDBNodeSpec --InstanceId mysql-xxx --StorageSpace 200 --body '{"NodeInfo": [...]}'` | Requires JSON body |
| Describe params | `ve rds_mysql DescribeDBInstanceParameters --InstanceId mysql-xxx` | MySQL config |
| Modify params | `ve rds_mysql ModifyDBInstanceParameter --InstanceId mysql-xxx --body '{"Parameters": [{"ParameterName":"max_connections","ParameterValue":"2000"}]}'` | |
| List regions | `ve rds_mysql DescribeRegions` | Available regions |
| List IP lists | `ve rds_mysql ListDBInstanceIPLists --InstanceId mysql-xxx` | Whitelist |
| Modify IP list | `ve rds_mysql ModifyDBInstanceIPList --InstanceId mysql-xxx --body '{"IPList": ["10.0.0.0/8"], "ModifyMode": "Cover"}'` | |
| List accounts | `ve rds_mysql DescribeDBAccounts --InstanceId mysql-xxx` | DB users |
| Create account | `ve rds_mysql CreateDBAccount --InstanceId mysql-xxx --AccountName myuser --AccountPassword '***'` | |
| List backups | `ve rds_mysql DescribeBackups --InstanceId mysql-xxx` | Backup history |
| Create backup | `ve rds_mysql CreateBackup --InstanceId mysql-xxx --BackupName manual-backup` | |
| Restore | `ve rds_mysql RestoreToNewInstance --body '{"BackupId":"bk-xxx","InstanceName":"restored-db"}'` | Creates new instance |
| Rebuild | `ve rds_mysql RebuildDBInstance --InstanceId mysql-xxx` | Rebuilds instance |

## Parameter Discovery

```bash
ve rds_mysql --help
ve rds_mysql CreateDBInstance --help
```

## JSON Body Passing

For complex nested parameters, use `--body` with JSON:

```bash
ve rds_mysql CreateDBInstance \
  --Region cn-beijing \
  --body '{
    "DBEngine": "Mysql",
    "DBEngineVersion": "MySQL_8_0",
    "InstanceName": "my-db",
    "NodeInfo": [
      {"NodeType": "Primary", "NodeSpec": "rds.mysql.2c4g", "ZoneId": "cn-beijing-a"},
      {"NodeType": "Secondary", "NodeSpec": "rds.mysql.2c4g", "ZoneId": "cn-beijing-b"}
    ],
    "StorageType": "LocalSSD",
    "StorageSpace": 100,
    "VpcId": "vpc-xxx",
    "SubnetId": "subnet-xxx",
    "ChargeType": "PostPaid"
  }'
```
