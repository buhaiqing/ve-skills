# CLI Usage — Volcengine RDS MySQL

> **Purpose:** CLI usage reference for RDS MySQL operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Instance Commands](#1-instance-commands)
2. [Database Commands](#2-database-commands)
3. [Account Commands](#3-account-commands)
4. [Parameter Commands](#4-parameter-commands)
5. [Backup Commands](#5-backup-commands)
6. [Allow List Commands](#6-allow-list-commands)
7. [Output Formatting](#7-output-formatting)
8. [Common Patterns](#8-common-patterns)

---

## 1. Instance Commands

### List All Instances

```bash
ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by VPC

```bash
ve rds_mysql DescribeDBInstances --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Create Instance (HA)

> **Note:** Many RDS MySQL operations use `--body` JSON format.

```bash
ve rds_mysql CreateDBInstance --body '{
  "RegionId": "{{user.region}}",
  "InstanceName": "prod-mysql",
  "DBEngineVersion": "MySQL_8_0",
  "InstanceType": "HA",
  "NodeSpec": "rds.mysql.2c4g",
  "StorageSpace": 100,
  "StorageType": "ESSD",
  "VpcId": "{{user.vpc_id}}",
  "SubnetId": "{{user.subnet_id}}",
  "ChargeType": "PostPaid",
  "NodeInfo": [
    {"ZoneId": "cn-beijing-a"},
    {"ZoneId": "cn-beijing-b"}
  ]
}'
```

### Restart Instance

```bash
ve rds_mysql RestartDBInstance --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Delete Instance

> 🔴 **IRREVERSIBLE.** See `SKILL.md §DeleteDBInstance` for safety gate and full flow (Region via `{{user.region}}` or `{{env.VOLCENGINE_REGION}}`).

---

## 2. Database Commands

### List Databases

```bash
ve rds_mysql DescribeDatabases --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Create Database

> See `SKILL.md §CreateDB` for full flow (CharacterSet defaults to `utf8mb4`).

### Delete Database

```bash
ve rds_mysql DeleteDB \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}"
```

---

## 3. Account Commands

### List Accounts

```bash
ve rds_mysql ListDBAccounts --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Create Account

> See `SKILL.md §CreateAccount` for full flow (defaults to AccountType `Normal`).

### Grant Privileges

> See `SKILL.md §CreateAccount → Grant Privileges` for full flow.

### Reset Password

```bash
ve rds_mysql ResetAccountPassword \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.account_name}}" \
  --AccountPassword "{{user.new_password}}"
```

---

## 4. Parameter Commands

### Query Parameters

```bash
# All parameters
ve rds_mysql DescribeDBInstanceParameters --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"

# Specific parameter
ve rds_mysql DescribeDBInstanceParameters --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --ParameterName "max_connections"
```

### Modify Parameters

```bash
ve rds_mysql ModifyDBInstanceParameters \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --Parameters '[{"ParameterName":"max_connections","ParameterValue":"2000"}]'
```

> **Warning:** Always check `ForceRestart` before modifying. If true, plan a maintenance window.

---

## 5. Backup Commands

### Create Backup

```bash
ve rds_mysql CreateBackup \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --BackupName "pre-deploy-backup" \
  --BackupStrategy "Snapshot"
```

### List Backups

```bash
ve rds_mysql DescribeBackups --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

---

## 6. Allow List Commands

### List Allow Lists

```bash
ve rds_mysql DescribeAllowLists --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"
```

### Create/Modify Allow List

```bash
ve rds_mysql CreateAllowList \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --AllowListName "app-servers" \
  --AllowListType "IPv4" \
  --AllowList '["10.0.2.0/24", "10.0.3.0/24"]'
```

---

## 7. Output Formatting

### Instance Table

```bash
ve rds_mysql DescribeDBInstances --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.Instances[] |
  [.InstanceId, .InstanceName, .InstanceStatus, .DBEngineVersion, .PrimaryIp, (.Port | tostring)] |
  @tsv
' | column -t -s $'\t'
```

### Database + Accounts Summary

```bash
INSTANCE_ID="mysql-xxx"
echo "=== Databases ==="
ve rds_mysql DescribeDatabases --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" | jq -r '.Result.Databases[] | "\(.DBName)\t\(.CharacterSet)"'
echo "=== Accounts ==="
ve rds_mysql ListDBAccounts --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" | jq -r '.Result.Accounts[] | "\(.AccountName)\t\(.AccountType)\t\(.AccountStatus)"'
```

---

## 8. Common Patterns

### Pattern: Full RDS Instance Setup

```bash
# Step 1: Create instance
INSTANCE_ID=$(ve rds_mysql CreateDBInstance --body '{
  "RegionId": "cn-beijing",
  "InstanceName": "prod-mysql",
  "DBEngineVersion": "MySQL_8_0",
  "InstanceType": "HA",
  "NodeSpec": "rds.mysql.2c4g",
  "StorageSpace": 100,
  "StorageType": "ESSD",
  "VpcId": "vpc-xxx",
  "SubnetId": "subnet-xxx",
  "ChargeType": "PostPaid",
  "NodeInfo": [{"ZoneId": "cn-beijing-a"}, {"ZoneId": "cn-beijing-b"}]
}' | jq -r '.Result.InstanceId')

# Step 2: Poll until Running
for i in {1..60}; do
  STATUS=$(ve rds_mysql DescribeDBInstances --Region "cn-beijing" --InstanceId "$INSTANCE_ID" | jq -r '.Result.Instances[0].InstanceStatus')
  [ "$STATUS" = "Running" ] && echo "Instance ready" && break
  sleep 10
done

# Step 3: Create database
ve rds_mysql CreateDB --Region "cn-beijing" --InstanceId "$INSTANCE_ID" --DBName "appdb" --CharacterSet "utf8mb4"

# Step 4: Create account and grant privileges
ve rds_mysql CreateAccount --Region "cn-beijing" --InstanceId "$INSTANCE_ID" --AccountName "app_user" --AccountPassword "SecureP@ss123"
ve rds_mysql GrantAccountPrivileges --Region "cn-beijing" --InstanceId "$INSTANCE_ID" --AccountName "app_user" --Privileges '[{"DBName":"appdb","AccountPrivilege":"ReadWrite"}]'

# Step 5: Set allow list
ve rds_mysql CreateAllowList --Region "cn-beijing" --InstanceId "$INSTANCE_ID" --AllowListName "default" --AllowListType "IPv4" --AllowList '["10.0.0.0/8"]'

# Connection info
PRIMARY_IP=$(ve rds_mysql DescribeDBInstances --Region "cn-beijing" --InstanceId "$INSTANCE_ID" | jq -r '.Result.Instances[0].PrimaryIp')
echo "Connect: mysql -h $PRIMARY_IP -P 3306 -u app_user -p appdb"
```

---

*This reference document is part of the ve-rds-ops skill.*
