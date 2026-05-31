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
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
fi
go version
```

### SLS SDK Package

| Resource | Go SDK Package |
|----------|---------------|
| SLS/TLS | `github.com/volcengine/volc-sdk-golang/service/tls` |

### Cross-Service Integration

- **ECS:** Logtail agent runs on ECS instances
- **TOS:** LogShipper delivers logs to TOS buckets
- **CMS:** Set up monitoring alerts for log volume anomalies
- **FunctionGraph:** Trigger serverless functions on log events
