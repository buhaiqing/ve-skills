# Integration — Volcengine ALB (应用型负载均衡)

> **Purpose:** Integration patterns between ALB and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-31

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [ALB ↔ VPC Integration](#2-alb--vpc-integration)
3. [ALB ↔ EIP Integration](#3-alb--eip-integration)
4. [ALB ↔ ECS Integration](#4-alb--ecs-integration)
5. [ALB ↔ Certificate Manager Integration](#5-alb--certificate-manager-integration)

---

## 1. Service Delegation Map

| Service | Skill | ALB Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| VPC | `ve-vpc-ops` | VPC + Subnet for ALB deployment | User asks to create VPC + ALB |
| EIP | `ve-eip-ops` | EIP for public ALB | User asks for public-facing ALB |
| ECS | `ve-ecs-ops` | ECS as ALB backend servers | User asks to add backend servers |
| Certificate | `ve-certificate-ops` | TLS certs for HTTPS listeners | User configures HTTPS listener |
| CMS | `ve-cms-ops` | Monitoring and alarm setup | User asks for ALB monitoring |
| CLB | `ve-clb-ops` | Classic Load Balancer | User mentions CLB instead of ALB |

---

## 2. ALB ↔ VPC Integration

ALB is deployed **inside** a VPC subnet. The VPC and subnet must exist before creating an ALB.

### Integration Pattern

```bash
# Step 1 (ve-vpc-ops): Create VPC + Subnets
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"

# Step 2 (ve-alb-ops): Create ALB in the subnet
ALB_ID=$(ve alb CreateLoadBalancer \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  --LoadBalancerName "prod-alb" \
  --Type "private" | jq -r '.Result.LoadBalancerId')
```

### Private ALB Access

- Private ALB is accessible only from within the same VPC
- All ECS instances in the same VPC can access the ALB via its internal IP (`Address`)
- Cross-VPC access requires VPC Peering or Cloud Connect

---

## 3. ALB ↔ EIP Integration

Public ALB requires an EIP for internet-facing access.

### Integration Pattern (Auto-Allocate)

```bash
# Create public ALB with EIP billing config (auto-allocates EIP)
ALB_ID=$(ve alb CreateLoadBalancer \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  --Type "public" \
  --EipBillingConfig '{"Bandwidth":10,"EipBillingType":"PayByTraffic","ISP":"BGP"}' \
  --LoadBalancerName "public-alb" | jq -r '.Result.LoadBalancerId')
```

### Integration Pattern (Manual EIP via ve-eip-ops)

```bash
# Step 1 (ve-eip-ops): Allocate EIP
EIP_ID=$(ve eip AllocateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --Bandwidth 50 | jq -r '.Result.AllocationId')

# Step 2 (ve-eip-ops): Associate EIP with ALB
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$ALB_ID" \
  --InstanceType "AlbInstance"
```

---

## 4. ALB ↔ ECS Integration

ECS instances serve as ALB backend servers through Server Groups.

### Integration Pattern

```bash
# Step 1 (ve-ecs-ops): Verify ECS instances exist
ve ecs DescribeInstances --Region "$VOLCENGINE_REGION" --InstanceIds '["i-xxx","i-yyy"]'

# Step 2 (ve-alb-ops): Add ECS instances as backend servers
ve alb AddServersToGroup \
  --Region "$VOLCENGINE_REGION" \
  --ServerGroupId "$SERVER_GROUP_ID" \
  --Servers '[
    {"ServerId":"i-xxx","Port":8080,"Weight":100,"ServerType":"ecs"},
    {"ServerId":"i-yyy","Port":8080,"Weight":100,"ServerType":"ecs"}
  ]'
```

### Security Group Configuration

Backend ECS security groups **MUST** allow inbound traffic on the backend port from the ALB subnet CIDR.

```bash
# Example: Add security group rule for ALB traffic
ve ecs AuthorizeSecurityGroupIngress \
  --Region "$VOLCENGINE_REGION" \
  --SecurityGroupId "sg-xxx" \
  --CidrIp "10.0.0.0/16" \
  --PortRange "8080/8080" \
  --Protocol "tcp"
```

---

## 5. ALB ↔ Certificate Manager Integration

HTTPS listeners require TLS certificates stored in Volcengine Certificate Manager.

### Integration Pattern

```bash
# Step 1: Create/upload certificate in Certificate Manager
# (via console or Certificate Manager API)

# Step 2 (ve-alb-ops): Create HTTPS listener with certificate
ve alb CreateListener \
  --Region "$VOLCENGINE_REGION" \
  --LoadBalancerId "$ALB_ID" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "https-listener" \
  --CertificateId "cert-xxx" \
  --TLSPolicy "tls-1-2"
```

---

*This reference document is part of the ve-alb-ops skill.*
