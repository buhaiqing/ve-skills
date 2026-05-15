# Execution Environment Setup

> **Purpose:** Detailed environment setup for executing `ve` CLI and JIT Go SDK operations. This file provides progressive depth for the [ve-skill-generator](../SKILL.md) meta-skill's Step 0.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15

---

## Table of Contents

1. [CLI-First with JIT Go SDK Fallback](#1-cli-first-with-jit-go-sdk-fallback)
2. [Phase 1: ve CLI Setup](#2-phase-1-ve-cli-setup)
3. [Phase 2: JIT Go SDK Setup](#3-phase-2-jit-go-sdk-setup)
4. [Credential Configuration](#4-credential-configuration)
5. [Credential Security (Mandatory)](#5-credential-security-mandatory)
6. [Environment Variable Sources](#6-environment-variable-sources)
7. [Verification](#7-verification)
8. [Enhanced Self-Healing](#8-enhanced-self-healing)

---

## 1. CLI-First with JIT Go SDK Fallback

The execution environment follows a **CLI-first with JIT Go SDK fallback** strategy:

1. **Primary path:** `ve` CLI (static Go binary, covers most APIs)
2. **Fallback path:** JIT Go SDK (dynamic script generation + `go run`)
3. **Go runtime:** JIT download if not present

---

## 2. Phase 1: ve CLI Setup

### Primary Path

**Install `ve` CLI from GitHub releases:**

The Volcengine CLI is distributed as pre-built binaries via [GitHub Releases](https://github.com/volcengine/volcengine-cli/releases). The CLI is a static Go binary with no runtime dependencies.

```bash
# Auto-detect OS and architecture, then download
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
[ "$ARCH" = "arm64" ] && [ "$OS" = "darwin" ] && ARCH="arm64"  # macOS Apple Silicon

# Latest version (update as needed)
VE_VERSION="v1.0.42"

curl -fsSL "https://github.com/volcengine/volcengine-cli/releases/download/${VE_VERSION}/ve-${OS}-${ARCH}" -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve
```

**Binary naming convention:**
- **macOS ARM64**: `ve-darwin-arm64`
- **macOS x86_64**: `ve-darwin-amd64`
- **Linux AMD64**: `ve-linux-amd64`
- **Linux ARM64**: `ve-linux-arm64`

**Verification after bootstrap:**
```bash
ve version
```

### Self-Healing Installation

See [enhanced-self-healing-framework.md](enhanced-self-healing-framework.md) for complete self-healing installation procedures including:
1. Pre-flight checks (network, disk, permissions, system compatibility)
2. Intelligent error classification
3. Multi-path self-healing (mirror switch, timeout adjustment, cache clear)
4. Health verification (binary check, permission check, PATH check, functional test)
5. Graceful degradation (fallback to JIT Go SDK)

---

## 3. Phase 2: JIT Go SDK Setup

When `ve` CLI is unavailable or does not support a specific API, **JIT build a Go SDK script** on-demand.

### Step 3.1: Bootstrap Go Runtime

**Check existing Go runtime:**
```bash
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    GO_MAJOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f1)
    GO_MINOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f2)
    if [ "$GO_MAJOR" -ge 1 ] && [ "$GO_MINOR" -ge 14 ]; then
        echo "Compatible Go runtime: $GO_VERSION"
    fi
fi
```

**JIT download Go 1.21+ (auto-detects OS and architecture):**
```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

mkdir -p /tmp/go-runtime
curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime

export PATH="/tmp/go-runtime/go/bin:$PATH"
export GOPATH="/tmp/go-workspace"
export GOCACHE="/tmp/go-cache"
export GOMODCACHE="/tmp/go-modcache"
export GOPROXY="https://goproxy.cn,direct"
```

**Go version strategy:**
- **Primary:** Go 1.21+ (stable, good performance)
- **Fallback:** Go 1.20 → 1.19 → 1.18 → 1.17 → 1.14 (minimum compatibility)
- **Mirrors:** `https://go.dev/dl`, `https://dl.google.com/go`, `https://mirrors.cloud.tencent.com/golang`, `https://golang.google.cn/dl`
- **Module proxy:** `GOPROXY=https://goproxy.cn,direct` (China CDN mirror)

### Step 3.2: Initialize Go Workspace

```bash
mkdir -p /tmp/ve-sdk-workspace
cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### Step 3.3: Get SDK Dependencies

```bash
# Volcengine Go SDK
go get -u github.com/volcengine/volc-sdk-golang
```

**Multi-GOPROXY strategy (self-healing):**
```bash
GOPROXY_MIRRORS=(
    "https://goproxy.cn,direct"      # China CDN (primary)
    "https://goproxy.io,direct"      # Alternative China CDN
    "https://proxy.golang.org,direct" # Official proxy
    "direct"                          # Direct download (fallback)
)
```

> **SDK package structure:** `github.com/volcengine/volc-sdk-golang/service/<product>`
> Find service packages at: https://github.com/volcengine/volc-sdk-golang/tree/main/service

### Step 3.4: Generate and Execute SDK Script

```go
// main.go (generated dynamically by Agent)
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ecs"
)

func main() {
    instance := ecs.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")

    resp, err := instance.Client.Request("ecs", "DescribeInstances", params)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(resp))
}
```

Execute:
```bash
cd /tmp/ve-sdk-workspace
go run ./main.go
```

### JIT Build Time Estimate

| Step | First Run | Subsequent Runs |
|------|-----------|-----------------|
| Download Go runtime | ~30s | 0s (cached) |
| `go get` dependencies | ~10s | ~2s (cached) |
| `go run` | ~5s | ~3s |
| **Total** | **~45s** | **~5s** |

---

## 4. Credential Configuration

### Environment Variables (Recommended for Agent Execution)

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
```

### Interactive CLI Configuration

```bash
ve configure set --profile default --region cn-beijing --access-key "{{user.access_key}}" --secret-key "{{user.secret_key}}"
```

### Config File (`~/.volcengine/config.json`)

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

### Custom Config Path (Sandbox / CI Environments)

```bash
mkdir -p /tmp/ve-home/.volcengine
# Write config to custom path
# Note: ve CLI does not support --config-path flag; use HOME override
HOME=/tmp/ve-home ve <service> <action>
```

> The `ve` CLI also supports: SSO profiles, Console Login mode, session tokens. See official CLI docs for details.

### `.env` File Support

For local development convenience, load environment variables from a `.env` file:

```ini
# Volcengine credentials (use VOLCENGINE_* prefix)
VOLCENGINE_ACCESS_KEY=your_access_key
VOLCENGINE_SECRET_KEY=your_secret_key
VOLCENGINE_REGION=cn-beijing
```

**Safety rules:**
- **NEVER** commit `.env` files to version control
- **NEVER** write `.env` values into generated skill documents
- Generated skills continue using `{{env.*}}` placeholders
- Shell environment variables **MUST** override `.env` values

---

## 5. Credential Security (Mandatory)

All generated skills MUST enforce these credential security rules across **every** execution path (CLI, JIT Go SDK, verification scripts, debugging output):

| Context | Required Behavior | Example |
|---------|------------------|---------|
| **Console output** (stdout/stderr) | Any field whose key matches `*secret*`, `*key*` (case-insensitive) MUST have its value replaced with `<masked>` or `***` | `VOLCENGINE_SECRET_KEY=<masked>` |
| **Local log files** | Same masking rule; log entries MUST NOT contain raw credential values | `[INFO] Credentials: AK=***, SK=***` |
| **Error messages** | Error objects containing credential fields MUST be sanitized before display | `Error: Request failed (credential omitted)` |
| **Debug/verbose mode** | Warn user that credential values may appear; recommend isolated environments | `⚠️ Debug mode may expose credential values in output` |
| **JIT Go SDK scripts** | SDK script reads credentials from env vars (safe); but `fmt.Println`, log, or error dump MUST NOT include credential fields | `instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))` — struct never printed |
| **Template generation** | Use `{{env.*}}` placeholders only; never include example values or real keys | `SetSecretKey("{{env.VOLCENGINE_SECRET_KEY}}")` |
| **Credential verification** | Verify existence only; never `echo` or print the value | `✅ VOLCENGINE_SECRET_KEY is set` |

**Masking patterns (use one of the following):**
- `VOLCENGINE_SECRET_KEY=<masked>`
- `secret_key=***`
- `"secret_key": "***"`
- `secret=****`

**Non-compliance consequence:** Any skill that outputs un-masked credential values in console or logs SHALL be treated as a **security incident** and blocked from merge.

---

## 6. Environment Variable Sources

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (highest) | CLI flags / `ve configure set` | Profile settings override everything |
| 2 | Shell environment | `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` |
| 3 | `~/.volcengine/config.json` | Persistent profile config (JSON format) |
| 4 (lowest) | Default profile | `default` profile from config file |

**Supported env var aliases:**
- **AK**: `VOLCENGINE_ACCESS_KEY`
- **SK**: `VOLCENGINE_SECRET_KEY`
- **Region**: `VOLCENGINE_REGION`
- **Endpoint**: `VOLCENGINE_ENDPOINT` (optional, defaults to `open.volcengineapi.com`)
- **Endpoint Resolver**: `VOLCENGINE_ENDPOINT_RESOLVER` (set to `standard` to use standard resolver)
- **Session Token**: `VOLCENGINE_SESSION_TOKEN` (for STS/temporary credentials)
- **Disable SSL**: `VOLCENGINE_DISABLE_SSL` (default: `false`)
- **Dual Stack**: `VOLCENGINE_USE_DUALSTACK` (default: `false`)

---

## 7. Verification

After credential setup, verify before proceeding:

```bash
# Primary: ve CLI validation
ve ecs DescribeInstances --Region "{{env.VOLCENGINE_REGION}}"
```

If `ve` validation fails (3 retries with backoff), proceed to JIT Go SDK verification:

```bash
# Go SDK credential check (in /tmp/ve-sdk-workspace)
cat > /tmp/ve-sdk-workspace/verify.go << 'EOF'
package main
import (
    "fmt"
    "os"
)
func main() {
    ak := os.Getenv("VOLCENGINE_ACCESS_KEY")
    sk := os.Getenv("VOLCENGINE_SECRET_KEY")
    if ak == "" || sk == "" {
        fmt.Println("Missing VOLCENGINE_ACCESS_KEY or VOLCENGINE_SECRET_KEY")
        os.Exit(1)
    }
    fmt.Println("Credentials OK (JIT Go SDK mode)")
}
EOF
go run /tmp/ve-sdk-workspace/verify.go
```

> **SECURITY:** The verification code above **ONLY checks for existence** of credentials. **NEVER** log, print, or expose secret values. Use masked placeholders for any credential status output.

If all verification paths fail:
- HALT with clear message: "Credentials invalid or environment not set up"
- Suggest: Check `.env` file or run `ve configure set`

---

## 8. Enhanced Self-Healing

See [enhanced-self-healing-framework.md](enhanced-self-healing-framework.md) for the complete self-healing framework covering:

- **CLI Installation:** Pre-flight checks → intelligent download → installation execution → health verification → graceful degradation
- **Go Runtime JIT Download:** Multi-version multi-mirror strategy → integrity check → version compatibility → PATH setup → health check
- **Dependency Download:** Multi-GOPROXY strategy → timeout control → cache management → build verification
- **Error Classification:** Network errors, permission errors, resource errors, configuration errors with specific recovery actions per category
- **Success Criteria:** Health score ≥ 8/10, self-healing duration < 30s, user intervention rate < 20%
