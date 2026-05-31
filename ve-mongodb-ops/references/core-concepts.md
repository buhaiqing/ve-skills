# MongoDB Core Concepts

## Architecture

Volcengine MongoDB provides managed MongoDB instances with two deployment architectures:

```
MongoDB Instance
├── Replica Set (3-node replica set)
│   ├── Primary Node (read-write)
│   ├── Secondary Node 1 (replica)
│   └── Secondary Node 2 (replica)
│
└── Sharded Cluster (distributed architecture)
    ├── Config Servers (3-node replica set)
    ├── Mongos Routers (1-32 nodes)
    └── Shard Nodes (2-32 shards, each 3-node replica set)

Common Components:
├── MongoDB Engine (WiredTiger)
├── Storage (ESSD PL0/PL1/PL2, 20-3000 GB)
├── Network (VPC, Subnet, Connection Strings)
├── Databases (logical containers)
├── Collections (document containers)
├── Indexes (query optimization)
├── Users (authentication & authorization)
├── IP Whitelist (access control)
└── Backup (automatic + manual)
```

## Deployment Architectures

### Replica Set (Recommended for most workloads)

| Component | Count | Role |
|-----------|-------|------|
| Primary | 1 | Handles all writes, readable |
| Secondary | 2 | Replicates from primary, readable |
| Hidden | 0-1 | Disaster recovery, not readable |

**Characteristics:**
- Automatic failover (10-30 seconds)
- Majority write concern by default
- Secondary nodes can serve read queries
- Minimum 3 nodes for high availability

### Sharded Cluster (For large-scale workloads)

| Component | Count | Role |
|-----------|-------|------|
| Mongos | 1-32 | Query routing layer |
| Config Server | 3 | Metadata and configuration |
| Shard | 2-32 | Data storage (each is a replica set) |

**Characteristics:**
- Horizontal scaling for data and queries
- Automatic data distribution via sharding
- Supports massive datasets (>TB)
- Higher complexity and cost

## MongoDB Versions

| Version | Features | Status | Recommendation |
|---------|----------|--------|----------------|
| 4.0 | Transactions (multi-doc), Change Streams | Supported | Legacy |
| 4.2 | On-demand materialized views, Wildcard indexes | Supported | Stable |
| 4.4 | Hedged reads, Hidden indexes, Union | Supported | Stable |
| 5.0 | Live resharding, Time Series collections, Client-side FLE | Supported | Recommended |
| 6.0 | Clustered collections, Time Series improvements | Supported | Latest |

## Node Specifications

Common specs (format varies by version):

| Spec | CPU | Memory | Max Connections | Recommended For |
|------|-----|--------|-----------------|-----------------|
| mongo.1c2g | 1 | 2GB | 500 | Development, testing |
| mongo.2c4g | 2 | 4GB | 1000 | Small applications |
| mongo.4c8g | 4 | 8GB | 2000 | Medium applications |
| mongo.8c16g | 8 | 16GB | 4000 | Large applications |
| mongo.16c32g | 16 | 32GB | 8000 | Enterprise workloads |
| mongo.16c64g | 16 | 64GB | 16000 | Memory-intensive |

## Storage Types

| Type | Performance | Use Case |
|------|-------------|----------|
| ESSD PL0 | Baseline (~10k IOPS) | Development, light workloads |
| ESSD PL1 | Enhanced (~50k IOPS) | Production, balanced workloads |
| ESSD PL2 | High performance (~100k IOPS) | High throughput, low latency |

## Connection Architecture

### Connection String Format

```
mongodb://[username]:[password]@[host]:[port]/[database]?replicaSet=[replicaSetName]
```

Example:
```
mongodb://myuser:mypass@mongo-xxx.mongodb.volces.com:27017/mydb?replicaSet=mgset-xxx
```

### Connection Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| Primary | Connect to primary only | Write-heavy workloads |
| SecondaryPreferred | Prefer secondaries for reads | Read-heavy, eventual consistency OK |
| PrimaryPreferred | Prefer primary, fallback to secondary | Read consistency required |
| Nearest | Connect to nearest node by latency | Geographic distribution |

## Database and Collection Model

```
Instance
├── Database (admin) - system database
├── Database (local) - replication data
├── Database (config) - sharding config (sharded only)
├── Database (custom)
│   ├── Collection A
│   │   ├── Documents (BSON)
│   │   └── Indexes
│   └── Collection B
└── Database (another_custom)
    └── Collections...
```

## User Privilege Model

| Role | Privileges | Use Case |
|------|------------|----------|
| `Read` | Read any database | Read-only applications |
| `ReadWrite` | Read and write any database | Standard applications |
| `dbAdmin` | Database administration | DBAs, schema management |
| `userAdmin` | User management | Security administration |
| `root` | Full instance access | Emergency, super admin |
| `ReadAnyDatabase` | Read all databases | Monitoring, analytics |
| `ReadWriteAnyDatabase` | Read/write all databases | Multi-tenant apps |

## Backup Types

| Type | Description | Retention |
|------|-------------|-----------|
| Automatic Backup | Scheduled daily backups | 7-730 days (configurable) |
| Manual Backup | On-demand backup at any time | Until deleted |
| Incremental Backup | Log-based incremental | Same as automatic |

## Limits

| Resource | Default Limit |
|----------|--------------|
| Instances per account | 50 |
| Databases per instance | 200 |
| Collections per database | 10,000 |
| Documents per collection | Unlimited (hardware limited) |
| Document size | 16 MB |
| Indexes per collection | 64 |
| Compound index fields | 32 |
| Users per instance | 500 |
| IP whitelist entries | 1000 |
| Storage per instance | 20-3000 GB |
| Connection limit | Spec-dependent (500-16000) |

## Default Ports

| Service | Port |
|---------|------|
| MongoDB | 27017 |
| MongoDB Sharded (mongos) | 27017 |
| MongoDB Config Server | 27019 |

## Best Practices

### Instance Creation
- Choose appropriate architecture (replica set vs sharded)
- Select MongoDB version based on feature requirements
- Size storage with 30% headroom for growth
- Deploy across multiple availability zones

### Performance
- Create indexes for frequently queried fields
- Use compound indexes for multi-field queries
- Monitor `db.currentOp()` for slow queries
- Enable profiling for query analysis

### Security
- Use VPC isolation
- Configure IP whitelists
- Enable SSL/TLS connections
- Use strong authentication (SCRAM)
- Rotate passwords regularly

### Backup Strategy
- Enable automatic backups with appropriate retention
- Test restore procedures regularly
- Keep manual backups before major changes
- Store critical backups cross-region
