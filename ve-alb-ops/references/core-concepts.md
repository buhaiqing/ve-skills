# Core Concepts — Volcengine ALB (应用型负载均衡)

> **Purpose:** Fundamental concepts for Volcengine ALB (Application Load Balancer).
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [What is ALB](#1-what-is-alb)
2. [ALB Types](#2-alb-types)
3. [Listeners](#3-listeners)
4. [Routing Rules](#4-routing-rules)
5. [Server Groups](#5-server-groups)
6. [Health Checks](#6-health-checks)
7. [Load Balancing Algorithms](#7-load-balancing-algorithms)
8. [Regions Supporting ALB](#8-regions-supporting-alb)
9. [Limits and Quotas](#9-limits-and-quotas)
10. [ALB vs CLB](#10-alb-vs-clb)

---

## 1. What is ALB

An **ALB (Application Load Balancer / 应用型负载均衡)** is a Layer 7 load balancing service that distributes application-level traffic (HTTP/HTTPS/gRPC) across multiple backend servers based on content (URL path, host header, HTTP method) rather than just IP and port.

### Key Characteristics

- **Layer 7 Only:** HTTP, HTTPS, and gRPC protocol support
- **Content-Based Routing:** Route traffic by URL path, host header, HTTP method, query string
- **HTTPS Termination:** Offload TLS/SSL processing to the load balancer
- **WebSocket Support:** Native WebSocket upgrade support
- **gRPC Support:** HTTP/2 based gRPC load balancing
- **Fine-Grained Health Checks:** Per-server-group health check configuration

### ALB Attributes

| Attribute | Description |
|-----------|-------------|
| `LoadBalancerId` | Unique identifier, format `alb-xxxxxxxxx` |
| `LoadBalancerName` | User-defined name |
| `VpcId` | Parent VPC ID |
| `SubnetId` | Deployment subnet ID |
| `Type` | `public` or `private` |
| `Address` | Internal IP address in VPC |
| `Status` | `active`, `inactive`, `creating`, `deleting` |
| `EipAddress` | EIP address (for public ALB) |
| `CreateTime` | Creation timestamp |

---

## 2. ALB Types

| Type | Code | Network Access | Use Case |
|------|------|---------------|----------|
| **Public** | `public` | Accessible from internet | Web applications, API endpoints |
| **Private** | `private` | Accessible only within VPC | Internal microservices, service mesh |

### Public ALB

- Automatically assigned EIP for internet access
- Suitable for external-facing applications
- Can restrict access with security group rules

### Private ALB

- Only accessible from within the same VPC
- No public endpoint, internal IP only
- Suitable for internal microservices communication

---

## 3. Listeners

A **Listener** defines the protocol and port on which the ALB accepts incoming connections.

### Protocol Types

| Protocol | Description | Use Case |
|----------|-------------|----------|
| HTTP | HTTP/1.1 protocol | Standard web applications |
| HTTPS | HTTP with TLS encryption | Secure web applications |
| gRPC | HTTP/2 based gRPC | Microservice RPC calls |

### Listener Attributes

| Attribute | Description |
|-----------|-------------|
| `ListenerId` | Unique listener ID, format `lsnr-xxxxxxxxx` |
| `Protocol` | HTTP, HTTPS, gRPC |
| `Port` | Listening port (1–65535) |
| `LoadBalancerId` | Parent ALB ID |
| `CertificateId` | TLS certificate ID (HTTPS only) |
| `TLSPolicy` | TLS version policy |
| `DefaultServerGroupId` | Default server group for unmatched requests |

---

## 4. Routing Rules

**Rules** define how the ALB directs traffic based on request content. ALB evaluates rules in priority order.

### Rule Match Criteria

| Criterion | Description | Example |
|-----------|-------------|---------|
| **Domain** | Match by host header | `api.example.com` |
| **URL** | Match by request path | `/api/*`, `/images/*` |
| **Method** | Match by HTTP method | `GET`, `POST` |
| **Query String** | Match by query parameters | `?version=v2` |

### Rule Actions

| Action | Description |
|--------|-------------|
| **Forward** | Forward to a server group |
| **Redirect** | Redirect to another URL |
| **Fixed Response** | Return a fixed HTTP response |

### Rule Priority

Rules are evaluated in order of priority (lower number = higher priority). Default listener forwarding is used when no rule matches.

---

## 5. Server Groups

A **Server Group** (also called target group) is a logical grouping of backend servers that receive traffic from the ALB.

### Server Group Attributes

| Attribute | Description |
|-----------|-------------|
| `ServerGroupId` | Unique ID, format `rsp-xxxxxxxxx` |
| `ServerGroupName` | User-defined name |
| `ServerGroupType` | `instance` (ECS) or `ip` (IP address) |
| `Servers` | List of backend servers |
| `HealthCheckConfig` | Health check configuration |

### Backend Server Attributes

| Attribute | Description |
|-----------|-------------|
| `ServerId` | ECS instance ID or IP address |
| `Port` | Backend service port |
| `Weight` | Traffic weight (0–100) |
| `ServerType` | `ecs` or `ip` |
| `Status` | `healthy`, `unhealthy`, or `unavailable` |

### Server Weight Behavior

| Weight | Behavior |
|--------|----------|
| 0 | No traffic (drain mode) |
| 1–50 | Proportionally less traffic |
| 51–100 | Proportionally more traffic |
| 100 | Equal distribution (with same-weight servers) |

---

## 6. Health Checks

Health checks monitor backend server availability. Unhealthy servers are automatically removed from the load balancing pool.

### Health Check Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `Enabled` | Enable/disable health checks | `true` |
| `Protocol` | Health check protocol: TCP, HTTP | HTTP |
| `Method` | HTTP method: GET, HEAD | GET |
| `Uri` | Health check path (HTTP only) | `/` |
| `Timeout` | Health check timeout in seconds | 3 |
| `Interval` | Seconds between checks | 5 |
| `HealthyThreshold` | Consecutive successes to mark healthy | 3 |
| `UnhealthyThreshold` | Consecutive failures to mark unhealthy | 3 |
| `HealthyHttpCode` | Expected HTTP status codes | `200` |

---

## 7. Load Balancing Algorithms

| Algorithm | Description |
|-----------|-------------|
| Weighted Round Robin | Distributes requests based on server weight (default) |
| Least Connections | Sends to server with fewest active connections |
| Source IP Hash | Same client IP → same backend server |

---

## 8. Regions Supporting ALB

| Region | RegionID | Status |
|--------|----------|--------|
| 华北2 (北京) | `cn-beijing` | Commercial |
| 华东2 (上海) | `cn-shanghai` | Commercial |
| 华南1 (广州) | `cn-guangzhou` | Commercial |
| 中国香港 | `cn-hongkong` | Commercial |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |

---

## 9. Limits and Quotas

| Resource | Default Quota | Notes |
|----------|---------------|-------|
| ALBs per account (per region) | 20 | Can request increase |
| Listeners per ALB | 50 | — |
| Rules per listener | 100 | — |
| Server groups per ALB | 50 | — |
| Servers per server group | 200 | — |
| Certificates per ALB | 20 | HTTPS listeners only |

---

## 10. ALB vs CLB

| Feature | ALB | CLB |
|---------|-----|-----|
| OSI Layer | Layer 7 (Application) | Layer 4 + Layer 7 |
| Protocols | HTTP, HTTPS, gRPC | TCP, UDP, HTTP, HTTPS |
| Content-Based Routing | ✅ URL path, host, query | ❌ |
| Server Groups | ✅ Separate resource | ❌ (direct backend) |
| WebSocket | ✅ Native | ❌ |
| gRPC | ✅ Native | ❌ |
| HTTPS Termination | ✅ | ✅ (limited) |
| Fixed Response | ✅ | ❌ |
| Redirect | ✅ | ❌ |
| Use Case | Microservices, API Gateway | General L4/L7 load balancing |

---

*This reference document is part of the ve-alb-ops skill.*