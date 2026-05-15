# Integration — TOS

## Environment Setup

TOS uses **dedicated credential env vars** (not the same as `VOLCENGINE_*`):

```bash
# TOS-specific credentials
export TOS_ACCESS_KEY="your_access_key"
export TOS_SECRET_KEY="your_secret_key"

# Region (shared with Volcengine ecosystem)
export VOLCENGINE_REGION="cn-beijing"
```

## TOS CLI Tool (tosutil)

**Download:** `https://github.com/volcengine/tosutil/releases`

**Initialization:**
```bash
tosutil config
# Interactive setup: enter AK, SK, endpoint, region
```

**Config file location:** `~/.tosconfig` (JSON format)

## JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/tos-sdk-workspace
   cd /tmp/tos-sdk-workspace
   go mod init tos-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   go get -u github.com/volcengine/ve-tos-golang-sdk/v2
   ```

3. **Set credentials:**
   ```bash
   export TOS_ACCESS_KEY="{{env.TOS_ACCESS_KEY}}"
   export TOS_SECRET_KEY="{{env.TOS_SECRET_KEY}}"
   export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
   ```

4. **Execute:**
   ```bash
   go run ./main.go
   ```

## SDK Client Pattern

```go
package main

import (
    "context"
    "fmt"
    "os"
    "log"

    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func main() {
    client, err := tos.NewClientV2(
        "https://tos-"+os.Getenv("VOLCENGINE_REGION")+".volces.com",
        tos.WithRegion(os.Getenv("VOLCENGINE_REGION")),
        tos.WithCredentials(tos.NewStaticCredentials(
            os.Getenv("TOS_ACCESS_KEY"),
            os.Getenv("TOS_SECRET_KEY"),
        )),
        tos.WithMaxRetryCount(3),
        tos.WithConnectionTimeout(10),
        tos.WithRequestTimeout(30),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Use client for TOS operations
    output, err := client.ListBuckets(context.Background(), &tos.ListBucketsInput{})
    if err != nil {
        log.Fatal(err)
    }

    for _, bucket := range output.Buckets {
        fmt.Printf("  - %s (region: %s)\n", bucket.Name, bucket.Location)
    }
}
```
