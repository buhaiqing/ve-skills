# Integration

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Go Runtime Bootstrap

```bash
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/{{env.GO_VERSION}}.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
fi
go version
```

### NAS SDK Package

| Resource | Go SDK Package |
|----------|---------------|
| NAS | `github.com/volcengine/volc-sdk-golang/service/nas` |

### Cross-Service Integration

- **ECS:** Mount NAS file systems on ECS instances
- **VPC:** Mount targets must be in the same VPC
- **TOS:** Archive cold NAS data to TOS for cost savings
- **CMS:** Set up monitoring alerts for capacity and performance
