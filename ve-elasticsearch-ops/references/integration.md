# Elasticsearch Integration

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
   mkdir -p /tmp/ve-es-workspace
   cd /tmp/ve-es-workspace
   go mod init es-script
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
       "fmt"
       "os"

       "github.com/volcengine/volc-sdk-golang/service/elasticsearch"
   )

   func main() {
       instance := elasticsearch.NewInstance()
       instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
       instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

       params := map[string]interface{}{
           "Region": os.Getenv("VOLCENGINE_REGION"),
       }

       resp, err := instance.Client.Request("elasticsearch", "ListInstances", params)
       if err != nil {
           panic(err)
       }

       fmt.Println(string(resp))
   }
   EOF

   go run ./main.go
   ```

## SDK Package Structure

| Product | Go SDK Package |
|---------|---------------|
| Elasticsearch | `github.com/volcengine/volc-sdk-golang/service/elasticsearch` |

## Common SDK Operations

### Initialize Client

```go
package main

import (
    "os"
    "github.com/volcengine/volc-sdk-golang/service/elasticsearch"
)

func main() {
    instance := elasticsearch.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
}
```

### Create Instance

```go
params := map[string]interface{}{
    "Region":       "cn-beijing",
    "InstanceName": "prod-search",
    "Version":      "7.16",
    "NodeSpec":     "es.x4.medium",
    "NodeNumber":   3,
    "StorageSpaceGb": 100,
    "VpcId":        "vpc-xxx",
    "SubnetId":     "subnet-xxx",
    "ChargeType":   "PostPaid",
}

resp, err := instance.Client.Request("elasticsearch", "CreateInstance", params)
if err != nil {
    panic(err)
}
```

### Create Index

```go
params := map[string]interface{}{
    "Region":      "cn-beijing",
    "InstanceId":  "es-xxx",
    "IndexName":   "logs-2024.05.27",
    "ShardCount":  5,
    "ReplicaCount": 1,
}

resp, err := instance.Client.Request("elasticsearch", "CreateIndex", params)
```

### Create Snapshot

```go
params := map[string]interface{}{
    "Region":       "cn-beijing",
    "InstanceId":   "es-xxx",
    "SnapshotName": "daily-backup",
    "Indices":      "*",
}

resp, err := instance.Client.Request("elasticsearch", "CreateSnapshot", params)
```

### Enable Kibana

```go
params := map[string]interface{}{
    "Region":        "cn-beijing",
    "InstanceId":    "es-xxx",
    "KibanaUser":    "admin",
    "KibanaPassword": os.Getenv("KIBANA_PASSWORD"), // Never hardcode
}

resp, err := instance.Client.Request("elasticsearch", "EnableKibana", params)
```

### Install Plugin

```go
params := map[string]interface{}{
    "Region":       "cn-beijing",
    "InstanceId":   "es-xxx",
    "PluginName":   "analysis-ik",
    "ForceRestart": true,
}

resp, err := instance.Client.Request("elasticsearch", "InstallPlugin", params)
```

## Elasticsearch Client Connection

### REST API Access

Elasticsearch instances expose a REST API endpoint for direct access:

```bash
# ES REST API endpoint
# https://es-{instance-id}.{region}.es.volces.com:9200

# Query cluster health
curl -u "username:password" \
  "https://es-xxx.cn-beijing.es.volces.com:9200/_cluster/health?pretty"

# List indices
curl -u "username:password" \
  "https://es-xxx.cn-beijing.es.volces.com:9200/_cat/indices?v"

# Search documents
curl -u "username:password" \
  -X GET "https://es-xxx.cn-beijing.es.volces.com:9200/my-index/_search?q=field:value"
```

### Python Client Example

```python
from elasticsearch import Elasticsearch

# Connect to ES instance
# Use credentials from Kibana user setup
es = Elasticsearch(
    ['https://es-xxx.cn-beijing.es.volces.com:9200'],
    http_auth=('admin', '<masked>'),  # Never hardcode
    verify_certs=True
)

# Index a document
doc = {
    'title': 'Elasticsearch Guide',
    'content': 'This is a sample document',
    'timestamp': '2024-05-27T10:00:00'
}
resp = es.index(index='my-index', id=1, document=doc)
print(resp['result'])

# Search
resp = es.search(index='my-index', query={"match": {"title": "guide"}})
for hit in resp['hits']['hits']:
    print(hit['_source'])
```

### Java Client Example

```java
import org.elasticsearch.client.RestClient;
import org.elasticsearch.client.RestHighLevelClient;
import org.apache.http.HttpHost;
import org.apache.http.auth.AuthScope;
import org.apache.http.auth.UsernamePasswordCredentials;
import org.apache.http.client.CredentialsProvider;
import org.apache.http.impl.client.BasicCredentialsProvider;

CredentialsProvider credentialsProvider = new BasicCredentialsProvider();
credentialsProvider.setCredentials(
    AuthScope.ANY,
    new UsernamePasswordCredentials("admin", "<masked>")
);

