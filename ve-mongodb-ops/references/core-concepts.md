# MongoDB Core Concepts

> **Version:** 1.1.0 | **Last Updated:** 2026-07-04

## Architecture

Volcengine MongoDB provides managed MongoDB instances with two deployment architectures:

```
MongoDB Instance
├── Replica Set (3-node replica set)
│   ├── Primary Node (read-write)
│   ├── Secondary Node 1 (replica)
│   └── Secondary Node 2 (replica)
│
└── Sharded Cluster (distributed)
    ├── Config Servers (3-node replica set)
    ├── Mongos Routers (1-32 nodes)
    └── Shard Nodes (2-32 shards, each 3-node replica set)
```

## Deployment Architectures

### Replica Set (recommended)

| Component | Count | Role |
|-----------|-------|------|
| Primary | 1 | Handles all writes |
| Secondary | 2 | Replicate from primary, readable |

- ✅ Auto failover (10-30s), majority write concern, minimum 3 nodes

### Sharded Cluster (large-scale workloads)

| Component | Count | Role |
|-----------|-------|------|
| Mongos | 1-32 | Query routing |
| Config Server | 3 | Metadata |
| Shard | 2-32 | Data storage (each = replica set) |

- ✅ Horizontal scaling, auto data distribution via sharding, supports >TB datasets

## MongoDB Versions

Query available versions: `ve mongodb DescribeDBInstanceSpecs --Region <region>`

| Version Range | Key Features | Status |
|--------------|-------------|--------|
| 4.x | Transactions, Change Streams, Materialized views | Legacy/Stable |
| 5.x | Live resharding, Time Series, Client-side FLE | Recommended |
| 6.x | Clustered collections, Time Series improvements | Latest |

## Node Specifications

Query available specs: `ve mongodb DescribeDBInstanceSpecs --Region <region> --MongoVersion <version>`

| Tier | vCPU | Memory | Use Case |
|------|------|--------|----------|
| Small | 1-2 | 2-4 GB | Dev/test, small apps |
| Medium | 4-8 | 8-16 GB | Production apps |
| Large | 16 | 32-64 GB | Enterprise, memory-intensive |

## Storage Types

| Type | Performance | Use Case |
|------|-------------|----------|
| ESSD PL0 | Baseline | Dev, light workloads |
| ESSD PL1 | Enhanced (~50k IOPS) | Production |
| ESSD PL2 | High (~100k IOPS) | High throughput |

## Connection Architecture

```
mongodb://[user]:[pass]@[host]:[port]/[db]?replicaSet=[rsName]
```

| Mode | Behavior |
|------|----------|
| Primary | Write-heavy workloads |
| SecondaryPreferred | Read-heavy, eventual consistency OK |
| Nearest | Geo-distribution |

## User Privilege Model

| Role | Scope |
|------|-------|
| `Read` | Read single DB |
| `ReadWrite` | Read/write single DB |
| `dbAdmin` | DB administration |
| `userAdmin` | User management |
| `root` | Full instance access |
| `ReadAnyDatabase` | Read all |
| `ReadWriteAnyDatabase` | Read/write all |

## Backup Types

| Type | Description | Retention |
|------|-------------|-----------|
| Auto | Scheduled daily | 7-730 days |
| Manual | On-demand | Until deleted |
| Incremental | Log-based | Same as auto |

## Key Limits

Query current limits via API. Typical defaults:
- Instances per account: 50
- Databases per instance: 200
- Collections per DB: 10,000
- Document size: 16 MB
- Indexes per collection: 64
- Users per instance: 500
- IP whitelist entries: 1,000
- Storage per instance: 20-3000 GB
- Connections: spec-dependent (500-16000)

## Default Ports
- MongoDB: **27017**
- Config Server (sharded): **27019**

## Best Practices

- ✅ Replica set for most workloads; sharded for >TB data
- ⚠️ 30% storage headroom for growth
- ✅ Deploy across multiple AZs
- ✅ Create indexes for queried fields
- ✅ Enable auto backups with adequate retention
- ✅ Use VPC + IP whitelists + SSL/TLS