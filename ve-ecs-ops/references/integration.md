# Integration — ECS

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Go Runtime Bootstrap

If Go runtime is not available:

```bash
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
```

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-sdk-workspace
   cd /tmp/ve-sdk-workspace
   go mod init ve-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate and execute:**
   ```bash
   go run ./main.go
   ```

### ECS SDK Package

| Product | Go SDK Package |
|---------|---------------|
| ECS | `github.com/volcengine/volc-sdk-golang/service/ecs` |

Full service list: https://github.com/volcengine/volc-sdk-golang/tree/main/service

### SDK Script Template for ECS

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ecs"
)

func main() {
    instance := ecs.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("DescribeInstances", nil, params)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(resp))
}
```

> Use `os.Getenv("KEY")` for all credentials. Never hardcode secrets in scripts.
