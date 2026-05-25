# Integration — Volcengine Redis

> **Purpose:** Integration patterns between Redis and other Volcengine cloud services.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Service Delegation Map](#1-service-delegation-map)
2. [Redis ↔ VPC Integration](#2-redis--vpc-integration)
3. [Redis ↔ ECS Integration](#3-redis--ecs-integration)
4. [Redis ↔ CLB Integration](#4-redis--clb-integration)

---

## 1. Service Delegation Map

| Service | Skill | Redis Dependency | Delegation Trigger |
|---------|-------|---------------|-------------------|
| VPC | `ve-vpc-ops` | VPC + Subnet for Redis deployment | User asks to create VPC + Redis |
| ECS | `ve-ecs-ops` | ECS connects to Redis as application server | App needs cache/session |

---

## 2. Redis ↔ VPC Integration

Redis is deployed inside a VPC subnet.

### Integration Pattern

```bash
# ve-vpc-ops: Create VPC + Subnets (see ve-vpc-ops)
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"

# ve-redis-ops: Create Redis instance in the subnet
ve redis CreateDBInstance --body '{
  "InstanceName": "prod-redis",
  "EngineVersion": "6.0",
  "InstanceClass": "PrimarySecondary",
  "NodeNumber": 2,
  "ShardCapacity": 4096,
  "Password": "SecureP@ss123",
  "VpcId": "'"$VPC_ID"'",
  "SubnetId": "'"$SUBNET_ID"'",
  "ChargeType": "PostPaid"
}'
```

---

## 3. Redis ↔ ECS Integration

ECS instances (application servers) connect to Redis as cache/session store.

### Security and Connectivity

1. ECS and Redis must be in the same VPC (or peered VPCs)
2. Add ECS subnet CIDR to Redis allow list:
   ```bash
   ve redis CreateAllowList --body '{
     "AllowListName": "ecs-apps",
     "AllowList": ["10.0.2.0/24"]
   }'
   ve redis AssociateAllowList --body '{
     "InstanceId": "redis-xxx",
     "AllowListIds": ["acl-xxx"]
   }'
   ```

3. Security group allows outbound traffic to Redis subnet on port 6379

### Application Connection String

```python
import redis
r = redis.Redis(host="redis-xxx.redis.ivolces.com", port=6379, password="SecureP@ss123", decode_responses=True)
```

---

## 4. Redis ↔ CLB Integration

Redis itself doesn't use CLB directly. However, applications connecting to Redis may be behind a CLB for load balancing.

```
Internet → CLB → [ECS App 1] ──┐
                 [ECS App 2] ──┴─▶ Redis (PrimarySecondary)
                 [ECS App 3] ──┘
```

All ECS instances share the same Redis connection pool. Use `PrimarySecondary` type with automatic failover for resilience.

---

*This reference document is part of the ve-redis-ops skill.*
