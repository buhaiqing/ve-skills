# Integration

## Environment Setup

### Primary Path: `ve` CLI

The `ve` CLI is the primary execution path — static Go binary with no runtime dependencies.

**Installation:**
```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify installation
ve version
```

### Fallback Path: JIT Go SDK

When CLI doesn't support a specific operation, use JIT Go SDK fallback.

### Go Runtime Bootstrap

If Go is not installed, JIT download:

```bash
# Check Go runtime
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
    export GOMODCACHE="/tmp/go-modcache"
    export GOPROXY="https://goproxy.cn,direct"
fi

go version
```

**Go version strategy:**
- **JIT download:** Go 1.21+ (stable)
- **Script compatibility:** Go 1.14+ (minimum)

## Credential Configuration

### Environment Variables (Recommended)

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="cn-beijing"
```

**Security — Credential Masking:**

| Execution | Safe | Unsafe |
|-----------|------|--------|
| Verification | `test -n "$VOLCENGINE_SECRET_KEY"` | `echo $VOLCENGINE_SECRET_KEY` |
| Logging | `VOLCENGINE_SECRET_KEY=<masked>` | `VOLCENGINE_SECRET_KEY=AKLT...` |
| Debug output | `SecretKey=***` | Printing actual value |

### CLI Configuration File

```bash
mkdir -p ~/.volcengine
cat > ~/.volcengine/config.json << 'CONFIGEOF'
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "YOUR_ACCESS_KEY",
      "secret_key": "YOUR_SECRET_KEY",
      "region": "cn-beijing"
    }
  ]
}
CONFIGEOF
chmod 600 ~/.volcengine/config.json
```

## JIT Go SDK Workflow

### 1. Initialize Workspace

```bash
mkdir -p /tmp/ve-sdk-workspace
cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
```

### 2. Get Dependencies

```bash
# Set proxy for China CDN mirror
export GOPROXY="https://goproxy.cn,direct"

# Volcengine SDK
go get -u github.com/volcengine/volc-sdk-golang
```

### 3. VPN SDK Package

```go
import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/vpn"
)
```

### 4. SDK Script Template

```go
// main.go (single-file script template)
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/vpn"
)

func main() {
    // Initialize VPN service instance
    instance := vpn.NewInstance()
    
    // Configure credentials from environment
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    
    // Add operation-specific parameters
    params["VpnGatewayId"] = "vgw-xxxxx"
    
    // Make API call
    resp, err := instance.Client.Request("DescribeVpnGateways", nil, params)
    if err != nil {
        log.Fatalf("API call failed: %v", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

### 5. Execute

```bash
go run ./main.go
```

## VPC Integration

VPN Gateway requires VPC and Subnet. Verify they exist:

```bash
# Check VPC
ve vpc DescribeVpcs --Region cn-beijing --VpcIds '["vpc-xxxxx"]'

# Check Subnet
ve vpc DescribeSubnets --Region cn-beijing --SubnetIds '["subnet-xxxxx"]'
```

### Required VPC Permissions

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpc:DescribeVpcs",
        "vpc:DescribeSubnets"
      ],
      "Resource": "*"
    }
  ]
}
```

## Route Table Integration

### Add Route for VPN Destination

After creating IPSec connection, add routes to VPC route table:

```bash
# Get route table for VPC
ve vpc DescribeRouteTables --Region cn-beijing --VpcId vpc-xxxxx

# Create route entry for remote subnet
ve vpc CreateRouteEntry \
  --Region cn-beijing \
  --RouteTableId rtb-xxxxx \
  --DestinationCidrBlock "192.168.0.0/16" \
  --NextHopType "VpnGateway" \
  --NextHopId "vgw-xxxxx"
```

## Security Group Integration

### Allow VPN Traffic

```bash
# Authorize IKE and IPSec traffic
ve ecs AuthorizeSecurityGroupIngress \
  --Region cn-beijing \
  --SecurityGroupId sg-xxxxx \
  --IpProtocol "udp" \
  --PortRange "500/500" \
  --SourceCidrIp "0.0.0.0/0" \
  --Description "IKE negotiation"

ve ecs AuthorizeSecurityGroupIngress \
  --Region cn-beijing \
  --SecurityGroupId sg-xxxxx \
  --IpProtocol "udp" \
  --PortRange "4500/4500" \
  --SourceCidrIp "0.0.0.0/0" \
  --Description "IPSec NAT-T"

# For SSL VPN
ve ecs AuthorizeSecurityGroupIngress \
  --Region cn-beijing \
  --SecurityGroupId sg-xxxxx \
  --IpProtocol "udp" \
  --PortRange "1194/1194" \
  --SourceCidrIp "0.0.0.0/0" \
  --Description "SSL VPN OpenVPN"
```

## Customer Gateway Integration

### On-Premises Device Requirements

**IPSec VPN Compatible Devices:**
- Cisco ASA/ISR/CSR
- Juniper SRX/NetScreen
- Palo Alto Networks
- Fortinet FortiGate
- Huawei USG/AR
- H3C MSR
- StrongSwan (Linux)
- pfSense/OPNsense

**Configuration Template:**

```
! IKE Configuration (IKEv2 recommended)
crypto ikev2 proposal VOLC-IKE-PROPOSAL
 encryption aes-256
 integrity sha256
 group 14

crypto ikev2 policy VOLC-IKE-POLICY
 proposal VOLC-IKE-PROPOSAL

crypto ikev2 keyring VOLC-KEYRING
 peer VOLC-PEER
  address <VPN_Gateway_IP>
  pre-shared-key <masked>

