# CLI Usage — Volcengine Redis

> **Purpose:** CLI usage reference for Redis operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-06-04

---

## Table of Contents

1. [Instance Commands](#1-instance-commands)
2. [Allow List Commands](#2-allow-list-commands)
3. [Connection Commands](#3-connection-commands)
4. [Performance Commands](#4-performance-commands)
5. [Output Formatting](#5-output-formatting)
6. [Common Patterns](#6-common-patterns)

---

## 1. Instance Commands

### Create Standalone Instance

> Query available versions via `ve redis DescribeDBEngineVersions`

```bash
ve redis CreateDBInstance --body '{
  "InstanceName": "dev-redis",
  "EngineVersion": "{{user.engine_version}}",
  "InstanceClass": "Standalone",
  "NodeNumber": 1,
  "ShardCapacity": 1024,
  "Password": "{{user.password}}",
  "VpcId": "{{user.vpc_id}}",
  "SubnetId": "{{user.subnet_id}}",
  "ChargeType": "PostPaid"
}'
```

### Create Primary-Secondary (HA) Instance

```bash
ve redis CreateDBInstance --body '{
  "InstanceName": "prod-redis",
  "EngineVersion": "{{user.engine_version}}",
  "InstanceClass": "PrimarySecondary",
  "NodeNumber": 2,
  "ShardCapacity": 4096,
  "MultiAZ": "enabled",
  "Password": "{{user.password}}",
  "VpcId": "{{user.vpc_id}}",
  "SubnetId": "{{user.subnet_id}}",
  "ConfigureNodes": [
    {"AZ": "cn-beijing-a"},
    {"AZ": "cn-beijing-b"}
  ],
  "ChargeType": "PostPaid"
}'
```

### Create Sharded Cluster Instance

```bash
ve redis CreateDBInstance --body '{
  "InstanceName": "prod-redis-cluster",
  "EngineVersion": "{{user.engine_version}}",
  "InstanceClass": "ShardedCluster",
  "ShardNumber": 4,
  "ShardCapacity": 8192,
  "Password": "{{user.password}}",
  "VpcId": "{{user.vpc_id}}",
  "SubnetId": "{{user.subnet_id}}",
  "ChargeType": "PostPaid"
}'
```

---

## 2. Allow List Commands

### Create Allow List

```bash
ve redis CreateAllowList --body '{
  "AllowListName": "app-servers",
  "AllowList": ["10.0.2.0/24", "10.0.3.0/24"]
}'
```

### Associate Allow List

```bash
ve redis AssociateAllowList --body '{
  "InstanceId": "{{user.instance_id}}",
  "AllowListIds": ["{{user.allow_list_id}}"]
}'
```

---

## 3. Connection Commands

### Enable Public Access

```bash
# ⚠️ SECURITY: configure allow list FIRST
ve redis CreateDBEndpointPublicAddress --body '{
  "InstanceId": "{{user.instance_id}}",
  "NetworkType": "SingleLine"
}'

# Get public address
ve redis DescribeDBInstanceDetail --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 4. Performance Commands

### Scan Big Keys

```bash
ve redis DescribeBigKeys --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Identify Hot Keys

```bash
ve redis DescribeHotKeys --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 5. Output Formatting

### Instance Table

```bash
ve redis DescribeDBInstances --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.Instances[] |
  [.InstanceId, .InstanceName, .Status, .EngineVersion, (.Capacity.Total | tostring) + "MB", .PrivateAddress] |
  @tsv
' | column -t -s $'\t'
```

---

## 6. Common Patterns

### Pattern: Full Redis Setup

```bash
INSTANCE_ID=$(ve redis CreateDBInstance --body '{
  "InstanceName": "prod-redis",
  "EngineVersion": "{{user.engine_version}}",
  "InstanceClass": "PrimarySecondary",
  "NodeNumber": 2,
  "ShardCapacity": 4096,
  "Password": "SecureP@ss123",
  "VpcId": "vpc-xxx",
  "SubnetId": "subnet-xxx",
  "ChargeType": "PostPaid",
  "ConfigureNodes": [{"AZ": "cn-beijing-a"}, {"AZ": "cn-beijing-b"}]
}' | jq -r '.Result.InstanceId')

# Poll until Running
for i in {1..60}; do
  STATUS=$(ve redis DescribeDBInstances --Region "cn-beijing" --InstanceId "$INSTANCE_ID" | jq -r '.Result.Instances[0].Status')
  [ "$STATUS" = "Running" ] && echo "Ready" && break
  sleep 5
done

# Get connection details
CONN=$(ve redis DescribeDBInstances --Region "cn-beijing" --InstanceId "$INSTANCE_ID" | jq -r '
  ".Result.Instances[0] | \(.PrivateAddress):\(.PrivatePort)"')
echo "Connect: redis-cli -h $CONN -a SecureP@ss123"
```

---

*This reference covers `--body JSON` variants and commands not in the main SKILL.md. Basic CRUD CLI → SKILL.md Execution Flows.*