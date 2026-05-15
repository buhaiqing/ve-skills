# Enhanced Self-Healing Framework for CLI Installation

> **Purpose:** 定义增强的CLI安装异常处理和自愈能力框架，确保在各种异常场景下都能自动恢复或提供明确的降级路径。
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15
> **Status:** MANDATORY — 所有生成的Skill必须遵循此自愈框架

---

## Table of Contents

1. [核心设计原则](#1-核心设计原则)
2. [错误分类体系](#2-错误分类体系)
3. [增强的自愈流程](#3-增强的自愈流程)
4. [降级路径和用户指导](#4-降级路径和用户指导)
5. [健康检查和状态验证](#5-健康检查和状态验证)
6. [自愈效果追踪和优化](#6-自愈效果追踪和优化)
7. [实施优先级](#7-实施优先级)
8. [合规性检查清单](#8-合规性检查清单)

---

## 1. 核心设计原则

### 1.1 自愈能力成熟度模型

| 等级 | 名称 | 特征 | 目标 |
|------|------|------|------|
| L1 | 基础重试 | 固定次数重试，无错误分类 | 当前状态 |
| L2 | 智能重试 | 错误分类，针对性重试策略 | 立即实现 |
| L3 | 多路径自愈 | 多种自愈路径，自动选择最优方案 | 短期目标 |
| L4 | 预防性自愈 | 预检异常，提前规避 | 中期目标 |
| L5 | 自学习自愈 | 历史数据分析，优化自愈策略 | 长期目标 |

### 1.2 自愈决策树原则

```
[异常发生]
    │
    ├── Step 1: 错误分类
    │   网络异常 / 权限异常 / 资源异常 / 配置异常 / 未知异常
    │
    ├── Step 2: 选择自愈路径
    │   根据错误类型选择对应的自愈策略
    │
    ├── Step 3: 执行自愈
    │   尝试自愈操作，记录结果
    │
    ├── Step 4: 验证自愈效果
    │   检查异常是否已解决
    │
    ├── Step 5: 自愈失败处理
    │   尝试下一级自愈路径或降级
    │
    └── Step 6: 用户指导
        提供明确的错误信息和修复建议
```

---

## 2. 错误分类体系

### 2.1 CLI安装错误分类

| 错误类别 | 错误代码 | 典型场景 | 自愈策略 |
|----------|---------|---------|---------|
| **网络异常** | `NET_TIMEOUT` | curl下载超时 | 切换镜像源，增加超时时间 |
| | `NET_DNS_FAIL` | DNS解析失败 | 使用IP直连或备用域名 |
| | `NET_CONNECTION_REFUSED` | 连接被拒绝 | 检查防火墙，切换网络 |
| | `NET_SSL_ERROR` | SSL证书错误 | 更新CA证书，使用--insecure |
| **权限异常** | `PERM_WRITE_FAIL` | 写入/usr/local/bin失败 | 使用用户目录，提示sudo |
| | `PERM_EXEC_FAIL` | 执行权限不足 | chmod +x，提示sudo |
| | `PERM_DIR_CREATE_FAIL` | 创建目录失败 | 使用/tmp目录，提示权限问题 |
| **资源异常** | `RES_DISK_FULL` | 磁盘空间不足 | 清理临时文件，提示用户 |
| | `RES_BINARY_CORRUPT` | 下载文件损坏 | 删除重新下载，校验完整性 |
| | `RES_VERSION_INCOMPATIBLE` | 版本不兼容 | 下载兼容版本 |
| **配置异常** | `CONF_PATH_NOT_FOUND` | PATH未包含安装路径 | 自动添加PATH，提示用户 |
| | `CONF_ENV_VAR_MISSING` | 环境变量缺失 | 设置临时环境变量 |
| **未知异常** | `UNKNOWN_ERROR` | 未分类错误 | 记录详细信息，提供诊断建议 |

### 2.2 Go Runtime JIT下载错误分类

| 错误类别 | 错误代码 | 典型场景 | 自愈策略 |
|----------|---------|---------|---------|
| **下载异常** | `GO_DOWNLOAD_FAIL` | Go runtime下载失败 | 切换镜像源，使用国内镜像 |
| | `GO_DOWNLOAD_INCOMPLETE` | 下载不完整 | 校验文件大小，重新下载 |
| | `GO_DOWNLOAD_TIMEOUT` | 下载超时 | 增加超时时间，使用更快的镜像 |
| **解压异常** | `GO_EXTRACT_FAIL` | tar解压失败 | 检查文件完整性，重新下载 |
| | `GO_EXTRACT_CORRUPT` | 解压后文件损坏 | 删除重新下载解压 |
| **版本异常** | `GO_VERSION_MISMATCH` | 版本不匹配 | 下载指定版本 |
| | `GO_VERSION_INCOMPATIBLE` | 版本不兼容 | 下载兼容版本(1.14+) |
| **环境异常** | `GO_PATH_SETUP_FAIL` | PATH设置失败 | 使用绝对路径调用 |
| | `GO_WORKSPACE_INIT_FAIL` | 工作空间初始化失败 | 清理重新初始化 |

### 2.3 依赖下载错误分类

| 错误类别 | 错误代码 | 典型场景 | 自愈策略 |
|----------|---------|---------|---------|
| **网络异常** | `DEP_NET_TIMEOUT` | go get超时 | 切换GOPROXY，增加超时 |
| | `DEP_NET_PROXY_FAIL` | 代理失败 | 切换镜像源 |
| **版本异常** | `DEP_VERSION_NOT_FOUND` | 版本不存在 | 使用最新稳定版本 |
| | `DEP_VERSION_INCOMPATIBLE` | 版本冲突 | 解决依赖冲突 |
| **权限异常** | `DEP_WRITE_FAIL` | 写入GOMODCACHE失败 | 使用/tmp目录 |
| **编译异常** | `DEP_BUILD_FAIL` | 编译失败 | 检查Go版本，清理缓存 |

---

## 3. 增强的自愈流程

### 3.1 CLI安装增强自愈流程

#### Phase 1: 预检阶段

```bash
# 预检1: 检查网络连通性
echo "=== Pre-flight Check: Network Connectivity ==="
if ! curl -fsSL --connect-timeout 5 https://github.com/volcengine/volcengine-cli > /dev/null 2>&1; then
    echo "⚠️  Network connectivity check failed"
    echo "Attempting alternative endpoints..."
    
    # 尝试备用端点
    ALT_ENDPOINTS=(
        "https://gitee.com/volcengine/volcengine-cli"
        "https://mirror.volcengine.com/"
    )
    
    for endpoint in "${ALT_ENDPOINTS[@]}"; do
        if curl -fsSL --connect-timeout 5 "$endpoint" > /dev/null 2>&1; then
            echo "✅ Alternative endpoint available: $endpoint"
            break
        fi
    done
    
    echo "❌ All endpoints unreachable. Network issue detected."
    ERROR_CODE="NET_CONNECTION_REFUSED"
fi

# 预检2: 检查磁盘空间
echo "=== Pre-flight Check: Disk Space ==="
REQUIRED_SPACE_MB=50
AVAILABLE_SPACE_KB=$(df -k /tmp | awk 'NR==2 {print $4}')
AVAILABLE_SPACE_MB=$((AVAILABLE_SPACE_KB / 1024))

if [ "$AVAILABLE_SPACE_MB" -lt "$REQUIRED_SPACE_MB" ]; then
    echo "⚠️  Insufficient disk space: ${AVAILABLE_SPACE_MB}MB available, ${REQUIRED_SPACE_MB}MB required"
    echo "Attempting self-healing: Cleaning temporary files..."
    
    rm -rf /tmp/ve-* /tmp/go-* /tmp/ve-sdk-* 2>/dev/null || true
    
    AVAILABLE_SPACE_KB=$(df -k /tmp | awk 'NR==2 {print $4}')
    AVAILABLE_SPACE_MB=$((AVAILABLE_SPACE_KB / 1024))
    
    if [ "$AVAILABLE_SPACE_MB" -lt "$REQUIRED_SPACE_MB" ]; then
        echo "❌ Self-healing failed: Still insufficient disk space"
        ERROR_CODE="RES_DISK_FULL"
        echo "User action required: Free up disk space or use alternative installation path"
    fi
fi

# 预检3: 检查安装路径权限
echo "=== Pre-flight Check: Installation Path Permissions ==="
INSTALL_PATH="/usr/local/bin"
if [ ! -w "$INSTALL_PATH" ]; then
    echo "⚠️  No write permission to $INSTALL_PATH"
    echo "Self-healing: Using alternative installation path..."
    
    USER_BIN="$HOME/.local/bin"
    mkdir -p "$USER_BIN"
    
    if [ -w "$USER_BIN" ]; then
        echo "✅ Alternative path available: $USER_BIN"
        INSTALL_PATH="$USER_BIN"
        
        if [[ ":$PATH:" != *":$USER_BIN:"* ]]; then
            export PATH="$USER_BIN:$PATH"
            echo "✅ Added $USER_BIN to PATH (temporary)"
        fi
    else
        echo "❌ Self-healing failed: No writable installation path"
        ERROR_CODE="PERM_WRITE_FAIL"
    fi
fi

# 预检4: 检查系统兼容性
echo "=== Pre-flight Check: System Compatibility ==="
OS=$(uname -s)
ARCH=$(uname -m)

if [ "$OS" != "Darwin" ] && [ "$OS" != "Linux" ]; then
    echo "❌ Unsupported OS: $OS (supported: macOS, Linux)"
    ERROR_CODE="RES_VERSION_INCOMPATIBLE"
fi

if [ "$ARCH" = "x86_64" ]; then
    ARCH_SUFFIX="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH_SUFFIX="arm64"
else
    echo "❌ Unsupported architecture: $ARCH"
    ERROR_CODE="RES_VERSION_INCOMPATIBLE"
fi

echo "✅ System compatible: $OS $ARCH_SUFFIX"
```

#### Phase 2: 智能下载阶段

```bash
download_ve_cli() {
    local attempt=1
    local max_attempts=5
    local version="v1.0.42"
    local mirrors=(
        "https://github.com/volcengine/volcengine-cli/releases/download/${version}/ve-${OS}-${ARCH_SUFFIX}"
        "https://mirror.volcengine.com/volcengine-cli/${version}/ve-${OS}-${ARCH_SUFFIX}"
    )
    
    while [ $attempt -le $max_attempts ]; do
        echo "=== Download Attempt $attempt/$max_attempts ==="
        
        for mirror in "${mirrors[@]}"; do
            echo "Trying mirror: $mirror"
            
            if curl -fsSL --connect-timeout 15 --max-time 120 "$mirror" -o /tmp/ve; then
                if [ -f /tmp/ve ] && [ -s /tmp/ve ]; then
                    echo "✅ Download successful"
                    mv /tmp/ve "${INSTALL_PATH}/ve"
                    chmod +x "${INSTALL_PATH}/ve"
                    return 0
                else
                    echo "⚠️  Downloaded file is empty or missing"
                    ERROR_CODE="RES_BINARY_CORRUPT"
                fi
            else
                echo "⚠️  Download failed from $mirror"
                ERROR_CODE="NET_TIMEOUT"
            fi
        done
        
        # 自愈策略
        if [ "$ERROR_CODE" = "NET_TIMEOUT" ]; then
            echo "Self-healing: Increasing timeout and retrying..."
            sleep $((attempt * 2))
        elif [ "$ERROR_CODE" = "RES_BINARY_CORRUPT" ]; then
            echo "Self-healing: Clearing cache and retrying..."
            rm -f /tmp/ve
        fi
        
        attempt=$((attempt + 1))
    done
    
    echo "❌ All download attempts failed after $max_attempts retries"
    return 1
}
```

#### Phase 3: 安装验证阶段

```bash
health_check_ve_cli() {
    echo "=== VE CLI Health Check ==="
    
    HEALTH_SCORE=0
    MAX_SCORE=10
    
    if [ -f "${INSTALL_PATH}/ve" ]; then
        echo "✅ Binary exists"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    else
        echo "❌ Binary missing"
    fi
    
    if [ -x "${INSTALL_PATH}/ve" ]; then
        echo "✅ Execute permission present"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    else
        echo "❌ Execute permission missing"
        chmod +x "${INSTALL_PATH}/ve" 2>/dev/null && HEALTH_SCORE=$((HEALTH_SCORE + 2))
    fi
    
    if command -v ve &> /dev/null; then
        echo "✅ In PATH"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    else
        echo "❌ Not in PATH"
        export PATH="${INSTALL_PATH}:$PATH"
        command -v ve &> /dev/null && HEALTH_SCORE=$((HEALTH_SCORE + 2))
    fi
    
    if ve version &> /dev/null; then
        echo "✅ Version command works"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    else
        echo "❌ Version command failed"
    fi
    
    if ve ecs DescribeInstances --Region cn-beijing &> /dev/null; then
        echo "✅ Basic API call works"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    else
        echo "⚠️  Basic API call failed (may be credential issue)"
    fi
    
    echo "Health Score: $HEALTH_SCORE/$MAX_SCORE"
    
    if [ "$HEALTH_SCORE" -ge 8 ]; then
        echo "✅ Health check passed"
        return 0
    elif [ "$HEALTH_SCORE" -ge 6 ]; then
        echo "⚠️  Health check partially passed"
        return 0
    else
        echo "❌ Health check failed"
        return 1
    fi
}
```

### 3.2 Go Runtime JIT下载增强自愈流程

```bash
bootstrap_go_runtime_enhanced() {
    echo "=== Go Runtime Bootstrap (Enhanced Self-Healing) ==="
    
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        GO_MAJOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f1)
        GO_MINOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f2)
        
        if [ "$GO_MAJOR" -ge 1 ] && [ "$GO_MINOR" -ge 14 ]; then
            echo "✅ Compatible Go runtime already installed: $GO_VERSION"
            return 0
        else
            echo "⚠️  Installed Go version $GO_VERSION is too old (minimum: go1.14)"
        fi
    fi
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
    if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi
    
    GO_VERSIONS=("go1.21.0" "go1.20.0" "go1.19.0" "go1.18.0" "go1.17.0" "go1.14.0")
    
    GO_MIRRORS=(
        "https://go.dev/dl"
        "https://dl.google.com/go"
        "https://mirrors.cloud.tencent.com/golang"
        "https://golang.google.cn/dl"
    )
    
    for go_version in "${GO_VERSIONS[@]}"; do
        for mirror in "${GO_MIRRORS[@]}"; do
            GO_URL="${mirror}/${go_version}.${OS}-${ARCH}.tar.gz"
            
            if curl -fsSL --connect-timeout 15 --max-time 120 "$GO_URL" -o /tmp/go-runtime.tar.gz; then
                FILE_SIZE=$(stat -f%z /tmp/go-runtime.tar.gz 2>/dev/null || stat -c%s /tmp/go-runtime.tar.gz 2>/dev/null)
                
                if [ "$FILE_SIZE" -gt 100000000 ]; then
                    mkdir -p /tmp/go-runtime
                    if tar -xzf /tmp/go-runtime.tar.gz -C /tmp/go-runtime; then
                        if [ -f "/tmp/go-runtime/go/bin/go" ]; then
                            export PATH="/tmp/go-runtime/go/bin:$PATH"
                            export GOPATH="/tmp/go-workspace"
                            export GOCACHE="/tmp/go-cache"
                            export GOMODCACHE="/tmp/go-modcache"
                            export GOPROXY="https://goproxy.cn,direct"
                            
                            echo "✅ Go runtime installed: $(go version | awk '{print $3}')"
                            rm -f /tmp/go-runtime.tar.gz
                            return 0
                        fi
                    fi
                    rm -rf /tmp/go-runtime /tmp/go-runtime.tar.gz
                else
                    rm -f /tmp/go-runtime.tar.gz
                fi
            fi
        done
    done
    
    echo "❌ Go runtime download failed after trying all versions and mirrors"
    return 1
}
```

---

## 4. 降级路径和用户指导

### 4.1 降级路径决策树

```
[CLI安装失败]
    │
    ├── 尝试自愈(最多5次)
    │   │
    │   ├── 自愈成功 → 继续执行
    │   │
    │   └── 自愈失败 → 进入降级路径
    │       │
    │       ├── 降级路径1: JIT Go SDK模式
    │       │   │
    │       │   ├── Go runtime可用 → 使用Go SDK
    │       │   │
    │       │   └── Go runtime不可用 → JIT下载Go
    │       │
    │       ├── 降级路径2: 控制台手动操作
    │       │   提供控制台链接和操作步骤
    │       │
    │       └── 降级路径3: 用户手动修复
    │           提供详细的错误信息和修复建议
```

### 4.2 用户指导模板

```markdown
## ❌ Installation Failed — Self-Healing Exhausted

### Error Summary
- **Error Code:** {{error_code}}
- **Error Category:** {{error_category}}
- **Failed Component:** {{failed_component}}
- **Attempted Self-Healing:** {{self_healing_attempts}} attempts

### Recommended Actions

#### Option 1: Manual Installation
```bash
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve
ve version
```

#### Option 2: Use JIT Go SDK Mode
The Agent will automatically use Go SDK fallback.

#### Option 3: Use Volcengine Console
- Console URL: https://console.volcengine.com/

### Support Escalation
If the issue persists:
1. Create a support ticket: https://console.volcengine.com/ticket
2. Include Error Code: {{error_code}}
```

---

## 5. 健康检查和状态验证

### 5.1 安装后健康检查

(See Phase 3 in section 3.1 above for the complete health check script)

### 5.2 Go Runtime健康检查

```bash
health_check_go_runtime() {
    echo "=== Go Runtime Health Check ==="
    
    HEALTH_SCORE=0
    MAX_SCORE=8
    
    if [ -f "/tmp/go-runtime/go/bin/go" ]; then
        echo "✅ Go binary exists"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    fi
    
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        GO_MAJOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f1)
        GO_MINOR=$(echo "$GO_VERSION" | sed 's/go//' | cut -d. -f2)
        
        if [ "$GO_MAJOR" -ge 1 ] && [ "$GO_MINOR" -ge 14 ]; then
            echo "✅ Go version compatible: $GO_VERSION"
            HEALTH_SCORE=$((HEALTH_SCORE + 2))
        fi
    fi
    
    if [ -f "/tmp/ve-sdk-workspace/go.mod" ]; then
        echo "✅ Workspace initialized"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    fi
    
    if [ -d "/tmp/go-workspace/pkg/mod/github.com/volcengine" ]; then
        echo "✅ SDK dependencies available"
        HEALTH_SCORE=$((HEALTH_SCORE + 2))
    fi
    
    echo "Health Score: $HEALTH_SCORE/$MAX_SCORE"
    [ "$HEALTH_SCORE" -ge 6 ] && return 0 || return 1
}
```

---

## 6. 自愈效果追踪和优化

### 6.1 自愈效果指标

| 指标 | 目标值 | 测量方法 |
|------|--------|---------|
| 自愈成功率 | > 80% | 成功自愈次数 / 总异常次数 |
| 平均自愈时间 | < 30s | 从异常发生到自愈完成的时间 |
| 用户干预率 | < 20% | 需要用户手动干预的异常比例 |
| 降级路径使用率 | < 10% | 进入降级路径的异常比例 |

### 6.2 自愈日志记录

```bash
log_self_healing_event() {
    local event_type="$1"
    local error_code="$2"
    local self_healing_action="$3"
    local result="$4"
    local duration="$5"
    
    LOG_FILE="/tmp/ve-self-healing.log"
    echo "$(date -Iseconds) | $event_type | $error_code | $self_healing_action | $result | $duration" >> "$LOG_FILE"
}
```

---

## 7. 实施优先级

### 7.1 立即实施 (P0)

1. **增强CLI安装预检阶段** — 添加网络、磁盘、权限预检
2. **实现智能错误分类** — 建立完整的错误代码体系
3. **增强下载健壮性** — 多镜像源、完整性校验、失败自愈
4. **实现健康检查机制** — 安装后验证和状态追踪

### 7.2 短期实施 (P1)

1. **优化Go runtime JIT下载** — 多版本、多镜像、完整性校验
2. **增强依赖下载容错** — 多GOPROXY、超时控制、缓存清理
3. **实现降级路径决策树** — 自动选择最优降级方案
4. **标准化用户指导模板** — 提供清晰的错误信息和修复建议

---

## 8. 合规性检查清单

- [ ] 所有CLI安装路径包含预检阶段
- [ ] 错误分类覆盖所有已知异常类型
- [ ] 每个错误类型有对应的自愈策略
- [ ] 自愈失败后有明确的降级路径
- [ ] 用户指导包含详细的错误信息和修复建议
- [ ] 安装后执行健康检查
- [ ] 自愈事件记录到日志
- [ ] 自愈成功率可追踪和测量

---

*This framework is mandatory for all generated skills. Update quarterly based on self-healing effectiveness data.*
