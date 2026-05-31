# Integration — Ark (方舟大模型平台)

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Go Runtime Bootstrap

If Agent Runtime lacks Go, JIT download from official source:

```bash
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

> **Go version strategy:**
> - **JIT download:** Go 1.21+ (stable)
> - **Script compatibility:** Go 1.14+ (minimum)

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-sdk-workspace
   cd /tmp/ve-sdk-workspace
   go mod init ve-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   export GOPROXY="https://goproxy.cn,direct"
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate script** (Agent dynamically creates operation-specific `.go` file)

4. **Execute:**
   ```bash
   go run ./main.go
   ```

### SDK Package

| Product | Go SDK Package |
|---------|---------------|
| Ark | `github.com/volcengine/volc-sdk-golang/service/ark` |

## SDK Script Template

```go
// main.go (single-file script template for Ark)
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    // Add operation-specific parameters

    // Change "ListEndpoints" to the desired action
    resp, err := instance.Client.Request("ark", "ListEndpoints", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

> Use `os.Getenv("KEY")` for all credentials. Never hardcode secrets in scripts.

## Related Service Integration

### VPC Integration for Private Endpoints

When creating endpoints with VPC access:
```bash
# Verify VPC exists (ve-vpc-ops)
ve vpc DescribeVpcs --VpcId "vpc-xxx" --Region "{{env.VOLCENGINE_REGION}}"

# Then create endpoint with VPC
ve ark CreateEndpoint \
  --EndpointName "{{user.endpoint_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --EndpointType "Inference" \
  --VpcId "vpc-xxx" \
  --SubnetIds '["subnet-xxx"]' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

### TOS Integration for Datasets

When creating datasets from TOS:
```bash
# Verify TOS bucket exists (ve-tos-ops)
ve tos HeadBucket --Bucket "{{user.bucket_name}}"

# Then create dataset
ve ark CreateDataset \
  --DatasetName "{{user.dataset_name}}" \
  --DatasetType "{{user.dataset_type}}" \
  --DataSourceType "TOS" \
  --TosPath "tos://{{user.bucket_name}}/{{user.object_prefix}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

### IAM Integration for Permissions

```bash
# Create IAM policy for Ark operations (ve-iam-ops)
ve iam CreatePolicy \
  --PolicyName "ark-endpoint-admin" \
  --PolicyDocument '{
    "Statement": [{
      "Effect": "Allow",
      "Action": ["ark:ListEndpoints", "ark:CreateEndpoint", "ark:DeleteEndpoint"],
      "Resource": ["*"]
    }]
  }' \
  --Region "{{env.VOLCENGINE_REGION}}"

# Attach to user
ve iam AttachUserPolicy \
  --UserName "{{user.user_name}}" \
  --PolicyName "ark-endpoint-admin" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

## Environment Verification Script

```bash
#!/bin/bash
# verify-ark-env.sh — Check Ark execution environment

echo "=== Ark Environment Check ==="

# Check CLI
echo -n "ve CLI: "
if command -v ve &> /dev/null; then
    echo "✅ $(ve version 2>&1 | head -1)"
else
    echo "❌ Not installed"
fi

# Check credentials
echo -n "VOLCENGINE_ACCESS_KEY: "
test -n "$VOLCENGINE_ACCESS_KEY" && echo "✅ Set" || echo "❌ Missing"
echo -n "VOLCENGINE_SECRET_KEY: "
test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ Set" || echo "❌ Missing"
echo -n "VOLCENGINE_REGION: "
test -n "$VOLCENGINE_REGION" && echo "✅ $VOLCENGINE_REGION" || echo "❌ Missing"

# Check Ark CLI access
echo -n "Ark API access: "
if ve ark ListEndpoints --Region "${VOLCENGINE_REGION:-cn-beijing}" &> /dev/null; then
    echo "✅"
else
    echo "❌ API call failed — check credentials and region"
fi

echo "=== Done ==="
```