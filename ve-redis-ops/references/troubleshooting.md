# Troubleshooting Guide — Volcengine Redis

> **Purpose:** Systematic troubleshooting guide for common Redis operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [Instance Creation Errors](#2-instance-creation-errors)
3. [Connection Errors](#3-connection-errors)
4. [Memory and Performance Issues](#4-memory-and-performance-issues)
5. [Cross-Service Authorization](#5-cross-service-authorization)

---

## 1. Error Taxonomy

| Category | Error Code | HALT or Retry | Example |
|----------|-----------|---------------|---------|
| **Parameter Error** | `Invalid*.Malformed` | HALT | `InvalidParameterValue.Malformed` |
| **Resource Not Found** | `InstanceNotFound` | HALT | `InstanceNotFound` |
| **Status Error** | `IncorrectInstanceStatus` | HALT | `IncorrectInstanceStatus` |
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.Redis` |
| **Billing Error** | `BalanceNotEnough` | HALT | `BalanceNotEnough` |
| **IAM Error** | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| **Rate Limit** | `FlowLimitExceeded` | Retry | `FlowLimitExceeded` |
| **Server Error** | `InternalError` | Retry | `InternalError` |

---

## 2. Instance Creation Errors

### Cross-Service Authorization Required

```
Error: Forbidden
Message: Cross-service access authorization is required. Please complete the authorization before calling CreateDBInstance.
```

**Root Cause:** Since 2022-05-17, Redis requires a service-linked role for cross-service access.

**Resolution:**
```bash
# Create service-linked role (use ve CLI)
ve iam CreateServiceLinkedRole --body '{"ServiceName": "Redis"}'
```

Or authorize via Volcengine console.

### SubnetNotFound

**Resolution:**
```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --SubnetIds "[\"$SUBNET_ID\"]"
```

---

## 3. Connection Errors

### Connection Refused

**Checklist:**
1. Instance status is `Running`
2. Client IP is in the allow list
3. Client is in the same VPC (for private access)
4. Security group allows Redis port (6379)

### NOAUTH Authentication Required

**Root Cause:** Not passing the password during connection.

**Resolution:**
```bash
# Using redis-cli
redis-cli -h $PRIVATE_ADDRESS -p $PORT -a "$PASSWORD"

# Or authenticate after connection
redis-cli -h $PRIVATE_ADDRESS -p $PORT
AUTH $PASSWORD
```

---

## 4. Memory and Performance Issues

### OOM Command Not Allowed

**Root Cause:** Redis memory is full and `maxmemory-policy` is set to `noeviction`.

**Resolution:**
1. Change eviction policy:
   ```bash
   ve redis ModifyDBInstanceParams --body '{"InstanceId":"xxx","Parameters":[{"Name":"maxmemory-policy","Value":"allkeys-lru"}]}'
   ```
2. Increase shard capacity via `ModifyDBInstanceShardCapacity`
3. Add more shards via `ModifyDBInstanceShardNumber`

### Slow Commands

Redis single-threaded operations mean slow commands (KEYS, SCAN with large result, SMEMBERS on large sets) block all other operations.

**Resolution:**
- Replace `KEYS pattern` with `SCAN cursor MATCH pattern COUNT 100`
- Move large set operations to pipeline

---

## 5. Cross-Service Authorization

Required for: CreateDBInstance, ModifyDBInstanceSubnet, CreateDBEndpointPublicAddress.

**Resolution:** Create the `Redis` service-linked role via console or:
```bash
ve iam CreateServiceLinkedRole --body '{"ServiceName": "Redis"}'
```

---

*This reference document is part of the ve-redis-ops skill.*
