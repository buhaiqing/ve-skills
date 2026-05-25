# Core Concepts — Volcengine CLB (负载均衡)

> **Purpose:** Fundamental concepts for Volcengine CLB (Classic Load Balancer).
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is CLB](#1-what-is-clb)
2. [CLB Types](#2-clb-types)
3. [Listeners](#3-listeners)
4. [Backend Servers](#4-backend-servers)
5. [Health Checks](#5-health-checks)
6. [Load Balancing Algorithms](#6-load-balancing-algorithms)
7. [Regions Supporting CLB](#7-regions-supporting-clb)
8. [Limits and Quotas](#8-limits-and-quotas)

---

## 1. What is CLB

A **CLB (Classic Load Balancer / 负载均衡)** distributes incoming network traffic across multiple backend servers to achieve high availability and horizontal scalability.

### Key Characteristics

- **Layer 4 + Layer 7:** Supports TCP/UDP (L4) and HTTP/HTTPS (L7)
- **High Availability:** Built-in HA with automatic failover
- **Health Checking:** Automatic backend server health monitoring
- **Auto Scaling Compatible:** Works with Auto Scaling groups

### CLB Attributes

| Attribute | Description |
|-----------|-------------|
| `LoadBalancerId` | Unique identifier, format `clb-xxxxxxxxx` |
| `LoadBalancerName` | User-defined name |
| `VpcId` | Parent VPC ID |
| `SubnetId` | Deployment subnet ID |
| `Type` | `public` or `private` |
| `Address` | Internal IP address in VPC |
| `Status` | `active`, `inactive`, or `allocating` |
| `EipAddress` | EIP address (for public CLB) |
| `Listeners` | List of listeners |
| `BackendServers` | List of backend servers |

---

## 2. CLB Types

| Type | Code | Network Access | Use Case |
|------|------|---------------|----------|
| **Public** | `public` | Accessible from internet | Web applications, API endpoints |
| **Private** | `private` | Accessible only within VPC | Internal microservices, database proxies |

### Public CLB Requirements

- Requires EIP bound for public access
- EIP can be bound during creation or after
- Public CLB has both internal VPC IP and public EIP

---

## 3. Listeners

A **Listener** monitors a specific port/protocol on the CLB for incoming connections.

### Protocol Types

| Protocol | Layer | Description | Use Case |
|----------|-------|-------------|----------|
| TCP | Layer 4 | Raw TCP traffic | Databases, SSH, custom protocols |
| UDP | Layer 4 | Raw UDP traffic | DNS, streaming, gaming |
| HTTP | Layer 7 | HTTP protocol | Web applications |
| HTTPS | Layer 7 | HTTPS protocol | Encrypted web applications |

### Listener Attributes

| Attribute | Description |
|-----------|-------------|
| `ListenerId` | Listener ID |
| `Protocol` | TCP, UDP, HTTP, HTTPS |
| `Port` | Listening port (1–65535) |
| `LoadBalancerId` | Parent CLB ID |
| `Status` | Listener status |
| `HealthCheck` | Health check configuration |

---

## 4. Backend Servers

**Backend servers** are the ECS instances that receive distributed traffic from the CLB.

### Backend Server Attributes

| Attribute | Description |
|-----------|-------------|
| `ServerId` | ECS instance ID |
| `Port` | Backend port |
| `Weight` | Traffic distribution weight (0–100) |
| `ServerType` | `ecs` (ECS instance) |
| `Status` | `healthy`, `unhealthy`, or `unavailable` |

### Server Weight Behavior

| Weight | Behavior |
|--------|----------|
| 0 | No traffic sent (maintenance mode) |
| 1–50 | Receives proportionally less traffic |
| 51–100 | Receives more traffic |
| 100 | Standard weight (equal distribution with same-weight servers) |

---

## 5. Health Checks

Health checks automatically detect unhealthy backend servers and stop sending traffic to them.

### Health Check Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `HealthyThreshold` | Consecutive successes to mark healthy | 3 |
| `UnhealthyThreshold` | Consecutive failures to mark unhealthy | 3 |
| `Interval` | Seconds between health checks | 5 |
| `Timeout` | Health check timeout in seconds | 3 |
| `HttpMethod` | HTTP method (HTTP/HTTPS only) | GET |
| `Uri` | Health check path (HTTP/HTTPS only) | / |
| `HealthyHttpCode` | Expected HTTP status codes | 200 |

---

## 6. Load Balancing Algorithms

| Algorithm | Protocol | Description |
|-----------|----------|-------------|
| Round Robin | TCP/UDP/HTTP | Rotates through backends in order |
| Weighted Round Robin | TCP/UDP/HTTP | Considers server weight |
| Least Connections | TCP | Sends to server with fewest connections |
| Source IP Hash | TCP/UDP | Same client → same server |
| URL Hash | HTTP | Same URL path → same server |

---

## 7. Regions Supporting CLB

| Region | RegionID | Status |
|--------|----------|--------|
| 华北2 (北京) | `cn-beijing` | Commercial |
| 华东2 (上海) | `cn-shanghai` | Commercial |
| 华南1 (广州) | `cn-guangzhou` | Commercial |
| 中国香港 | `cn-hongkong` | Commercial |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |

---

## 8. Limits and Quotas

| Resource | Default Quota | Notes |
|----------|---------------|-------|
| CLBs per account (per region) | 20 | Can request increase |
| Listeners per CLB | 50 | — |
| Backend servers per CLB | 50 | — |
| Backend servers per listener | 50 | — |
| Ports per listener | 1 | — |

---

*This reference document is part of the ve-clb-ops skill.*
