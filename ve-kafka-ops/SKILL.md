---
name: ve-kafka-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) Kafka (消息队列 Kafka版) — instance lifecycle, topic/partition
  management, producer/consumer groups, ACLs, SASL authentication, monitoring,
  and scaling operations. User mentions Kafka, 消息队列, or describes messaging
  scenarios (e.g., message streaming, topic creation, consumer lag, SASL setup)
  even without naming the product directly. Not for RabbitMQ, RocketMQ, or other
  message queue services.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Kafka endpoints (kafka.{region}.volces.com).
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-27"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Kafka API is accessible via `ve kafka --help`. CLI supports instance,
    topic, consumer group, user, and ACL management operations.
    API docs: https://www.volcengine.com/docs/6410
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine Kafka Operations Skill

## Overview

Kafka (消息队列 Kafka版) on Volcengine (火山引擎) provides fully managed Apache Kafka service with high throughput, low latency, and horizontal scalability. This skill is an **operational runbook** for agents with **dual-path execution**: `ve` CLI for API calls and JIT Go SDK fallback.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Kafka operations.
  - **`ve kafka`**: Instance, topic, consumer group, user, and ACL management

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 Kafka-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (Kafka), one primary resource model; cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine Kafka", "火山引擎 Kafka", "消息队列 Kafka", or "Kafka"
- Task involves instance operations: CreateInstance, DescribeInstance, ModifyInstance, DeleteInstance, ListInstances
- Task involves topic operations: CreateTopic, DescribeTopic, DeleteTopic, ListTopics, ModifyTopic
- Task involves partition operations: create partitions, describe partition distribution
- Task involves consumer group operations: ListGroups, DescribeGroup, ResetOffset
- Task involves ACL management: CreateACL, DeleteACL, ListACLs
- Task involves user/SASL management: CreateUser, DeleteUser, ModifyUser
- Task involves monitoring: consumer lag, throughput metrics, broker health
- Task involves scaling: horizontal scaling (partitions), vertical scaling (instance specs)
- Task involves SASL/SCRAM authentication setup

### SHOULD NOT Use This Skill When

- Task is about RabbitMQ → delegate to: `ve-rabbitmq-ops` (when present)
- Task is about RocketMQ → delegate to: `ve-rocketmq-ops` (when present)
- Task is about MQTT → delegate to: `ve-mqtt-ops` (when present)
- Task is purely billing → delegate to billing ops

### Delegation Rules

- Kafka instances depend on VPC networking → reference `ve-vpc-ops` for subnet/VPC setup
- Kafka authentication depends on IAM for access control → reference `ve-iam-ops` (when present)
- Kafka metrics are collected via CMS → reference `ve-cms-ops` for monitoring setup

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Access key from runtime | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Secret key from runtime | NEVER ask user; fail if unset; **ALWAYS mask as `<masked>`** |
| `{{env.VOLCENGINE_REGION}}` | Region (e.g., cn-beijing) | Use documented default if skill allows |
| `{{user.instance_id}}` | Kafka instance ID | Ask once; starts with `kafka-` |
| `{{user.instance_name}}` | Kafka instance name | Ask once; globally unique per region |
| `{{user.topic_name}}` | Topic name | Ask once; unique within instance |
| `{{user.partition_count}}` | Number of partitions | Ask; default 3; max 300 per topic |
| `{{user.replication_factor}}` | Replication factor | Ask; default 3; max available brokers |
| `{{user.username}}` | SASL username | Ask once; alphanumeric and hyphens |
| `{{user.password}}` | SASL password | Ask once; min 8 chars, mask in output |
| `{{user.consumer_group_id}}` | Consumer group ID | Ask once |
| `{{output.instance_id}}` | Instance ID from create response | Parse from response |
| `{{output.request_id}}` | Request ID for tracing | Parse from response |

> **Security Warning (Credential Masking):** NEVER echo or log `VOLCENGINE_SECRET_KEY` or any credential values. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.

> **Password Masking:** NEVER display SASL passwords in output. Always show as `<masked>`.

## API and Response Conventions (Agent-Readable)

