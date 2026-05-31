# Kafka Troubleshooting Guide

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| `InvalidParameter` / 400 | Request parameter invalid | Align parameters with OpenAPI spec |
| `InvalidInstance.NotFound` / 404 | Instance does not exist | Verify instance ID; may be deleted |
| `InvalidVpc.NotFound` / 400 | VPC does not exist | Create VPC via `ve-vpc-ops` first |
| `InvalidSubnet.NotFound` / 400 | Subnet does not exist | Create subnet first |
| `InvalidZone.NotFound` / 400 | Zone not available | Select different zone |
| `TopicAlreadyExists` / 400 | Topic already exists | Use different name or delete first |
| `TopicNotFound` / 404 | Topic does not exist | Verify topic name |
| `UserAlreadyExists` / 400 | SASL user already exists | Use different username |
| `UserNotFound` / 404 | SASL user does not exist | Verify username |
| `GroupNotFound` / 404 | Consumer group not found | Verify group ID |
| `ACLAlreadyExists` / 400 | ACL rule already exists | Skip or delete first |
| `QuotaExceeded` / 400 | Resource quota exceeded | Request quota increase |
| `QuotaExceeded.Partition` / 400 | Partition quota exceeded | Scale instance or delete unused topics |
| `QuotaExceeded.User` / 400 | User quota exceeded | Delete unused users |
| `InsufficientBalance` / 400 | Account balance insufficient | Recharge account |
| `Unauthorized` / 403 | IAM permission denied | Check IAM policies |
| `Forbidden.RAM` / 403 | RAM policy denies access | Add Kafka permissions to IAM policy |
| `InvalidInstanceStatus` / 400 | Instance status invalid | Wait for instance to reach stable state |
| `DeleteTopicPartial` / 500 | Partial topic deletion failure | Check broker health |
| `InternalError` / 500 | Server-side error | Retry with backoff; escalate if persists |
| `Throttling` / 429 | Rate limit exceeded | Back off and retry |
| `ServiceUnavailable` / 503 | Service temporarily unavailable | Retry after delay |

## Diagnostic Order

### 1. Instance Issues

**Symptom: Instance creation fails**

```bash
# Check VPC exists
ve vpc DescribeVpc --Region cn-beijing --VpcId vpc-xxx

# Check subnet exists
ve vpc DescribeSubnet --Region cn-beijing --SubnetId subnet-xxx

# Check quota
ve kafka ListInstances --Region cn-beijing | jq '.Result.TotalCount'

# Check account balance
ve billing DescribeBalance
```

**Symptom: Instance stuck in Creating/Starting state**

```bash
# Check instance status
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '.Result.Status'

# Wait and poll
for i in {1..60}; do
  STATUS=$(ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Status')
  echo "Status: $STATUS (attempt $i)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done
```

### 2. Topic Issues

**Symptom: Cannot create topic**

```bash
# Check instance status
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '.Result.Status'

# Check existing topics
ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx | jq '.Result.Topics[].TopicName'

# Check partition quota
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '.Result | {Quota: .PartitionQuota, Used: .UsedPartition}'

# Check for duplicate name
ve kafka DescribeTopic --Region cn-beijing --InstanceId kafka-xxx --TopicName "target-topic" 2>&1 | grep -q "TopicNotFound" && echo "Name available" || echo "Topic exists"
```

**Symptom: Topic deletion fails**

```bash
# Check if topic exists
ve kafka DescribeTopic --Region cn-beijing --InstanceId kafka-xxx --TopicName "target-topic"

# Check for active producers/consumers
ve kafka ListGroups --Region cn-beijing --InstanceId kafka-xxx

# Check if deletion is partial
ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx | grep "target-topic"
```

### 3. SASL/Authentication Issues

**Symptom: Authentication fails**

```bash
# Check user exists
ve kafka ListUsers --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Users[].UserName'

# Check user details
ve kafka DescribeUser --Region cn-beijing --InstanceId kafka-xxx --UserName "target-user"

# Check ACLs for user
ve kafka ListACLs --Region cn-beijing --InstanceId kafka-xxx --ResourceType Topic --ResourceName "*"

# Verify password (if using kcat/kafka-console-producer)
echo "Test auth with kcat:"
kcat -b kafka-xxx.cn-beijing.kafka.volces.com:9093 -L -X security.protocol=SASL_SSL -X sasl.mechanism=SCRAM-SHA-512 -X sasl.username="user" -X sasl.password="<masked>"
```

**Symptom: Authorization fails (ACL)**

```bash
# List all ACLs
ve kafka ListACLs --Region cn-beijing --InstanceId kafka-xxx --ResourceType Topic --ResourceName "target-topic"

# Check user has required permission
ve kafka ListACLs --Region cn-beijing --InstanceId kafka-xxx --ResourceType Topic --ResourceName "target-topic" | jq '.Result.ACLs[] | select(.UserName=="target-user")'

# Verify resource type matches operation (Topic for produce/consume, Group for consumer groups)
```

