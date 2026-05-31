# Kafka Monitoring

## Key Metrics

### Throughput Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| MessagesInPerSec | `Volcengine_Kafka` | count/s | Messages received per second |
| BytesInPerSec | `Volcengine_Kafka` | bytes/s | Data in per second |
| BytesOutPerSec | `Volcengine_Kafka` | bytes/s | Data out per second |
| ProduceRequestsPerSec | `Volcengine_Kafka` | count/s | Produce requests per second |
| FetchRequestsPerSec | `Volcengine_Kafka` | count/s | Fetch requests per second |

### Consumer Lag Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| ConsumerLag | `Volcengine_Kafka` | count | Lag for consumer group |
| ConsumerLagPerPartition | `Volcengine_Kafka` | count | Lag per partition |

### Broker Health Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| BrokerDiskUsage | `Volcengine_Kafka` | percent | Disk usage percentage |
| BrokerCPUUsage | `Volcengine_Kafka` | percent | CPU usage |
| BrokerMemoryUsage | `Volcengine_Kafka` | percent | Memory usage |
| UnderReplicatedPartitions | `Volcengine_Kafka` | count | Under-replicated partitions |
| OfflinePartitions | `Volcengine_Kafka` | count | Offline partitions |
| ActiveControllerCount | `Volcengine_Kafka` | count | Active controllers |

### Topic Metrics

| Metric | Namespace | Unit | Description |
|--------|-----------|------|-------------|
| TopicMessagesInPerSec | `Volcengine_Kafka` | count/s | Messages in per topic |
| TopicBytesInPerSec | `Volcengine_Kafka` | bytes/s | Bytes in per topic |
| TopicBytesOutPerSec | `Volcengine_Kafka` | bytes/s | Bytes out per topic |

## Alert Examples

### Consumer Lag Alert

```json
{
  "AlertName": "KafkaConsumerLagHigh",
  "Metric": "ConsumerLag",
  "Namespace": "Volcengine_Kafka",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "kafka-xxx"},
    {"Name": "GroupId", "Value": "order-processors"}
  ],
  "Condition": "Average > 10000",
  "Duration": 300,
  "Severity": "Warning",
  "Description": "Consumer lag exceeds 10000 messages"
}
```

### Disk Usage Alert

```json
{
  "AlertName": "KafkaDiskUsageHigh",
  "Metric": "BrokerDiskUsage",
  "Namespace": "Volcengine_Kafka",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "kafka-xxx"},
    {"Name": "BrokerId", "Value": "1"}
  ],
  "Condition": "Average > 80",
  "Duration": 300,
  "Severity": "Critical",
  "Description": "Broker disk usage exceeds 80%"
}
```

### Under-Replicated Partitions Alert

```json
{
  "AlertName": "KafkaUnderReplicatedPartitions",
  "Metric": "UnderReplicatedPartitions",
  "Namespace": "Volcengine_Kafka",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "kafka-xxx"}
  ],
  "Condition": "Average > 0",
  "Duration": 60,
  "Severity": "Critical",
  "Description": "Partitions are under-replicated"
}
```

### Instance Status Alert

```json
{
  "AlertName": "KafkaInstanceNotRunning",
  "Metric": "InstanceStatus",
  "Namespace": "Volcengine_Kafka",
  "Dimensions": [
    {"Name": "InstanceId", "Value": "kafka-xxx"}
  ],
  "Condition": "Status != Running",
  "Duration": 60,
  "Severity": "Critical",
  "Description": "Kafka instance is not in Running state"
}
```

## Monitoring via CMS

### List Available Metrics

```bash
# List Kafka metrics in CMS
ve cms ListMetrics --Region cn-beijing --Namespace Volcengine_Kafka
```

### Query Metric Data

```bash
# Query consumer lag
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_Kafka \
  --MetricName ConsumerLag \
  --Dimensions '[{"Name":"InstanceId","Value":"kafka-xxx"},{"Name":"GroupId","Value":"order-processors"}]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 300

# Query disk usage
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine_Kafka \
  --MetricName BrokerDiskUsage \
  --Dimensions '[{"Name":"InstanceId","Value":"kafka-xxx"},{"Name":"BrokerId","Value":"1"}]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 300
```

### Create Alarm Rule

```bash
# Create consumer lag alert
ve cms CreateAlarmRule \
  --Region cn-beijing \
  --RuleName "kafka-consumer-lag-alert" \
  --Namespace Volcengine_Kafka \
  --MetricName ConsumerLag \
  --Dimensions '[{"Name":"InstanceId","Value":"kafka-xxx"}]' \
  --EvaluationCount 1 \
  --ComparisonOperator GreaterThan \
  --Threshold 10000 \
  --Period 300 \
  --ContactGroupId "cg-xxx"
```

## Monitoring Best Practices

### Threshold Recommendations

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| ConsumerLag | > 1000 | > 10000 | Scale consumers |
| BrokerDiskUsage | > 70% | > 80% | Scale storage |
| UnderReplicatedPartitions | > 0 | > 5 | Check broker health |
| OfflinePartitions | > 0 | — | Immediate investigation |
| ActiveControllerCount | != 1 | — | Check controller health |

