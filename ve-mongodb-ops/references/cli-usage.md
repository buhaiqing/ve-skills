# CLI — MongoDB (`ve mongodb`)

> **jq Paths (`.Result.*`):** `.Instances[].InstanceId` (instance ID)

## Install and Config

- **Install:** See [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **Credentials:** `ve` CLI reads from `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` env vars OR `~/.volcengine/config.json`

## Conventions

- Output is **JSON by default**
- Help: `ve mongodb --help` or `ve mongodb <action> --help`
- CLI: `ve mongodb <action> --parameter value`
- JSON body: `ve mongodb <action> --body '{"key":"value"}'`

## Metadata Queries (not in SKILL.md)

| Goal | Command | Notes |
|------|---------|-------|
| List regions | `ve mongodb DescribeRegions` | All supported regions |
| List zones | `ve mongodb DescribeAvailabilityZones --Region cn-beijing` | AZs in region |
| List specs | `ve mongodb DescribeDBInstanceSpecs --Region cn-beijing --MongoVersion "{{user.mongo_version}}"` | Available node specs |

> Basic CRUD commands → SKILL.md Execution Flows (CLI `--Param` style) + api-sdk-usage.md (`--body` JSON variants)

## CLI Output Examples

### DescribeDBInstanceDetail

```json
{
  "Result": {
    "InstanceId": "mongo-xxx",
    "InstanceName": "my-mongo",
    "InstanceStatus": "RUNNING",
    "MongoVersion": "{{user.mongo_version}}",
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
        "MongoVersion": "{{user.mongo_version}}",
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
3. **JSON parsing:** Pipe to `jq` — `ve mongodb DescribeDBInstances --Region cn-beijing | jq '.Result.Instances[].InstanceId'`
4. **Help on demand:** Add `--help` to any command for parameter reference