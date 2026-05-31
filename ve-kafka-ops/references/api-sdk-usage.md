# Kafka API & SDK Usage

## OpenAPI Reference

- **API Documentation**: https://www.volcengine.com/docs/6410
- **Base Path**: `https://kafka.{region}.volces.com`
- **API Version**: 2020-05-01

## SDK Operations Map

| Goal | API Operation | CLI Command | SDK Method |
|------|---------------|-------------|------------|
| Create Instance | CreateInstance | `ve kafka CreateInstance` | `CreateInstance` |
| Describe Instance | DescribeInstance | `ve kafka DescribeInstance` | `DescribeInstance` |
| Modify Instance | ModifyInstance | `ve kafka ModifyInstance` | `ModifyInstance` |
| Delete Instance | DeleteInstance | `ve kafka DeleteInstance` | `DeleteInstance` |
| List Instances | ListInstances | `ve kafka ListInstances` | `ListInstances` |
| Create Topic | CreateTopic | `ve kafka CreateTopic` | `CreateTopic` |
| Describe Topic | DescribeTopic | `ve kafka DescribeTopic` | `DescribeTopic` |
| Delete Topic | DeleteTopic | `ve kafka DeleteTopic` | `DeleteTopic` |
| List Topics | ListTopics | `ve kafka ListTopics` | `ListTopics` |
| Modify Topic | ModifyTopic | `ve kafka ModifyTopic` | `ModifyTopic` |
| Create Partitions | CreatePartitions | `ve kafka CreatePartitions` | `CreatePartitions` |
| List Groups | ListGroups | `ve kafka ListGroups` | `ListGroups` |
| Describe Group | DescribeGroup | `ve kafka DescribeGroup` | `DescribeGroup` |
| Reset Offset | ResetOffset | `ve kafka ResetOffset` | `ResetOffset` |
| Describe Consumer Lag | DescribeConsumerLag | `ve kafka DescribeConsumerLag` | `DescribeConsumerLag` |
| Create User | CreateUser | `ve kafka CreateUser` | `CreateUser` |
| Delete User | DeleteUser | `ve kafka DeleteUser` | `DeleteUser` |
| List Users | ListUsers | `ve kafka ListUsers` | `ListUsers` |
| Describe User | DescribeUser | `ve kafka DescribeUser` | `DescribeUser` |
| Create ACL | CreateACL | `ve kafka CreateACL` | `CreateACL` |
| Delete ACL | DeleteACL | `ve kafka DeleteACL` | `DeleteACL` |
| List ACLs | ListACLs | `ve kafka ListACLs` | `ListACLs` |

## Request / Response Examples

### CreateInstance

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceName": "prod-events",
  "InstanceType": "kafka.n1.x2.small",
  "StorageSpace": 300,
  "PartitionQuota": 1000,
  "VpcId": "vpc-xxx",
  "SubnetId": "subnet-xxx",
  "ZoneId": "cn-beijing-a",
  "Version": "2.6"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123456",
    "Action": "CreateInstance",
    "Version": "2020-05-01",
    "Service": "kafka",
    "Region": "cn-beijing"
  },
  "Result": {
    "InstanceId": "kafka-xxx"
  }
}
```

### DescribeInstance

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123457",
    "Action": "DescribeInstance",
    "Version": "2020-05-01"
  },
  "Result": {
    "InstanceId": "kafka-xxx",
    "InstanceName": "prod-events",
    "InstanceType": "kafka.n1.x2.small",
    "Status": "Running",
    "Version": "2.6",
    "StorageSpace": 300,
    "UsedStorageSpace": 45,
    "PartitionQuota": 1000,
    "UsedPartition": 45,
    "VpcId": "vpc-xxx",
    "SubnetId": "subnet-xxx",
    "ZoneId": "cn-beijing-a",
    "BrokerList": [
      {
        "BrokerId": 1,
        "Endpoint": "kafka-xxx.cn-beijing.kafka.volces.com:9092"
      }
    ],
    "CreateTime": "2024-05-20T10:00:00+08:00",
    "ExpiredTime": "2025-05-20T10:00:00+08:00",
    "ChargeType": "PostPaid"
  }
}
```

