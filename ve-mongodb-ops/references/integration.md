# Integration

> **Version:** 1.1.0 | **Last Updated:** 2026-07-04

## Environment Setup

**Primary:** `ve` CLI (static Go binary, no runtime)

**Fallback:** JIT Go SDK (dynamic script + `go run`)

### Go Runtime Bootstrap

```bash
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
fi
go version
```

### JIT Go SDK Workflow

```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
go run ./main.go
```

## SDK Package

| Product | Package |
|---------|---------|
| MongoDB | `github.com/volcengine/volc-sdk-golang/service/mongodb` |

## Script Template

All operations follow this pattern:

```go
// main.go
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

    params := map[string]interface{}{
        // Set params for target operation (see SKILL.md Execution Flows)
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("mongodb", "<Action>", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

> Full operation-specific examples are available in SKILL.md (CreateDBInstance JIT fallback). The same pattern applies to all other operations by changing `<Action>` and `params`.

## Connection Examples

### Go Driver

```go
// go.mongodb.org/mongo-driver/mongo
uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?replicaSet=%s",
    os.Getenv("MONGO_USER"), os.Getenv("MONGO_PASSWORD"),
    os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"),
    os.Getenv("MONGO_DB"), os.Getenv("MONGO_REPLICA_SET"))

client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
if err = client.Ping(context.TODO(), nil); err != nil {
    panic(err)
}
```

### Python (pymongo)

```python
from pymongo import MongoClient
client = MongoClient(host=os.getenv("MONGO_HOST"), port=int(os.getenv("MONGO_PORT", 27017)),
    username=os.getenv("MONGO_USER"), password=os.getenv("MONGO_PASSWORD"),
    authSource="admin")
print(client[os.getenv("MONGO_DB")].command("ping"))
```

## Best Practices

- **Security:** SSL/TLS, env vars for credentials, IP whitelist, rotate passwords
- **Perf:** Connection pooling, set timeouts, use read preferences for secondaries
- **Error handling:** Retry transient errors, log with context, circuit breaker