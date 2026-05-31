# Elasticsearch Monitoring

## Key Metrics

### Cluster Health Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| ClusterStatus | `Volcengine_ES` | enum | Green/Yellow/Red — overall cluster health |
| ActiveShards | `Volcengine_ES` | count | Number of active primary shards |
| ActiveShardsPercent | `Volcengine_ES` | percent | Percentage of active shards |
| UnassignedShards | `Volcengine_ES` | count | Shards not assigned to any node |
| RelocatingShards | `Volcengine_ES` | count | Shards being relocated |
| InitializingShards | `Volcengine_ES` | count | Shards being initialized |
| DelayedUnassignedShards | `Volcengine_ES` | count | Shards with delayed unassignment |

### Performance Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| SearchLatency | `Volcengine_ES` | ms | Average search query latency |
| SearchRate | `Volcengine_ES` | qps | Search queries per second |
| IndexingLatency | `Volcengine_ES` | ms | Average indexing latency |
| IndexingRate | `Volcengine_ES` | docs/s | Documents indexed per second |
| BulkIndexingRate | `Volcengine_ES` | docs/s | Bulk indexing throughput |
| BulkRejectionCount | `Volcengine_ES` | count | Bulk requests rejected (circuit breaker) |

### Resource Utilization Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| CPUUsage | `Volcengine_ES` | percent | CPU usage percentage |
| MemoryUsage | `Volcengine_ES` | percent | Memory usage percentage |
| JVMHeapUsage | `Volcengine_ES` | percent | JVM heap usage percentage |
| JVMHeapMax | `Volcengine_ES` | bytes | Maximum JVM heap size |
| DiskUsage | `Volcengine_ES` | percent | Disk usage percentage per node |
| DiskFreeBytes | `Volcengine_ES` | bytes | Free disk space |
| NetworkIn | `Volcengine_ES` | bytes/s | Inbound network traffic |
| NetworkOut | `Volcengine_ES` | bytes/s | Outbound network traffic |

### Document and Index Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| DocumentsCount | `Volcengine_ES` | count | Total document count across all indices |
| DocumentsDeleted | `Volcengine_ES` | count | Deleted document count |
| IndexStorageBytes | `Volcengine_ES` | bytes | Total index storage size |
| IndexCount | `Volcengine_ES` | count | Total number of indices |

### Thread Pool Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| SearchThreadPoolQueue | `Volcengine_ES` | count | Search thread pool queue size |
| SearchThreadPoolRejected | `Volcengine_ES` | count | Search thread pool rejections |
| IndexThreadPoolQueue | `Volcengine_ES` | count | Index thread pool queue size |
| IndexThreadPoolRejected | `Volcengine_ES` | count | Index thread pool rejections |
| BulkThreadPoolQueue | `Volcengine_ES` | count | Bulk thread pool queue size |
| BulkThreadPoolRejected | `Volcengine_ES` | count | Bulk thread pool rejections |

## Alert Examples

### Cluster Health Alert

```json
{
  "AlertName": "ESClusterHealthNotGreen",
  "Metric": "ClusterStatus",
  "Namespace": "Volcengine_ES",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "es-xxx"}
  ],
  "Condition": "Value != Green",
  "Duration": 60,
  "Severity": "Critical",
  "Description": "Elasticsearch cluster health is not Green"
}
```

### Disk Usage Alert

```json
{
  "AlertName": "ESDiskUsageHigh",
  "Metric": "DiskUsage",
  "Namespace": "Volcengine_ES",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "es-xxx"}
  ],
  "Condition": "Average > 80",
  "Duration": 300,
  "Severity": "Warning",
  "Description": "ES node disk usage exceeds 80%"
}
```

### JVM Heap Alert

```json
{
  "AlertName": "ESJVMHeapHigh",
  "Metric": "JVMHeapUsage",
  "Namespace": "Volcengine_ES",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "es-xxx"}
  ],
  "Condition": "Average > 85",
  "Duration": 300,
  "Severity": "Warning",
  "Description": "JVM heap usage exceeds 85%"
}
```

### Search Latency Alert

```json
{
  "AlertName": "ESSearchLatencyHigh",
  "Metric": "SearchLatency",
  "Namespace": "Volcengine_ES",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "es-xxx"}
  ],
  "Condition": "Average > 1000",
  "Duration": 300,
  "Severity": "Warning",
  "Description": "Average search latency exceeds 1000ms"
}
```

### Circuit Breaker Alert

```json
{
  "AlertName": "ESCircuitBreakerTripped",
  "Metric": "BulkRejectionCount",
  "Namespace": "Volcengine_ES",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "es-xxx"}
  ],
  "Condition": "Sum > 0",
  "Duration": 60,
  "Severity": "Critical",
  "Description": "ES circuit breaker has tripped — bulk requests rejected"
}
```

## Monitoring via CMS

### List Available Metrics