RestHighLevelClient client = new RestHighLevelClient(
    RestClient.builder(
        new HttpHost("es-xxx.cn-beijing.es.volces.com", 9200, "https")
    ).setHttpClientConfigCallback(
        httpClientBuilder -> httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider)
    )
);

// Use client for search/index operations
// client.index(...)
// client.search(...)
```

### Go Client Example (elastic/go-elasticsearch)

```go
package main

import (
    "crypto/tls"
    "net/http"
    "github.com/elastic/go-elasticsearch/v8"
)

func main() {
    cfg := elasticsearch.Config{
        Addresses: []string{
            "https://es-xxx.cn-beijing.es.volces.com:9200",
        },
        Username: "admin",
        Password: "<masked>",
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: false,
            },
        },
    }

    es, _ := elasticsearch.NewClient(cfg)
    
    // Ping cluster
    res, _ := es.Ping()
    defer res.Body.Close()
}
```

## VPC Integration

### Network Architecture

```
+----------------------------------------------------------+
|                      Your VPC                              |
|  +----------------------------------------------------+  |
|  |                ES Subnet                             |  |
|  |  +----------------------------------------------+  |  |
|  |  |         ES Instance (Managed Cluster)         |  |  |
|  |  |  +----------+  +----------+  +----------+    |  |  |
|  |  |  | Data N1  |  | Data N2  |  | Data N3  |    |  |  |
|  |  |  | :9200    |  | :9200    |  | :9200    |    |  |  |
|  |  |  +----------+  +----------+  +----------+    |  |  |
|  |  +----------------------------------------------+  |  |
|  +----------------------------------------------------+  |
|                                                           |
|  +----------------------------------------------------+  |
|  |              Application Subnet                     |  |
|  |     (Elasticsearch Clients, Logstash, Beats)       |  |
|  +----------------------------------------------------+  |
+----------------------------------------------------------+
```

### Security Group Rules

| Direction | Protocol | Port | Source | Description |
|-----------|----------|------|--------|-------------|
| Inbound | TCP | 9200 | Application subnet | ES REST API |
| Inbound | TCP | 9300 | Internal | ES node transport |
| Inbound | TCP | 5601 | Admin IPs | Kibana UI |
| Outbound | All | All | 0.0.0.0/0 | Allow all outbound |

## CI/CD Integration

### GitHub Actions Example

```yaml
name: ES Index Management

on:
  push:
    paths:
      - 'es-indices/**'

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

      - name: Create indices
        run: |
          for index_file in es-indices/*.json; do
            index_name=$(jq -r '.name' "$index_file")
            shards=$(jq -r '.shards' "$index_file")
            replicas=$(jq -r '.replicas' "$index_file")

            ve elasticsearch CreateIndex \
              --Region ${{ env.VOLCENGINE_REGION }} \
              --InstanceId ${{ secrets.ES_INSTANCE_ID }} \
              --IndexName "$index_name" \
              --ShardCount "$shards" \
              --ReplicaCount "$replicas"
          done

      - name: Create snapshot
        run: |
          ve elasticsearch CreateSnapshot \
            --Region ${{ env.VOLCENGINE_REGION }} \
            --InstanceId ${{ secrets.ES_INSTANCE_ID }} \
            --SnapshotName "ci-$(date +%Y%m%d%H%M%S)" \
            --Indices "*"
```

## Terraform Integration

```hcl
# Note: This is conceptual — Volcengine provider syntax may vary

resource "volcengine_elasticsearch_instance" "main" {
  instance_name  = "prod-search"
  version        = "7.16"
  node_spec      = "es.x4.medium"
  node_number    = 3
  storage_space_gb = 100
  storage_type   = "ESSD"
  vpc_id         = volcengine_vpc.main.id
  subnet_id      = volcengine_subnet.main.id
  charge_type    = "PostPaid"
}

resource "volcengine_elasticsearch_index" "logs" {
  instance_id  = volcengine_elasticsearch_instance.main.id
  index_name   = "logs"
  shard_count  = 5
  replica_count = 1
}

resource "volcengine_elasticsearch_snapshot" "daily" {
  instance_id   = volcengine_elasticsearch_instance.main.id
  snapshot_name = "daily-backup"
  indices       = "*"
}
```

## FinOps Considerations

| Billing Model | Use Case | Cost Savings |
|---------------|----------|--------------|
| **PostPaid** | Development, variable workloads | Flexible, no upfront |
| **PrePaid** | Stable production workloads | ~15-30% savings |
| **Reserved** | Long-term steady state | ~30-50% savings |

### Cost Optimization Tips

1. **Right-size nodes:** Monitor CPU/Memory/Disk usage and adjust specs accordingly
2. **Use PrePaid for predictable workloads:** Stable production workloads benefit from subscription pricing
3. **Delete unused indices:** Old or orphaned indices consume storage and memory
4. **Optimize shard count:** Too many small shards waste resources; target 10-50 GB per shard
5. **Manage snapshot retention:** Delete old snapshots to reduce storage costs
6. **Monitor storage growth:** Set up alerts for unexpected storage increases
