# Integration — Volcengine NAT Gateway

> **Purpose:** Integration patterns between NAT Gateway and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [NAT Gateway ↔ VPC Integration](#2-nat-gateway--vpc-integration)
3. [NAT Gateway ↔ EIP Integration](#3-nat-gateway--eip-integration)
4. [NAT Gateway ↔ ECS Integration](#4-nat-gateway--ecs-integration)
5. [NAT Gateway ↔ RDS/Redis Integration](#5-nat-gateway--rdsredis-integration)
6. [Complete Architecture Pattern](#6-complete-architecture-pattern)

---

## 1. Service Delegation Map

| Service | Skill | NAT Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| VPC | `ve-vpc-ops` | VPC + Subnet for NAT deployment | User asks to create VPC + NAT |
| EIP | `ve-eip-ops` | EIP for NAT SNAT/DNAT | User asks to allocate EIP for NAT |
| ECS | `ve-ecs-ops` | NAT provides outbound for ECS in private subnet | Private ECS needs internet |
| RDS | `ve-rds-ops` | NAT provides outbound for RDS backup | Database needs external backup |

---

## 2. NAT Gateway ↔ VPC Integration

NAT Gateway is deployed **inside** a VPC subnet.

```
VPC (10.0.0.0/16)
├── Public Subnet (10.0.1.0/24) → NAT Gateway
└── Private Subnet (10.0.2.0/24) → ECS, RDS → routes through NAT → Internet
```

### Integration Pattern

```bash
# ve-vpc-ops: Create VPC + Subnets (see ve-vpc-ops)
# ve-nat-ops: Create NAT Gateway in public subnet

ve nat Gateway CreateNatGateway \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$PUBLIC_SUBNET_ID" \
  --NatGatewaySpec "Small" \
  --NatGatewayName "prod-nat"
```

### Route Table Configuration

For private subnets to use NAT Gateway, the route table must include a `0.0.0.0/0` route:

```bash
ve vpc CreateRouteEntry \
  --Region "$VOLCENGINE_REGION" \
  --RouteTableId "$PRIVATE_RT_ID" \
  --DestinationCidrBlock "0.0.0.0/0" \
  --NextHopType "NatGateway" \
  --NextHopId "$NAT_ID"
```

---

## 3. NAT Gateway ↔ EIP Integration

NAT Gateway **requires** at least one EIP bound for translation.

### Integration Pattern

```bash
# Step 1 (ve-eip-ops): Allocate EIP
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 20 | jq -r '.Result.AllocationId')
EIP_ADDR=$(ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" --AllocationIds "[\"$EIP_ID\"]" | jq -r '.Result.EipAddresses[0].EipAddress')

# Step 2 (ve-nat-ops): Create NAT Gateway
NAT_ID=$(ve nat Gateway CreateNatGateway --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --NatGatewaySpec "Small" | jq -r '.Result.NatGatewayId')

# Step 3 (ve-eip-ops): Bind EIP to NAT Gateway
ve eip AssociateEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" --InstanceId "$NAT_ID" --InstanceType "Nat"

# Step 4 (ve-nat-ops): Create SNAT rule
ve nat Gateway CreateSnatRule --Region "$VOLCENGINE_REGION" --NatGatewayId "$NAT_ID" --SourceCidr "10.0.2.0/24" --EipAddresses "[\"$EIP_ADDR\"]"
```

---

## 4. NAT Gateway ↔ ECS Integration

ECS instances in private subnets use NAT Gateway for outbound internet access.

### Architecture

```
Private ECS (10.0.2.10)
       │
       │ (VPC internal)
       ▼
NAT Gateway (10.0.1.10)
       │
       │ (SNAT translation)
       ▼
EIP (120.78.x.x)
       │
       ▼
   Internet
```

### ECS Setup in Private Subnet

```bash
# ve-ecs-ops: Create ECS in private subnet (no public IP)
ve ecs RunInstances \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$PRIVATE_SUBNET_ID" \
  --...

# Verify outbound connectivity
# SSH via bastion or VPC internal, then:
curl -I https://api.example.com  # Should succeed via NAT SNAT
```

---

## 5. NAT Gateway ↔ RDS/Redis Integration

Databases in private subnets may need internet access for:

- Backup to external storage
- Software updates
- License validation
- Cloud service integration

### Integration Pattern

Same as ECS — ensure RDS/Redis subnet has route table pointing `0.0.0.0/0` → NAT Gateway.

---

## 6. Complete Architecture Pattern

### Production Multi-AZ Setup

```
                          Internet
                              │
                              ▼
                          EIP (50Mbps)
                              │
                    ┌─────────┴─────────┐
                    │   Public CLB      │
                    │  (10.0.1.0/24)    │
                    └─────────┬─────────┘
                              │
┌─────────────────────────────┼─────────────────────────────┐
│                           VPC                              │
│                                                             │
│  ┌─────────────┐    ┌─────────────────┐   ┌─────────────┐ │
│  │ NAT Gateway │    │ App Subnet      │   │ DB Subnet   │ │
│  │ 10.0.1.10   │◀───│ 10.0.2.0/24     │   │ 10.0.3.0/24 │ │
│  │ + EIP       │    │ [App Server]    │   │ [RDS]       │ │
│  └──────┬──────┘    │ [App Server]    │   │ [Redis]     │ │
│         │           └─────────────────┘   └─────────────┘ │
│         │                                                  │
│         ▼                                                  │
│      Internet Outbound                                     │
└───────────────────────────────────────────────────────────┘
```

---

*This reference document is part of the ve-nat-ops skill.*