```bash
ve cms ListMetrics --Region cn-beijing --Namespace Volcengine_ES
```

### Query Metric Data

```bash
# Query cluster health
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_ES \
  --MetricName ClusterStatus \
  --Dimensions '[]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 300

# Query disk usage
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_ES \
  --MetricName DiskUsage \
  --Dimensions '[{"Name":"InstanceId","Value":"es-xxx"}]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 300
```

### Create Alarm Rule

```bash
# Create disk usage alert
ve cms CreateAlarmRule \
  --Region cn-beijing \
  --RuleName "es-disk-usage-alert" \
  --Namespace Volcengine_ES \
  --MetricName DiskUsage \
  --Dimensions '[{"Name":"InstanceId","Value":"es-xxx"}]' \
  --EvaluationCount 1 \
  --ComparisonOperator GreaterThan \
  --Threshold 80 \
  --Period 300 \
  --ContactGroupId "cg-xxx"
```

## Monitoring Best Practices

### Threshold Recommendations

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| ClusterStatus | Yellow | Red | Investigate shard allocation |
| DiskUsage | > 75% | > 85% | Scale storage or delete indices |
| JVMHeapUsage | > 75% | > 90% | Scale node specs or tune queries |
| SearchLatency | > 500ms | > 1000ms | Optimize queries, add nodes |
| IndexingLatency | > 200ms | > 500ms | Bulk index, scale nodes |
| UnassignedShards | > 0 | > 10 | Investigate node health |
| CircuitBreakerTripped | > 0 | > 5 | Reduce bulk sizes, scale cluster |
| ThreadPoolQueue | > 50 | > 200 | Scale nodes, check for slow queries |
| DocumentsDeleted | > 10% of total | > 25% of total | Force segment merge |

### Dashboard Structure

```
ES Monitoring Dashboard
├── Cluster Overview
│   ├── Cluster Health Status
│   ├── Active/Unassigned Shards
│   ├── Total Documents & Indices
│   └── Data Node Count
├── Performance
│   ├── Search Rate & Latency
│   ├── Indexing Rate & Latency
│   ├── Bulk Rejection Count
│   └── Thread Pool Queue Sizes
├── Resource Utilization
│   ├── CPU Usage
│   ├── JVM Heap Usage
│   ├── Disk Usage per Node
│   └── Network I/O
├── Index Health
│   ├── Top Indices by Size
│   ├── Top Indices by Document Count
│   ├── Index Growth Rate
│   └── Index Health Status
└── Snapshot Status
    ├── Last Snapshot Time
    ├── Snapshot Success Rate
    └── Snapshot Storage Usage
```

## Health Check Script

```bash
#!/bin/bash
# Elasticsearch Health Check

INSTANCE_ID="es-xxx"
REGION="cn-beijing"

echo "=== ES Health Check for $INSTANCE_ID ==="

# Check instance status
STATUS=$(ve elasticsearch DescribeInstances --Region $REGION --InstanceId $INSTANCE_ID | jq -r '.Result.Instances[0].Status')
echo "Instance Status: $STATUS"

if [ "$STATUS" != "Running" ]; then
  echo "ERROR: Instance is not running!"
  exit 1
fi

# Check version
VERSION=$(ve elasticsearch DescribeInstances --Region $REGION --InstanceId $INSTANCE_ID | jq -r '.Result.Instances[0].Version')
echo "ES Version: $VERSION"

# Check storage
echo ""
echo "=== Storage ==="
ve elasticsearch DescribeInstances --Region $REGION --InstanceId $INSTANCE_ID | jq '.Result.Instances[0] | {Total: .StorageSpaceGb, Spec: .NodeSpec, Nodes: .NodeNumber}'

# List indices
echo ""
echo "=== Index Summary ==="
ve elasticsearch ListIndices --Region $REGION --InstanceId $INSTANCE_ID | jq '{TotalIndices: .Result.TotalCount}'

# Check Kibana status
echo ""
echo "=== Kibana ==="
KIBANA=$(ve elasticsearch DescribeKibana --Region $REGION --InstanceId $INSTANCE_ID 2>&1)
if echo "$KIBANA" | grep -q "KibanaEndpoint"; then
  echo "Kibana: Enabled"
  echo "Kibana URL: $(echo "$KIBANA" | jq -r '.Result.KibanaEndpoint')"
else
  echo "Kibana: Disabled"
fi

echo ""
echo "=== Health Check Complete ==="
```

## Log Analysis

### Slow Query Logs

Monitor slow query logs through the Elasticsearch REST API or Kibana. Key fields to track:

- `query`: The search query that was executed
- `query_time_ms`: Query execution time
- `index`: Target index
- `shard`: Shard that executed the query
- `node`: Node that handled the query

### Audit Logs

Audit logs track management operations:
- Instance creation/deletion
- Index creation/deletion
- Plugin installations
- Kibana access
- Snapshot operations
