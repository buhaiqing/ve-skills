# Kafka CLI Usage

## Installation and Configuration

### Install ve CLI

```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify installation
ve version
```

### Configure Credentials

**Environment Variables (Recommended for Agents):**

```bash
export VOLCENGINE_ACCESS_KEY="your-access-key"
export VOLCENGINE_SECRET_KEY="<masked>"  # Never display in output
export VOLCENGINE_REGION="cn-beijing"
```

**Verify credentials are set:**
```bash
test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "Credentials configured"
```

**Config File:**

```bash
mkdir -p ~/.volcengine
cat > ~/.volcengine/config.json << 'CONFIGEOF'
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "your-access-key",
      "secret_key": "<masked>",
      "region": "cn-beijing"
    }
  ]
}
CONFIGEOF
```

## CLI Conventions

- **Output is JSON by default**
- **Service prefix**: `ve kafka`
- **Help**: `ve kafka --help` or `ve kafka <action> --help`
- **Region is required** for most operations

## Command Reference

### Instance Management

```bash
# List instances
ve kafka ListInstances --Region cn-beijing

# Describe instance
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx

# Create instance
ve kafka CreateInstance \
  --Region cn-beijing \
  --InstanceName "prod-kafka" \
  --InstanceType "kafka.n1.x2.small" \
  --StorageSpace 300 \
  --PartitionQuota 1000 \
  --VpcId "vpc-xxx" \
  --SubnetId "subnet-xxx" \
  --ZoneId "cn-beijing-a"

# Modify instance
ve kafka ModifyInstance \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --InstanceName "new-name"

# Delete instance
ve kafka DeleteInstance \
  --Region cn-beijing \
  --InstanceId kafka-xxx
```

### Topic Management

```bash
# List topics
ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx

# Describe topic
ve kafka DescribeTopic \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --TopicName "orders"

# Create topic
ve kafka CreateTopic \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --TopicName "orders" \
  --PartitionNumber 6 \
  --ReplicaNumber 3 \
  --MinInsyncReplicas 2 \
  --RetentionHours 168 \
  --MaxMessageSize 1048576

# Delete topic
ve kafka DeleteTopic \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --TopicName "orders"

# Modify topic
ve kafka ModifyTopic \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --TopicName "orders" \
  --RetentionHours 72
```

### Partition Operations

```bash
# Add partitions
ve kafka CreatePartitions \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --TopicName "orders" \
  --PartitionNumber 12
```

### User/SASL Management

```bash
# List users
ve kafka ListUsers --Region cn-beijing --InstanceId kafka-xxx

# Describe user
ve kafka DescribeUser \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --UserName "app-producer"

# Create user
ve kafka CreateUser \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --UserName "app-producer" \
  --Password "<masked>" \
  --Mechanism SCRAM-SHA-512

# Delete user
ve kafka DeleteUser \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --UserName "app-producer"
```

### ACL Management

```bash
# List ACLs for a resource
ve kafka ListACLs \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --ResourceType Topic \
  --ResourceName "orders"

# Create ACL (allow produce)
ve kafka CreateACL \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --ResourceType Topic \
  --ResourceName "orders" \
  --UserName "app-producer" \
  --PermissionType Allow \
  --Operation Write \
  --Host "*"

# Create ACL (allow consume)
ve kafka CreateACL \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --ResourceType Topic \
  --ResourceName "orders" \
  --UserName "app-consumer" \
  --PermissionType Allow \
  --Operation Read \
  --Host "*"

# Create ACL (allow group join)
ve kafka CreateACL \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --ResourceType Group \
  --ResourceName "order-processors" \
  --UserName "app-consumer" \
  --PermissionType Allow \
  --Operation Read \
  --Host "*"

# Delete ACL
ve kafka DeleteACL \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --ResourceType Topic \
  --ResourceName "orders" \
  --UserName "app-producer" \
  --PermissionType Allow \
  --Operation Write \
  --Host "*"
```

### Consumer Group Management

```bash
# List consumer groups
ve kafka ListGroups --Region cn-beijing --InstanceId kafka-xxx

# Describe consumer group
ve kafka DescribeGroup \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --GroupId "order-processors"

# Describe consumer lag
ve kafka DescribeConsumerLag \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --GroupId "order-processors"

# Reset offset to earliest
ve kafka ResetOffset \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --GroupId "order-processors" \
  --TopicName "orders" \
  --ResetType earliest

# Reset offset to latest
ve kafka ResetOffset \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --GroupId "order-processors" \
  --TopicName "orders" \
  --ResetType latest

# Reset offset to specific timestamp
ve kafka ResetOffset \
  --Region cn-beijing \
  --InstanceId kafka-xxx \
  --GroupId "order-processors" \
  --TopicName "orders" \
  --ResetType timestamp \
  --Timestamp 1716777600
```

## CLI vs API Coverage Gap

| Operation | Available via `ve` CLI | Notes |
|-----------|------------------------|-------|
| CreateInstance | Yes | — |
| DescribeInstance | Yes | — |
| ModifyInstance | Yes | Limited fields |
| DeleteInstance | Yes | — |
| ListInstances | Yes | — |
| CreateTopic | Yes | — |
| DescribeTopic | Yes | — |
| DeleteTopic | Yes | — |
| ListTopics | Yes | — |
| ModifyTopic | Yes | Limited configs |
| CreatePartitions | Yes | — |
| ListGroups | Yes | — |
| DescribeGroup | Yes | — |
| ResetOffset | Yes | — |
| DescribeConsumerLag | Yes | — |
| CreateUser | Yes | — |
| DeleteUser | Yes | — |
| ListUsers | Yes | — |
| DescribeUser | Yes | — |
| CreateACL | Yes | — |
| DeleteACL | Yes | — |
| ListACLs | Yes | — |

## Common Patterns

### Check Instance Status

```bash
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Status'
```

### Get Bootstrap Endpoints

```bash
ve kafka DescribeInstance --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.BrokerList[].Endpoint'
```

### Filter Topics by Name

```bash
ve kafka ListTopics --Region cn-beijing --InstanceId kafka-xxx | jq -r '.Result.Topics[].TopicName' | grep "prod-"
```

### Get Consumer Group State

```bash
ve kafka DescribeGroup --Region cn-beijing --InstanceId kafka-xxx --GroupId "order-processors" | jq -r '.Result.State'
```

### Calculate Total Consumer Lag

```bash
ve kafka DescribeConsumerLag --Region cn-beijing --InstanceId kafka-xxx --GroupId "order-processors" | jq -r '.Result.TotalLag'
```

## JSON Path Quick Reference

| Field | JSON Path |
|-------|-----------|
| Instance ID | `.Result.InstanceId` |
| Instance Status | `.Result.Status` |
| Instance Name | `.Result.InstanceName` |
| Broker Endpoints | `.Result.BrokerList[].Endpoint` |
| Topic Name | `.Result.TopicName` |
| Partition Count | `.Result.PartitionNum` |
| Replication Factor | `.Result.ReplicaNum` |
| Consumer Group State | `.Result.State` |
| Total Consumer Lag | `.Result.TotalLag` |
| Request ID | `.ResponseMetadata.RequestId` |
| Error Code | `.Error.Code` |
| Error Message | `.Error.Message` |
