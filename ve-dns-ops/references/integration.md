# Integration — Volcengine DNS

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Installing ve CLI

```bash
# macOS (ARM64 — Apple Silicon)
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-arm64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# macOS (x86_64 — Intel)
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Linux (x86_64)
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Linux (ARM64)
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-arm64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve
```

> **Note:** Check [releases page](https://github.com/volcengine/volcengine-cli/releases) for the latest version.

### Go Runtime Bootstrap

If Agent Runtime lacks Go, JIT download from official source:

```bash
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"

    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/{{env.GO_VERSION}}.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
fi

go version
```

### Credential Configuration

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
```

**Alternative — Config File (`~/.volcengine/config.json`):**
```bash
mkdir -p ~/.volcengine
cat > ~/.volcengine/config.json << 'CONFIGEOF'
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "{{user.access_key}}",
      "secret_key": "{{user.secret_key}}",
      "region": "{{user.region}}"
    }
  ]
}
CONFIGEOF
```

### Verification

```bash
# Verify CLI and credentials
ve dns ListDomains --Region "{{env.VOLCENGINE_REGION}}"
```

## JIT Go SDK Workflow

### Initialize Workspace

```bash
mkdir -p /tmp/ve-sdk-workspace
cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### Get Dependencies

```bash
# Set proxy for China CDN mirror (faster download)
export GOPROXY="https://goproxy.cn,direct"

go get -u github.com/volcengine/volc-sdk-golang
```

### Execute Script

```bash
cd /tmp/ve-sdk-workspace
DOMAIN_NAME="example.com" go run ./main.go
```

### Clean Up

```bash
rm -rf /tmp/ve-sdk-workspace
```

## SDK Package Structure

```
github.com/volcengine/volc-sdk-golang/
  ├── service/
  │   ├── dns/           # DNS service package
  │   ├── ecs/
  │   ├── vpc/
  │   └── ...
  ├── base/              # Base client, auth, config
  └── volc/              # Volume utilities
```

## Integration with Other Skills

| Scenario | DNS Operation | Related Skill |
|----------|--------------|---------------|
| Point domain to ECS IP | Add A/AAAA record | `ve-ecs-ops` (verify IP) |
| Point domain to CLB | Add CNAME/A record | `ve-clb-ops` (get CLB endpoint) |
| Configure CDN custom domain | Add CNAME to CDN | `ve-cdn-ops` |
| SSL certificate validation | Add TXT record for domain verification | `ve-ssl-ops` |
| Email server setup | Add MX + TXT (SPF/DKIM) | N/A (app-level) |

## DNS Propagation Notes

- TTL determines maximum propagation time
- Changes take effect immediately on Volcengine authoritative servers
- Global DNS resolvers may cache records for the full TTL duration
- To minimize propagation delay: reduce TTL to 60-300 seconds **before** making changes, then restore to normal values
