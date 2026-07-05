# API & SDK Usage — Volcengine ALB

> **Purpose:** Detailed API reference for ALB operations.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

# Common JSON Paths:
# CreateLoadBalancer: $.Result.LoadBalancerId
# DescribeLoadBalancers: $.Result.LoadBalancers[]
# DescribeListeners: $.Result.Listeners[]
# DescribeRules: $.Result.Rules[]

## Table of Contents

1. [API Overview](#1-api-overview)
2. [Load Balancer Operations](#2-load-balancer-operations)
3. [Listener Operations](#3-listener-operations)
4. [Rule Operations](#4-rule-operations)
5. [Server Group Operations](#5-server-group-operations)
6. [Response Parsing](#6-response-parsing)
7. [Go SDK Examples](#7-go-sdk-examples)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `alb` |
| API Version | `2022-03-01` |
| Endpoint | `open.volcengineapi.com` |
| Protocol | HTTPS |
| Go SDK | `github.com/volcengine/volc-sdk-golang/service/alb` |

### Common Response Structure

```json
{
  "ResponseMetadata": {
    "RequestId": "...",
    "Action": "...",
    "Service": "alb"
  },
  "Result": { ... }
}
```

---

## 2. Load Balancer Operations

### 2.1 CreateLoadBalancer

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `VpcId` | String | Yes | Parent VPC ID | `vpc-xxx` |
| `SubnetId` | String | Yes | Deployment subnet | `subnet-xxx` |
| `LoadBalancerName` | String | No | ALB name | `prod-alb` |
| `Type` | String | No | `public` or `private` | `private` |
| `Description` | String | No | Description | — |
| `EipBillingConfig` | Object | No | EIP billing config (public) | See below |

**EipBillingConfig format (for public type):**
```json
{
  "Bandwidth": 10,
  "EipBillingType": "PayByTraffic",
  "ISP": "BGP"
}
```

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `LoadBalancerId` | String | ALB ID, format `alb-xxx` |

### 2.2 DescribeLoadBalancers

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerIds.N` | String Array | No | Filter by ALB IDs |
| `VpcId` | String | No | Filter by VPC |
| `LoadBalancerName` | String | No | Filter by name |
| `Type` | String | No | Filter by type |
| `PageNumber` | Integer | No | Pagination |
| `PageSize` | Integer | No | Page size (max 100) |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `LoadBalancers` | Array | ALB list |
| `LoadBalancers[].LoadBalancerId` | String | ALB ID |
| `LoadBalancers[].LoadBalancerName` | String | ALB name |
| `LoadBalancers[].VpcId` | String | Parent VPC |
| `LoadBalancers[].Type` | String | `public` or `private` |
| `LoadBalancers[].Status` | String | `active`, `inactive` |
| `LoadBalancers[].Address` | String | Internal IP |
| `LoadBalancers[].EipAddress` | String | Public IP (public type) |
| `LoadBalancers[].CreateTime` | String | ISO 8601 timestamp |
| `TotalCount` | Integer | Total matching records |

### 2.3 DeleteLoadBalancer

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | ALB ID to delete |

> **Note:** All listeners must be removed before deleting an ALB.

### 2.4 ModifyLoadBalancerAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | ALB ID |
| `LoadBalancerName` | String | No | New name |
| `Description` | String | No | New description |

---

## 3. Listener Operations

### 3.1 CreateListener

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `LoadBalancerId` | String | Yes | ALB ID | `alb-xxx` |
| `Protocol` | String | Yes | HTTP, HTTPS, gRPC | `HTTP` |
| `Port` | Integer | Yes | Listening port | `80` |
| `ListenerName` | String | No | Listener name | `http-listener` |
| `CertificateId` | String | No | TLS cert ID (HTTPS) | `cert-xxx` |
| `TLSPolicy` | String | No | TLS version policy | `tls-1-2` |
| `ServerGroupId` | String | No | Default server group | `rsp-xxx` |
| `Description` | String | No | Description | — |

**TLSPolicy Values:** `tls-1-0`, `tls-1-1`, `tls-1-2`

### 3.2 DescribeListeners

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | ALB ID |
| `ListenerIds.N` | String Array | No | Filter by listener IDs |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Listeners` | Array | Listener list |
| `Listeners[].ListenerId` | String | Listener ID |
| `Listeners[].Protocol` | String | HTTP, HTTPS, gRPC |
| `Listeners[].Port` | Integer | Listening port |
| `Listeners[].DefaultServerGroupId` | String | Default server group |
| `Listeners[].Status` | String | Status |

### 3.3 DeleteListener

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ListenerId` | String | Yes | Listener ID |

### 3.4 ModifyListenerAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ListenerId` | String | Yes | Listener ID |
| `ListenerName` | String | No | New name |
| `Description` | String | No | New description |
| `CertificateId` | String | No | New TLS cert (HTTPS) |
| `TLSPolicy` | String | No | New TLS policy |

---

## 4. Rule Operations

### 4.1 CreateRule

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `ListenerId` | String | Yes | Listener ID | `lsnr-xxx` |
| `ServerGroupId` | String | Yes | Target server group | `rsp-xxx` |
| `Domain` | String | No | Host header match | `api.example.com` |
| `Url` | String | No | URL path pattern | `/api/*` |
| `RuleName` | String | No | Rule name | `api-route` |
| `Priority` | Integer | No | Rule priority | `100` |

> **Note:** At least one of `Domain` or `Url` must be specified.

**URL Pattern Examples:**
- `/api/*` — Match all paths starting with `/api/`
- `/*.jpg` — Match all JPEG images
- `/exact-path` — Exact match

### 4.2 DescribeRules

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ListenerId` | String | Yes | Listener ID |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Rules` | Array | Rule list |
| `Rules[].RuleId` | String | Rule ID |
| `Rules[].Domain` | String | Host match |
| `Rules[].Url` | String | URL path match |
| `Rules[].ServerGroupId` | String | Target server group |
| `Rules[].RuleName` | String | Rule name |
| `Rules[].Priority` | Integer | Priority |

### 4.3 DeleteRule

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RuleId` | String | Yes | Rule ID |

### 4.4 ModifyRuleAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RuleId` | String | Yes | Rule ID |
| `Domain` | String | No | New domain |
| `Url` | String | No | New URL pattern |
| `ServerGroupId` | String | No | New server group |

---

## 5. Server Group Operations

### 5.1 CreateServerGroup

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `ServerGroupName` | String | Yes | Server group name | `web-servers` |
| `ServerGroupType` | String | No | `instance` or `ip` | `instance` |
| `Description` | String | No | Description | — |
| `HealthCheckConfig` | Object | No | Health check config | See below |

**HealthCheckConfig format:**
```json
{
  "Enabled": true,
  "Protocol": "HTTP",
  "Method": "GET",
  "Uri": "/health",
  "Timeout": 3,
  "Interval": 5,
  "HealthyThreshold": 3,
  "UnhealthyThreshold": 3,
  "HealthyHttpCode": "200"
}
```

### 5.2 DescribeServerGroups

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ServerGroupIds.N` | String Array | No | Filter by IDs |
| `ServerGroupName` | String | No | Filter by name |
| `LoadBalancerId` | String | No | Filter by associated ALB |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `ServerGroups` | Array | Server group list |
| `ServerGroups[].ServerGroupId` | String | Server group ID |
| `ServerGroups[].ServerGroupName` | String | Name |
| `ServerGroups[].ServerGroupType` | String | `instance` or `ip` |
| `ServerGroups[].Servers` | Array | Backend server list |
| `ServerGroups[].HealthCheckConfig` | Object | Health check config |

### 5.3 DeleteServerGroup

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ServerGroupId` | String | Yes | Server group ID |

### 5.4 AddServersToGroup

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ServerGroupId` | String | Yes | Server group ID |
| `Servers` | Array | Yes | Server list (JSON) |

**Server format:**
```json
[
  {
    "ServerId": "i-xxx",
    "Port": 8080,
    "Weight": 100,
    "ServerType": "ecs"
  },
  {
    "ServerId": "10.0.1.5",
    "Port": 8080,
    "Weight": 100,
    "ServerType": "ip"
  }
]
```

### 5.5 RemoveServersFromGroup

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ServerGroupId` | String | Yes | Server group ID |
| `ServerIds` | Array | Yes | Server IDs to remove |

**ServerIds format:**
```json
[
  {
    "ServerId": "i-xxx",
    "ServerType": "ecs"
  }
]
```

### 5.6 ModifyServerGroupAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ServerGroupId` | String | Yes | Server group ID |
| `ServerGroupName` | String | No | New name |
| `HealthCheckConfig` | Object | No | Updated health check config |
| `Description` | String | No | New description |

---

## 6. Response Parsing

### Extract ALB ID from CreateLoadBalancer

```bash
ALB_ID=$(ve alb CreateLoadBalancer --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --Type "private" | jq -r '.Result.LoadBalancerId')
```

### List ALBs with Listeners and Rules

```bash
ALB_ID="alb-xxx"
echo "=== ALB ==="
ve alb DescribeLoadBalancers --Region "$VOLCENGINE_REGION" --LoadBalancerIds "[\"$ALB_ID\"]" | jq '.Result.LoadBalancers[0]'

echo "=== Listeners ==="
ve alb DescribeListeners --Region "$VOLCENGINE_REGION" --LoadBalancerId "$ALB_ID" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'

echo "=== Rules ==="
for LSNR_ID in $(ve alb DescribeListeners --Region "$VOLCENGINE_REGION" --LoadBalancerId "$ALB_ID" | jq -r '.Result.Listeners[].ListenerId'); do
  echo "Listener $LSNR_ID rules:"
  ve alb DescribeRules --Region "$VOLCENGINE_REGION" --ListenerId "$LSNR_ID" | jq -r '.Result.Rules[] | "  \(.RuleId) domain=\(.Domain) url=\(.Url)"'
done
```

---

## 7. Go SDK Examples

### CreateLoadBalancer

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/alb"
)

func main() {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    // CreateLoadBalancer: creates an ALB instance
    resp, err := instance.CreateLoadBalancer(&alb.CreateLoadBalancerInput{
        VpcId:            "vpc-xxx",
        SubnetId:         "subnet-xxx",
        LoadBalancerName: "prod-alb",
        Type:             "private",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("ALB ID: %s\n", resp.Result.LoadBalancerId)
}
```

### DescribeLoadBalancers

```go
// listALBs: lists all ALB instances
func listALBs() {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.DescribeLoadBalancers(&alb.DescribeLoadBalancersInput{})
    if err != nil {
        panic(err)
    }
    for _, lb := range resp.Result.LoadBalancers {
        fmt.Printf("%s: %s (%s)\n", lb.LoadBalancerId, lb.LoadBalancerName, lb.Status)
    }
}
```

### CreateListener

```go
// createHTTPListener: creates an HTTP listener on the ALB
func createHTTPListener(albID string, port int) {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateListener(&alb.CreateListenerInput{
        LoadBalancerId: albID,
        Protocol:       "HTTP",
        Port:           port,
        ListenerName:   "http-listener",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Listener ID: %s\n", resp.Result.ListenerId)
}
```

### CreateServerGroup

```go
// createServerGroup: creates a server group
func createServerGroup(name string) {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateServerGroup(&alb.CreateServerGroupInput{
        ServerGroupName: name,
        ServerGroupType: "instance",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Server Group ID: %s\n", resp.Result.ServerGroupId)
}
```

### CreateRule

```go
// createPathRule: creates a path-based routing rule
func createPathRule(listenerID, serverGroupID, path string) {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateRule(&alb.CreateRuleInput{
        ListenerId:    listenerID,
        ServerGroupId: serverGroupID,
        Url:           path,
        RuleName:      "path-rule",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Rule ID: %s\n", resp.Result.RuleId)
}
```

### AddServersToGroup

```go
// addServers: adds ECS instances to a server group
func addServers(groupID string, serverIDs []string, port int) {
    instance := alb.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    var servers []alb.ServerForAdd
    for _, id := range serverIDs {
        servers = append(servers, alb.ServerForAdd{
            ServerId:   id,
            Port:       port,
            Weight:     100,
            ServerType: "ecs",
        })
    }

    resp, err := instance.AddServersToGroup(&alb.AddServersToGroupInput{
        ServerGroupId: groupID,
        Servers:       servers,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Added %d servers\n", len(resp.Result.Servers))
}
```

---

*This reference document is part of the ve-alb-ops skill.*