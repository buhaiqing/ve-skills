# Troubleshooting Guide — Volcengine CLB

> **Purpose:** Systematic troubleshooting guide for common CLB operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [CLB Creation Errors](#2-clb-creation-errors)
3. [Listener Errors](#3-listener-errors)
4. [Backend Server Errors](#4-backend-server-errors)
5. [Health Check Issues](#5-health-check-issues)
6. [Rate Limiting](#6-rate-limiting)

---

## 1. Error Taxonomy

| Category | Code Pattern | HALT/Retry | Example |
|----------|-------------|------------|---------|
| Parameter Error | `Invalid*.*` | HALT | `InvalidLoadBalancerId.NotFound` |
| Resource Error | `*.NotFound` | HALT | `InvalidSubnetId.NotFound` |
| Conflict Error | `*.Conflict` | HALT | `PortConflict.Listener` |
| Quota Error | `QuotaExceeded.*` | HALT | `QuotaExceeded.LoadBalancer` |
| Status Error | `IncorrectStatus.*` | HALT | `IncorrectStatus.LoadBalancer` |
| Dependency Error | `DependencyViolation` | HALT | `DependencyViolation.Listener` |
| IAM Error | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| Rate Limit | `Throttling` | Retry | `Throttling` |
| Server Error | `InternalError` | Retry | `InternalError` |

---

## 2. CLB Creation Errors

### QuotaExceeded.LoadBalancer

`Error: QuotaExceeded.LoadBalancer — max CLBs per region reached.`

**Resolution:** Request quota increase via Volcengine console.

### InvalidSubnetId.NotFound

`Error: InvalidSubnetId.NotFound — SubnetId does not exist.`

**Resolution:**
```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID"
```

---

## 3. Listener Errors

### PortConflict.Listener

`Error: PortConflict.Listener — same protocol+port already exists.`

**Resolution:** Use a different port, or delete the conflicting listener.

### ProtocolNotSupported

`Error: ProtocolNotSupported — invalid protocol for this CLB.`

**Resolution:** Use valid protocols: `TCP`, `UDP`, `HTTP`, `HTTPS`.

---

## 4. Backend Server Errors

### BackendServer.NotFound

`Error: BackendServer.NotFound — backend server does not exist.`

**Resolution:** Verify ECS instance exists:
```bash
ve ecs DescribeInstances --Region "$VOLCENGINE_REGION" --InstanceIds "[\"$ECS_ID\"]"
```

### InvalidPort.Range

`Error: InvalidPort.Range — port out of range (1–65535).`

**Resolution:** Port range is 1–65535.

---

## 5. Health Check Issues

### All Backend Servers Unhealthy

**Checklist:**
1. Backend server process is running on the expected port
2. Security group allows CLB health check IP range
3. Health check path/protocol matches the backend service
4. Firewall on the server is not blocking health check traffic

### Health Check Never Passes

```bash
# Manually test health check from within the same VPC
curl -s --connect-timeout 5 http://10.0.2.10:8080/health

# Test TCP connectivity
nc -zv 10.0.2.10 8080
```

---

## 6. Rate Limiting

### Throttling

Standard backoff retry: 2s, 4s, 8s. Max 3 retries.

---

*This reference document is part of the ve-clb-ops skill.*
