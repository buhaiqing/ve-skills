# Core Concepts — Volcengine NAT Gateway (NAT网关)

> **Purpose:** Fundamental concepts for Volcengine NAT Gateway, SNAT rules, and DNAT rules.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is NAT Gateway](#1-what-is-nat-gateway)
2. [SNAT (Source NAT)](#2-snat-source-nat)
3. [DNAT (Destination NAT)](#3-dnat-destination-nat)
4. [NAT Gateway Specifications](#4-nat-gateway-specifications)
5. [NAT Gateway Lifecycle](#5-nat-gateway-lifecycle)
6. [Regions Supporting NAT](#6-regions-supporting-nat)
7. [Limits and Quotas](#7-limits-and-quotas)

---

## 1. What is NAT Gateway

A **NAT Gateway (NAT网关)** is a managed network address translation service that enables resources in private subnets (without public IPs) to access the internet (via SNAT) and be accessed from the internet on specific ports (via DNAT).

### Key Characteristics

- **Managed Service:** Fully managed by Volcengine — no infrastructure to maintain
- **High Availability:** Built-in HA within the availability zone
- **SNAT + DNAT:** Supports both outbound and inbound address translation
- **EIP Dependency:** Requires at least one EIP bound for translation

### NAT Gateway Attributes

| Attribute | Description |
|-----------|-------------|
| `NatGatewayId` | Unique identifier, format `ngw-xxxxxxxxx` |
| `NatGatewayName` | User-defined name |
| `Description` | User-defined description |
| `NatGatewaySpec` | Specification: `Small`, `Medium`, `Large` |
| `VpcId` | Parent VPC ID |
| `SubnetId` | Deployment subnet ID |
| `Status` | `Available` or `Pending` |
| `BandwidthPackageIds` | Associated bandwidth packages |

---

## 2. SNAT (Source NAT)

**SNAT (源地址转换)** translates private source IP addresses to public IP addresses for outbound internet traffic.

### How SNAT Works

```
Private ECS (10.0.2.10) → NAT Gateway → EIP (120.78.x.x) → Internet
```

### SNAT Rule Attributes

| Attribute | Description |
|-----------|-------------|
| `SnatRuleId` | SNAT rule ID, format `snat-xxxxxxxxx` |
| `SnatRuleName` | User-defined name |
| `NatGatewayId` | Parent NAT Gateway ID |
| `SourceCidr` | Source CIDR block (private subnet) |
| `EipAddresses` | List of EIP addresses used for translation |
| `Status` | Rule status (`Available`) |

### SNAT Use Cases

| Scenario | Configuration |
|----------|--------------|
| Single subnet outbound | SourceCidr = subnet CIDR, 1 EIP |
| Multiple subnets outbound | Multiple SNAT rules, shared pool of EIPs |
| High-throughput outbound | Multiple EIPs in SNAT rule for load distribution |

### SNAT with Multiple EIPs

When multiple EIPs are bound to a SNAT rule, outbound connections are distributed across the EIPs. This increases the maximum concurrent connection count:

| EIPs per SNAT Rule | Max Concurrent Connections |
|--------------------|---------------------------|
| 1 | ~200,000 |
| 2 | ~400,000 |
| N | ~N × 200,000 |

---

## 3. DNAT (Destination NAT)

**DNAT (目的地址转换)** maps external public IP:port to internal private IP:port for inbound traffic.

### How DNAT Works

```
Internet → EIP:ExternalPort → NAT Gateway → Private ECS:InternalPort
```

### DNAT Rule Attributes

| Attribute | Description |
|-----------|-------------|
| `DnatRuleId` | DNAT rule ID, format `dnat-xxxxxxxxx` |
| `DnatRuleName` | User-defined name |
| `NatGatewayId` | Parent NAT Gateway ID |
| `EipAddress` | EIP address for external access |
| `IpProtocol` | Protocol type: `TCP` or `UDP` |
| `ExternalPort` | External (public) port |
| `InternalIp` | Target private IP |
| `InternalPort` | Target private port |
| `Status` | Rule status (`Available`) |

### DNAT Examples

| Rule | External | Internal | Use Case |
|------|----------|----------|----------|
| 1 | `120.78.x.x:8080` | `10.0.2.10:80` | Access web server via port 8080 |
| 2 | `120.78.x.x:2222` | `10.0.2.10:22` | SSH via non-standard port |
| 3 | `120.78.x.x:3306` | `10.0.3.10:3306` | Remote MySQL access |

---

## 4. NAT Gateway Specifications

| Specification | Max Bandwidth | Max Connections | Use Case |
|---------------|---------------|-----------------|----------|
| `Small` | 1 Gbps | 100,000 | Development, testing, low traffic |
| `Medium` | 5 Gbps | 500,000 | Production, medium traffic |
| `Large` | 10 Gbps | 1,000,000 | High-traffic production |

### Spec Upgrade

- Spec can be changed at any time (upgrade or downgrade)
- Change takes effect immediately
- Existing connections are NOT interrupted

---

## 5. NAT Gateway Lifecycle

```
CreateNatGateway ──────▶ Pending ──────▶ Available
    │                                          │
    │                                          ▼
    │                                    ModifyNatGatewayAttribute
    │                                          │
    │                                          ▼
    │                                    DeleteNatGateway
    │                                          │
    ▼                                          ▼
  Delete requires: No SNAT rules, no DNAT rules,
  EIP unbound
```

---

## 6. Regions Supporting NAT

| Region | RegionID | Status |
|--------|----------|--------|
| 华北2 (北京) | `cn-beijing` | Commercial |
| 华东2 (上海) | `cn-shanghai` | Commercial |
| 华南1 (广州) | `cn-guangzhou` | Commercial |
| 中国香港 | `cn-hongkong` | Commercial |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |
| 亚太东南 (雅加达) | `ap-southeast-3` | Commercial |

---

## 7. Limits and Quotas

| Resource | Default Quota (per region) | Notes |
|----------|---------------------------|-------|
| NAT Gateways per VPC | 5 | Can request increase |
| SNAT rules per NAT Gateway | 20 | — |
| DNAT rules per NAT Gateway | 50 | — |
| EIPs per NAT Gateway (via binding) | 20 | Limited by EIP quota |
| SNAT rule max concurrent connections | ~200,000 per EIP | Scales with EIP count |

---

*This reference document is part of the ve-nat-ops skill.*
