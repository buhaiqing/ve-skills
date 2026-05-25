# API & SDK Usage — Volcengine CLB

> **Purpose:** Detailed API reference for CLB operations.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [API Overview](#1-api-overview)
2. [Load Balancer Operations](#2-load-balancer-operations)
3. [Listener Operations](#3-listener-operations)
4. [Backend Server Operations](#4-backend-server-operations)
5. [Health Check Operations](#5-health-check-operations)
6. [Response Parsing](#6-response-parsing)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `clb` |
| API Version | `2020-04-01` |
| Endpoint | `clb.volcengineapi.com` |
| Protocol | HTTPS |

### Common Response Structure

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "...", "Service": "clb" },
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
| `LoadBalancerName` | String | No | CLB name | `prod-clb` |
| `Type` | String | No | CLB type | `public`, `private` |
| `Description` | String | No | Description | — |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `LoadBalancerId` | String | CLB ID |

### 2.2 DescribeLoadBalancers

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerIds.N` | String Array | No | Filter by CLB IDs |
| `VpcId` | String | No | Filter by VPC |
| `LoadBalancerName` | String | No | Filter by name |
| `Type` | String | No | Filter by type |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `LoadBalancers` | Array | CLB list |
| `LoadBalancers[].LoadBalancerId` | String | CLB ID |
| `LoadBalancers[].LoadBalancerName` | String | CLB name |
| `LoadBalancers[].VpcId` | String | Parent VPC |
| `LoadBalancers[].Type` | String | Type |
| `LoadBalancers[].Status` | String | Status (`active`) |
| `LoadBalancers[].Address` | String | Internal IP |
| `LoadBalancers[].EipAddress` | String | Public IP |

### 2.3 DeleteLoadBalancer

**Prerequisites:** All listeners and backend servers removed.

### 2.4 ModifyLoadBalancerAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | CLB ID |
| `LoadBalancerName` | String | No | New name |
| `Description` | String | No | New description |

---

## 3. Listener Operations

### 3.1 CreateListener

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `LoadBalancerId` | String | Yes | CLB ID | `clb-xxx` |
| `Protocol` | String | Yes | Protocol | `TCP`, `UDP`, `HTTP`, `HTTPS` |
| `Port` | Integer | Yes | Listening port | `80` |
| `ListenerName` | String | No | Listener name | `http-listener` |

### 3.2 DescribeListeners

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Listeners` | Array | Listener list |
| `Listeners[].ListenerId` | String | Listener ID |
| `Listeners[].Protocol` | String | Protocol |
| `Listeners[].Port` | Integer | Port |
| `Listeners[].Status` | String | Status |

### 3.3 DeleteListener

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ListenerId` | String | Yes | Listener ID |

### 3.4 ModifyListenerAttributes

For modifying listener name, description, and other attributes.

---

## 4. Backend Server Operations

### 4.1 AddBackendServers

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | CLB ID |
| `BackendServers` | String Array | Yes | Backend server list (JSON) |

Backend server format:

```json
[
  {
    "ServerId": "i-xxx",
    "Port": 8080,
    "Weight": 100
  }
]
```

### 4.2 DescribeBackendServers

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `BackendServers` | Array | Backend server list |
| `BackendServers[].ServerId` | String | ECS instance ID |
| `BackendServers[].Port` | Integer | Backend port |
| `BackendServers[].Weight` | Integer | Weight |
| `BackendServers[].Status` | String | `healthy`, `unhealthy`, `unavailable` |

### 4.3 RemoveBackendServers

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `LoadBalancerId` | String | Yes | CLB ID |
| `ServerIds.N` | String Array | Yes | Server IDs to remove |

---

## 5. Health Check Operations

### 5.1 SetHealthCheckConfig

**Request Parameters:**

| Parameter | Type | Required | Description | Default |
|-----------|------|----------|-------------|---------|
| `LoadBalancerId` | String | Yes | CLB ID | — |
| `ListenerId` | String | Yes | Listener ID | — |
| `HealthyThreshold` | Integer | No | Healthy count | 3 |
| `UnhealthyThreshold` | Integer | No | Unhealthy count | 3 |
| `Interval` | Integer | No | Check interval (s) | 5 |
| `Timeout` | Integer | No | Check timeout (s) | 3 |
| `HttpMethod` | String | No | GET/HEAD/POST | GET |
| `Uri` | String | No | Health check path | / |
| `HealthyHttpCode` | String | No | Expected codes | 200 |

---

## 6. Response Parsing

### Extract CLB ID

```bash
CLB_ID=$(ve clb CreateLoadBalancer --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --Type "private" \
  | jq -r '.Result.LoadBalancerId')
```

### List CLB + Listeners + Backends

```bash
CLB_ID="clb-xxx"
echo "=== Load Balancer ==="
ve clb DescribeLoadBalancers --Region "$VOLCENGINE_REGION" --LoadBalancerIds "[\"$CLB_ID\"]" | jq '.Result.LoadBalancers[0]'

echo "=== Listeners ==="
ve clb DescribeListeners --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'

echo "=== Backend Servers ==="
ve clb DescribeBackendServers --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" | jq -r '.Result.BackendServers[] | "\(.ServerId):\(.Port) weight=\(.Weight) status=\(.Status)"'
```

---

*This reference document is part of the ve-clb-ops skill.*
