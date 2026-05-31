# Elasticsearch Core Concepts

## Architecture Overview

Volcengine Elasticsearch Service provides fully managed Elasticsearch clusters with search, analytics, and indexing capabilities.

### Key Components

| Component | Description |
|-----------|-------------|
| **Instance** | A complete Elasticsearch cluster with data nodes, master nodes, and Kibana |
| **Data Node** | Stores data and executes search/indexing operations |
| **Master Node** | Manages cluster state and coordinates operations |
| **Kibana** | Visualization and management interface for Elasticsearch |
| **Index** | Collection of documents with similar characteristics |
| **Shard** | Subdivision of an index for distributed storage and processing |
| **Replica** | Copy of a shard for high availability and query throughput |
| **Snapshot** | Backup of an index or entire cluster to TOS-compatible storage |

### Node Types

| Type | Description | Recommended Count |
|------|-------------|-------------------|
| **Data Node** | Stores and indexes data; runs queries | 3+ for production |
| **Master Node** | Cluster management; dedicated in production | 3 for HA |
| **Kibana** | Not a node; managed separately | 1 per instance |

### Instance Specifications

| Spec | vCPU | Memory (GB) | Use Case |
|------|------|-------------|----------|
| es.x2.small | 2 | 4 | Development, low-volume indexing |
| es.x2.medium | 4 | 8 | Small production, moderate workloads |
| es.x4.medium | 4 | 16 | General production |
| es.x4.large | 8 | 32 | High-throughput production |
| es.x8.large | 8 | 64 | Memory-intensive workloads |
| es.x8.xlarge | 16 | 128 | Large-scale enterprise |

### Storage Options

| Type | Description | Max IOPS | Use Case |
|------|-------------|----------|----------|
| **ESSD PL0** | Basic ESSD | 10,000 | Development, cold data |
| **ESSD PL1** | Standard ESSD | 50,000 | General production |
| **ESSD PL2** | High-performance ESSD | 100,000 | IOPS-intensive workloads |
| **ESSD PL3** | Ultra-high ESSD | 1,000,000 | Write-heavy, log ingestion |

## Index and Shard Design

### Shard Strategy

| Factor | Recommendation |
|--------|----------------|
| **Shard size** | 10-50 GB per shard (target ~30 GB) |
| **Max shards per node** | < 1000 per GB of heap |
| **Primary shards** | Set at index creation; cannot be changed |
| **Replica count** | 1 for production (2 for critical) |

### Index Lifecycle

| Phase | Description |
|-------|-------------|
| **Hot** | Active indexing and querying — SSD storage |
| **Warm** | Less frequent queries — lower-cost storage |
| **Cold** | Rarely accessed — read-only, compressed |
| **Delete** | Data retention expired — permanent removal |

## Elasticsearch Versions

| Version | Status | Notes |
|---------|--------|-------|
| 7.10 | GA | Widely supported |
| 7.16 | Recommended | Production-ready, many plugins available |
| 8.5 | Latest | Newest features, check plugin compatibility |
| 8.10 | Preview | Latest version, limited plugin support |

## Snapshot and Restore

### Snapshot Repository

Snapshots are stored in TOS-compatible object storage buckets. Each instance has a default repository configured automatically.

### Snapshot Types

| Type | Description |
|------|-------------|
| **Manual** | On-demand snapshots created via API |
| **Automatic** | Scheduled snapshots (configurable interval) |

### Restore Options

| Option | Description |
|--------|-------------|
| Full cluster restore | Restore all indices from snapshot |
| Selective index restore | Restore specific indices only |
| Rename on restore | Restore with different index name |

## Security

### Authentication

| Method | Description |
|--------|-------------|
| **IAM** | Volcengine IAM for API management access |
| **Kibana auth** | Built-in username/password for Kibana |
| **IP whitelist** | Restrict access by IP address |

### Network Security

- Deployed in user VPC
- Private subnet deployment recommended
- Security group rules control inbound access

## Quotas and Limits

| Resource | Default Limit |
|----------|---------------|
| Instances per region | 20 |
| Indices per instance | 1000 |
| Shards per instance | 10000 |
| Snapshot retention | 7 days (auto) |
| Concurrent snapshots | 5 per instance |
| Kibana sessions | 50 concurrent |
| Storage per node | 20-2000 GB |
| Node count per instance | 1-50 |

## Monitoring Metrics

### Key Metrics

| Metric | Namespace | Description |
|--------|-----------|-------------|
| ClusterHealth | `Volcengine_ES` | Green/Yellow/Red status |
| SearchLatency | `Volcengine_ES` | Average search latency (ms) |
| IndexingLatency | `Volcengine_ES` | Average indexing latency (ms) |
| DiskUsage | `Volcengine_ES` | Disk usage percentage |
| JVMHeapUsage | `Volcengine_ES` | JVM heap usage percentage |
| CPUUsage | `Volcengine_ES` | CPU usage percentage |
| DocumentsCount | `Volcengine_ES` | Total document count |
| QueryRate | `Volcengine_ES` | Queries per second |
| IndexingRate | `Volcengine_ES` | Documents indexed per second |
| CircuitBreakerTripped | `Volcengine_ES` | Circuit breaker trips count |

## Version Compatibility for Upgrades

| From | To | Supported | Notes |
|------|----|-----------|-------|
| 7.10 | 7.16 | Yes | Simple upgrade |
| 7.10 | 8.5 | No | Must go through 7.16 first |
| 7.16 | 8.5 | Yes | May require reindexing |
| 8.5 | 8.10 | Yes | Minor upgrade |
