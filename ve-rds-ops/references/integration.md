# Integration — Volcengine RDS MySQL

> **Purpose:** Integration patterns between RDS MySQL and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [RDS ↔ VPC Integration](#2-rds--vpc-integration)
3. [RDS ↔ ECS Integration](#3-rds--ecs-integration)
4. [RDS ↔ CLB Integration](#4-rds--clb-integration)

---

## 1. Service Delegation Map

| Service | Skill | RDS Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| VPC | `ve-vpc-ops` | VPC + Subnet for RDS deployment | User asks to create VPC + RDS |
| ECS | `ve-ecs-ops` | ECS connects to RDS as application server | App needs DB access |
| CLB | `ve-clb-ops` | CLB distributes to app servers connecting to RDS | Database-powered web app |

---

## 2. RDS ↔ VPC Integration

RDS is deployed **inside** a VPC subnet.

### Integration Pattern

```bash
# ve-vpc-ops: Create VPC + Subnets (see ve-vpc-ops)
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"

# ve-rds-ops: Create RDS instance in the subnet
ve rds_mysql CreateDBInstance --body '{
  "RegionId": "cn-beijing",
  "InstanceName": "prod-mysql",
  "DBEngineVersion": "MySQL_8_0",
  "InstanceType": "HA",
  "NodeSpec": "rds.mysql.2c4g",
  "StorageSpace": 100,
  "StorageType": "ESSD",
  "VpcId": "'"$VPC_ID"'",
  "SubnetId": "'"$SUBNET_ID"'",
  "ChargeType": "PostPaid",
  "NodeInfo": [{"ZoneId": "cn-beijing-a"}, {"ZoneId": "cn-beijing-b"}]
}'
```

### Network Connectivity

RDS instances are accessible only from within the same VPC or peered VPCs.

---

## 3. RDS ↔ ECS Integration

ECS instances typically connect to RDS as application/database servers.

### Security Group Configuration

The ECS application security group must allow outbound traffic to the RDS subnet on port 3306. The RDS allow list must include the ECS private IP or subnet CIDR.

```bash
# Add ECS subnet to RDS allow list
ve rds_mysql CreateAllowList --Region "$VOLCENGINE_REGION" --InstanceId "$RDS_INSTANCE_ID" \
  --AllowListName "ecs-app" --AllowListType "IPv4" --AllowList '["10.0.2.0/24"]'
```

### Application Connection String

```
mysql://app_user:password@${RDS_PRIVATE_IP}:3306/appdb
```

---

## 4. Read-Only Proxy Pattern

For read-heavy workloads, use `MultiNode` instance type with read replicas, or configure application-level read/write splitting.

```
                    ┌──────────────┐
                    │    CLB       │ (optional, for read splitting)
                    │ :3306        │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         ┌─────────┐ ┌─────────┐ ┌─────────┐
         │ Primary │ │ Replica │ │ Replica │
         │  (RW)   │ │  (RO)   │ │  (RO)   │
         └─────────┘ └─────────┘ └─────────┘
```

---

*This reference document is part of the ve-rds-ops skill.*
