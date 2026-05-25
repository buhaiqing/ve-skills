# Troubleshooting Guide — Volcengine RDS MySQL

> **Purpose:** Systematic troubleshooting guide for common RDS MySQL operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [Instance Creation Errors](#2-instance-creation-errors)
3. [Connection Errors](#3-connection-errors)
4. [Database and Account Errors](#4-database-and-account-errors)
5. [Backup Errors](#5-backup-errors)
6. [Parameter Modification Issues](#6-parameter-modification-issues)
7. [Debugging Strategies](#7-debugging-strategies)

---

## 1. Error Taxonomy

| Category | Error Code | HALT or Retry | Example |
|----------|-----------|---------------|---------|
| **Parameter Error** | `Invalid*.Malformed` | HALT | `InvalidParameterValue.Malformed` |
| **Resource Not Found** | `*.NotFound` | HALT | `InstanceNotFound` |
| **Status Error** | `IncorrectInstanceStatus` | HALT | `IncorrectInstanceStatus` |
| **Conflict Error** | `*.Conflict` | HALT | `DBAlreadyExists` |
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.Instance` |
| **IAM Error** | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| **Billing Error** | `BalanceNotEnough` | HALT | `BalanceNotEnough` |
| **Rate Limit** | `FlowLimitExceeded` | Retry | `FlowLimitExceeded` |
| **Server Error** | `InternalError` | Retry | `InternalError` |

---

## 2. Instance Creation Errors

### IncorrectInstanceStatus for Subnet

```
Error: IncorrectInstanceStatus
Message: The specified subnet is not in the correct status for this operation.
```

**Root Cause:** The specified subnet is not `Available` or has no resources.

**Resolution:**
```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --SubnetIds "[\"$SUBNET_ID\"]"
```

### BalanceNotEnough

**Resolution:** Recharge the account via Volcengine console → Billing Management.

---

## 3. Connection Errors

### Cannot Connect to RDS Instance

**Checklist:**
1. Instance status is `Running`:
   ```bash
   ve rds_mysql DescribeDBInstances --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" | jq '.Result.Instances[0].InstanceStatus'
   ```

2. Allow list includes client IP:
   ```bash
   ve rds_mysql DescribeAllowLists --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID"
   ```

3. Security group allows inbound traffic on port 3306 from client subnet

4. VPC subnet route tables are properly configured

### Access Denied for User

**Root Cause:** Missing privileges or wrong credentials.

**Resolution:**
```bash
# Verify account exists and has privileges
ve rds_mysql ListDBAccounts --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID"

# Grant privileges if missing
ve rds_mysql GrantAccountPrivileges --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" --AccountName "$ACCOUNT" --Privileges '[{"DBName":"mydb","AccountPrivilege":"ReadWrite"}]'
```

---

## 4. Database and Account Errors

### DBAlreadyExists

```
Error: DBAlreadyExists
Message: The database already exists.
```

**Resolution:** Use a different database name.

### AccountAlreadyExists

**Resolution:** Use a different account name or delete existing account first.

---

## 5. Backup Errors

### BackupInProgress

```
Error: BackupInProgress
Message: A backup is already in progress for this instance.
```

**Resolution:** Wait for the existing backup to complete, then retry.

---

## 6. Parameter Modification Issues

### Parameter Requires Restart

Some parameters require a restart to take effect. Always check the `ForceRestart` field:

```bash
ve rds_mysql DescribeDBInstanceParameters --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" --ParameterName "innodb_buffer_pool_size" \
  | jq '.Result.Parameters[0].ForceRestart'
```

If `true`, restart after modification:
```bash
ve rds_mysql ModifyDBInstanceParameters --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" --Parameters '[{"ParameterName":"innodb_buffer_pool_size","ParameterValue":"2G"}]'
ve rds_mysql RestartDBInstance --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID"
```

---

## 7. Debugging Strategies

### Capture RequestId

```bash
ve rds_mysql CreateDBInstance --Region "$VOLCENGINE_REGION" --body '...' 2>&1 \
  | jq -r '.ResponseMetadata.RequestId'
```

### Test Database Connectivity from CLI

```bash
# From within the same VPC
mysql -h "$RDS_PRIVATE_IP" -P 3306 -u "$ACCOUNT" -p"$PASSWORD" -e "SELECT 1"
```

### Query Slow Logs

```bash
ve rds_mysql DescribeSlowLogs --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" --StartTime "2026-05-25T00:00:00Z" --EndTime "2026-05-25T23:59:59Z"
```

---

*This reference document is part of the ve-rds-ops skill.*
