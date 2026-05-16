# RDS PostgreSQL Integration — Go SDK Setup

## Go SDK Package

| Product | Go SDK Package |
|---------|---------------|
| RDS PostgreSQL | `github.com/volcengine/volc-sdk-golang/service/rds_postgresql` |

## JIT Go SDK Workflow

### Initialize
```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### Get Dependencies
```bash
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
```

### SDK Script Template

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/rds_postgresql"
)

func main() {
    instance := rds_postgresql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "DbEngineVersion": os.Getenv("PG_VERSION"),
        "NodeSpec":        os.Getenv("NODE_SPEC"),
        "PrimaryZoneId":   os.Getenv("PRIMARY_ZONE"),
        "SecondaryZoneId": os.Getenv("SECONDARY_ZONE"),
        "StorageSpace":    os.Getenv("STORAGE_SPACE"),
        "SubnetId":        os.Getenv("SUBNET_ID"),
        "InstanceName":    os.Getenv("INSTANCE_NAME"),
        "ChargeInfo":      map[string]interface{}{"ChargeType": "PostPaid"},
    }

    resp, err := instance.Client.Request("rds_postgresql", "CreateDBInstance", params)
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
go run ./main.go
```