crypto ikev2 profile VOLC-IKE-PROFILE
 match identity remote address <VPN_Gateway_IP> 255.255.255.255
 authentication remote pre-share
 authentication local pre-share
 keyring local VOLC-KEYRING

crypto ipsec transform-set VOLC-TRANSFORM esp-aes 256 esp-sha-hmac
 mode tunnel

crypto ipsec profile VOLC-IPSEC-PROFILE
 set transform-set VOLC-TRANSFORM
 set ikev2-profile VOLC-IKE-PROFILE
 set pfs group14

crypto ipsec transform-set VOLC-TRANSFORM esp-aes 256 esp-sha-hmac
 mode tunnel

! Tunnel Interface
interface Tunnel1
 ip address <LOCAL_TUNNEL_IP> <MASK>
 tunnel source <LOCAL_PUBLIC_IP>
 tunnel destination <VPN_Gateway_IP>
 tunnel mode ipsec ipv4
 tunnel protection ipsec profile VOLC-IPSEC-PROFILE
 ip virtual-reassembly

! Routing
ip route <CLOUD_CIDR> <MASK> Tunnel1
```

## SSL VPN Client Configuration

### OpenVPN Client Config

After creating SSL VPN client certificate, configure client:

```
# client.ovpn
client
dev tun
proto udp
remote <VPN_Gateway_IP> 1194
resolv-retry infinite
nobind
persist-key
persist-tun
ca ca.crt
cert client.crt
key client.key
remote-cert-tls server
cipher AES-256-CBC
auth SHA256
verb 3
```

### Client Setup Commands

```bash
# Linux/macOS
sudo openvpn --config client.ovpn

# Windows (OpenVPN GUI)
# Copy files to C:\Program Files\OpenVPN\config\
# Right-click OpenVPN GUI icon and connect
```

## Multi-Region Deployment

### Cross-Region VPN Setup

```bash
# Region A: Create VPN Gateway
ve vpn CreateVpnGateway \
  --Region cn-beijing \
  --VpcId vpc-beijing \
  --SubnetId subnet-beijing \
  --Bandwidth 100 \
  --VpnGatewayName "beijing-gateway"

# Region B: Create VPN Gateway
ve vpn CreateVpnGateway \
  --Region cn-shanghai \
  --VpcId vpc-shanghai \
  --SubnetId subnet-shanghai \
  --Bandwidth 100 \
  --VpnGatewayName "shanghai-gateway"

# Create Customer Gateway in each region pointing to the other
# Create IPSec connections
```

## Automation Examples

### Terraform Provider

```hcl
resource "volcengine_vpn_gateway" "main" {
  vpc_id       = volcengine_vpc.main.id
  subnet_id    = volcengine_subnet.main.id
  bandwidth    = 10
  gateway_name = "prod-vpn-gateway"
  description  = "Production VPN Gateway"
  charge_type  = "PostPaid"
}

resource "volcengine_customer_gateway" "hq" {
  ip_address   = "203.0.113.1"
  gateway_name = "hq-office-gateway"
}

resource "volcengine_vpn_connection" "main" {
  vpn_gateway_id    = volcengine_vpn_gateway.main.id
  customer_gateway_id = volcengine_customer_gateway.hq.id
  connection_name   = "hq-connection"
  pre_shared_key    = var.vpn_psk  # Use sensitive variable
  local_subnet      = ["10.0.0.0/16"]
  remote_subnet     = ["192.168.0.0/16"]
  
  ike_config {
    version    = "ikev2"
    mode       = "main"
    enc_alg    = "aes256"
    auth_alg   = "sha256"
    dh_group   = "group14"
    lifetime   = 86400
  }
  
  ipsec_config {
    enc_alg  = "aes256"
    auth_alg = "sha256"
    pfs      = "group14"
    lifetime = 3600
  }
}
```

### Python SDK Example

```python
import os
from volcengine.vpc.VpcService import VpcService

# Initialize client
client = VpcService()
client.set_ak(os.getenv("VOLCENGINE_ACCESS_KEY"))
client.set_sk(os.getenv("VOLCENGINE_SECRET_KEY"))
client.set_region("cn-beijing")

# Describe VPN Gateways
resp = client.describe_vpn_gateways({
    "RegionId": "cn-beijing",
    "VpnGatewayIds": ["vgw-xxxxx"]
})
print(resp)
```

## IAM Policy Examples

### Read-Only Access

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpn:DescribeVpnGateways",
        "vpn:DescribeCustomerGateways",
        "vpn:DescribeVpnConnections",
        "vpn:DescribeSslVpnServers",
        "vpn:DescribeSslVpnClientCerts"
      ],
      "Resource": "*"
    }
  ]
}
```

### Administrator Access

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpn:*"
      ],
      "Resource": "*"
    }
  ]
}
```

### Operator Access (Create/Modify Only)

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpn:Describe*",
        "vpn:CreateVpnGateway",
        "vpn:CreateCustomerGateway",
        "vpn:CreateVpnConnection",
        "vpn:CreateSslVpnServer",
        "vpn:CreateSslVpnClientCert",
        "vpn:Modify*"
      ],
      "Resource": "*"
    }
  ]
}
```

## SDK Package Reference

| Product | Go SDK Package |
|---------|---------------|
| VPN | `github.com/volcengine/volc-sdk-golang/service/vpn` |
| VPC | `github.com/volcengine/volc-sdk-golang/service/vpc` |
| ECS | `github.com/volcengine/volc-sdk-golang/service/ecs` |

Find all packages at: https://github.com/volcengine/volc-sdk-golang/tree/main/service
