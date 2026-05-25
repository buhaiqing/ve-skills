# Integration — Volcengine VPC

> **Purpose:** Integration patterns and delegation workflows between VPC operations and other Volcengine cloud services. Documents how VPC serves as the networking foundation for compute, database, and load balancing resources.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Integration Architecture](#1-integration-architecture)
2. [Service Delegation Map](#2-service-delegation-map)
3. [VPC ↔ ECS Integration](#3-vpc--ecs-integration)
4. [VPC ↔ CLB Integration](#4-vpc--clb-integration)
5. [VPC ↔ EIP Integration](#5-vpc--eip-integration)
6. [VPC ↔ NAT Gateway Integration](#6-vpc--nat-gateway-integration)
7. [VPC ↔ RDS Integration](#7-vpc--rds-integration)
8. [VPC Peering](#8-vpc-peering)
9. [Cross-Account Patterns](#9-cross-account-patterns)

---

## 1. Integration Architecture

VPC is the foundational networking layer that all compute and application resources depend on:

```
                    ┌──────────────┐
                    │   VPC        │
                    │  (Networking) │
                    └──────┬───────┘
                           │
        ┌──────────┬───────┼───────┬──────────┐
        │          │       │       │          │
        ▼          ▼       ▼       ▼          ▼
     ┌──────┐  ┌─────┐  ┌─────┐ ┌─────┐  ┌──────┐
     │ ECS  │  │ CLB │  │ EIP │ │ NAT │  │  RDS │
     │(ve-ecs)│ │(ve-clb)│ │(ve-eip)│ │(ve-nat)│ │(ve-rds) │
     └──────┘  └─────┘  └─────┘ └─────┘  └──────┘
```

All services below the VPC layer delegate networking setup to this skill.

---

## 2. Service Delegation Map

| Service | Skill | VPC Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| ECS | `ve-ecs-ops` | Requires VPC + Subnet | User asks to create ECS instance |
| CLB | `ve-clb-ops` | Requires VPC + Subnet | User asks to create load balancer |
| EIP | `ve-eip-ops` | Binds to VPC-resident resources | User asks to allocate EIP |
| NAT Gateway | `ve-nat-ops` | Requires VPC + Subnet | User asks to create NAT gateway |
| RDS | `ve-rds-ops` | Requires VPC + Subnet | User asks to create database instance |
| Redis | `ve-redis-ops` | Requires VPC + Subnet | User asks to create Redis instance |

### Delegation Flow

When `ve-vpc-ops` receives a request that involves another service:

1. **Verify** VPC/subnet exists using this skill
2. **Capture** VPC ID and Subnet ID as `{{output.*}}` placeholders
3. **Delegate** to the appropriate service skill with the captured IDs

**Example Flow: Create ECS in new VPC**

```
ve-vpc-ops:  CreateVpc → {{output.vpc_id}}
ve-vpc-ops:  CreateSubnet → {{output.subnet_id}}
     │
     └──▶ delegate to ve-ecs-ops with VpcId={{output.vpc_id}}, SubnetId={{output.subnet_id}}
```

---

## 3. VPC ↔ ECS Integration

### ECS Instance Creation Requirements

| Requirement | Source | API |
|-------------|--------|-----|
| VPC exists | `ve-vpc-ops` | `DescribeVpcs` |
| Subnet exists | `ve-vpc-ops` | `DescribeSubnets` |
| Security group exists | VPC auto-created or `ve-ecs-ops` | `DescribeSecurityGroups` |

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): Ensure VPC + Subnet exist
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"

# Step 2 (ve-ecs-ops): Create ECS instance in the subnet
ve ecs RunInstances \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  --InstanceType "ecs.g2i.large" \
  ...
```

### Shared Network ACL

Security groups are VPC-scoped. Instances in the same VPC share the same security group pool:

```bash
# List security groups in the VPC
ve ecs DescribeSecurityGroups --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID"
```

---

## 4. VPC ↔ CLB Integration

### CLB Network Configuration

| CLB Type | VPC Requirement | Network |
|----------|----------------|---------|
| Public CLB | Requires VPC + Subnet for internal ENI | CLB gets VPC-internal IP + EIP |
| Internal CLB | Requires VPC + Subnet | CLB accessible only within VPC |

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): Ensure VPC + Subnet exist
# Step 2 (ve-clb-ops): Create CLB
ve clb CreateLoadBalancer \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  ...

# Step 3 (ve-clb-ops): Add backend ECS instances from same VPC
ve clb AddBackendServers \
  --LoadBalancerId "$CLB_ID" \
  --BackendServers '[{"ServerId":"i-xxx", "Port":80}]'
```

### Health Check Considerations

CLB health checks use VPC internal networking. Ensure:
- Backend ECS instances are in the same VPC or a peered VPC
- Security group rules allow CLB health check IP ranges
- No network ACL blocks health check traffic

---

## 5. VPC ↔ EIP Integration

### EIP Binding Types

| Binding Target | VPC Requirement | Delegation |
|---------------|----------------|------------|
| ECS instance | ECS must be in a VPC | `ve-eip-ops` → `ve-ecs-ops` |
| NAT Gateway | NAT must be in a VPC | `ve-eip-ops` → `ve-nat-ops` |
| CLB (public) | CLB must be in a VPC | `ve-eip-ops` → `ve-clb-ops` |

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): VPC and Subnet ready
# Step 2 (ve-ecs-ops): ECS instance created

# Step 3 (ve-eip-ops): Allocate and bind EIP
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" | jq -r '.Result.AllocationId')
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "{{ecs.instance_id}}" \
  --InstanceType EcsInstance
```

---

## 6. VPC ↔ NAT Gateway Integration

### NAT Gateway Network Setup

```
┌─────────────────────────────────────────┐
│                    VPC                   │
│  ┌─────────────┐    ┌────────────────┐  │
│  │ Public Subnet│   │ Private Subnet │  │
│  │ 10.0.1.0/24 │   │ 10.0.2.0/24    │  │
│  │             │   │                │  │
│  │ [NAT GW]    │──▶│ [App Servers]  │  │
│  │ [EIP bound] │   │ outbound→Internet│ │
│  └─────────────┘    └────────────────┘  │
└─────────────────────────────────────────┘
```

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): Create VPC + subnets
# Step 2 (ve-eip-ops): Allocate EIP for NAT Gateway

# Step 3 (ve-nat-ops): Create NAT Gateway in public subnet
ve nat CreateNatGateway \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$PUBLIC_SUBNET_ID" \
  --NatGatewayName "nat-prod"

# Step 4 (ve-nat-ops): Create SNAT rule for private subnet
ve nat CreateSnatRule \
  --Region "$VOLCENGINE_REGION" \
  --NatGatewayId "$NAT_ID" \
  --SnatRuleName "snat-private" \
  --SourceCidr "$PRIVATE_SUBNET_CIDR" \
  --EipAddress "$EIP_ADDRESS"
```

---

## 7. VPC ↔ RDS Integration

### RDS Network Requirements

RDS MySQL instances must be deployed within a VPC subnet.

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): Ensure VPC + Subnet exist in target region
# Step 2 (ve-rds-ops): Create RDS instance
ve rds_mysql CreateDBInstance \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  ...
```

### Private Connectivity

RDS instances in the same VPC as application servers can communicate via internal IPs without EIP.

---

## 8. VPC Peering

### Overview

VPC Peering enables private network communication between two VPCs (same account or cross-account, same region or cross-region).

### Peering Setup

```bash
# Step 1: Create VPC peering connection
ve vpc CreateVpcPeer \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_A_ID" \
  --PeerVpcAccountId "{{user.peer_account_id}}" \
  --PeerVpcId "{{user.peer_vpc_id}}" \
  --PeerVpcRegion "{{user.peer_region}}"

# Step 2: Accept peering (in peer account)
ve vpc AcceptVpcPeer \
  --Region "{{user.peer_region}}" \
  --VpcPeerName "{{output.vpc_peer_name}}"

# Step 3: Add route entries in both VPCs
ve vpc CreateRouteEntry \
  --Region "$VOLCENGINE_REGION" \
  --RouteTableId "$VPC_A_ROUTE_TABLE" \
  --DestinationCidrBlock "$VPC_B_CIDR" \
  --NextHopType VpcPeer \
  --NextHopId "{{output.vpc_peer_id}}"
```

### Route Table Update Requirement

After creating a peering connection, **both VPCs** must add route entries pointing to the peer's CIDR block via `NextHopType=VpcPeer`.

---

## 9. Cross-Account Patterns

### Resource Sharing in a VPC

Multiple Volcengine accounts can share resources within a VPC through:

1. **VPC Peering** — Cross-account, cross-region VPC connectivity
2. **CEN (Cloud Enterprise Network)** — Hub-and-spoke VPC interconnection
3. **Shared VPC** — Share subnets across accounts within the same organization

### Best Practices

| Pattern | Use Case | Complexity |
|---------|----------|------------|
| Shared VPC | Same org, shared infrastructure | Low |
| VPC Peering | Two VPCs, point-to-point | Medium |
| CEN | Multiple VPCs (hub-and-spoke) | High |

---

*This reference document is part of the ve-vpc-ops skill. For API details, see [api-sdk-usage.md](api-sdk-usage.md).*