- **Kafka uses RESTful API** with JSON responses
- **Endpoint:** `kafka.{region}.volces.com` for API calls; `kafka-{instance-id}.{region}.kafka.volces.com:{port}` for Kafka protocol
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/kafka`
- **Error responses:** JSON with `Error` object containing `Code` and `Message`

### Key Response Fields

| Operation | Response Field | Type | Description |
|-----------|---------------|------|-------------|
| CreateInstance | `$.Result.InstanceId` | string | New instance ID |
| DescribeInstance | `$.Result.Status` | string | Instance state |
| DescribeInstance | `$.Result.BrokerList` | array | Broker endpoints |
| ListInstances | `$.Result.Instances[].InstanceId` | array | Instance IDs |
| ListTopics | `$.Result.Topics[].TopicName` | array | Topic names |
| DescribeTopic | `$.Result.PartitionNum` | integer | Partition count |
| DescribeTopic | `$.Result.ReplicaNum` | integer | Replication factor |
| ListGroups | `$.Result.Groups[].GroupId` | array | Consumer group IDs |
| DescribeGroup | `$.Result.Members[]` | array | Active consumers |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| Create | — | `Running` | 30s | 1800s |
| Start | `Stopped` | `Running` | 30s | 600s |
| Stop | `Running` | `Stopped` | 30s | 600s |
| Delete | any stable | `Deleting` → absent | 30s | 1800s |
| Scale | `Running` | `Scaling` → `Running` | 30s | 1800s |

## Quick Start

### What This Skill Does
This skill enables you to manage Volcengine Kafka instances and resources — create instances, manage topics, configure ACLs, set up SASL authentication, monitor consumer lag, and scale resources.

### Prerequisites
- [ ] `ve` CLI installed
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
# List Kafka instances
ve kafka ListInstances --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all Kafka instances
ve kafka ListInstances --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateInstance | Create a new Kafka instance | Medium | Low |
| DescribeInstance | View instance details | Low | None |
| ModifyInstance | Change instance configuration | Medium | Medium |
| DeleteInstance | Remove an instance | Low | **High** — irreversible |
| ListInstances | View all instances | Low | None |
| CreateTopic | Create a topic | Low | Low |
| DescribeTopic | View topic details | Low | None |
| DeleteTopic | Delete a topic | Low | **High** — data loss |
| ListTopics | List all topics | Low | None |
| ModifyTopic | Change topic config | Medium | Medium |
| CreatePartitions | Add partitions | Medium | Low |
| ListGroups | List consumer groups | Low | None |
| DescribeGroup | View group details | Low | None |
| ResetOffset | Reset consumer offset | Medium | **High** — data reprocessing |
| CreateUser | Create SASL user | Low | Medium |
| DeleteUser | Delete SASL user | Low | High |
| ListUsers | List SASL users | Low | None |
| CreateACL | Create ACL rule | Medium | Medium |
| DeleteACL | Delete ACL rule | Medium | Medium |
| ListACLs | List ACL rules | Low | None |
| DescribeConsumerLag | Monitor lag metrics | Low | None |
| ScaleInstance | Scale instance specs | High | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with instance, topic, ACL, and user management |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateInstance — Create a Kafka Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` is valid | Region supported | Suggest valid region |
| VPC/Subnet | `ve vpc DescribeVpcs` | VPC exists | HALT; create VPC first |
| Name unique | Instance name not in use | No conflict | Use different name |
| Quota | `ve kafka ListInstances` | Under quota | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create a Kafka instance
ve kafka CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --InstanceType "kafka.n1.x2.small" \
  --StorageSpace 300 \
  --PartitionQuota 1000 \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ZoneId "{{user.zone_id}}"

# Create with specific version
ve kafka CreateInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceName "{{user.instance_name}}" \
  --InstanceType "kafka.n1.x2.small" \
  --StorageSpace 300 \
  --PartitionQuota 1000 \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ZoneId "{{user.zone_id}}" \
  --Version "2.6"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/base"
    "github.com/volcengine/volc-sdk-golang/service/kafka"
)

func main() {
    instance := kafka.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":         os.Getenv("VOLCENGINE_REGION"),
        "InstanceName":   os.Getenv("INSTANCE_NAME"),
        "InstanceType":   "kafka.n1.x2.small",
        "StorageSpace":   300,
        "PartitionQuota": 1000,
        "VpcId":          os.Getenv("VPC_ID"),
        "SubnetId":       os.Getenv("SUBNET_ID"),
        "ZoneId":         os.Getenv("ZONE_ID"),
        "Version":        "2.6",
    }

    resp, err := instance.Client.Request("kafka", "CreateInstance", params)
    if err != nil {
        log.Fatalf("CreateInstance failed: %v", err)
    }

    fmt.Println(string(resp))
}
```

#### Validation

```bash
# Poll until instance is Running
for i in $(seq 1 60); do
  STATUS=$(ve kafka DescribeInstance --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq -r '.Result.Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Running" ] && break
  sleep 30
done

# Verify broker endpoints
ve kafka DescribeInstance --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{output.instance_id}}" | jq '.Result.BrokerList'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidVpc.NotFound` | HALT; VPC does not exist — create VPC first via `ve-vpc-ops` |
| `InvalidSubnet.NotFound` | HALT; Subnet does not exist — create subnet first |
| `InvalidParameter` | HALT; fix parameter values per error message |
| `QuotaExceeded` | HALT; instance quota exceeded — request quota increase |
| `InsufficientBalance` | HALT; account balance insufficient — recharge |
| `InvalidZone.NotFound` | HALT; zone not available — select different zone |

