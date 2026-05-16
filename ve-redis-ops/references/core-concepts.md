# Redis Core Concepts

## Architecture

Volcengine Cache for Redis provides Redis-compatible managed instances:

```
Redis Instance
├── Engine Version (4.0 / 5.0 / 6.0)
├── Capacity (256MB - 64GB per shard)
├── Network (VPC, Subnet, Private Address)
├── Node Configuration
│   ├── NodeNumber (2 for primary-secondary, 4+ for cluster)
│   ├── ShardNumber (for sharded cluster)
│   └── ShardCapacity (per shard memory)
├── Security (AllowList, Accounts)
└── Backup Configuration
```

## Instance Types

| Type | Description | Nodes | Use Case |
|------|-------------|-------|----------|
| Primary-Secondary | 1 master + 1 replica | 2 | General purpose |
| Sharded Cluster | Multiple shards with replicas | 4+ | High-throughput, large datasets |

## Engine Versions

| Version | Features | EOL |
|---------|----------|-----|
| 4.0 | Basic Redis 4.0 features | Legacy |
| 5.0 | Streams, modules support | Supported |
| 6.0 | ACL, RESP3, SSL/TLS | Latest |

## Capacity Planning

| Specification | Memory (MB) | Connections | Bandwidth (MB/s) |
|--------------|-------------|-------------|-----------------|
| 256MB | 256 | 10,000 | 48 |
| 1024MB (1GB) | 1024 | 20,000 | 96 |
| 2048MB (2GB) | 2048 | 20,000 | 96 |
| 4096MB (4GB) | 4096 | 20,000 | 192 |
| 8192MB (8GB) | 8192 | 20,000 | 384 |
| 16384MB (16GB) | 16384 | 40,000 | 768 |

## Connection

- **Private Address**: `redis-{instance_id}.redis.ivolces.com:6379`
- **VPC Only**: Instances are accessible only within the configured VPC
- **AllowList**: IP whitelist controls access (default: no IPs allowed)

## Data Persistence

| Method | Description |
|--------|-------------|
| RDB | Point-in-time snapshots |
| AOF | Append-only file (continuous) |
| Automatic Backup | Scheduled daily backups |

## Deletion Protection

Instances support deletion protection. Must be disabled before deletion.

## Limits

| Resource | Default Limit |
|----------|--------------|
| Instances per account | 50 |
| Allowlists per instance | 50 |
| IPs per allowlist | 300 |
| Accounts per instance | 10 |
