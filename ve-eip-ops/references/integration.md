# Integration — Volcengine EIP

> **Purpose:** Integration patterns between EIP and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [EIP ↔ ECS Integration](#2-eip--ecs-integration)
3. [EIP ↔ CLB Integration](#3-eip--clb-integration)
4. [EIP ↔ NAT Gateway Integration](#4-eip--nat-gateway-integration)
5. [EIP ↔ VPC Integration](#5-eip--vpc-integration)
6. [EIP ↔ ENI Integration](#6-eip--eni-integration)

---

## 1. Service Delegation Map

| Service | Skill | EIP Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| ECS | `ve-ecs-ops` | Optional EIP for public access | User asks ECS to get public IP |
| CLB | `ve-clb-ops` | Optional EIP for public CLB | User asks for public load balancer |
| NAT Gateway | `ve-nat-ops` | Requires EIP for SNAT/DNAT | User asks to create NAT with EIP |
| VPC | `ve-vpc-ops` | VPC context for EIP allocation | User asks to create VPC + EIP |
| ENI | `ve-ecs-ops` | Optional EIP for ENI | User binds EIP to network interface |

---

## 2. EIP ↔ ECS Integration

### Integration Pattern

```bash
# Step 1 (ve-ecs-ops): Create ECS instance (in VPC private subnet)
ECS_ID="i-xxx"

# Step 2 (ve-eip-ops): Allocate and bind EIP
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 5 | jq -r '.Result.AllocationId')
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$ECS_ID" \
  --InstanceType "EcsInstance"

echo "ECS $ECS_ID now accessible at: $(ve eip DescribeEipAddressAttributes --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" | jq -r '.Result.EipAddress')"
```

### ECS IP Considerations

| Scenario | Network Config |
|----------|---------------|
| ECS with EIP | Public IP = EIP address |
| ECS without EIP | Private subnet only, use NAT Gateway for outbound |
| ECS with ELB | Internal IP via CLB, CLB has EIP for public access |

---

## 3. EIP ↔ CLB Integration

### Public CLB Architecture

```
Internet → EIP → CLB (VPC internal IP) → Backend ECS (private subnet)
```

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): VPC + Subnet
# Step 2 (ve-clb-ops): Create internal CLB
CLB_ID="clb-xxx"

# Step 3 (ve-eip-ops): Allocate EIP and bind to CLB
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 20 | jq -r '.Result.AllocationId')
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$CLB_ID" \
  --InstanceType "ClbInstance"
```

---

## 4. EIP ↔ NAT Gateway Integration

### NAT Gateway EIP Requirement

NAT Gateway **requires** at least one EIP bound for SNAT/DNAT functionality.

### Integration Pattern

```bash
# Step 1 (ve-eip-ops): Allocate EIP for NAT
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 20 | jq -r '.Result.AllocationId')
EIP_ADDR=$(ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" --AllocationIds "[\"$EIP_ID\"]" | jq -r '.Result.EipAddresses[0].EipAddress')

# Step 2 (ve-nat-ops): Create NAT Gateway
NAT_ID=$(ve nat Gateway CreateNatGateway --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" | jq -r '.Result.NatGatewayId')

# Step 3 (ve-eip-ops): Bind EIP to NAT Gateway
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$NAT_ID" \
  --InstanceType "Nat"

# Step 4 (ve-nat-ops): Create SNAT rule with EIP
ve nat Gateway CreateSnatRule \
  --Region "$VOLCENGINE_REGION" \
  --NatGatewayId "$NAT_ID" \
  --SourceCidr "10.0.2.0/24" \
  --EipAddresses "[\"$EIP_ADDR\"]"
```

---

## 5. EIP ↔ VPC Integration

EIP allocation does not require VPC directly but is used in conjunction with VPC resources. All EIP-bound resources must be in the same region as the EIP.

### Region Alignment

```
EIP (cn-beijing) ──────┬─────▶ ECS (cn-beijing)  ✅ Same region
                        ├─────▶ CLB (cn-beijing)  ✅ Same region
                        └─────▶ ECS (cn-shanghai) ❌ Different region - NOT ALLOWED
```

---

## 6. EIP ↔ ENI Integration

Binding EIP to an Elastic Network Interface allows the EIP to follow the ENI even when the underlying instance changes.

```bash
# Bind EIP to ENI
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$ENI_ID" \
  --InstanceType "NetworkInterface"
```

---

*This reference document is part of the ve-eip-ops skill.*
