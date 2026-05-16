# VKE Integration — Go SDK Setup

## Environment Setup

**Primary path:** `ve` CLI (static Go binary)
**Fallback path:** JIT Go SDK script

## Go SDK Package

| Product | Go SDK Package |
|---------|---------------|
| VKE | `github.com/volcengine/volc-sdk-golang/service/vke` |

## JIT Go SDK Workflow

### 1. Initialize Workspace

```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### 2. Bootstrap Go Runtime (if needed)

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
```

### 3. Get Dependencies

```bash
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
```

### 4. SDK Script Template (VKE)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/vke"
)

func main() {
    instance := vke.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // Set action-specific parameters
    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
        // Add action-specific fields here
    }

    resp, err := instance.Client.Request("vke", os.Getenv("VKE_ACTION"), params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### 5. Execute

```bash
cd /tmp/ve-sdk-workspace
export VKE_ACTION="ListClusters"  # or CreateCluster, DescribeCluster, etc.
go run ./main.go
```

## Build Time Estimate

| Step | First Run | Cached |
|------|-----------|--------|
| Go runtime download | ~30s | 0s |
| go get dependencies | ~10s | ~2s |
| go run | ~5s | ~3s |
| **Total** | **~45s** | **~5s** |
