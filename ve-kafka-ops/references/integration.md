# Kafka Integration

## Environment Setup

### Primary Path: ve CLI

The `ve` CLI is a static Go binary with no runtime dependencies.

**Installation:**

```bash
# Download from GitHub releases
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

curl -fsSL "https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-${OS}-${ARCH}" -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify
ve version
```

**Credential Configuration:**

```bash
# Environment variables (recommended for agents)
export VOLCENGINE_ACCESS_KEY="${VOLCENGINE_ACCESS_KEY}"
export VOLCENGINE_SECRET_KEY="${VOLCENGINE_SECRET_KEY}"  # Never display in output
export VOLCENGINE_REGION="cn-beijing"

# Verify
test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "Credentials configured"
```

### Fallback Path: JIT Go SDK

When `ve` CLI does not support a specific operation, dynamically build a Go SDK script.

**Go Runtime Bootstrap:**

```bash
# Check if Go exists
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/{{env.GO_VERSION}}.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
    export GOMODCACHE="/tmp/go-modcache"
    export GOPROXY="https://goproxy.cn,direct"
fi

go version
```

**JIT Go SDK Workflow:**

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-kafka-workspace
   cd /tmp/ve-kafka-workspace
   go mod init kafka-script
   ```

2. **Get dependencies:**
   ```bash
   export GOPROXY="https://goproxy.cn,direct"
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate and execute script:**
   ```bash
   cat > main.go << 'EOF'
   package main
   
   import (
       "context"
       "fmt"
       "log"
       "os"
       
       "github.com/volcengine/volc-sdk-golang/service/kafka"
   )
   
   func main() {
       instance := kafka.NewInstance()
       instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
       instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
       
       params := map[string]interface{}{
           "Region": os.Getenv("VOLCENGINE_REGION"),
       }
       
       resp, err := instance.Client.Request("kafka", "ListInstances", params)
       if err != nil {
           log.Fatalf("Request failed: %v", err)
       }
       
       fmt.Println(string(resp))
   }
   EOF
   
   go run ./main.go
   ```

## SDK Package Structure

| Product | Go SDK Package |
|---------|---------------|
| Kafka | `github.com/volcengine/volc-sdk-golang/service/kafka` |

## Common SDK Operations

### Initialize Client

```go
package main

import (
    "os"
    "github.com/volcengine/volc-sdk-golang/service/kafka"
)

func main() {
    instance := kafka.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    // instance.Client.SetHost("https://kafka.cn-beijing.volces.com")
}
```

### Create Instance

```go
params := map[string]interface{}{
    "Region":         "cn-beijing",
    "InstanceName":   "prod-kafka",
    "InstanceType":   "kafka.n1.x2.small",
    "StorageSpace":   300,
    "PartitionQuota": 1000,
    "VpcId":          "vpc-xxx",
    "SubnetId":       "subnet-xxx",
    "ZoneId":         "cn-beijing-a",
    "Version":        "2.6",
}

resp, err := instance.Client.Request("kafka", "CreateInstance", params)
if err != nil {
    log.Fatalf("CreateInstance failed: %v", err)
}
```

### Create Topic

```go
params := map[string]interface{}{
    "Region":          "cn-beijing",
    "InstanceId":      "kafka-xxx",
    "TopicName":       "orders",
    "PartitionNumber": 6,
    "ReplicaNumber":   3,
    "MinInsyncReplicas": 2,
    "RetentionHours":  168,
    "MaxMessageSize":  1048576,
}

resp, err := instance.Client.Request("kafka", "CreateTopic", params)
```

### Create User

```go
params := map[string]interface{}{
    "Region":     "cn-beijing",
    "InstanceId": "kafka-xxx",
    "UserName":   "app-producer",
    "Password":   os.Getenv("KAFKA_USER_PASSWORD"), // Never hardcode
    "Mechanism":  "SCRAM-SHA-512",
}

resp, err := instance.Client.Request("kafka", "CreateUser", params)
```

### Create ACL

```go
params := map[string]interface{}{
    "Region":         "cn-beijing",
    "InstanceId":     "kafka-xxx",
    "ResourceType":   "Topic",
    "ResourceName":   "orders",
    "UserName":       "app-producer",
    "PermissionType": "Allow",
    "Operation":      "Write",
    "Host":           "*",
}

resp, err := instance.Client.Request("kafka", "CreateACL", params)
```

## Kafka Client Connection

### Connection Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| Bootstrap Servers | Broker endpoints | `kafka-xxx.cn-beijing.kafka.volces.com:9092` |
| Security Protocol | PLAINTEXT or SASL_SSL | `SASL_SSL` |
| SASL Mechanism | SCRAM-SHA-256/512 | `SCRAM-SHA-512` |
| SASL Username | SASL user | `app-producer` |
| SASL Password | SASL password | `<masked>` |

### Python Client Example (kafka-python)

```python
from kafka import KafkaProducer, KafkaConsumer

# SASL configuration
producer = KafkaProducer(
    bootstrap_servers=['kafka-xxx.cn-beijing.kafka.volces.com:9093'],
    security_protocol='SASL_SSL',
    sasl_mechanism='SCRAM-SHA-512',
    sasl_plain_username='app-producer',
    sasl_plain_password='<masked>',  # Never hardcode
    value_serializer=lambda v: v.encode('utf-8')
)

# Send message
producer.send('orders', b'order-data')
producer.flush()

# Consumer
consumer = KafkaConsumer(
    'orders',
    bootstrap_servers=['kafka-xxx.cn-beijing.kafka.volces.com:9093'],
    security_protocol='SASL_SSL',
    sasl_mechanism='SCRAM-SHA-512',
    sasl_plain_username='app-consumer',
    sasl_plain_password='<masked>',
    group_id='order-processors',
    auto_offset_reset='earliest'
)

for msg in consumer:
    print(f"Received: {msg.value}")
```

