# Troubleshooting Guide — Volcengine Redis

> **Purpose:** Systematic troubleshooting guide for common Redis operational issues.
> **Version:** 1.1.0
> **Last Updated:** 2026-07-04

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [Instance Creation Errors](#2-instance-creation-errors)
3. [Connection Errors](#3-connection-errors)
4. [Memory and Performance Issues](#4-memory-and-performance-issues)
5. [Cross-Service Authorization](#5-cross-service-authorization)

---

## 1. Error Taxonomy

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `Invalid*.Malformed` | **HALT** — param val failed | Check values against API docs |
| `InstanceNotFound` | **HALT** — instance missing | Verify ID via describe |
| `IncorrectInstanceStatus` | **HALT** — wrong state | Wait for transition or cancel current op |
| `QuotaExceeded.*` | **HALT** — quota exhausted | Delete unused instances or request increase |
| `BalanceNotEnough` | **HALT** — insufficient balance | Recharge via billing console |
| `Forbidden.RAM` | **HALT** — insufficient IAM perms | Verify IAM policy |
| `FlowLimitExceeded` | RETRY w/ exponential backoff | Max 3 retries |
| `InternalError` | RETRY w/ exponential backoff | Max 3 retries; capture RequestId |

---

## 2. Instance Creation Errors

### Cross-Service Authorization Required

```
Error: Forbidden
Message: Cross-service access authorization is required.
```

**Root Cause:** Since 2022-05-17, Redis requires a service-linked role.

**Resolution:**
```bash
ve iam CreateServiceLinkedRole --body '{"ServiceName": "Redis"}'
```

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
2. Client IP in allow list
3. Client in same VPC (private access)
4. Security group allows port 6379

### NOAUTH Authentication Required

**Root Cause:** Missing password on connection.

**Resolution:**
```bash
redis-cli -h $PRIVATE_ADDRESS -p $PORT -a "$PASSWORD"
# Or authenticate after connect
AUTH $PASSWORD
```

---

## 4. Memory and Performance Issues

### OOM Command Not Allowed

**Root Cause:** Memory full + `maxmemory-policy` = `noeviction`.

**Resolution:**
1. Change eviction policy:
   ```bash
   ve redis ModifyDBInstanceParams --body '{"InstanceId":"xxx","Parameters":[{"Name":"maxmemory-policy","Value":"allkeys-lru"}]}'
   ```
2. Increase shard capacity via `ModifyDBInstanceShardCapacity`
3. Add more shards via `ModifyDBInstanceShardNumber`

### Slow Commands

Redis single-threaded ⇒ slow cmds (KEYS, SCAN large results, SMEMBERS on large sets) block all ops.

**Resolution:**
- Replace `KEYS pattern` w/ `SCAN cursor MATCH pattern COUNT 100`
- Move large set ops to pipeline

---

## 5. Cross-Service Authorization

Required for: CreateDBInstance, ModifyDBInstanceSubnet, CreateDBEndpointPublicAddress.

**Resolution:** Create `Redis` service-linked role:
```bash
ve iam CreateServiceLinkedRole --body '{"ServiceName": "Redis"}'
```

---

*This reference document is part of the ve-redis-ops skill.*