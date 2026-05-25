# CLI Usage — Volcengine Redis

> **Purpose:** CLI usage reference for Redis operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Instance Commands](#1-instance-commands)
2. [Account Commands](#2-account-commands)
3. [Parameter Commands](#3-parameter-commands)
4. [Backup Commands](#4-backup-commands)
5. [Allow List Commands](#5-allow-list-commands)
6. [Connection Commands](#6-connection-commands)
7. [Performance Commands](#7-performance-commands)
8. [Output Formatting](#8-output-formatting)
9. [Common Patterns](#9-common-patterns)

---

## 1. Instance Commands

### List All Instances

```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by VPC

```bash
ve redis DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Create Standalone Instance

```bash
ve redis CreateDBInstance --body '{
  "InstanceName": "dev-redis",
  "EngineVersion": "7.0",
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
  "EngineVersion": "6.0",
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
  "EngineVersion": "7.0",
  "InstanceClass": "ShardedCluster",
  "ShardNumber": 4,
  "ShardCapacity": 8192,
  "Password": "{{user.password}}",
  "VpcId": "{{user.vpc_id}}",
  "SubnetId": "{{user.subnet_id}}",
  "ChargeType": "PostPaid"
}'
```

### Delete Instance

```bash
# ⚠️ IRREVERSIBLE
ve redis DeleteDBInstance --body '{
  "InstanceId": "{{user.instance_id}}",
  "BackupPointName": "pre-delete-backup"
}'
```

---

## 2. Account Commands

### List Accounts

```bash
ve redis ListDBAccount --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Create Account

```bash
ve redis CreateDBAccount --body '{
  "InstanceId": "{{user.instance_id}}",
  "AccountName": "app_cache",
  "AccountPassword": "{{user.password}}",
  "Role": "ReadWrite"
}'
```

---

## 3. Parameter Commands

### Query Parameters

```bash
ve redis DescribeDBInstanceParams --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Modify Parameters

```bash
ve redis ModifyDBInstanceParams --body '{
  "InstanceId": "{{user.instance_id}}",
  "Parameters": [
    {"Name": "maxmemory-policy", "Value": "allkeys-lru"},
    {"Name": "timeout", "Value": "300"}
  ]
}'
```

---

## 4. Backup Commands

### Create Backup

```bash
ve redis CreateBackup --body '{
  "InstanceId": "{{user.instance_id}}",
  "BackupName": "pre-deploy-backup"
}'
```

### List Backups

```bash
ve redis DescribeBackups --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 5. Allow List Commands

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

## 6. Connection Commands

### Enable Public Access

```bash
# ⚠️ SECURITY: Always configure allow list FIRST
ve redis CreateDBEndpointPublicAddress --body '{
  "InstanceId": "{{user.instance_id}}",
  "NetworkType": "SingleLine"
}'

# Get public address
ve redis DescribeDBInstanceDetail --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 7. Performance Commands

### Scan Big Keys

```bash
ve redis DescribeBigKeys --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Identify Hot Keys

```bash
ve redis DescribeHotKeys --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 8. Output Formatting

### Instance Table

```bash
ve redis DescribeDBInstances --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.Instances[] |
  [.InstanceId, .InstanceName, .Status, .EngineVersion, (.Capacity.Total | tostring) + "MB", .PrivateAddress] |
  @tsv
' | column -t -s $'\t'
```

---

## 9. Common Patterns

### Pattern: Full Redis Setup

```bash
INSTANCE_ID=$(ve redis CreateDBInstance --body '{
  "InstanceName": "prod-redis",
  "EngineVersion": "6.0",
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

*This reference document is part of the ve-redis-ops skill.*
