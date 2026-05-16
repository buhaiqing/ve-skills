# CLI — Redis (`ve redis`)

## Install and Config

- Install: `ve` CLI from https://github.com/volcengine/volcengine-cli/releases
- Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
- Output: **JSON by default**

## Command Map

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List instances | `ve redis DescribeDBInstances --Region cn-beijing` | Paginated |
| Instance details | `ve redis DescribeDBInstanceDetail --InstanceId redis-xxx` | Full info |
| Create instance | `ve redis CreateDBInstance --Region cn-beijing --InstanceName my-redis --EngineVersion 5.0 --NodeNumber 2 --ShardCapacity 2048 --VpcId vpc-xxx --SubnetId subnet-xxx --ChargeType PostPaid --Password '***'` | JSON output |
| Delete instance | `ve redis DeleteDBInstance --InstanceId redis-xxx` | **Irreversible** |
| Modify spec | `ve redis ModifyDBInstanceSpec --InstanceId redis-xxx --body '{"ShardCapacity": 4096}'` | May cause downtime |
| Restart | `ve redis RestartDBInstance --InstanceId redis-xxx` | Brief interruption |
| List allowlists | `ve redis DescribeAllowLists --Region cn-beijing` | Whitelist info |
| Create allowlist | `ve redis CreateAllowList --Region cn-beijing --Name my-whitelist --IPList '["10.0.0.0/8"]'` | |
| List accounts | `ve redis DescribeAccounts --InstanceId redis-xxx` | DB accounts |
| Create account | `ve redis CreateAccount --InstanceId redis-xxx --AccountName myuser --AccountPassword '***' --AccountRole Standard` | |
| List backups | `ve redis DescribeBackups --InstanceId redis-xxx` | Backup history |
| Create backup | `ve redis CreateBackup --InstanceId redis-xxx --BackupName manual-backup` | Manual backup |
| Describe params | `ve redis DescribeDBInstanceParameters --InstanceId redis-xxx` | Config params |

## Parameter Discovery

```bash
ve redis --help
ve redis CreateDBInstance --help
ve redis DescribeDBInstances --help
```

## JSON Output Parsing

```bash
# Parse instance ID
ve redis CreateDBInstance ... | jq -r '.Result.InstanceId'

# Parse instance status
ve redis DescribeDBInstanceDetail --InstanceId redis-xxx | jq -r '.Result.Status'

# Parse connection address
ve redis DescribeDBInstanceDetail --InstanceId redis-xxx | jq -r '.Result.PrivateAddress'
```
