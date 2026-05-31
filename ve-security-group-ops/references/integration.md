# Integration

## Environment Setup

**Primary path:** `ve` CLI via VPC service (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

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

### Security Group SDK Package

| Resource | Go SDK Package |
|----------|---------------|
| VPC (SG APIs) | `github.com/volcengine/volc-sdk-golang/service/vpc` |

### Cross-Service Integration

- **ECS:** Security groups are applied to ECS instances
- **VPC:** Security groups belong to a VPC
- **CLB:** Load balancer security groups
- **CMS:** Monitor ECS network traffic