---

### Operation: DeleteInstance — Delete a Kafka Instance

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of instance `{{user.instance_id}}` ({{user.instance_name}})
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all topics, messages, and configurations will be permanently deleted
- **MUST** verify instance status is `Running` or `Stopped`

```bash
# Verify instance exists and get details
ve kafka DescribeInstance --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"

# List topics to warn about data loss
echo "WARNING: The following topics and all their data will be permanently deleted:"
ve kafka ListTopics --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | jq -r '.Result.Topics[].TopicName'
```

#### Execution

```bash
# Delete the instance
ve kafka DeleteInstance \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"
```

#### Validation

```bash
# Poll until instance is deleted (Describe returns NotFound)
for i in $(seq 1 60); do
  if ve kafka DescribeInstance --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" 2>&1 | grep -q "InvalidInstance.NotFound"; then
    echo "Instance deleted successfully"
    break
  fi
  echo "Waiting for deletion... (attempt $i/60)"
  sleep 30
done
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | Instance already deleted; skip |
| `InvalidInstanceStatus` | HALT; instance is being modified — wait and retry |
| `DependencyViolation` | HALT; resources still attached — remove dependencies first |
| `Unauthorized` | HALT; check IAM permissions |

---

### Operation: CreateTopic — Create a Topic

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve kafka DescribeInstance` | Instance Running | HALT |
| Topic name unique | `ve kafka ListTopics` | Name not in list | Use different name |
| Partition quota | Instance partition quota | Under quota | HALT; scale instance |
| Replication valid | `replication_factor <= available_brokers` | Valid | Reduce replication |

#### Execution

```bash
# Create a topic
ve kafka CreateTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}" \
  --PartitionNumber {{user.partition_count}} \
  --ReplicaNumber {{user.replication_factor}}

# Create with configuration
ve kafka CreateTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}" \
  --PartitionNumber 6 \
  --ReplicaNumber 3 \
  --MinInsyncReplicas 2 \
  --RetentionHours 168 \
  --MaxMessageSize 1048576
```

#### Validation

```bash
# Verify topic exists
ve kafka DescribeTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}"

# Verify partition distribution
ve kafka DescribeTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}" | jq '.Result.PartitionDetails'
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `TopicAlreadyExists` | HALT; topic exists — use different name or delete first |
| `QuotaExceeded.Partition` | HALT; partition quota exceeded — scale instance |
| `InvalidParameter` | HALT; check partition count and replication factor limits |

---

### Operation: DeleteTopic — Delete a Topic

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of topic `{{user.topic_name}}`
- **MUST NOT** proceed without clear user assent
- **MUST** warn about data loss: all messages in the topic will be permanently deleted
- **MUST** check if consumers are active on this topic

```bash
# Check for active consumer groups
echo "Checking for active consumers on topic {{user.topic_name}}..."
ve kafka ListGroups --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}"

# Warn user
echo "WARNING: Deleting topic '{{user.topic_name}}' will permanently delete all messages."
echo "Active consumers may fail if this topic is deleted."
```

#### Execution

```bash
# Delete the topic
ve kafka DeleteTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}"
```

#### Validation

```bash
# Verify topic is deleted (should return TopicNotFound)
if ve kafka DescribeTopic --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" --TopicName "{{user.topic_name}}" 2>&1 | grep -q "TopicNotFound"; then
  echo "Topic deleted successfully"
else
  echo "Topic may still exist — checking..."
  ve kafka ListTopics --Region "{{env.VOLCENGINE_REGION}}" --InstanceId "{{user.instance_id}}" | grep "{{user.topic_name}}" || echo "Topic deleted"
fi
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `TopicNotFound` | Topic already deleted; skip |
| `InvalidInstance.NotFound` | HALT; instance does not exist |
| `DeleteTopicPartial` | Some partitions failed — investigate broker health |

---

