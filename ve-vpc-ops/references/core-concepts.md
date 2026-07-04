# Core Concepts — Volcengine VPC (私有网络)

> **Purpose:** Fundamental concepts for Volcengine VPC, Subnets, Route Tables, and Route Entries. This reference provides the conceptual foundation needed to understand and operate the `ve-vpc-ops` skill.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is VPC](#1-what-is-vpc)
2. [CIDR Block Planning](#2-cidr-block-planning)
3. [Subnet](#3-subnet)
4. [Route Table and Route Entries](#4-route-table-and-route-entries)
5. [Network Topology Patterns](#5-network-topology-patterns)
6. [Regions and Availability Zones](#6-regions-and-availability-zones)
7. [Resource Lifecycle](#7-resource-lifecycle)
8. [Limits and Quotas](#8-limits-and-quotas)

---

## 1. What is VPC

A **VPC (Virtual Private Cloud / 私有网络)** is a logically isolated, private network on Volcengine (火山引擎). It provides the foundation for deploying cloud resources (ECS, RDS, Redis, CLB).

### Key Characteristics

- 🔒 **Isolation** — each VPC is completely isolated from others
- ✏️ **Customizable** — define IP range, subnets, and routing
- 🛡️ **Secure** — built-in network ACL and security group support
- 🔄 **Flexible** — support for IPv4 and IPv6 dual-stack

### VPC Attributes

| Attribute | Description |
|-----------|-------------|
| `VpcId` | Unique identifier, format `vpc-xxxxxxxxx` |
| `VpcName` | User-defined name (1–128 characters) |
| `CidrBlock` | IPv4 CIDR block (e.g., `10.0.0.0/16`) |
| `Ipv6CidrBlock` | IPv6 CIDR block (if enabled) |
| `Status` | `Available` or `Pending` |
| `SecondaryCidrBlocks` | Additional CIDR blocks attached to the VPC |
| `RouteTableIds` | Associated route table IDs |
| `SubnetIds` | List of subnet IDs within this VPC |
| `NatGatewayIds` | List of NAT gateway IDs |
| `SecurityGroupIds` | List of security group IDs |
| `CreationTime` | ISO 8601 timestamp |
| `UpdateTime` | ISO 8601 timestamp |

### VPC Status Lifecycle

```
[Creating/Pending] ──────▶ [Available] ──────▶ [Deleting] ──────▶ [Deleted]
```

- ✅ **Available** — ready for use
- 🔄 **Deleting** — being removed (must be empty: no subnets, instances, or other resources)

---

## 2. CIDR Block Planning

### Allowed VPC CIDR Ranges

| CIDR Block | Usable IP Addresses | Suitable For |
|------------|---------------------|--------------|
| `10.0.0.0/8` | 16,777,216 | Large enterprise networks |
| `172.16.0.0/12` | 1,048,576 | Medium deployments |
| `192.168.0.0/16` | 65,536 | Small deployments |

### Planning Best Practices

- ❌ **No Overlapping CIDRs** — VPC CIDRs must not overlap with each other or on-premises networks (Direct Connect)
- 📈 **Plan for Growth** — reserve sufficient address space for future subnets & resources
- 🏗️ **Segment by Tier** — separate subnets for web, app, and database tiers
- 🔗 **Consider Interconnection** — if using VPC Peering or Transit Gateway, plan non-overlapping ranges across all VPCs

### Example CIDR Plan

```
VPC: 10.0.0.0/16
├── Subnet-A (10.0.1.0/24)  → Web servers (cn-beijing-a)
├── Subnet-B (10.0.2.0/24)  → Application servers (cn-beijing-b)
├── Subnet-C (10.0.3.0/24)  → Database servers (cn-beijing-c)
└── Subnet-D (10.0.4.0/24)  → NAT Gateway / Management (cn-beijing-a)
```

### Subnet CIDR Rules

- 📏 Mask length: `/16` to `/29`
- 🎯 Must be a subset of the VPC CIDR
- 🚫 Cannot overlap with route entry destination CIDRs
- 🔒 First IP and last 3 IPs in each subnet are reserved by the system:
  - `192.168.0.0` — Network address
  - `192.168.0.253`, `192.168.0.254`, `192.168.0.255` — Reserved by system

---

## 3. Subnet

A **Subnet (子网)** is a subdivision of a VPC's CIDR block within a specific availability zone.

### Key Attributes

| Attribute | Description |
|-----------|-------------|
| `SubnetId` | Unique identifier, format `subnet-xxxxxxxxx` |
| `SubnetName` | User-defined name |
| `CidrBlock` | Subnet IPv4 CIDR |
| `ZoneId` | Availability zone (e.g., `cn-beijing-a`) |
| `VpcId` | Parent VPC ID |
| `AvailableIpAddressCount` | Number of remaining usable IPs |
| `TotalIpAddressCount` | Total number of IPs in the subnet |
| `Status` | `Available` or `Pending` |
| `IsDefault` | Whether it's a default subnet |
| `Ipv6CidrBlock` | IPv6 CIDR (if enabled) |
| `RouteTable` | Associated route table (ID + type) |
| `NetworkAclId` | Associated network ACL |

### Subnet Types

| Type | Description |
|------|-------------|
| **Default Subnet** | Automatically created when creating an ECS instance without specifying a subnet |
| **Custom Subnet** | Manually created by the user |

### Availability Zone Deployment

Deploy subnets across multiple AZs for high availability:

```
VPC: 10.0.0.0/16
├── cn-beijing-a: 10.0.1.0/24 (Web tier)
├── cn-beijing-b: 10.0.2.0/24 (App tier)
└── cn-beijing-c: 10.0.3.0/24 (DB tier)
```

---

## 4. Route Table and Route Entries

### Route Table (路由表)

A **Route Table** controls network traffic routing for resources associated with it. Each subnet is associated with exactly one route table.

| Attribute | Description |
|-----------|-------------|
| `RouteTableId` | Unique identifier |
| `RouteTableName` | User-defined name |
| `VpcId` | Parent VPC ID |
| `RouteTableType` | `System` (default) or `Custom` |
| `RouteEntryCount` | Number of route entries |

### Route Entry Types

| Route Entry Type | Description | Auto-created |
|------------------|-------------|--------------|
| **System Route** | Manages traffic to the internet and within the VPC | Yes |
| **Custom Route** | User-defined routes for specific traffic patterns | No |

### System Routes

Each VPC and subnet automatically have system route entries for:

| Destination | Target | Purpose |
|-------------|--------|---------|
| VPC CIDR (e.g., `10.0.0.0/16`) | Local | Communication within the VPC |
| `0.0.0.0/0` | NAT Gateway | Internet access via NAT |

### Common Route Entry Next Hop Types

| Next Hop Type | Description |
|---------------|-------------|
| `Local` | Traffic to the local VPC CIDR (system route) |
| `NatGateway` | Traffic through a NAT Gateway |
| `.Instance` | Traffic to a specific ECS instance |
| `HaVip` | Traffic to a High-Availability Virtual IP |
| `NetworkInterface` | Traffic to a specific ENI |
| `VpnConnection` | Traffic through a VPN connection |

---

## 5. Network Topology Patterns

### Pattern 1: Single VPC, Single AZ

```
┌─────────────────────────────┐
│         VPC (10.0.0.0/16)   │
│  ┌─────────────────────┐    │
│  │  Subnet-A           │    │
│  │  (cn-beijing-a)     │    │
│  │  10.0.1.0/24        │    │
│  │                     │    │
│  │  [ECS][ECS][DB]     │    │
│  └─────────────────────┘    │
└─────────────────────────────┘
```

- **🎯 Use case:** Development, testing, single-AZ deployment
- **📊 Complexity:** Low

### Pattern 2: Multi-AZ High Availability

```
┌─────────────────────────────────────────┐
│              VPC (10.0.0.0/16)          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │ Subnet-A│ │ Subnet-B│ │ Subnet-C│   │
│  │ AZ-a    │ │ AZ-b    │ │ AZ-c    │   │
│  │ Web     │ │ App     │ │ DB      │   │
│  │ 10.0.1  │ │ 10.0.2  │ │ 10.0.3  │   │
│  └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────┘
```

- **🎯 Use case:** Production, high availability required
- **📊 Complexity:** Medium

### Pattern 3: Multi-tier with NAT Gateway

```
┌─────────────────────────────────────────────────────┐
│                    VPC (10.0.0.0/16)                │
│  ┌──────────────┐    ┌────────────────────────┐     │
│  │  Public Subnet│    │   Private Subnets       │     │
│  │  (10.0.1.0/24)│    │   ┌────────────────┐   │     │
│  │              │    │   │ 10.0.2.0/24    │   │     │
│  │  [NAT GW]    │    │   │   App Servers   │   │     │
│  │  [CLB]       │    │   │ 10.0.3.0/24    │   │     │
│  └──────────────┘    │   │   DB Servers    │   │     │
│                      │   └────────────────┘   │     │
│                      └────────────────────────┘     │
└─────────────────────────────────────────────────────┘
```

- **🎯 Use case:** Production with internet access for private subnets
- **📊 Complexity:** High
- **🔗 Requires:** `ve-nat-ops` for NAT Gateway, `ve-clb-ops` for CLB

### Pattern 4: Multiple VPCs with Peering

```
┌──────────────────┐         ┌──────────────────┐
│  VPC-A: 10.0.0.0/16 │◀─────▶│  VPC-B: 172.16.0.0/16│
│  ┌──────────┐    │  Peering │    ┌──────────┐    │
│  │ Web Tier │    │ ────────▶│    │ App Tier │    │
│  │ 10.0.1.x │    │          │    │ 172.16.1.x│    │
│  └──────────┘    │          │    └──────────┘    │
└──────────────────┘         └──────────────────┘
```

- **🎯 Use case:** Environment isolation (dev + prod), multi-project
- **📊 Complexity:** High

---

## 6. Regions and Availability Zones

### Regions Supporting VPC (Commercial)

| Region | RegionID | Available Zones | Status |
|--------|----------|-----------------|--------|
| 华北2 (北京) | `cn-beijing` | a, b, c, d | Commercial |
| 华东2 (上海) | `cn-shanghai` | a, b, c, e | Commercial |
| 华南1 (广州) | `cn-guangzhou` | a, b, c | Commercial |
| 中国香港 | `cn-hongkong` | a, b | Commercial |

### Additional Regions (Specialized)

| Region | RegionID | Type |
|--------|----------|------|
| 华北4 (大同) | `cn-datong` | Dedicated |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |
| 亚太东南 (雅加达) | `ap-southeast-3` | Commercial |

### How to List Available Regions

```bash
ve vpc DescribeRegions
```

---

## 7. Resource Lifecycle

### VPC Lifecycle

```
CreateVpc ──────▶ Pending ──────▶ Available
    │                                    │
    │                                    ▼
    │                              ModifyVpcAttribute
    │                                    │
    │                                    ▼
    │                              DeleteVpc
    │                                    │
    ▼                                    ▼
  Delete requires: No subnets, no route table associations,
  no attached resources in the VPC
```

### Subnet Lifecycle

```
CreateSubnet ──────▶ Pending ──────▶ Available
    │                                     │
    │                                     ▼
    │                               DeleteSubnet
    │                                     │
    ▼                                     ▼
  Delete requires: No instances, no ENIs,
  no other resources in the subnet
```

### Route Table Lifecycle

```
CreateRouteTable ──────▶ Available ──────▶ AssociateRouteTable (with subnet)
    │                                          │
    │                                          ▼
    │                                    DisassociateRouteTable
    │                                          │
    │                                          ▼
    │                                    DeleteRouteTable
    ▼
  Delete requires: No route entries (except system),
  no subnet associations
```

---

## 8. Limits and Quotas

| Resource | Default Quota (per region) | Notes |
|----------|---------------------------|-------|
| VPCs per region | 10 | Can request increase |
| Subnets per VPC | 200 | — |
| Route tables per VPC | 50 | — |
| Route entries per route table | 100 | — |
| Secondary CIDR blocks per VPC | 5 | — |
| IPv4 CIDR blocks options | `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` | And their subnets |

> ⚠️ Quotas may vary by region and account type. Use `DescribeVpcs` to check current usage, or contact support for increases.

---

*This reference document is part of the ve-vpc-ops skill. For operational procedures, see the main SKILL.md.*
