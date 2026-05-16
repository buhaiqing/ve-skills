# RDS PostgreSQL Core Concepts

## Architecture

Volcengine RDS for PostgreSQL provides managed PostgreSQL instances:

```
RDS PostgreSQL Instance
├── DB Engine: PostgreSQL (11 / 12 / 13 / 14 / 15 / 16 / 17)
├── Primary Node (read-write)
├── Secondary Node (standby, automatic failover)
├── Read-Only Instances (optional, for read scaling)
├── Storage (LocalSSD, 20-3000 GB)
├── Network (VPC, Subnet, Endpoints)
├── Parameters (kernel + non-kernel, e.g., shared_buffers, wal_level)
├── Accounts & Schemas (PostgreSQL-specific)
├── Allow Lists (access control)
└── Backup (automatic + manual, PITR support)
```

## Engine Versions

| Version | Key Features | Status |
|---------|-------------|--------|
| PostgreSQL 11 | Partitioning improvements, JIT | Supported |
| PostgreSQL 12 | CTE inlining, parallel queries | Supported |
| PostgreSQL 13 | Incremental sort, duplicate indexes | Supported |
| PostgreSQL 14 | Subscriptions, security improvements | Recommended |
| PostgreSQL 15 | Schema ACL improvements, sorting | Supported |
| PostgreSQL 16 | Logical replication enhancements | Latest |
| PostgreSQL 17 | Latest release | Preview/New |

## Node Specifications

Common specs (format: `rds.postgres.{cpu}c{mem}g`):
- `rds.postgres.1c2g` — 1 CPU, 2GB RAM
- `rds.postgres.2c4g` — 2 CPU, 4GB RAM
- `rds.postgres.4c8g` — 4 CPU, 8GB RAM
- `rds.postgres.8c16g` — 8 CPU, 16GB RAM
- `rds.postgres.16c32g` — 16 CPU, 32GB RAM
- `rds.postgres.16c64g` — 16 CPU, 64GB RAM

## Storage

**Storage Type:** LocalSSD (fixed for PostgreSQL on Volcengine)
- Range: 20-3000 GB
- Step: 10 GB
- Default: 100 GB

## Node Types

| Type | Role | Description |
|------|------|-------------|
| Primary | Read-Write | Master handles all writes and replication source |
| Secondary | Standby | Automatic failover target; not directly queryable |
| ReadOnly | Read-Only | Separate instance for read scaling; has own endpoint |

## Connection & Endpoints

Instances provide endpoint info:
- **Primary Endpoint:** Read-write connection
- **Read-Only Endpoint:** For read-only node(s)
- All within VPC; no public access by default

## Parameters

PostgreSQL parameters include:
- **Kernel parameters:** `shared_buffers`, `wal_level`, `max_connections`, `work_mem`, `effective_cache_size`
- **Non-kernel parameters:** Extension-specific settings
- Each has `ForceRestart` flag indicating if restart needed

## Read-Only Instances

Read-only instances are created from the primary instance:
- Replicated via streaming replication
- Separate billing
- Own node spec and zone configuration
- Separate endpoint for connections

## Limits

| Resource | Default Limit |
|----------|--------------|
| Instances per account | 50 |
| Read-only instances per primary | 5 |
| Storage per instance | 20-3000 GB |
| Accounts per instance | 500 |
