# SLS Core Concepts

## Architecture

SLS (Simple Log Service / TLS) provides log collection, storage, indexing, analysis, and delivery.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Project** | Container for log topics, defines region and billing |
| **Topic** | A log stream with a specific schema and TTL |
| **Shard** | Partition for parallel ingestion and querying |
| **Index** | Enables full-text and key-value search on log content |
| **LogShipper** | Delivers logs from SLS to TOS or other destinations |
| **Machine Group** | Group of ECS instances running Logtail agent |
| **Logtail** | Agent for collecting logs from ECS instances |

## How it Works

1. Logtail agent on ECS collects log files
2. Logs are sent to SLS topic for storage and indexing
3. Indexed logs are searchable via SearchLogs API
4. LogShipper delivers logs to TOS for long-term retention
