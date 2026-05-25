# CLI Usage — Volcengine CLB

> **Purpose:** CLI usage reference for CLB operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Load Balancer Commands](#1-load-balancer-commands)
2. [Listener Commands](#2-listener-commands)
3. [Backend Server Commands](#3-backend-server-commands)
4. [Health Check Commands](#4-health-check-commands)
5. [Output Formatting](#5-output-formatting)
6. [Common Patterns](#6-common-patterns)

---

## 1. Load Balancer Commands

### List All CLBs

```bash
ve clb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by VPC

```bash
ve clb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Filter by Type

```bash
ve clb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --Type "public"
```

### Create CLB (Internal)

```bash
ve clb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.clb_name}}" \
  --Type "private"
```

### Create CLB (Public — requires EIP)

```bash
ve clb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.clb_name}}" \
  --Type "public"
```

### Delete CLB

```bash
# ⚠️ IRREVERSIBLE: All listeners and backends must be removed first
ve clb DeleteLoadBalancer --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.clb_id}}"
```

---

## 2. Listener Commands

### List Listeners for a CLB

```bash
ve clb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.clb_id}}"
```

### Create TCP Listener

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "TCP" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"
```

### Create HTTP Listener

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "HTTP" \
  --Port 80 \
  --ListenerName "http-listener"
```

### Create HTTPS Listener

```bash
ve clb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --Protocol "HTTPS" \
  --Port 443 \
  --ListenerName "https-listener"
```

### Delete Listener

```bash
ve clb DeleteListener --Region "{{env.VOLCENGINE_REGION}}" --ListenerId "{{user.listener_id}}"
```

---

## 3. Backend Server Commands

### List Backend Servers

```bash
ve clb DescribeBackendServers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.clb_id}}"
```

### Add Backend Servers

```bash
ve clb AddBackendServers \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --BackendServers '[{"ServerId":"{{ecs.instance_id}}","Port":8080,"Weight":100}]'
```

### Remove Backend Servers

```bash
ve clb RemoveBackendServers \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --ServerIds '["{{ecs.instance_id}}"]'
```

---

## 4. Health Check Commands

### Configure Health Check

```bash
ve clb SetHealthCheckConfig \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.clb_id}}" \
  --ListenerId "{{user.listener_id}}" \
  --HealthyThreshold 3 \
  --UnhealthyThreshold 3 \
  --Interval 5 \
  --Timeout 3 \
  --HttpMethod "GET" \
  --Uri "/health" \
  --HealthyHttpCode "200"
```

---

## 5. Output Formatting

### CLB Table

```bash
ve clb DescribeLoadBalancers --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.LoadBalancers[] |
  [.LoadBalancerId, .LoadBalancerName, .Type, .Status, (.Address // "-"), (.EipAddress // "-")] |
  @tsv
' | column -t -s $'\t'
```

### Backend Server Status

```bash
ve clb DescribeBackendServers --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" | jq -r '
  .Result.BackendServers[] |
  "\(.ServerId):\(.Port) weight=\(.Weight) status=\(.Status)"
'
```

---

## 6. Common Patterns

### Pattern: Full CLB Setup with TCP Listener and Backend

```bash
# Step 1: Create CLB
CLB_ID=$(ve clb CreateLoadBalancer --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --Type "private" --LoadBalancerName "prod-clb" | jq -r '.Result.LoadBalancerId')

# Step 2: Poll until active
for i in {1..30}; do
  STATUS=$(ve clb DescribeLoadBalancers --Region "$VOLCENGINE_REGION" --LoadBalancerIds "[\"$CLB_ID\"]" | jq -r '.Result.LoadBalancers[0].Status')
  [ "$STATUS" = "active" ] && break
  sleep 2
done

# Step 3: Create TCP listener
LISTENER_ID=$(ve clb CreateListener --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" --Protocol "TCP" --Port 80 --ListenerName "tcp-80" | jq -r '.Result.ListenerId')

# Step 4: Add backend servers
ve clb AddBackendServers --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" \
  --BackendServers '[{"ServerId":"i-ecs-1","Port":8080,"Weight":100},{"ServerId":"i-ecs-2","Port":8080,"Weight":100}]'

# Step 5: Configure health check
ve clb SetHealthCheckConfig --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" --ListenerId "$LISTENER_ID" --HealthyThreshold 3 --UnhealthyThreshold 3 --Interval 5 --Timeout 3

echo "CLB setup complete: $CLB_ID"
```

---

*This reference document is part of the ve-clb-ops skill.*
