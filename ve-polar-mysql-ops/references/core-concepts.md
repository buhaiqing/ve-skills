# PolarDB MySQL Core Concepts

## Architecture

Volcengine PolarDB for MySQL uses a **compute-storage separation** architecture that enables independent scaling of compute and storage resources.

```
PolarDB MySQL Cluster
├── Compute Layer (Nodes)
│   ├── Primary Node (read-write, 1 required)
│   ├── Secondary Node (read-only, HA standby, 1 required)
│   └── Read-Only Nodes (0-14, horizontal read scaling)
├── Storage Layer (Shared)
│   └── Shared Storage Pool (distributed, multi-AZ)
├── Network
│   ├── VPC (Virtual Private Cloud)
│   ├── Subnet
│   └── Cluster Endpoints (read-write, read-only)
├── Parameters
│   └── Parameter Groups (cluster-level configuration)
└── Backup
    ├── Automatic Backups (scheduled, retained)
    └── Manual Backups (on-demand)
```

## Compute-Storage Separation

PolarDB's key architectural advantage:

| Aspect | Traditional RDS | PolarDB |
|--------|-----------------|---------|
| Compute & Storage | Coupled | Separated |
| Storage scaling | Requires data migration | Instant, online |
| Compute scaling | Minutes to hours | Seconds to minutes |
| Max storage | Limited by node | 100TB+ shared pool |
| Read replicas | Async replication | Shared storage, zero lag |

### Benefits

1. **Elastic Scalability**: Scale compute and storage independently
2. **Cost Efficiency**: Pay for compute when needed; storage is shared
3. **High Performance**: Shared storage eliminates replication lag
4. **High Availability**: Automatic failover with shared storage

## Node Types

| Node Type | Role | Count | Description |
|-----------|------|-------|-------------|
| Primary | Read-Write | 1 | Master node handles all writes and reads |
| Secondary | Read-Only | 1 | High availability standby; automatic failover target |
| Read-Only | Read-Only | 0-14 | Horizontal read scaling (shared storage, no lag) |

### Node Roles

- **Primary**: Handles all write operations; readable for load balancing
- **Secondary**: Hot standby for failover; readable for load balancing
- **Read-Only**: Additional read capacity; no replication lag

### Node Status

| Status | Description |
|--------|-------------|
| `RUNNING` | Node operational |
| `CREATING` | Node being created |
| `DELETING` | Node being deleted |
| `RESTARTING` | Node restarting |
| `ERROR` | Node error state |
| `FAILOVERING` | Failover in progress |

## Storage Pool

PolarDB uses a shared distributed storage pool:

| Feature | Specification |
|---------|--------------|
| Min size | 100 GB |
| Max size | 100,000 GB (100 TB) |
| Scaling | Online, no downtime |
| Redundancy | Multi-AZ replication |
| Performance | Automatic tiering |

### Storage Metrics

- **StorageSpace**: Total allocated storage
- **StorageUsed**: Actual data size
- **Free Space**: Available for growth

## Engine Versions

| Version | Features | Status |
|---------|----------|--------|
| MySQL 5.7 | Widely used, stable | Supported |
| MySQL 8.0 | CTEs, Window functions, JSON improvements | Recommended |

## Node Classes

Common node classes (format varies by region):

| Class | vCPU | Memory | Use Case |
|-------|------|--------|----------|
| polar.mysql.x2.medium | 2 | 4 GB | Development, small workloads |
| polar.mysql.x4.large | 4 | 16 GB | Medium production |
| polar.mysql.x4.xlarge | 8 | 32 GB | Large production |
| polar.mysql.x4.2xlarge | 16 | 64 GB | High throughput |
| polar.mysql.x8.2xlarge | 16 | 128 GB | Memory-intensive |

> Use `DescribeDBNodeClasses` API to get available classes in your region.

## Connection Endpoints

PolarDB provides multiple endpoints for different access patterns:

| Endpoint Type | Description | Use Case |
|--------------|-------------|----------|
| Read-Write Endpoint | Primary node + Secondary node | Write operations, consistent reads |
| Read-Only Endpoint | All read-only nodes | Load-balanced read queries |
| Direct Node Endpoints | Individual node access | Debugging, specific node targeting |

### Endpoint Features

- **Auto-Failover**: Read-write endpoint automatically switches to new primary
- **Load Balancing**: Read-only endpoint distributes traffic across nodes
- **Auto-Add**: New read-only nodes can be automatically added to endpoints

## Failover

PolarDB supports automatic and manual failover:

### Automatic Failover

- Triggers when primary node fails
- Secondary node promoted to primary
- Typically completes in < 30 seconds
- No data loss (shared storage)

### Manual Failover

- User-initiated via `FailoverDBCluster` API
- Useful for maintenance windows
- Secondary node promoted to primary
- Old primary becomes secondary

## High Availability

| HA Feature | Implementation |
|------------|---------------|
| Node redundancy | Minimum 2 nodes (primary + secondary) |
| Storage redundancy | Multi-AZ replication |
| Automatic failover | Secondary promotion on primary failure |
| Connection continuity | Endpoint abstraction masks failover |

## Limits

| Resource | Default Limit |
|----------|--------------|
| Clusters per account | 50 |
| Nodes per cluster | 16 (1 primary + 1 secondary + 14 RO) |
| Read-only nodes per cluster | 14 |
| Storage per cluster | 100-100,000 GB |
| Parameter groups per account | 100 |
| Backups per cluster | 7 automatic + unlimited manual |
| IP whitelist entries | 1000 |

## Comparison: PolarDB vs RDS MySQL

| Feature | PolarDB MySQL | RDS MySQL |
|---------|---------------|-----------|
| Architecture | Compute-storage separation | Traditional coupled |
| Max storage | 100 TB | 3 TB |
| Read replicas | 14 (zero lag) | 10 (async replication) |
| Storage scaling | Online, instant | Requires migration |
| Compute scaling | Fast | Slower |
| Failover time | < 30s | 60-120s |
| Price model | Compute + storage separately | Bundled instance |
| Best for | Large datasets, variable workloads | Smaller, stable workloads |
