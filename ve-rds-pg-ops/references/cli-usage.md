# CLI — RDS PostgreSQL (`ve rds_postgresql`)

## Install and Config

- Install: `ve` CLI from https://github.com/volcengine/volcengine-cli/releases
- Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
- Output: **JSON by default**

## Command Map

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List instances | `ve rds_postgresql DescribeDBInstances --Region cn-beijing` | Paginated |
| Instance details | `ve rds_postgresql DescribeDBInstanceDetail --InstanceId pg-xxx` | Full info |
| Create instance | `ve rds_postgresql CreateDBInstance --Region cn-beijing --DbEngineVersion PostgreSQL_14 --NodeSpec rds.postgres.2c4g --PrimaryZoneId cn-beijing-a --SecondaryZoneId cn-beijing-b --StorageSpace 100 --SubnetId subnet-xxx --InstanceName my-pg --ChargeInfo.ChargeType PostPaid` | JSON output |
| Delete instance | `ve rds_postgresql DeleteDBInstance --InstanceId pg-xxx` | **Irreversible** |
| Modify spec | `ve rds_postgresql ModifyDBInstanceSpec --InstanceId pg-xxx --NodeSpec rds.postgres.4c8g --StorageSpace 200` | Downtime expected |
| Describe params | `ve rds_postgresql DescribeDBInstanceParameters --InstanceId pg-xxx` | PG config |
| Modify params | `ve rds_postgresql ModifyDBInstanceParameter --InstanceId pg-xxx --body '{"Parameters": [{"Name":"shared_buffers","Value":"2GB"}]}'` | |
| Create RO node | `ve rds_postgresql CreateReadonlyInstance --SrcInstanceInstanceId pg-xxx --NodeSpec rds.postgres.2c4g --ZoneId cn-beijing-a` | Read scaling |
| List accounts | `ve rds_postgresql DescribeAccounts --InstanceId pg-xxx` | DB users |
| Create account | `ve rds_postgresql CreateAccount --InstanceId pg-xxx --AccountName myuser --AccountPassword '***' --AccountType Normal` | |
| List backups | `ve rds_postgresql DescribeBackups --InstanceId pg-xxx` | Backup history |
| Create backup | `ve rds_postgresql CreateBackup --InstanceId pg-xxx --BackupName manual-backup` | |
| Restore | `ve rds_postgresql RestoreToNewInstance --body '{"BackupId":"bk-xxx","InstanceName":"restored-pg",...}'` | New instance |
| Rebuild | `ve rds_postgresql RebuildDBInstance --InstanceId pg-xxx` | Rebuild |

## Key Differences from MySQL CLI

| Aspect | MySQL | PostgreSQL |
|--------|-------|-----------|
| Service slug | `rds_mysql` | `rds_postgresql` |
| Version param | `--DBEngineVersion MySQL_8_0` | `--DbEngineVersion PostgreSQL_14` |
| Spec format | `rds.mysql.2c4g` | `rds.postgres.2c4g` |
| Zone config | `--NodeInfo [...]` | `--PrimaryZoneId` + `--SecondaryZoneId` |
| Storage type | LocalSSD or ESSD | LocalSSD only |
