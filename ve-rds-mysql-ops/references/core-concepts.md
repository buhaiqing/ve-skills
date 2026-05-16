# RDS MySQL Core Concepts

## Architecture

Volcengine RDS for MySQL provides managed MySQL database instances:

```
RDS MySQL Instance
├── DB Engine: MySQL (5.7 / 8.0)
├── Primary Node (read-write)
├── Secondary Node (read-only replica, high availability)
├── Read-Only Nodes (0-10 optional, for read scaling)
├── Storage (LocalSSD / ESSD, 20-3000 GB)
├── Network (VPC, Subnet, Connection Strings)
├── Parameters (tunable MySQL config)
├── Accounts (DB users with privileges)
├── IP Whitelist (access control)
└── Backup (automatic + manual)
```

## Node Types

| Node Type | Role | Count | Description |
|-----------|------|-------|-------------|
| Primary | Read-Write | 1 | Master node handles all writes |
| Secondary | Read-Only Replica | 1 | High availability, automatic failover |
| Read-Only | Read-Only | 0-10 | Horizontal read scaling |

## Engine Versions

| Version | Features | Status |
|---------|----------|--------|
| MySQL 5.7 | Widely used, stable | Supported |
| MySQL 8.0 | CTEs, Window functions, JSON improvements | Recommended |

## Storage Types

| Type | Performance | Use Case |
|------|-------------|----------|
| LocalSSD | Lowest latency, highest IOPS | Production, OLTP workloads |
| ESSD (PL0/PL1/PL2) | Flexible capacity, elastic | General purpose, development |

## Node Specifications

Common specs (format: `rds.mysql.{cpu}c{mem}g`):
- `rds.mysql.1c2g` — 1 CPU, 2GB RAM
- `rds.mysql.2c4g` — 2 CPU, 4GB RAM
- `rds.mysql.4c8g` — 4 CPU, 8GB RAM
- `rds.mysql.8c16g` — 8 CPU, 16GB RAM
- `rds.mysql.16c32g` — 16 CPU, 32GB RAM
- `rds.mysql.16c64g` — 16 CPU, 64GB RAM

## Connection

Instances provide connection strings:
- **Primary**: Read-write endpoint
- **Secondary**: Read-only endpoint (for replica)
- **Read-Only**: Separate endpoint for read-only nodes

All connections are within VPC only.

## Parameters

Each MySQL parameter has:
- `ParameterValue`: Current value
- `ParameterDefaultValue`: Default value
- `ForceRestart`: Whether change requires restart
- `CheckingCode`: Valid value range (e.g., `[1-65535]`)
- `ParameterDescription`: Human-readable description

## Backup and Restore

| Feature | Description |
|---------|-------------|
| Automatic Backup | Scheduled daily backups with configurable retention |
| Manual Backup | On-demand backup at any time |
| Point-in-Time Recovery | Restore to any second within backup retention period |
| Restore to New Instance | Creates a new instance from backup |

## Limits

| Resource | Default Limit |
|----------|--------------|
| Instances per account | 50 |
| Read-only nodes per instance | 10 |
| Storage per instance | 20-3000 GB |
| Accounts per instance | 500 |
| IP list entries | 1000 |
