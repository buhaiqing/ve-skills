# Redis Integration — Go SDK Setup

## Go SDK Package

| Product | Go SDK Package |
|---------|---------------|
| Redis | `github.com/volcengine/volc-sdk-golang/service/redis` |

## JIT Go SDK Workflow

### Initialize Workspace
```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### Get Dependencies
```bash
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
```

### SDK Script Template (Redis)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/redis"
)

func main() {
    instance := redis.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":        os.Getenv("VOLCENGINE_REGION"),
        "Action":        os.Getenv("REDIS_ACTION"),
        // Add action-specific fields
    }

    resp, err := instance.Client.Request("redis", os.Getenv("REDIS_ACTION"), params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Execute
```bash
cd /tmp/ve-sdk-workspace
export REDIS_ACTION="CreateDBInstance"  # or DescribeDBInstances, etc.
go run ./main.go
```

## Build Time Estimate

| Step | First Run | Cached |
|------|-----------|--------|
| Go runtime | ~30s | 0s |
| go get deps | ~10s | ~2s |
| go run | ~5s | ~3s |
| **Total** | **~45s** | **~5s** |
