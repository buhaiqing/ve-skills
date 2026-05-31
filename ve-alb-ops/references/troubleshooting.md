# Troubleshooting Guide — Volcengine ALB (应用型负载均衡)

> **Purpose:** Systematic troubleshooting guide for common ALB operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [ALB Creation Errors](#2-alb-creation-errors)
3. [Listener Errors](#3-listener-errors)
4. [Rule Errors](#4-rule-errors)
5. [Server Group Errors](#5-server-group-errors)
6. [Health Check Issues](#6-health-check-issues)
7. [Rate Limiting](#7-rate-limiting)

---

## 1. Error Taxonomy

| Category | Code Pattern | HALT or Retry | Example |
|----------|-------------|---------------|---------|
| **Parameter Error** | `Invalid*.*` | HALT | `InvalidLoadBalancer.NotFound` |
| **Resource Error** | `*.NotFound` | HALT | `InvalidVpc.NotFound`, `InvalidCertificate.NotFound` |
| **Conflict Error** | `*.Conflict` | HALT | `PortConflict.Listener` |
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.LoadBalancer`, `QuotaExceeded.Listener` |
| **Status Error** | `IncorrectStatus.*` | HALT | `IncorrectStatus.LoadBalancer` |
| **Dependency Error** | `DependencyViolation` | HALT | `DependencyViolation.Listener` |
| **IAM Error** | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| **Rate Limit** | `Throttling` | Retry (3×) | `Throttling` |
| **Server Error** | `InternalError` | Retry (3×) | `InternalError` |

---

## 2. ALB Creation Errors

### QuotaExceeded.LoadBalancer

```
Error Code: QuotaExceeded.LoadBalancer
Message: The maximum number of ALBs per region has been reached.
```

**Resolution:** Request quota increase via Volcengine console.

---

### InvalidVpc.NotFound

```
Error Code: InvalidVpc.NotFound
Message: The specified VpcId does not exist.
```

**Resolution:**
```bash
# Verify VPC exists
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}" --VpcIds '[{"{{user.vpc_id}}"}]'
```

---

### InvalidSubnet.NotFound

```
Error Code: InvalidSubnet.NotFound
Message: The specified subnet does not exist in the region.
```

**Resolution:**
```bash
# Verify subnet exists in the VPC
ve vpc DescribeSubnets --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

---

### InvalidParameter (CreateLoadBalancer)

```
Error Code: InvalidParameter
Message: The request parameter "Type" is invalid.
```

**Resolution:** The `Type` field must be `public` or `private`. For `public` type, ensure `EipBillingConfig` is provided correctly.

---

## 3. Listener Errors

### PortConflict.Listener

```
Error Code: PortConflict.Listener
Message: A listener with the same protocol and port already exists on this ALB.
```

**Resolution:** Use a different port, or delete the conflicting listener first.

```bash
# Check existing listeners and their ports
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" | jq '.Result.Listeners[] | {ListenerId, Protocol, Port}'
```

---

### InvalidCertificate.NotFound

```
Error Code: InvalidCertificate.NotFound
Message: The specified CertificateId does not exist.
```

**Resolution:** Verify the certificate ID. HTTPS listeners require a valid TLS certificate stored in Volcengine Certificate Manager.

---

### QuotaExceeded.Listener

```
Error Code: QuotaExceeded.Listener
Message: The maximum number of listeners has been reached.
```

**Resolution:** Delete unused listeners or request a quota increase.

---

## 4. Rule Errors

### InvalidListener.NotFound

```
Error Code: InvalidListener.NotFound
Message: The specified ListenerId does not exist.
```

**Resolution:**
```bash
# Verify listener exists
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}"
```

---

### InvalidServerGroup.NotFound

```
Error Code: InvalidServerGroup.NotFound
Message: The specified ServerGroupId does not exist.
```

**Resolution:**
```bash
# Verify server group exists
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds '["{{user.server_group_id}}"]'
```

---

### InvalidRule.Pattern

```
Error Code: InvalidRule.Pattern
Message: The URL pattern is invalid.
```

**Resolution:** URL patterns must follow wildcard format:
- `/api/*` — matches all paths under /api/
- `/images/*.jpg` — matches .jpg files under /images/
- `/*` — matches all paths

---

### QuotaExceeded.Rule

```
Error Code: QuotaExceeded.Rule
Message: The maximum number of rules has been reached.
```

**Resolution:** Delete unused rules or request a quota increase.

---

## 5. Server Group Errors

### Backend Server Unhealthy (AddServersToGroup)

```
Error Code: BackendServer.Unhealthy
Message: The backend server failed health check immediately after registration.
```

**Resolution:**
- Verify the backend process is running on the specified port
- Check security group rules allow traffic from ALB
- Verify health check URI returns expected HTTP code

---

### QuotaExceeded.ServerGroup

```
Error Code: QuotaExceeded.ServerGroup
Message: The maximum number of server groups has been reached.
```

**Resolution:** Delete unused server groups or request a quota increase.

---

## 6. Health Check Issues

### All Backend Servers Unhealthy

**Checklist:**
1. Backend service is running on the expected port
2. Security group on ECS allows inbound traffic from ALB subnet
3. Health check path matches the actual service endpoint
4. Health check protocol matches (HTTP vs TCP)
5. Network ACLs between ALB subnet and backend subnet allow traffic

```bash
# Simulate health check from within VPC (SSH to a jump box)
curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 http://10.0.1.10:8080/health
```

---

### Intermittent Health Check Failures

**Possible Causes:**
- Backend server CPU/memory usage spiking during health check windows
- Connection pool exhaustion on backend server
- Health check interval too aggressive (recommended: 5s minimum)
- Timeout too low (recommended: 3s minimum)

**Resolution:** Increase `UnhealthyThreshold` to 5 to reduce flapping.

---

## 7. Rate Limiting

### Throttling (429)

```
Error Code: Throttling
Message: Request has been throttled due to rate limiting.
```

**Resolution:** Standard exponential backoff — retry at 1s, 2s, 4s intervals. Max 3 retries.

---

### InternalError (500)

```
Error Code: InternalError
Message: An internal error occurred.
```

**Resolution:** Retry with backoff (2s, 4s, 8s). If persistent, contact Volcengine support with `RequestId`.

---

*This reference document is part of the ve-alb-ops skill.*
