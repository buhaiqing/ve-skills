# Integration — Volcengine CLB

> **Purpose:** Integration patterns between CLB and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [CLB ↔ VPC Integration](#2-clb--vpc-integration)
3. [CLB ↔ EIP Integration](#3-clb--eip-integration)
4. [CLB ↔ ECS Integration](#4-clb--ecs-integration)

---

## 1. Service Delegation Map

| Service | Skill | CLB Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| VPC | `ve-vpc-ops` | VPC + Subnet for CLB deployment | User asks to create VPC + CLB |
| EIP | `ve-eip-ops` | EIP for public CLB | User asks for public-facing CLB |
| ECS | `ve-ecs-ops` | ECS as CLB backend | User asks to add backends |

---

## 2. CLB ↔ VPC Integration

CLB is deployed **inside** a VPC subnet.

### Integration Pattern

```bash
# ve-vpc-ops: Create VPC + Subnets (see ve-vpc-ops)
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"

# ve-clb-ops: Create CLB in the subnet
CLB_ID=$(ve clb CreateLoadBalancer --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --LoadBalancerName "prod-clb" --Type "private" | jq -r '.Result.LoadBalancerId')
```

### Internal CLB Access

Private CLB is accessible only from within the VPC. All ECS instances in the same VPC can access the CLB via its internal address.

---

## 3. CLB ↔ EIP Integration

Public CLB requires an EIP for internet access.

### Integration Pattern

```bash
# Step 1 (ve-clb-ops): Create public CLB
CLB_ID=$(ve clb CreateLoadBalancer --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --Type "public" --LoadBalancerName "public-clb" | jq -r '.Result.LoadBalancerId')

# Step 2 (ve-eip-ops): Allocate EIP if not auto-allocated
# Note: Public CLB may auto-allocate EIP depending on region/config

# Step 3 (ve-eip-ops): If manual EIP needed, bind it
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 50 | jq -r '.Result.AllocationId')
ve eip AssociateEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" --InstanceId "$CLB_ID" --InstanceType "ClbInstance"
```

---

## 4. CLB ↔ ECS Integration

ECS instances serve as CLB backend servers.

### Integration Pattern

```bash
# ve-clb-ops: Add ECS instances as backends
ve clb AddBackendServers --Region "$VOLCENGINE_REGION" --LoadBalancerId "$CLB_ID" \
  --BackendServers '[
    {"ServerId":"i-ecs-1","Port":8080,"Weight":100},
    {"ServerId":"i-ecs-2","Port":8080,"Weight":100},
    {"ServerId":"i-ecs-3","Port":8080,"Weight":50}
  ]'
```

### Security Group Configuration

CLB health checks originate from specific IP ranges. The ECS security group must allow inbound traffic on the backend port from the CLB subnet.

---

*This reference document is part of the ve-clb-ops skill.*
