# Integration — CMS

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script + `go run`)

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-cms-workspace
   cd /tmp/ve-cms-workspace
   go mod init ve-cms-script
   ```

2. **Get dependencies:**
   ```bash
   export GOPROXY="https://goproxy.cn,direct"
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Execute:**
   ```bash
   go run ./main.go
   ```

### SDK Client Pattern for CMS

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/base"
)

func main() {
    client := base.NewClient(
        os.Getenv("VOLCENGINE_ACCESS_KEY"),
        os.Getenv("VOLCENGINE_SECRET_KEY"),
    )
    client.SetHost("open.volcengineapi.com")

    params := map[string]string{
        "Action":    "GetMetricData",
        "Version":   "2018-03-14",
        "Namespace": os.Getenv("NAMESPACE"),
        "MetricName": os.Getenv("METRIC_NAME"),
    }

    resp, err := client.Get("metrics_v2", params)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(resp))
}
```