### Dashboard Structure

```
Kafka Monitoring Dashboard
├── Overview
│   ├── Instance Status
│   ├── Total Topics
│   ├── Total Partitions
│   ├── Active Consumer Groups
│   └── SASL Users
├── Throughput
│   ├── Messages In/sec
│   ├── Bytes In/sec
│   ├── Bytes Out/sec
│   └── Request Rates
├── Consumer Health
│   ├── Lag per Group
│   ├── Lag per Topic
│   ├── Active Consumers
│   └── Rebalance Rate
├── Broker Health
│   ├── Disk Usage per Broker
│   ├── CPU Usage per Broker
│   ├── Memory Usage per Broker
│   └── Under-Replicated Partitions
└── Topic Metrics
    ├── Messages In per Topic
    ├── Top Topics by Throughput
    └── Partition Distribution
```

## Log Analysis

### Access Logs

Access logs are available via TLS (Tencent Log Service) or can be streamed to your log aggregation system.

Key log fields:
- `timestamp`: Request timestamp
- `client_id`: Client identifier
- `client_host`: Client IP
- `request_type`: Produce/Fetch/Metadata/etc.
- `topic`: Target topic
- `response_time_ms`: Request latency
- `error_code`: Kafka error code

### Audit Logs

Audit logs track management operations:
- Instance creation/deletion
- Topic creation/deletion
- User/ACL modifications
- Configuration changes

## Consumer Lag Investigation

### Step-by-Step Diagnosis

1. **Identify affected group:**
   ```bash
   ve kafka DescribeConsumerLag \
     --Region cn-beijing \
     --InstanceId kafka-xxx \
     --GroupId "target-group"
   ```

2. **Check consumer count:**
   ```bash
   ve kafka DescribeGroup \
     --Region cn-beijing \
     --InstanceId kafka-xxx \
     --GroupId "target-group" | jq '.Result.Members | length'
   ```

3. **Compare with partition count:**
   ```bash
   ve kafka DescribeTopic \
     --Region cn-beijing \
     --InstanceId kafka-xxx \
     --TopicName "target-topic" | jq '.Result.PartitionNum'
   ```

4. **Calculate lag per partition:**
   ```bash
   ve kafka DescribeConsumerLag \
     --Region cn-beijing \
     --InstanceId kafka-xxx \
     --GroupId "target-group" | jq '.Result.Lags[] | {partition: .Partition, lag: .Lag}'
   ```

### Resolution Matrix

| Symptom | Cause | Resolution |
|---------|-------|------------|
| Lag increasing, consumers < partitions | Insufficient consumers | Add consumer instances |
| Lag increasing, consumers = partitions | Slow processing | Optimize consumer code |
| Lag spike then recover | Temporary slowdown | Monitor; may auto-resolve |
| Lag on specific partitions | Hot partition | Increase partition count |
| Lag on all partitions equally | Global slowdown | Check network/broker health |

## Health Check Script

```bash
#!/bin/bash
# Kafka Health Check

INSTANCE_ID="kafka-xxx"
REGION="cn-beijing"

echo "=== Kafka Health Check for $INSTANCE_ID ==="

# Check instance status
STATUS=$(ve kafka DescribeInstance --Region $REGION --InstanceId $INSTANCE_ID | jq -r '.Result.Status')
echo "Instance Status: $STATUS"

if [ "$STATUS" != "Running" ]; then
  echo "ERROR: Instance is not running!"
  exit 1
fi

# Check disk usage
echo ""
echo "=== Disk Usage ==="
ve kafka DescribeInstance --Region $REGION --InstanceId $INSTANCE_ID | jq '.Result | {Total: .StorageSpace, Used: .UsedStorageSpace, Percent: ((.UsedStorageSpace / .StorageSpace) * 100)}'

# Check partition usage
echo ""
echo "=== Partition Usage ==="
ve kafka DescribeInstance --Region $REGION --InstanceId $INSTANCE_ID | jq '.Result | {Quota: .PartitionQuota, Used: .UsedPartition, Available: (.PartitionQuota - .UsedPartition)}'

# List topics
echo ""
echo "=== Topics ==="
ve kafka ListTopics --Region $REGION --InstanceId $INSTANCE_ID | jq -r '.Result.Topics[].TopicName'

# Check consumer groups with high lag
echo ""
echo "=== Consumer Groups with Lag ==="
ve kafka ListGroups --Region $REGION --InstanceId $INSTANCE_ID | jq -r '.Result.Groups[].GroupId' | while read group; do
  LAG=$(ve kafka DescribeConsumerLag --Region $REGION --InstanceId $INSTANCE_ID --GroupId "$group" | jq -r '.Result.TotalLag')
  echo "Group: $group, Lag: $LAG"
done

echo ""
echo "=== Health Check Complete ==="
```