### CreateTopic

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "TopicName": "orders",
  "PartitionNumber": 6,
  "ReplicaNumber": 3,
  "MinInsyncReplicas": 2,
  "RetentionHours": 168,
  "MaxMessageSize": 1048576
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123458",
    "Action": "CreateTopic"
  },
  "Result": {}
}
```

### DescribeTopic

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "TopicName": "orders"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123459",
    "Action": "DescribeTopic"
  },
  "Result": {
    "TopicName": "orders",
    "PartitionNum": 6,
    "ReplicaNum": 3,
    "Status": "Running",
    "CreateTime": "2024-05-20T10:05:00+08:00",
    "PartitionDetails": [
      {
        "PartitionId": 0,
        "LeaderBrokerId": 1,
        "ReplicaBrokerIds": [1, 2, 3],
        "IsrBrokerIds": [1, 2, 3]
      }
    ],
    "Config": {
      "retention.ms": "604800000",
      "min.insync.replicas": "2",
      "max.message.bytes": "1048576"
    }
  }
}
```

### CreateUser

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "UserName": "app-producer",
  "Password": "SecurePass123!",
  "Mechanism": "SCRAM-SHA-512"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123460",
    "Action": "CreateUser"
  },
  "Result": {}
}
```

### CreateACL

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "ResourceType": "Topic",
  "ResourceName": "orders",
  "UserName": "app-producer",
  "PermissionType": "Allow",
  "Operation": "Write",
  "Host": "*"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123461",
    "Action": "CreateACL"
  },
  "Result": {}
}
```

### DescribeGroup

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "GroupId": "order-processors"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123462",
    "Action": "DescribeGroup"
  },
  "Result": {
    "GroupId": "order-processors",
    "State": "Stable",
    "Protocol": "range",
    "Members": [
      {
        "MemberId": "consumer-1",
        "ClientId": "order-service",
        "ClientHost": "10.0.1.5",
        "Assignments": [
          {
            "Topic": "orders",
            "Partitions": [0, 1, 2]
          }
        ]
      }
    ]
  }
}
```

### DescribeConsumerLag

**Request:**
```json
{
  "Region": "cn-beijing",
  "InstanceId": "kafka-xxx",
  "GroupId": "order-processors"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "202405271234567890123463",
    "Action": "DescribeConsumerLag"
  },
  "Result": {
    "GroupId": "order-processors",
    "Lags": [
      {
        "Topic": "orders",
        "Partition": 0,
        "ConsumerOffset": 1000000,
        "LogEndOffset": 1000100,
        "Lag": 100
      },
      {
        "Topic": "orders",
        "Partition": 1,
        "ConsumerOffset": 2000000,
        "LogEndOffset": 2000150,
        "Lag": 150
      }
    ],
    "TotalLag": 250
  }
}
```

## Pagination

List operations support pagination via `Limit` and `Offset` parameters:

```json
{
  "Region": "cn-beijing",
  "Limit": 20,
  "Offset": 0
}
```

Response includes pagination info:
```json
{
  "Result": {
    "TotalCount": 100,
    "Instances": [...]
  }
}
```

## Required Fields Summary

| Operation | Required Fields |
|-----------|-----------------|
| CreateInstance | Region, InstanceName, InstanceType, StorageSpace, PartitionQuota, VpcId, SubnetId, ZoneId |
| DescribeInstance | Region, InstanceId |
| DeleteInstance | Region, InstanceId |
| CreateTopic | Region, InstanceId, TopicName, PartitionNumber, ReplicaNumber |
| DeleteTopic | Region, InstanceId, TopicName |
| CreatePartitions | Region, InstanceId, TopicName, PartitionNumber |
| CreateUser | Region, InstanceId, UserName, Password |
| DeleteUser | Region, InstanceId, UserName |
| CreateACL | Region, InstanceId, ResourceType, ResourceName, UserName, PermissionType, Operation |
| DeleteACL | Region, InstanceId, ResourceType, ResourceName, UserName, PermissionType, Operation |
| ResetOffset | Region, InstanceId, GroupId, TopicName, ResetType |
