# Integration

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Go Runtime Bootstrap

If Agent Runtime lacks Go, JIT download from official source:

```bash
# Check Go runtime
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
fi

go version
```

**Go version strategy:**
- **JIT download:** Go 1.21+ (stable)
- **Script compatibility:** Go 1.14+ (minimum)

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-sdk-workspace
   cd /tmp/ve-sdk-workspace
   go mod init ve-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   # Set proxy for China CDN mirror (faster download)
   export GOPROXY="https://goproxy.cn,direct"
   
   # Volcengine SDK
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate script** (Agent dynamically creates operation-specific .go file)

4. **Execute:**
   ```bash
   go run ./main.go
   ```

### SDK Package Structure

| Product | Go SDK Package |
|---------|---------------|
| MongoDB | `github.com/volcengine/volc-sdk-golang/service/mongodb` |

Find package names at: https://github.com/volcengine/volc-sdk-golang/tree/main/service

## SDK Script Templates

### Basic Template

```go
// main.go (single-file script template)
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

func main() {
    // Initialize service instance
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    
    // Make API call
    resp, err := instance.Client.Request("mongodb", "DescribeDBInstances", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

### Create Instance Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

func main() {
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":         os.Getenv("VOLCENGINE_REGION"),
        "InstanceName":   os.Getenv("INSTANCE_NAME"),
        "MongoVersion":   os.Getenv("MONGO_VERSION"),
        "NodeSpec":       os.Getenv("NODE_SPEC"),
        "StorageSpaceGB": os.Getenv("STORAGE_GB"),
        "NodeNumber":     3,
        "VpcId":          os.Getenv("VPC_ID"),
        "SubnetId":       os.Getenv("SUBNET_ID"),
        "ChargeType":     "PostPaid",
    }

    resp, err := instance.Client.Request("mongodb", "CreateDBInstance", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create instance: %v\n", err)
        os.Exit(1)
    }

    var result struct {
        Result struct {
            InstanceId string `json:"InstanceId"`
        } `json:"Result"`
    }
    
    if err := json.Unmarshal(resp, &result); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Instance created: %s\n", result.Result.InstanceId)
}
```

### Describe Instances Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

type Instance struct {
    InstanceId       string `json:"InstanceId"`
    InstanceName     string `json:"InstanceName"`
    InstanceStatus   string `json:"InstanceStatus"`
    MongoVersion     string `json:"MongoVersion"`
    NodeSpec         string `json:"NodeSpec"`
    StorageSpaceGB   int    `json:"StorageSpaceGB"`
    ConnectionString string `json:"ConnectionString"`
    Port             int    `json:"Port"`
}

type Response struct {
    Result struct {
        Instances []Instance `json:"Instances"`
        Total     int        `json:"Total"`
    } `json:"Result"`
}

func main() {
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":     os.Getenv("VOLCENGINE_REGION"),
        "PageNumber": 1,
        "PageSize":   100,
    }

    resp, err := instance.Client.Request("mongodb", "DescribeDBInstances", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to list instances: %v\n", err)
        os.Exit(1)
    }

    var result Response
    if err := json.Unmarshal(resp, &result); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Found %d instances:\n", result.Result.Total)
    for _, inst := range result.Result.Instances {
        fmt.Printf("  - %s (%s): %s [%s]\n", 
            inst.InstanceId, 
            inst.InstanceName, 
            inst.InstanceStatus,
            inst.MongoVersion)
    }
}
```

### User Management Example

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

func main() {
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Create account
    params := map[string]interface{}{
        "InstanceId":       os.Getenv("INSTANCE_ID"),
        "AccountName":      os.Getenv("ACCOUNT_NAME"),
        "AccountPassword":  os.Getenv("ACCOUNT_PASSWORD"),
        "AccountPrivilege": "ReadWrite",
    }

    resp, err := instance.Client.Request("mongodb", "CreateDBAccount", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create account: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

### Backup Management Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/mongodb"
)

func main() {
    instance := mongodb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Create backup
    params := map[string]interface{}{
        "InstanceId": os.Getenv("INSTANCE_ID"),
        "BackupName": os.Getenv("BACKUP_NAME"),
    }

    resp, err := instance.Client.Request("mongodb", "CreateBackup", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create backup: %v\n", err)
        os.Exit(1)
    }

    var result struct {
        Result struct {
            BackupId string `json:"BackupId"`
        } `json:"Result"`
    }
    
    if err := json.Unmarshal(resp, &result); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Backup created: %s\n", result.Result.BackupId)
}
```

## Connection Examples

### Go Driver Connection

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Connection URI
    uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?replicaSet=%s",
        os.Getenv("MONGO_USER"),
        os.Getenv("MONGO_PASSWORD"),
        os.Getenv("MONGO_HOST"),
        os.Getenv("MONGO_PORT"),
        os.Getenv("MONGO_DB"),
        os.Getenv("MONGO_REPLICA_SET"),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        panic(err)
    }
    defer client.Disconnect(ctx)

    // Ping the database
    if err := client.Ping(ctx, nil); err != nil {
        panic(err)
    }

    fmt.Println("Connected to MongoDB!")
}
```

### Python Connection

```python
from pymongo import MongoClient
import os

client = MongoClient(
    host=os.getenv("MONGO_HOST"),
    port=int(os.getenv("MONGO_PORT", 27017)),
    username=os.getenv("MONGO_USER"),
    password=os.getenv("MONGO_PASSWORD"),
    authSource="admin",
    replicaSet=os.getenv("MONGO_REPLICA_SET")
)

# Test connection
db = client[os.getenv("MONGO_DB")]
print(db.command("ping"))
```

## Best Practices

### Security
- Use SSL/TLS for connections
- Store credentials in environment variables
- Use IP whitelists for access control
- Rotate passwords regularly

### Performance
- Use connection pooling
- Set appropriate timeouts
- Monitor connection count
- Use read preferences for secondaries

### Error Handling
- Implement retry logic for transient errors
- Handle connection failures gracefully
- Log errors with context
- Use circuit breakers for resilience
