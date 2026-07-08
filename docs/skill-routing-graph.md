# Skill Routing Graph — Cross-Skill Diagnosis Orchestration

> **Purpose**: Machine-readable routing table for Agent cross-skill delegation during multi-product
> fault diagnosis. Links alarm patterns → primary skill → secondary skill(s).
>
> **Usage**: GCL Orchestrator loads this during pre-flight to auto-route when primary skill
> detects cross-product symptoms.
>
> **TE-7 / Token Efficiency**: Content lives here (not duplicated in each aiops.md).
> Each skill's aiops.md links here via `[...](docs/skill-routing-graph.md)` for delegation rules.

---

## Alarm Pattern → Skill Routing Table

> 被 [AGENTS.md](../AGENTS.md) §Cross-Skill Delegation 和 §Runtime Quality Gates 引用。

| Alarm Pattern | Primary Skill | Secondary Skills | Trigger Condition |
|--------------|--------------|-----------------|-------------------|
| ECS CPU>90% + app slow query | `ve-ecs-ops` | `ve-rds-mysql-ops` | CPU spike with concurrent DB latency |
| ECS CPU>90% + high connections | `ve-ecs-ops` | `ve-ecs-ops` (cloud assistant) | Single process runaway vs system-wide |
| ECS Disk>90% | `ve-ecs-ops` | `ve-ecs-ops` (log cleanup) | Log-related vs data-related |
| ECS Disk IOPS>90% | `ve-ecs-ops` | `ve-rds-mysql-ops` / `ve-redis-ops` | DB query / backup causing IO |
| ECS Network latency | `ve-ecs-ops` | `ve-vpc-ops` | Packet loss / upstream dependency |
| ECS Security group change detected | `ve-ecs-ops` | `ve-security-group-ops` | Recent SG rule change → rollback |
| ECS + DB connection error | `ve-ecs-ops` | `ve-rds-mysql-ops` | App server connection pool leak |
| RDS CPU>90% | `ve-rds-mysql-ops` | `ve-ecs-ops` | App connection storm from ECS |
| RDS Lock waits | `ve-rds-mysql-ops` | `ve-rds-mysql-ops` (SHOW ENGINE) | Lock wait timeout events |
| RDS Disk>85% + binlog growth | `ve-rds-mysql-ops` | `ve-cms-ops` | Cleanup policy via alarm rule |
| Redis Memory>90% | `ve-redis-ops` | `ve-redis-ops` (eviction policy) | Memory fragmentation vs dataset size |
| Redis Slow commands | `ve-redis-ops` | `ve-redis-ops` (MONITOR) | O(N) command scan |
| VPC Route table change | `ve-vpc-ops` | `ve-security-group-ops` | Recent VPC change → rollback |
| VPC ENI attachment failure | `ve-vpc-ops` | `ve-ecs-ops` | ENI not attaching → check instance state |
| VPC NAT gateway latency | `ve-vpc-ops` | `ve-nat-ops` | NAT SNAT rule issue |
| EIP quota exceeded | `ve-eip-ops` | `ve-vpc-ops` | ENI quota or VPC CIDR exhaustion |
| ELB 5xx spike | `ve-clb-ops` | `ve-ecs-ops` / `ve-alb-ops` | Backend health check failure |
| ELB connection limit | `ve-clb-ops` | `ve-ecs-ops` | Backend connection pool exhaustion |
| Alarm storm (>10 alarms/5min) | `ve-cms-ops` | all | Correlation analysis + root cause |
| Unknown pattern | `ve-cms-ops` | any skill | No match in table |

---

## Dependency Graph

```
[Alarm triggered]
    │
    ├── ve-cms-ops (alarm correlation — always check first if pattern unknown)
    │
    ├── ve-ecs-ops
    │     ├── ve-vpc-ops (network latency / route table)
    │     ├── ve-security-group-ops (SG change → rollback)
    │     ├── ve-rds-mysql-ops (DB connection error)
    │     ├── ve-redis-ops (cache miss → app fallback)
    │     └── ve-clb-ops / ve-alb-ops (load balancer backend)
    │
    ├── ve-rds-mysql-ops
    │     ├── ve-ecs-ops (connection storm from app)
    │     ├── ve-cms-ops (cleanup policy)
    │     └── ve-vpc-ops (VPC peering / security group)
    │
    ├── ve-redis-ops
    │     └── ve-ecs-ops (app-level eviction tuning)
    │
    ├── ve-vpc-ops
    │     ├── ve-security-group-ops (SG rule conflicts)
    │     ├── ve-eip-ops (IP quota)
    │     ├── ve-nat-ops (NAT gateway SNAT)
    │     └── ve-ecs-ops (ENI attachment)
    │
    ├── ve-clb-ops / ve-alb-ops
    │     └── ve-ecs-ops (backend health check failure)
    │
    └── ve-kms-ops / ve-iam-ops (permission errors — always delegate on AuthFailure)
```

---

## Critical Routing Rules

### Rule 1: AuthFailure always delegates to IAM
```
Any skill + AuthFailure / UnauthorizedOperation
  → ve-iam-ops (check policy) → ve-kms-ops (if key-related)
```

### Rule 2: Alarm storm triggers CMS correlation first
```
>10 alarms in 5min from same resource group
  → ve-cms-ops (suppress + correlate) → primary skill (root cause)
```

### Rule 3: Network issue → VPC first
```
Any skill + network latency / packet loss / connection timeout
  → ve-vpc-ops (route table, ENI, ACL) → downstream skill
```

### Rule 4: Cost spike → FinOps
```
Any skill + unexpected cost increase
  → ve-billing-ops (cost breakdown) → relevant skill (optimize)
```

### Rule 5: Unknown pattern → CMS
```
Symptom doesn't match any rule above
  → ve-cms-ops (alarm correlation + log analysis)
```

---

## Implementation Notes

- **TE-6**: Cross-skill delegation rules live here; individual `aiops.md` files link here.
  Do NOT duplicate routing rules in per-skill `aiops.md`.
- **GCL Orchestrator**: Load this graph during pre-flight; match alarm/error symptoms
  against the routing table before executing primary skill.
- **Token budget**: This graph is ~100 lines. If individual skill's aiops.md exceeds
  compression budget, trim prose and link here instead.
