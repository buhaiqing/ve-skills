# Core Concepts — Volcengine Redis (缓存数据库 Redis 版)

> **Purpose:** Fundamental concepts for Volcengine Redis.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [What is Redis](#1-what-is-redis)
2. [Instance Classes](#2-instance-classes)
3. [Engine Versions](#3-engine-versions)
4. [Architecture Types](#4-architecture-types)
5. [Data Layout](#5-data-layout)
6. [Sharding](#6-sharding)
7. [Instance Lifecycle](#7-instance-lifecycle)
8. [Regions Supporting Redis](#8-regions-supporting-redis)
9. [Limits and Quotas](#9-limits-and-quotas)

---

## 1. What is Redis

A **Redis (缓存数据库 Redis 版)** is a fully managed in-memory data store compatible with the Redis protocol. It provides sub-millisecond latency for caching, session management, leaderboards, and real-time analytics.

---

## 2. Instance Classes

| Class | Code | HA | Persistence | Use Case |
|-------|------|-----|-------------|----------|
| **Standalone** | `Standalone` | No | None | Dev/test, non-critical cache |
| **Primary-Secondary** | `PrimarySecondary` | Auto failover | AOF + RDB | Production, HA required |
| **Sharded Cluster** | `ShardedCluster` | Multi-shard HA | AOF + RDB | High throughput, large datasets |

---

## 3. Engine Versions

| Version | Status | Notable Features |
|---------|--------|-----------------|
| `5.0` | Stable | Stream datatype, modules |
| `6.0` | Stable | ACL, RESP3, SSL, multi-part keys |
| `7.0` | Latest | Functions, new commands, improved eviction |

---

## 4. Architecture Types

### Standalone (单节点)

- Single node, in-memory only
- Data lost on node restart/failure
- Lowest cost, highest risk

### Primary-Secondary (主从)

- Primary handles all writes; replica handles failover + optional reads
- Automatic failover if primary fails
- Persistent storage via AOF/RDB snapshots

### Sharded Cluster (分片集群)

- Multiple shards distribute data via hash slots
- Each shard is a primary-secondary pair
- Scales both memory and throughput

---

## 5. Data Layout

| Layout | Description | Use Case |
|--------|-------------|----------|
| **RAM** | In-memory only, highest performance | Low-latency cache |
| **DRAM** | DRAM + disk, larger capacity | Large datasets, cost-effective |

---

## 6. Sharding

Each shard has a maximum capacity (1 GB, 2 GB, 4 GB, 8 GB, 16 GB, 32 GB, 64 GB). Total instance capacity = `ShardNumber × ShardCapacity`.

---

## 7. Instance Lifecycle

| Status | Description | Allowed Operations |
|--------|-------------|-------------------|
| `Creating` | Being provisioned | None |
| `Running` | Operational | All |
| `Modifying` | Spec change in progress | Restricted |
| `Restarting` | Being restarted | None |
| `Error` | Error state | Debug or recreate |
| `Released` | Deleted | None |

---

## 8. Regions Supporting Redis

Query available regions via API:
```bash
ve redis DescribeRegions
```
---

## 9. Limits and Quotas

| Resource | Default Quota | Notes |
|----------|---------------|-------|
| Redis instances per region | 20 | Can request increase |
| Max keys per instance | Memory limited | Bound by maxmemory |
| Max connections per instance | Varies by spec | 10,000–100,000 |
| Allow lists per instance | 100 | — |
| Database accounts per instance | 50 | — |

---

*This reference document is part of the ve-redis-ops skill.*
