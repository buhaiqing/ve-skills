# PolarDB MySQL Integration — Go SDK Setup

## Go SDK Package

| Product | Go SDK Package |
|---------|---------------|
| PolarDB MySQL | `github.com/volcengine/volc-sdk-golang/service/polardb_mysql` |

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

### SDK Script Template (PolarDB MySQL)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
        // Add action-specific fields
    }

    resp, err := instance.Client.Request("polardb_mysql", os.Getenv("POLAR_ACTION"), params)
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
export POLAR_ACTION="CreateDBCluster"
go run ./main.go
```

## Build Time Estimate

| Step | First Run | Cached |
|------|-----------|--------|
| Go runtime | ~30s | 0s |
| go get deps | ~10s | ~2s |
| go run | ~5s | ~3s |
| **Total** | **~45s** | **~5s** |

## Operation-Specific Examples

### Create Cluster

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":          os.Getenv("VOLCENGINE_REGION"),
        "ZoneId":          "cn-beijing-a",
        "VpcId":           os.Getenv("VPC_ID"),
        "SubnetId":        os.Getenv("SUBNET_ID"),
        "ClusterName":     os.Getenv("CLUSTER_NAME"),
        "DBEngineVersion": "MySQL_8_0",
        "NodeClass":       "polar.mysql.x4.large",
        "NodeNumber":      2,
        "StorageSpace":    100,
        "ChargeType":      "PostPaid",
    }

    resp, err := instance.Client.Request("polardb_mysql", "CreateDBCluster", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Add Read-Only Nodes

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "ClusterId":  os.Getenv("CLUSTER_ID"),
        "NodeClass":  os.Getenv("NODE_CLASS"),
        "NodeNumber": 2,
    }

    resp, err := instance.Client.Request("polardb_mysql", "CreateDBNode", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Scale Storage

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "ClusterId":    os.Getenv("CLUSTER_ID"),
        "StorageSpace": 500,
    }

    resp, err := instance.Client.Request("polardb_mysql", "ScaleStorage", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Modify Parameters

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "ClusterId": os.Getenv("CLUSTER_ID"),
        "Parameters": []map[string]interface{}{
            {"ParameterName": "max_connections", "ParameterValue": "2000"},
            {"ParameterName": "innodb_buffer_pool_size", "ParameterValue": "8589934592"},
        },
    }

    resp, err := instance.Client.Request("polardb_mysql", "ModifyDBClusterParameters", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Failover Cluster

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "ClusterId": os.Getenv("CLUSTER_ID"),
    }

    resp, err := instance.Client.Request("polardb_mysql", "FailoverDBCluster", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

### Restore from Backup

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "BackupId":      os.Getenv("BACKUP_ID"),
        "ClusterName":   os.Getenv("NEW_CLUSTER_NAME"),
        "VpcId":         os.Getenv("VPC_ID"),
        "SubnetId":      os.Getenv("SUBNET_ID"),
        "NodeClass":     "polar.mysql.x4.large",
        "NodeNumber":    2,
        "StorageSpace":  100,
        "ChargeType":    "PostPaid",
    }

    resp, err := instance.Client.Request("polardb_mysql", "RestoreDBCluster", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

## Polling Helpers

### Wait for Cluster Running

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/polardb_mysql"
)

func waitForCluster(instance *polardb_mysql.PolarDBMySQL, clusterId string) error {
    for i := 0; i < 90; i++ {
        params := map[string]interface{}{
            "ClusterId": clusterId,
        }

        resp, err := instance.Client.Request("polardb_mysql", "DescribeDBClusterDetail", params)
        if err != nil {
            return err
        }

        var result map[string]interface{}
        if err := json.Unmarshal(resp, &result); err != nil {
            return err
        }

        status := result["Result"].(map[string]interface{})["ClusterStatus"].(string)
        fmt.Printf("Status: %s (attempt %d/90)\n", status, i+1)

        if status == "RUNNING" {
            return nil
        }
        if status == "ERROR" {
            return fmt.Errorf("cluster entered ERROR state")
        }

        time.Sleep(10 * time.Second)
    }
    return fmt.Errorf("timeout waiting for cluster")
}

func main() {
    instance := polardb_mysql.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    clusterId := os.Getenv("CLUSTER_ID")
    if err := waitForCluster(instance, clusterId); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Cluster is RUNNING")
}
```

## Environment Setup Script

```bash
#!/bin/bash
# setup-polar-env.sh

# Check Go
if ! command -v go &> /dev/null; then
    echo "Installing Go..."
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"

    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
fi

# Check credentials
if [ -z "$VOLCENGINE_ACCESS_KEY" ] || [ -z "$VOLCENGINE_SECRET_KEY" ]; then
    echo "Error: VOLCENGINE_ACCESS_KEY and VOLCENGINE_SECRET_KEY must be set"
    exit 1
fi

# Create workspace
mkdir -p /tmp/ve-sdk-workspace
cd /tmp/ve-sdk-workspace

# Initialize if needed
if [ ! -f go.mod ]; then
    go mod init ve-sdk-script
fi

# Get SDK
go get -u github.com/volcengine/volc-sdk-golang

echo "Setup complete. Ready to run PolarDB MySQL SDK scripts."
```
