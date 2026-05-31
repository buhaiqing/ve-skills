# CLI Usage — Volcengine ALB (应用型负载均衡)

> **Purpose:** CLI usage reference for ALB operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [Load Balancer Commands](#1-load-balancer-commands)
2. [Listener Commands](#2-listener-commands)
3. [Rule Commands](#3-rule-commands)
4. [Server Group Commands](#4-server-group-commands)
5. [Output Formatting](#5-output-formatting)
6. [Common Patterns](#6-common-patterns)

---

## 1. Load Balancer Commands

### List All ALBs

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by VPC

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Filter by Type

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --Type "public"
# or: --Type "private"
```

### Filter by Name

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerName "{{user.alb_name}}"
```

### Get Single ALB Details

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerIds '[{"{{user.alb_id}}"}]'
```

### Create ALB (Private — Internal)

```bash
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "private"
```

### Create ALB (Public — Internet-Facing)

```bash
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "public" \
  --EipBillingConfig '{"Bandwidth":10,"EipBillingType":"PayByTraffic","ISP":"BGP"}'
```

### Rename ALB

```bash
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --LoadBalancerName "{{user.new_alb_name}}"
```

### Update ALB Description

```bash
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Description "New description here"
```

### Delete ALB

```bash
# ⚠️ IRREVERSIBLE: Listeners and rules must be removed first
ve alb DeleteLoadBalancer --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}"
```

---

## 2. Listener Commands

### List Listeners for an ALB

```bash
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}"
```

### Create HTTP Listener

```bash
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTP" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"
```

### Create HTTPS Listener (with Certificate)

```bash
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "{{user.listener_name}}" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2"
```

**TLSPolicy Values:** `tls-1-0`, `tls-1-1`, `tls-1-2`

### Create HTTPS Listener with Server Group Association

```bash
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "https-listener" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2" \
  --ServerGroupId "{{user.server_group_id}}"
```

### Create gRPC Listener

```bash
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "gRPC" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"
```

### Delete Listener

```bash
ve alb DeleteListener --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

### Modify Listener Attributes

```bash
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ListenerName "updated-listener-name"
```

---

## 3. Rule Commands

### List Rules for a Listener

```bash
ve alb DescribeRules --Region "{{env.VOLCENGINE_REGION}}" --ListenerId "{{user.listener_id}}"
```

### Create Path-Based Routing Rule

```bash
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Url "/api/*" \
  --RuleName "api-route"
```

### Create Host-Based Routing Rule

```bash
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "api.example.com" \
  --RuleName "host-route"
```

### Create Combined (Host + Path) Rule

```bash
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "api.example.com" \
  --Url "/v1/*" \
  --RuleName "api-v1-route"
```

### Delete Rule

```bash
ve alb DeleteRule --Region "{{user.region}}" --RuleId "{{user.rule_id}}"
```

---

## 4. Server Group Commands

### List Server Groups

```bash
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}"

# Filter by ID
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds '["{{user.server_group_id}}"]'

# Filter by Name
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupName "{{user.server_group_name}}"
```

### Create Server Group (HTTP Health Check)

```bash
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "HTTP",
    "Method": "GET",
    "Uri": "/health",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3,
    "HealthyHttpCode": "200"
  }'
```

### Create Server Group (TCP Health Check)

```bash
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "TCP",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3
  }'
```

**ServerGroupType Values:** `instance` (ECS instance), `ip` (IP address)

### Add Backend Servers to Server Group

```bash
ve alb AddServersToGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Servers '[
    {"ServerId":"i-xxx","Port":8080,"Weight":100,"ServerType":"ecs"},
    {"ServerId":"i-yyy","Port":8080,"Weight":100,"ServerType":"ecs"}
  ]'
```

### Remove Backend Servers from Server Group

```bash
ve alb RemoveServersFromGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerIds '[{"ServerId":"i-xxx","ServerType":"ecs"}]'
```

### Modify Server Group Name

```bash
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerGroupName "new-server-group-name"
```

### Modify Server Group Health Check

```bash
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --HealthCheckConfig '{"Enabled":true,"Uri":"/healthz","Interval":10}'
```

### Delete Server Group

```bash
# ⚠️ Ensure no listeners reference this server group before deletion
ve alb DeleteServerGroup --Region "{{user.region}}" --ServerGroupId "{{user.server_group_id}}"
```

---

## 5. Output Formatting

### JSON Pretty-Print with `jq`

```bash
# Get ALB name and status only
echo "=== ALB Summary ==="
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '
  .Result.LoadBalancers[] |
  "\(.LoadBalancerId) | \(.LoadBalancerName) | \(.Type) | \(.Status) | \(.Address // "-") | \(.EipAddress // "-")"
'

# Table header
echo "ID | Name | Type | Status | Internal IP | Public IP"
echo "---+------+------+--------+-------------+----------"
```

### Count Resources

```bash
# Count ALBs by type
echo "Public ALBs: $(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq '[.Result.LoadBalancers[] | select(.Type=="public")] | length')"
echo "Private ALBs: $(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq '[.Result.LoadBalancers[] | select(.Type=="private")] | length')"
```

### List Listeners for All ALBs

```bash
for ALB_ID in $(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[].LoadBalancerId'); do
  echo "ALB: $ALB_ID"
  ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq -r '.Result.Listeners[] | "  \(.ListenerId) \(.Protocol):\(.Port)"'
  echo
done
```

---

## 6. Common Patterns

### Quick Health Check of an ALB

```bash
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerIds '["{{user.alb_id}}"]' | jq '.Result.LoadBalancers[0] | {Status, Address, EipAddress, CreateTime}'
```

### Find ALBs Without Listeners (Idle Detection)

```bash
for ALB_ID in $(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[].LoadBalancerId'); do
  CNT=$(ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq '.Result.Listeners | length')
  [ "$CNT" -eq 0 ] && echo "IDLE ALB: $ALB_ID"
done
```

### Check Backend Server Health Across All Server Groups

```bash
for SG_ID in $(ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.ServerGroups[].ServerGroupId'); do
  echo "Server Group: $SG_ID"
  ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds "[\"$SG_ID\"]" | jq -r '.Result.ServerGroups[0].Servers[] | "  \(.ServerId):\(.Port) weight=\(.Weight) status=\(.Status)"'
done
```

---

*This reference document is part of the ve-alb-ops skill.*
