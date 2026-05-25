# Core Concepts — Volcengine EIP (弹性公网IP)

> **Purpose:** Fundamental concepts for Volcengine EIP (Elastic IP Address). This reference provides the conceptual foundation needed to understand and operate the `ve-eip-ops` skill.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is EIP](#1-what-is-eip)
2. [EIP Billing Types](#2-eip-billing-types)
3. [Bandwidth Types](#3-bandwidth-types)
4. [Line Types](#4-line-types)
5. [EIP Lifecycle and Status](#5-eip-lifecycle-and-status)
6. [Binding Types](#6-binding-types)
7. [Regions Supporting EIP](#7-regions-supporting-eip)
8. [Limits and Quotas](#8-limits-and-quotas)

---

## 1. What is EIP

An **EIP (Elastic IP Address / 弹性公网IP)** is a static public IP address that can be dynamically associated with and disassociated from cloud resources such as ECS instances, CLB, NAT Gateways, and ENIs.

### Key Characteristics

- **Dynamic Binding:** Can be bound to any supported resource type, enabling flexible network topologies
- **Independent Lifecycle:** Exists independently of the resource it's bound to
- **Static IP:** The IP address does not change unless explicitly released and reallocated
- **Bandwidth Adjustable:** Bandwidth can be modified at any time without releasing the EIP

### EIP Attributes

| Attribute | Description |
|-----------|-------------|
| `AllocationId` | Unique identifier, format `eipalloc-xxxxxxxxx` |
| `EipAddress` | Actual public IP address (e.g., `120.78.x.x`) |
| `Status` | `Available` or `InUse` |
| `LineType` | Line type: `BGP` |
| `Bandwidth` | Bandwidth in Mbps |
| `ChargeType` | Billing type: `PostPaid` or `PrePaid` |
| `InstanceType` | Type of bound instance (e.g., `EcsInstance`, `ClbInstance`, `Nat`) |
| `InstanceId` | ID of the bound instance (empty when unbound) |
| `Name` | User-defined name |
| `Description` | User-defined description |
| `AvailableRegions` | Available regions for this EIP |

---

## 2. EIP Billing Types

| Billing Type | Code | Description |
|--------------|------|-------------|
| **PostPaid (按量计费)** | `PostPaid` | Pay-as-you-go; billed by actual usage time |
| **PrePaid (包年包月)** | `PrePaid` | Monthly/yearly subscription; discounted rate |

### PostPaid Billing Details

| Billing Item | Unit | Description |
|--------------|------|-------------|
| IP Holding Fee | CNY/hour | Charged while EIP is unbound (`Available`) |
| Traffic Fee | CNY/GB | Charged when billing by traffic |
| Bandwidth Fee | CNY/Mbps/hour | Charged when billing by bandwidth |

> **Tip:** EIPs bound to running resources (`InUse`) typically have no IP holding fee — only the bound resource's billing applies.

---

## 3. Bandwidth Types

| Type | Code | Description |
|------|------|-------------|
| **PayByTraffic (按流量计费)** | `PayByTraffic` | Billed per GB of data transferred |
| **PayByBandwidth (按带宽计费)** | `PayByBandwidth` | Billed per Mbps of allocated bandwidth |

### Bandwidth Limits

| Type | Range per EIP | Default Max |
|------|---------------|-------------|
| PayByTraffic | 1–200 Mbps | 200 Mbps |
| PayByBandwidth | 1–500 Mbps | 500 Mbps |

---

## 4. Line Types

| Line Type | Code | Description | Regions |
|-----------|------|-------------|---------|
| **BGP** | `BGP` | Border Gateway Protocol, multi-ISP routing | All commercial regions |

BGP is the default and only line type in most Volcengine regions. It provides optimal routing across China Unicom, China Telecom, China Mobile, and other ISPs.

---

## 5. EIP Lifecycle and Status

```
[AllocateEipAddress] ──────▶ Available ──────▶ [AssociateEipAddress] ──────▶ InUse
                                │                     ▲                          │
                                │                     │                          │
                                │              [DisassociateEipAddress]          │
                                │                                                │
                                ▼                                                ▼
                          [ReleaseEipAddress]                            [DisassociateEipAddress]
                                │                                                │
                                ▼                                                ▼
                            Deleted                                        Available
```

### Status Definitions

| Status | Meaning | Allowed Operations |
|--------|---------|-------------------|
| `Available` | EIP is allocated but not bound | Associate, Modify, Release, Tag |
| `InUse` | EIP is bound to an instance | Disassociate, Modify, Tag |

---

## 6. Binding Types

EIP can be bound to these resource types:

| Instance Type | Description | Use Case |
|---------------|-------------|----------|
| `EcsInstance` | ECS instance | Direct internet access for compute |
| `ClbInstance` | CLB load balancer | Public-facing load balancer |
| `Nat` | NAT Gateway | SNAT/DNAT for private subnets |
| `HaVip` | High-availability VIP | HA cluster floating IP |
| `NetworkInterface` | Elastic network interface | Custom network attachment |

### Binding Rules

- One EIP → one resource at a time
- One resource → one EIP at a time (for ECS)
- Must be in the same region
- EIP must be `Available` to bind
- Resource must not already have an EIP (for ECS)

---

## 7. Regions Supporting EIP

| Region | RegionID | Status |
|--------|----------|--------|
| 华北2 (北京) | `cn-beijing` | Commercial |
| 华东2 (上海) | `cn-shanghai` | Commercial |
| 华南1 (广州) | `cn-guangzhou` | Commercial |
| 中国香港 | `cn-hongkong` | Commercial |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |
| 亚太东南 (雅加达) | `ap-southeast-3` | Commercial |

### Query Supported Regions

```bash
ve eip DescribeRegions
```

---

## 8. Limits and Quotas

| Resource | Default Quota (per region) | Notes |
|----------|---------------------------|-------|
| EIPs per account | 20 | Can request increase |
| Max bandwidth per EIP (traffic) | 200 Mbps | — |
| Max bandwidth per EIP (bandwidth billing) | 500 Mbps | — |
| Bindings per EIP | 1 | One resource at a time |
| Bandwidth modifications per day | Unlimited | — |

> **Note:** Quotas may vary by region and account type. Contact Volcengine support for quota increases.

---

*This reference document is part of the ve-eip-ops skill. For operational procedures, see the main SKILL.md.*