### Operation: CreatePartitions — Add Partitions to Topic

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Topic exists | `ve kafka DescribeTopic` | Topic found | HALT |
| Instance has quota | Partition quota check | Sufficient quota | HALT; scale instance |
| Partition increase valid | New count > current | Valid increase | HALT; cannot reduce |

#### Execution

```bash
# Add partitions
ve kafka CreatePartitions \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}" \
  --PartitionNumber {{user.new_partition_count}}
```

#### Validation

```bash
# Verify new partition count
ve kafka DescribeTopic \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --TopicName "{{user.topic_name}}" | jq '.Result.PartitionNum'
```

---

### Operation: CreateUser — Create SASL User

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve kafka DescribeInstance` | Running | HALT |
| Username unique | `ve kafka ListUsers` | Username not in list | Use different name |
| Password valid | Length and complexity | Min 8 chars | Require stronger password |

#### Execution

```bash
# Create SASL user
ve kafka CreateUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --UserName "{{user.username}}" \
  --Password "{{user.password}}"

# Create with specific mechanism
ve kafka CreateUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --UserName "{{user.username}}" \
  --Password "{{user.password}}" \
  --Mechanism SCRAM-SHA-512
```

> **Security:** Password is masked in all output as `<masked>`.

#### Validation

```bash
# Verify user was created
ve kafka ListUsers \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | jq '.Result.Users[].UserName'

# Check user details
ve kafka DescribeUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --UserName "{{user.username}}"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `UserAlreadyExists` | HALT; user exists — use different name or delete first |
| `InvalidPassword` | HALT; password does not meet complexity requirements |
| `QuotaExceeded.User` | HALT; user quota exceeded |

---

### Operation: DeleteUser — Delete SASL User

#### Pre-flight (Safety Gate)

- **MUST** confirm: delete SASL user `{{user.username}}`
- **MUST** warn: connected clients using this user will be disconnected
- **MUST NOT** delete the last admin user without alternative authentication

#### Execution

```bash
# Delete the user
ve kafka DeleteUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --UserName "{{user.username}}"
```

#### Validation

```bash
# Verify user is deleted
ve kafka ListUsers \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" | grep "{{user.username}}" || echo "User deleted"
```

---

### Operation: CreateACL — Create ACL Rule

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Instance exists | `ve kafka DescribeInstance` | Running | HALT |
| User exists | `ve kafka DescribeUser` | User found | HALT; create user first |
| Resource valid | Topic/Group exists | Resource found | HALT |

#### Execution

```bash
# Allow user to produce to a topic
ve kafka CreateACL \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ResourceType Topic \
  --ResourceName "{{user.topic_name}}" \
  --UserName "{{user.username}}" \
  --PermissionType Allow \
  --Operation Write

# Allow user to consume from a topic
ve kafka CreateACL \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ResourceType Topic \
  --ResourceName "{{user.topic_name}}" \
  --UserName "{{user.username}}" \
  --PermissionType Allow \
  --Operation Read

# Allow user to join a consumer group
ve kafka CreateACL \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ResourceType Group \
  --ResourceName "{{user.consumer_group_id}}" \
  --UserName "{{user.username}}" \
  --PermissionType Allow \
  --Operation Read
```

#### Validation

```bash
# List ACLs to verify
ve kafka ListACLs \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --ResourceType Topic \
  --ResourceName "{{user.topic_name}}"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidUser.NotFound` | HALT; user does not exist |
| `InvalidResource.NotFound` | HALT; topic/group does not exist |
| `ACLAlreadyExists` | HALT; ACL rule already exists |

---

### Operation: DescribeConsumerLag — Monitor Consumer Lag

#### Execution

```bash
# List consumer groups
ve kafka ListGroups \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}"

# Describe specific consumer group with lag info
ve kafka DescribeGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}"

# Get detailed lag per partition
ve kafka DescribeConsumerLag \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}"
```

#### Lag Analysis

| Lag Level | Threshold | Recommendation |
|-----------|-----------|----------------|
| Healthy | < 1000 | Normal operation |
| Warning | 1000 - 10000 | Monitor; may need consumer scaling |
| Critical | > 10000 | Scale consumers; check for errors |

---

### Operation: ResetOffset — Reset Consumer Offset

#### Pre-flight (Safety Gate)

- **MUST** confirm: reset offset for consumer group `{{user.consumer_group_id}}`
- **MUST** warn: will cause message reprocessing or skipping
- **MUST** verify consumer group has no active members

```bash
# Check for active members
ve kafka DescribeGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}" | jq '.Result.Members'
```

#### Execution

