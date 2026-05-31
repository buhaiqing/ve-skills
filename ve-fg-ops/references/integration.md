# Integration

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
fi
go version
```

### FunctionGraph SDK Package

| Resource | Go SDK Package |
|----------|---------------|
| FunctionGraph | `github.com/volcengine/volc-sdk-golang/service/functiongraph` |

### Cross-Service Integration

- **VPC:** Functions can be attached to VPC for private network access
- **APIG:** Trigger functions via API Gateway
- **CTS:** Trigger functions on Cloud Trail Service events
- **RocketMQ:** Trigger functions on message queue events
- **SLS:** Collect function logs