### 4. Consumer Group Issues

**Symptom: High consumer lag**

```bash
# Check lag
ve kafka DescribeConsumerLag --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group"

# Check group members
ve kafka DescribeGroup --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group" | jq '.Result.Members | length'

# Check partition count
ve kafka DescribeTopic --Region cn-beijing --InstanceId kafka-xxx --TopicName "target-topic" | jq '.Result.PartitionNum'

# Recommendation: consumer count should match partition count
```

**Symptom: Consumer group rebalancing frequently**

```bash
# Check consumer session timeout
ve kafka DescribeGroup --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group" | jq '.Result.Members[].ClientHost'

# Multiple members from same host may indicate over-scaling
# Check for consumer errors in application logs
```

**Symptom: Cannot reset offset**

```bash
# Check group has no active members
ve kafka DescribeGroup --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group" | jq '.Result.Members'

# If members exist, consumers must be stopped first
# Then retry reset
```

### 5. Connection Issues

**Symptom: Cannot connect to brokers**

```bash
# Get correct endpoints
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.BrokerList[].Endpoint'

# Check VPC connectivity
# Ensure client is in same VPC or VPC peering is configured
# For SASL: use port 9093, for plaintext: use port 9092

# Test with kcat (if available)
kcat -b kafka-xxx.cn-beijing.kafka.volces.com:9092 -L
```

### 6. Performance Issues

**Symptom: High latency**

```bash
# Check instance specs
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '.Result.InstanceType'

# Check disk usage
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '.Result | {Total: .StorageSpace, Used: .UsedStorageSpace}'

# If disk > 80%, scale storage
```

**Symptom: Throughput bottlenecks**

```bash
# Check partition count
echo "Partitions per topic:"
ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Topics[].TopicName' | while read topic; do
  COUNT=$(ve kafka DescribeTopic --Region cn-beijing --InstanceId kafka-xxx --TopicName "$topic" | jq -r '.Result.PartitionNum')
  echo "$topic: $COUNT partitions"
done

# Recommendation: ensure sufficient partitions for throughput
```

## Error Recovery Procedures

### Instance Creation Failure

1. **Check VPC/Subnet:**
   ```bash
   ve vpc DescribeVpc --Region cn-beijing --VpcId vpc-xxx
   ve vpc DescribeSubnet --Region cn-beijing --SubnetId subnet-xxx
   ```

2. **Verify quota:**
   ```bash
   CURRENT=$(ve kafka ListInstances --Region cn-beijing | jq -r '.Result.TotalCount')
   echo "Current instances: $CURRENT"
   ```

3. **Check balance:**
   ```bash
   ve billing DescribeBalance
   ```

### Topic Creation Failure

1. **Check instance status:**
   ```bash
   ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Status'
   ```

2. **Check partition quota:**
   ```bash
   ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq '{quota: .Result.PartitionQuota, used: .Result.UsedPartition, available: (.Result.PartitionQuota - .Result.UsedPartition)}'
   ```

3. **Check for duplicate:**
   ```bash
   ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Topics[].TopicName' | grep -i "target-topic"
   ```

### Authentication Failure

1. **Verify user exists:**
   ```bash
   ve kafka ListUsers --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Users[].UserName' | grep "target-user"
   ```

2. **Check mechanism:**
   ```bash
   ve kafka DescribeUser --Region cn-beijing --InstanceId kafka-xxx --UserName "target-user" | jq '.Result.Mechanism'
   ```

3. **Verify ACLs:**
   ```bash
   ve kafka ListACLs --Region cn-beijing --InstanceId kafka-xxx --ResourceType Topic --ResourceName "target-topic"
   ```

### Consumer Lag Issues

1. **Check current lag:**
   ```bash
   ve kafka DescribeConsumerLag --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group"
   ```

2. **Check consumer count:**
   ```bash
   ve kafka DescribeGroup --Region cn-beijing --InstanceId kafka-xxx --GroupId "target-group" | jq '.Result.Members | length'
   ```

3. **Check partition count:**
   ```bash
   ve kafka DescribeTopic --Region cn-beijing --InstanceId kafka-xxx --TopicName "target-topic" | jq '.Result.PartitionNum'
   ```

4. **Resolution:**
   - If lag > 10000: Scale consumers to match partition count
   - If consumers < partitions: Add more consumer instances
   - If processing is slow: Optimize consumer logic

## Prevention Checklist

- [ ] Instance naming follows convention (env-purpose)
- [ ] VPC and subnet verified before instance creation
- [ ] Partition quota planned based on expected throughput
- [ ] SASL users created with strong passwords
- [ ] ACLs follow least privilege principle
- [ ] Consumer group names are descriptive
- [ ] Monitoring alerts configured for consumer lag
- [ ] Retention configured based on data requirements
- [ ] Replication factor appropriate for environment
- [ ] min.insync.replicas configured for durability needs
