# Core Concepts — Volcengine RDS MySQL (云数据库 MySQL)

> **Purpose:** Fundamental concepts for Volcengine RDS MySQL.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is RDS MySQL](#1-what-is-rds-mysql)
2. [Instance Types](#2-instance-types)
3. [Storage Types](#3-storage-types)
4. [Instance Specifications](#4-instance-specifications)
5. [MySQL Engine Versions](#5-mysql-engine-versions)
6. [Database Accounts and Privileges](#6-database-accounts-and-privileges)
7. [Backup and Recovery](#7-backup-and-recovery)
8. [Instance Lifecycle](#8-instance-lifecycle)
9. [Regions Supporting RDS](#9-regions-supporting-rds)
10. [Limits and Quotas](#10-limits-and-quotas)

---

## 1. What is RDS MySQL

A **RDS MySQL (云数据库 MySQL 版)** is a fully managed relational database service on Volcengine that is compatible with MySQL. It handles provisioning, patching, backup, recovery, and monitoring.

### Key Characteristics

- **Managed Service:** Automated backups, patching, monitoring
- **High Availability:** Primary-replica architecture with automatic failover
- **Performance:** ESSD storage for high IOPS, optimized MySQL configurations
- **Security:** VPC isolation, IP allow lists, SSL, encryption at rest

---

## 2. Instance Types

| Instance Type | Code | Architecture | HA | Use Case |
|---------------|------|-------------|-----|----------|
| **Single Node** | `Single` | 1 node | No | Dev/test, non-critical workloads |
| **High Availability** | `HA` | Primary + Secondary | Auto failover | Production, standard workloads |
| **Multi-Node** | `MultiNode` | Primary + Secondary + Read Replicas | Auto failover + read scaling | Production, read-heavy workloads |

---

## 3. Storage Types

| Storage Type | Code | IOPS | Max Capacity | Use Case |
|--------------|------|------|--------------|----------|
| **Local SSD** | `LocalSSD` | High (up to 100K) | Up to 3 TB | Low latency, I/O intensive |
| **ESSD** | `ESSD` | Very high (up to 1M) | Up to 65 TB | High throughput, large datasets, flexible sizing |

### Storage Scaling

- ESSD can be expanded online without downtime
- Minimum storage: 20 GB
- Auto-scaling available for ESSD

---

## 4. Instance Specifications

| Node Spec | vCPU | Memory | Max Connections | Suitable For |
|-----------|------|--------|-----------------|--------------|
| `rds.mysql.1c1g` | 1 | 1 GB | 100 | Dev/test, very light |
| `rds.mysql.1c2g` | 1 | 2 GB | 200 | Dev/test, light |
| `rds.mysql.2c4g` | 2 | 4 GB | 500 | Small production |
| `rds.mysql.4c8g` | 4 | 8 GB | 1,000 | Medium production |
| `rds.mysql.4c16g` | 4 | 16 GB | 2,000 | Medium-heavy |
| `rds.mysql.8c16g` | 8 | 16 GB | 4,000 | Heavy |
| `rds.mysql.8c32g` | 8 | 32 GB | 8,000 | Very heavy |
| `rds.mysql.16c64g` | 16 | 64 GB | 16,000 | Enterprise |

> Note: Actual specs vary by region. Use `DescribeAvailableClasses` to query available specs.

---

## 5. MySQL Engine Versions

| Version | Status | End of Support | Notable Features |
|---------|--------|---------------|-----------------|
| **MySQL 5.7** | Stable | 2025-10+ | JSON, generated columns, multi-source replication |
| **MySQL 8.0** | Current | 2026+ | CTE, window functions, JSON enhancements, authentication plugin |

### Version Selection Guidance

- Use **MySQL 8.0** for new projects
- Use **MySQL 5.7** for compatibility with existing applications
- In-place version upgrade from 5.7 → 8.0 is supported

---

## 6. Database Accounts and Privileges

### Account Types

| Account Type | Code | Privileges | Use Case |
|--------------|------|-----------|----------|
| **Super** | `Super` | Full access to all databases | Administrator |
| **Normal** | `Normal` | Selected database privileges | Application accounts |

### Privilege Levels

| Privilege | Code | Description |
|-----------|------|-------------|
| Read-Only | `ReadOnly` | SELECT only |
| Read-Write | `ReadWrite` | SELECT, INSERT, UPDATE, DELETE |
| Read-Write DDL | `ReadWriteDDL` | ReadWrite + CREATE, ALTER, DROP |

---

## 7. Backup and Recovery

### Backup Types

| Backup Type | Trigger | Retention | Use Case |
|-------------|---------|-----------|----------|
| **Automated Backup** | Scheduled | Configurable (7–730 days) | Continuous protection |
| **Manual Backup** | User-initiated | Until manually deleted | Before risky operations |
| **PITR (Point-in-Time Recovery)** | Automated | Within backup retention | Precise recovery to any second |

### Data Keep Policy on Deletion

| Policy | Code | Description |
|--------|------|-------------|
| Last Backup | `Last` | Keep the last backup |
| All Backups | `All` | Keep all backups |
| No Backup | `None` | Delete everything |

---

## 8. Instance Lifecycle

```
CreateDBInstance ──────▶ Creating ──────▶ Running
     │                                             │
     │                                             ├─────▶ RestartDBInstance
     │                                             │
     │                                             ├─────▶ ModifyDBInstanceSpec
     │                                             │
     │                                             ▼
     │                                       DeleteDBInstance
     │                                             │
     ▼                                             ▼
  Failure → Recovery or recreation          Instance data cleared
                                              (backups retained per policy)
```

### Status Values

| Status | Description | Allowed Operations |
|--------|-------------|-------------------|
| `Creating` | Instance being created | None |
| `Running` | Instance operational | All operations |
| `Modifying` | Spec or parameters changing | Restricted |
| `Restarting` | Instance being restarted | None |
| `Error` | Error state | Debug or recreate |
| `Released` | Instance deleted | None |

---

## 9. Regions Supporting RDS

| Region | RegionID | Status |
|--------|----------|--------|
| 华北2 (北京) | `cn-beijing` | Commercial |
| 华东2 (上海) | `cn-shanghai` | Commercial |
| 华南1 (广州) | `cn-guangzhou` | Commercial |
| 中国香港 | `cn-hongkong` | Commercial |
| 亚太东南 (新加坡) | `ap-southeast-1` | Commercial |
| 亚太东南 (雅加达) | `ap-southeast-3` | Commercial |

---

## 10. Limits and Quotas

| Resource | Default Quota | Notes |
|----------|---------------|-------|
| RDS instances per region | 10 | Can request increase |
| Databases per instance | 200 | — |
| Accounts per instance | 50 | — |
| Max connections (depends on spec) | Varies | 100–64,000 |
| Backup retention | 7–730 days | Configurable |
| Parameter groups | 50 | — |
| IP allow lists | 100 | — |

---

*This reference document is part of the ve-rds-ops skill.*
