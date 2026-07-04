# Knowledge Base — Volcengine ALB (应用型负载均衡)

> **Purpose:** Fault pattern library and operational knowledge for ALB.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [Fault Pattern 1: ALB in "Inactive" State](#1-fault-pattern-1-alb-in-inactive-state)
2. [Fault Pattern 2: All Backend Servers Unhealthy](#2-fault-pattern-2-all-backend-servers-unhealthy)
3. [Fault Pattern 3: High 502/504 Error Rate](#3-fault-pattern-3-high-502504-error-rate)
4. [Fault Pattern 4: Listener Port Conflict](#4-fault-pattern-4-listener-port-conflict)
5. [Fault Pattern 5: Rule Not Matching Traffic](#5-fault-pattern-5-rule-not-matching-traffic)
6. [Fault Pattern 6: Rate Limiting (Throttling)](#6-fault-pattern-6-rate-limiting-throttling)
7. [Operational Tips](#7-operational-tips)

---

## 1. Fault Pattern 1: ALB in "Inactive" State

**Symptoms:**
- ALB status shows `inactive` instead of `active`
- No traffic is forwarded to backend servers
- Health check status shows all backends as unhealthy

**Root Causes:**
- ALB was created but no listeners configured
- All listeners have been deleted
- ALB was manually stopped or suspended
- Account in arrears

**Diagnosis:**

```bash
# Check ALB status and details
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerIds '["{{user.alb_id}}"]' | jq '.Result.LoadBalancers[0] | {Status, Type, VpcId}'

# Check listeners
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" | jq '.Result.Listeners'

# Check account balance if suspected
```

**Resolution:**
1. Create at least one listener with a server group
2. If account in arrears, recharge and wait for ALB to resume
3. If suspended, contact support to reinstate

---

## 2. Fault Pattern 2: All Backend Servers Unhealthy

**Symptoms:**
- Health check shows all backends `unhealthy`
- Traffic returns 502/503 errors
- ALB is `active` but no successful requests

**Root Causes:**
- Backend service not running on expected port
- Security group blocks ALB health check traffic
- Health check path/port mismatches backend configuration
- Backend server overloaded (CPU/memory saturation)

**Diagnosis:**

```bash
# Check which servers are healthy vs unhealthy
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds '["{{user.server_group_id}}"]' | jq '.Result.ServerGroups[0].Servers[] | {ServerId, Status, Port}'

# Verify backend server is running (via SSH or other monitoring)
```

**Resolution Checklist:**

| Check | Command | Expected Result |
|-------|---------|----------------|
| Backend process running | `ssh user@host 'systemctl status app'` | Service active |
| Security group allows ALB | `ve ecs DescribeSecurityGroups --InstanceId i-xxx` | ALB subnet CIDR allowed on backend port |
| Health check path correct | `curl http://10.0.1.10:8080/health` | Returns 200 |
| Backend port matches listener | Verify listener config | Port matches backend service port |

---

## 3. Fault Pattern 3: High 502/504 Error Rate

**Symptoms:**
- `alb_http_502` or `alb_http_504` metrics spike
- Users report "Bad Gateway" or "Gateway Timeout" errors
- Application logs show upstream connection failures

**Root Causes:**
- Backend server timeout (default ALB upstream timeout: 60s)
- Backend server connections exhausted
- Backend server unresponsive due to GC pause or deadlock
- Network connectivity issues between ALB and backend

**Resolution:**

```bash
# 1. Increase backend timeout if needed
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --RequestTimeout 120

# 2. Check backend server health under load
# 3. Scale backend server count if capacity issue
# 4. Review application logs for errors
```

**Prevention:**
- Set up alarms for 5xx rates > threshold
- Configure proper health check intervals
- Use connection pooling on backend applications

---

## 4. Fault Pattern 4: Listener Port Conflict

**Symptoms:**
- Error `PortConflict.Listener` when creating a listener
- Cannot create listener on desired port

**Root Causes:**
- Port already in use by another listener on the same ALB
- Previous listener was not fully deleted

**Diagnosis:**

```bash
# Check existing listeners and their ports
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'
```

**Resolution:**
1. Use a different port
2. Or delete the conflicting listener and recreate

---

## 5. Fault Pattern 5: Rule Not Matching Traffic

**Symptoms:**
- Traffic reaches the listener but is not forwarded to expected server group
- Default action (if configured) handles traffic instead of custom rule
- Rule seems correctly configured but doesn't match

**Root Causes:**
- Domain/host header mismatch (case sensitivity, trailing dots)
- URL path pattern incorrect (wildcard placement)
- Rule priority: more specific rules should be evaluated first
- HTTP method match not configured when needed

**Diagnosis:**

```bash
# List all rules for the listener
ve alb DescribeRules --Region "{{env.VOLCENGINE_REGION}}" --ListenerId "{{user.listener_id}}"

# Verify rule patterns match expected traffic
```

**Rule Matching Order:** ALB evaluates rules in order of specificity (most specific first).

| Pattern Type | Example | Match Priority |
|-------------|---------|----------------|
| Exact path | `/api/v1/users` | Highest |
| Wildcard prefix | `/api/*` | Medium |
| Wildcard suffix | `*.jpg` | Medium |
| Host only | `api.example.com` | Lower |
| Default (catch-all) | `/*` | Lowest |

---

## 6. Fault Pattern 6: Rate Limiting (Throttling)

**Symptoms:**
- Error `Throttling` or HTTP 429 responses
- API calls intermittently fail
- Batch operations fail partway through

**Root Causes:**
- API rate limit reached (varies by account tier)
- Too many concurrent API calls
- Automated scripts sending requests too quickly

**Resolution:**

```bash
# Implement exponential backoff in scripts:
for i in 1 2 3; do
  RESPONSE=$(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" 2>&1)
  if echo "$RESPONSE" | grep -q "Throttling"; then
    SLEEP=$((2 ** i))
    echo "Rate limited. Retrying in ${SLEEP}s... (attempt $i/3)"
    sleep $SLEEP
  else
    echo "$RESPONSE"
    break
  fi
done
```

---

## 7. Operational Tips

### Idle ALB Detection

> See [`SKILL.md §FinOps Operation: DescribeIdleLoadBalancers`](../SKILL.md#finops-operation-describeidleloadbalancers-find-idle-albs) for the complete idle ALB detection flow.

### Migrating from CLB to ALB

| Capability | CLB | ALB |
|-----------|-----|-----|
| Layer | L4 (TCP/UDP) + L7 | L7 (HTTP/HTTPS/gRPC) |
| Content-based routing | No | Yes (host + path) |
| HTTPS termination | Limited | Full support |
| WebSocket | No | Native |
| gRPC | No | Native |
| Health check granularity | Per listener | Per server group |

### Cost Optimization

1. Delete idle ALBs with no listeners (hourly billing still applies for inactive ALBs)
2. Right-size backend server weights — remove zero-weight servers
3. Consolidate listeners — use path-based routing instead of separate ALBs for each service
4. Monitor `alb_out_bps` to detect unexpected traffic spikes

---

*This reference document is part of the ve-alb-ops skill.*