### Java Client Example

```java
Properties props = new Properties();
props.put("bootstrap.servers", "kafka-xxx.cn-beijing.kafka.volces.com:9093");
props.put("security.protocol", "SASL_SSL");
props.put("sasl.mechanism", "SCRAM-SHA-512");
props.put("sasl.jaas.config", 
    "org.apache.kafka.common.security.scram.ScramLoginModule required " +
    "username=\"app-producer\" " +
    "password=\"<masked>\";");

// Producer
props.put("key.serializer", "org.apache.kafka.common.serialization.StringSerializer");
props.put("value.serializer", "org.apache.kafka.common.serialization.StringSerializer");
Producer<String, String> producer = new KafkaProducer<>(props);

producer.send(new ProducerRecord<>("orders", "key", "value"));
producer.close();
```

### Go Client Example (sarama)

```go
package main

import (
    "crypto/tls"
    "github.com/IBM/sarama"
)

func main() {
    config := sarama.NewConfig()
    config.Net.SASL.Enable = true
    config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
    config.Net.SASL.User = "app-producer"
    config.Net.SASL.Password = "<masked>"
    config.Net.TLS.Enable = true
    config.Net.TLS.Config = &tls.Config{InsecureSkipVerify: false}
    
    producer, err := sarama.NewSyncProducer(
        []string{"kafka-xxx.cn-beijing.kafka.volces.com:9093"},
        config,
    )
    if err != nil {
        panic(err)
    }
    defer producer.Close()
    
    msg := &sarama.ProducerMessage{
        Topic: "orders",
        Value: sarama.StringEncoder("order-data"),
    }
    
    _, _, err = producer.SendMessage(msg)
}
```

## VPC Integration

### Network Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Your VPC                            │
│  ┌───────────────────────────────────────────────────┐  │
│  │                 Kafka Subnet                       │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │            Kafka Instance                    │  │  │
│  │  │  ┌─────────┐  ┌─────────┐  ┌─────────┐      │  │  │
│  │  │  │ Broker 1│  │ Broker 2│  │ Broker 3│      │  │  │
│  │  │  │ :9092   │  │ :9092   │  │ :9092   │      │  │  │
│  │  │  └─────────┘  └─────────┘  └─────────┘      │  │  │
│  │  └─────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Application Subnet                    │  │
│  │     (Producers, Consumers, Admin Tools)           │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Security Group Rules

| Direction | Protocol | Port | Source | Description |
|-----------|----------|------|--------|-------------|
| Inbound | TCP | 9092 | Application subnet | Kafka plaintext |
| Inbound | TCP | 9093 | Application subnet | Kafka SASL/SSL |
| Inbound | TCP | 2181 | Internal | ZooKeeper (internal) |
| Outbound | All | All | 0.0.0.0/0 | Allow all outbound |

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Kafka Topic Management

on:
  push:
    paths:
      - 'kafka-topics/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install ve CLI
        run: |
          curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
          chmod +x /usr/local/bin/ve
      
      - name: Configure credentials
        env:
          VOLCENGINE_ACCESS_KEY: ${{ secrets.VOLCENGINE_ACCESS_KEY }}
          VOLCENGINE_SECRET_KEY: ${{ secrets.VOLCENGINE_SECRET_KEY }}
        run: |
          echo "VOLCENGINE_ACCESS_KEY=${VOLCENGINE_ACCESS_KEY}" >> $GITHUB_ENV
          echo "VOLCENGINE_SECRET_KEY=${VOLCENGINE_SECRET_KEY}" >> $GITHUB_ENV
          echo "VOLCENGINE_REGION=cn-beijing" >> $GITHUB_ENV
      
      - name: Create topics
        run: |
          for topic_file in kafka-topics/*.json; do
            topic_name=$(jq -r '.name' "$topic_file")
            partitions=$(jq -r '.partitions' "$topic_file")
            replicas=$(jq -r '.replicas' "$topic_file")
            
            ve kafka CreateTopic \
              --Region ${{ env.VOLCENGINE_REGION }} \
              --InstanceId ${{ secrets.KAFKA_INSTANCE_ID }} \
              --TopicName "$topic_name" \
              --PartitionNumber "$partitions" \
              --ReplicaNumber "$replicas"
          done
```

## Terraform Integration

```hcl
# Note: This is conceptual - Volcengine provider syntax may vary

resource "volcengine_kafka_instance" "main" {
  instance_name   = "prod-kafka"
  instance_type   = "kafka.n1.x2.small"
  storage_space   = 300
  partition_quota = 1000
  vpc_id          = volcengine_vpc.main.id
  subnet_id       = volcengine_subnet.main.id
  zone_id         = "cn-beijing-a"
  version         = "2.6"
}

resource "volcengine_kafka_topic" "orders" {
  instance_id       = volcengine_kafka_instance.main.id
  topic_name        = "orders"
  partition_number  = 6
  replica_number    = 3
  retention_hours   = 168
}

resource "volcengine_kafka_user" "producer" {
  instance_id = volcengine_kafka_instance.main.id
  user_name   = "app-producer"
  password    = var.kafka_producer_password
  mechanism   = "SCRAM-SHA-512"
}

resource "volcengine_kafka_acl" "producer_write" {
  instance_id     = volcengine_kafka_instance.main.id
  resource_type   = "Topic"
  resource_name   = volcengine_kafka_topic.orders.topic_name
  user_name       = volcengine_kafka_user.producer.user_name
  permission_type = "Allow"
  operation       = "Write"
}
```