```bash
# Reset to earliest offset
ve kafka ResetOffset \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}" \
  --TopicName "{{user.topic_name}}" \
  --ResetType earliest

# Reset to latest offset
ve kafka ResetOffset \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}" \
  --TopicName "{{user.topic_name}}" \
  --ResetType latest

# Reset to specific timestamp
ve kafka ResetOffset \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}" \
  --TopicName "{{user.topic_name}}" \
  --ResetType timestamp \
  --Timestamp 1716777600
```

#### Validation

```bash
# Verify offset was reset
ve kafka DescribeConsumerLag \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --InstanceId "{{user.instance_id}}" \
  --GroupId "{{user.consumer_group_id}}"
```

---

## Error Taxonomy (≥ 10 Codes)

| Error Code | HTTP | Meaning | Max Retries | Backoff | Agent Action | UX Template |
|------------|------|---------|-------------|---------|--------------|-------------|
| `InvalidParameter` | 400 | Request parameter invalid | 0 | — | Fix parameter and retry | `[ERROR] InvalidParameter: Request parameter is invalid. Check the parameter against API docs.` |
| `InvalidInstance.NotFound` | 404 | Instance does not exist | 0 | — | HALT; verify instance ID | `[ERROR] Instance not found. Verify the instance ID.` |
| `InvalidVpc.NotFound` | 400 | VPC does not exist | 0 | — | HALT; create VPC first | `[ERROR] VPC not found. Create VPC via ve-vpc-ops first.` |
| `InvalidSubnet.NotFound` | 400 | Subnet does not exist | 0 | — | HALT; create subnet first | `[ERROR] Subnet not found. Create subnet first.` |
| `TopicAlreadyExists` | 400 | Topic already exists | 0 | — | Use different name or delete first | `[ERROR] Topic already exists. Use a different name.` |
| `TopicNotFound` | 404 | Topic does not exist | 0 | — | Verify topic name or create topic | `[ERROR] Topic not found. Verify the topic name.` |
| `UserAlreadyExists` | 400 | SASL user already exists | 0 | — | Use different username | `[ERROR] User already exists. Use a different username.` |
| `UserNotFound` | 404 | SASL user does not exist | 0 | — | Verify username | `[ERROR] User not found. Verify the username.` |
| `QuotaExceeded` | 400 | Resource quota exceeded | 0 | — | HALT; request quota increase | `[ERROR] Quota exceeded. Request quota increase from support.` |
| `InsufficientBalance` | 400 | Account balance insufficient | 0 | — | HALT; recharge account | `[ERROR] Insufficient balance. Recharge your account.` |
| `Unauthorized` | 403 | IAM permission denied | 0 | — | HALT; check IAM policies | `[ERROR] Unauthorized. Check IAM permissions.` |
| `InvalidInstanceStatus` | 400 | Instance status not valid for operation | 3 | 10s | Wait and retry | `[WARNING] Instance busy. Retrying...` |
| `InternalError` | 500 | Server-side error | 3 | 2s, 4s, 8s | Retry with backoff | `[ERROR] Internal error. Retrying...` |
| `Throttling` | 429 | Rate limit exceeded | 3 | 1s, 2s, 4s | Back off and retry | `[WARNING] Rate limit. Retrying in {backoff}s...` |
| `ACLAlreadyExists` | 400 | ACL rule already exists | 0 | — | Skip or delete first | `[ERROR] ACL already exists.` |

## Prerequisites

1. **Install `ve` CLI**:

   ```bash
   # Download from GitHub releases
   curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
   chmod +x /usr/local/bin/ve
   ve version
   ```

2. **Configure Credentials**:

   ```bash
   export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
   export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
   export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
   ```

3. **Verify Configuration**:

   ```bash
   ve kafka ListInstances --Region {{env.VOLCENGINE_REGION}}
   ```

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring & Alerts](references/monitoring.md)
- [Integration](references/integration.md)

## Operational Best Practices

- **Instance naming:** use descriptive names with environment prefix (e.g., `prod-events`, `dev-logs`)
- **Topic naming:** use hierarchical naming (e.g., `service.event.v1`)
- **Partition count:** start with 3-6 partitions, scale based on throughput needs
- **Replication factor:** use 3 for production (tolerates 2 broker failures)
- **SASL authentication:** enable for production instances; rotate passwords regularly
- **ACLs:** follow least privilege principle; grant only necessary permissions
- **Consumer groups:** use descriptive names; monitor lag continuously
- **Retention:** configure appropriate retention hours based on data requirements
- **Scaling:** scale partitions before increasing broker specs
- **Monitoring:** set up alerts for consumer lag > 10000 and broker disk usage > 80%
