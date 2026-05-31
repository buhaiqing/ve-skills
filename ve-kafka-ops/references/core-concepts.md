# Kafka Core Concepts

## Architecture Overview

Volcengine Kafka is a fully managed Apache Kafka service providing distributed streaming platform capabilities.

### Key Components

| Component | Description |
|-----------|-------------|
| **Instance** | A complete Kafka cluster with brokers, ZooKeeper, and management plane |
| **Broker** | Kafka server that handles message storage and retrieval |
| **Topic** | Logical channel for message publication and subscription |
| **Partition** | Subdivision of a topic for parallel processing |
| **Producer** | Client that publishes messages to topics |
| **Consumer** | Client that subscribes to topics and processes messages |
| **Consumer Group** | Group of consumers that jointly consume from a topic |
| **ZooKeeper** | Coordination service for broker management (internal) |

### Instance Types

| Type | Specs | Use Case |
|------|-------|----------|
| kafka.n1.x2.small | 2 vCPU, 4GB RAM | Development, small workloads |
| kafka.n1.x4.small | 4 vCPU, 8GB RAM | Medium throughput |
| kafka.n1.x4.medium | 4 vCPU, 16GB RAM | High throughput |
| kafka.n1.x8.medium | 8 vCPU, 32GB RAM | Enterprise workloads |

### Storage Options

- **Cloud Disk**: Standard cloud storage, scalable
- **ESSD**: Enhanced SSD for high IOPS workloads

## Topic and Partition Design

### Partition Strategy

| Factor | Recommendation |
|--------|----------------|
| **Throughput** | 1 partition ≈ 10MB/s write throughput |
| **Parallelism** | Match partition count to consumer count |
| **Ordering** | Messages within partition are ordered |
| **Maximum** | 300 partitions per topic, 1000 per instance |

### Replication Factor

| Environment | Recommended RF | Tolerance |
|-------------|----------------|-----------|
| Development | 1 | No broker failure tolerance |
| Staging | 2 | 1 broker failure |
| Production | 3 | 2 broker failures |

### min.insync.replicas

- **Set to 2** when RF=3 for strong durability
- **Set to 1** for higher availability, lower durability

## SASL Authentication

### Supported Mechanisms

| Mechanism | Description | Use Case |
|-----------|-------------|----------|
| **PLAIN** | Simple username/password | Development, internal networks |
| **SCRAM-SHA-256** | Salted challenge-response | Production environments |
| **SCRAM-SHA-512** | Stronger salted challenge-response | High-security production |

### User Management

- Maximum 100 SASL users per instance
- Username: 1-64 characters, alphanumeric and hyphens
- Password: 8-64 characters, must include uppercase, lowercase, and number

## ACL Authorization

### Resource Types

| Type | Description | Example |
|------|-------------|---------|
| **Topic** | Message channels | `orders`, `events.clicks` |
| **Group** | Consumer groups | `analytics-consumers` |
| **Cluster** | Cluster-level operations | — |
| **TransactionalId** | Transactional producer IDs | `payment-producer` |

### Operations

| Operation | Description |
|-----------|-------------|
| **Read** | Consume messages |
| **Write** | Produce messages |
| **Create** | Create topics |
| **Delete** | Delete topics |
| **Describe** | View metadata |
| **DescribeConfigs** | View configurations |
| **Alter** | Modify configurations |
| **AlterConfigs** | Change dynamic configs |

### Permission Types

- **Allow**: Grant permission
- **Deny**: Explicitly deny (deny overrides allow)

## Consumer Groups and Offset Management

### Offset Storage

- Offsets stored in internal `__consumer_offsets` topic
- Automatic commit or manual commit modes
- Retention: 7 days default

### Reset Types

| Type | Description |
|------|-------------|
| **earliest** | Reset to beginning of topic |
| **latest** | Reset to end of topic |
| **timestamp** | Reset to specific time |

### Consumer Lag

| Lag Level | Action |
|-----------|--------|
| < 1000 | Healthy |
| 1000 - 10000 | Warning: scale consumers |
| > 10000 | Critical: investigate |

## Network Architecture

### Endpoints

| Endpoint Type | Format | Purpose |
|---------------|--------|---------|
| **API Endpoint** | `kafka.{region}.volces.com` | Management API |
| **Bootstrap** | `kafka-{instance-id}.{region}.kafka.volces.com:9092` | Kafka protocol (plaintext) |
| **Bootstrap SASL** | `kafka-{instance-id}.{region}.kafka.volces.com:9093` | Kafka protocol (SASL) |

### VPC Integration

- Kafka instances deploy in user VPC
- Supports private subnet deployment
- Cross-VPC access via VPC peering

## Monitoring Metrics

### Key Metrics

| Metric | Namespace | Description |
|--------|-----------|-------------|
| MessagesInPerSec | `Volcengine_Kafka` | Messages in per second |
| BytesInPerSec | `Volcengine_Kafka` | Bytes in per second |
| BytesOutPerSec | `Volcengine_Kafka` | Bytes out per second |
| ConsumerLag | `Volcengine_Kafka` | Consumer lag per group |
| BrokerDiskUsage | `Volcengine_Kafka` | Disk usage percentage |
| UnderReplicatedPartitions | `Volcengine_Kafka` | Under-replicated count |
| ActiveControllerCount | `Volcengine_Kafka` | Active controllers |

## Quotas and Limits

| Resource | Default Limit |
|----------|---------------|
| Instances per region | 10 |
| Topics per instance | 100 |
| Partitions per topic | 300 |
| Partitions per instance | 1000 |
| SASL users per instance | 100 |
| ACL rules per instance | 500 |
| Consumer groups per instance | 500 |
| Message size | 10MB |
| Retention period | 7 days |

## Version Support

| Kafka Version | Status |
|---------------|--------|
| 2.6 | Recommended |
| 2.5 | Supported |
| 2.4 | Supported |
| 2.3 | Deprecated |
