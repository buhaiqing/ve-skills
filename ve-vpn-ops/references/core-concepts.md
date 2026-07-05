# VPN Core Concepts

## Overview

Volcengine VPN (虚拟专用网络) provides secure, encrypted network connections between your on-premises data centers, remote offices, and your Volcengine VPC infrastructure. VPN enables hybrid cloud architectures and secure remote access.

## Architecture Components

### VPN Gateway (VPN网关)

The VPN Gateway is the entry point for VPN connections in your VPC.

| Attribute | Description |
|-----------|-------------|
| **Resource Type** | Regional resource |
| **Association** | Bound to one VPC and one Subnet |
| **Capabilities** | IPSec connections, SSL VPN Server |
| **Bandwidth** | 1-1000 Mbps configurable |
| **Billing** | PostPaid (pay-as-you-go) or PrePaid (subscription) |

**Gateway Types:**
- **Standard**: Supports IPSec connections only
- **SSL**: Supports both IPSec connections and SSL VPN Server

### Customer Gateway (客户网关)

Represents your on-premises VPN device or gateway.

| Attribute | Description |
|-----------|-------------|
| **Resource Type** | Regional resource |
| **Identification** | Static public IP address |
| **Association** | Used with IPSec connections |
| **Multiple Connections** | One Customer Gateway can connect to multiple VPN Gateways |

**Requirements:**
- Public IPv4 address (static IP recommended)
- IKEv1 or IKEv2 support
- IPSec support with configurable encryption/authentication

### IPSec Connection (IPSec连接)

Site-to-site VPN tunnel between VPN Gateway and Customer Gateway.

| Attribute | Description |
|-----------|-------------|
| **Topology** | Site-to-site (network-to-network) |
| **Protocol** | IPSec with IKE negotiation |
| **Redundancy** | Supports dual tunnels (active/standby) |
| **Encryption** | Configurable (AES-128/192/256) |
| **Authentication** | Pre-shared key (PSK) or certificates |

**Tunnel States:**
- `Available` ✅ — Tunnel established and operational
- `Pending` ⏳ — Tunnel is being created
- `Modifying` ⏳ — Tunnel configuration is being updated
- `Deleting` ⏳ — Tunnel is being deleted
- `Deleted` ❌ — Tunnel has been deleted

### SSL VPN Server (SSL VPN服务端)

Provides client-to-site VPN access for remote users.

| Attribute | Description |
|-----------|-------------|
| **Topology** | Client-to-site (remote access) |
| **Protocol** | SSL/TLS over UDP or TCP |
| **Port** | Configurable (default: 1194) |
| **Client Support** | OpenVPN compatible clients |
| **Authentication** | Client certificates |

**Supported Clients:**
- OpenVPN Connect (Windows, macOS, Linux, iOS, Android)
- Tunnelblick (macOS)
- OpenVPN GUI (Windows)

### SSL VPN Client Certificate (SSL VPN客户端证书)

Authentication credential for SSL VPN clients.

| Attribute | Description |
|-----------|-------------|
| **Type** | X.509 certificate with private key |
| **Format** | PEM encoded |
| **Lifecycle** | Issued by SSL VPN Server |
| **Security** | Private key returned only once at creation |

## IKE and IPsec Configuration

### IKE (Internet Key Exchange)

IKE negotiates the security association (SA) for the IPSec tunnel.

**Supported Versions:**
- IKEv1 (Main Mode, Aggressive Mode)
- IKEv2 (recommended)

**IKE Parameters:**
| Parameter | Options | Recommended |
|-----------|---------|-------------|
| Version | ikev1, ikev2 | ikev2 |
| Mode | main, aggressive (IKEv1 only) | main |
| Encryption | aes128, aes192, aes256 | aes256 |
| Authentication | md5, sha1, sha256, sha384, sha512 | sha256 |
| DH Group | group1, group2, group5, group14, group24 | group14 |
| Lifetime | 86400 seconds (24 hours) | 86400 |

### IPsec

IPsec provides the encrypted tunnel for data transmission.

**IPsec Parameters:**
| Parameter | Options | Recommended |
|-----------|---------|-------------|
| Protocol | esp | esp |
| Encryption | aes128, aes192, aes256 | aes256 |
| Authentication | md5, sha1, sha256, sha384, sha512 | sha256 |
| PFS (Perfect Forward Secrecy) | disabled, group1, group2, group5, group14 | group14 |
| Lifetime | 3600 seconds (1 hour) | 3600 |

### Security Recommendations

| Aspect | Recommendation |
|--------|----------------|
| IKE Version | ✅ Use IKEv2 (more secure, faster reconnection) |
| Encryption | ✅ Use AES-256-GCM or AES-256-CBC |
| Authentication | ✅ Use SHA-256 or higher |
| DH Group | ✅ Use group14 (2048-bit) or higher |
| PFS | ✅ Enable PFS with group14 or higher |
| PSK | ⚠️ Use strong pre-shared key (20+ random characters) |
| Certificate | ✅ Use certificates instead of PSK when possible |

## Network Requirements

### Subnet Planning

**Local Subnets (Volcengine VPC):**
- Define CIDR blocks accessible through VPN
- Can specify multiple subnets
- Must not overlap with remote subnets

**Remote Subnets (On-premises):**
- Define CIDR blocks at customer side
- Can specify multiple subnets
- Must not overlap with local subnets

**SSL VPN Client Pool:**
- Dedicated CIDR for VPN clients
- Must not overlap with local or remote subnets
- Recommended: /24 or smaller for client pools

### Port Requirements

**IPSec VPN:**
| Protocol | Port | Direction | Purpose |
|----------|------|-----------|---------|
| UDP | 500 | Inbound/Outbound | IKE negotiation |
| UDP | 4500 | Inbound/Outbound | IPSec NAT-T |
| ESP | 50 | Inbound/Outbound | Encapsulated payload |

**SSL VPN:**
| Protocol | Port | Direction | Purpose |
|----------|------|-----------|---------|
| UDP | 1194 | Inbound/Outbound | Default OpenVPN |
| TCP | 443 | Inbound/Outbound | Alternative OpenVPN |

## Quotas and Limits

| Resource | Default Quota | Adjustable |
|----------|---------------|------------|
| VPN Gateways per region | 5 | Yes |
| IPSec Connections per gateway | 10 | Yes |
| Customer Gateways per region | 20 | Yes |
| SSL VPN Servers per gateway | 5 | Yes |
| SSL VPN Client Certs per server | 100 | Yes |
| Routes per VPN Gateway | 100 | Yes |

## Pricing

**VPN Gateway:**
- Instance fee: Per hour based on bandwidth
- Traffic fee: Outbound traffic (inbound is free)

**IPSec Connection:**
- Connection fee: Per hour per connection
- Traffic fee: Outbound through connection

**SSL VPN Server:**
- Server fee: Per hour per server
- Client fee: Per connected client per hour
- Traffic fee: Outbound through SSL VPN

## Regional Availability

VPN is available in all Volcengine commercial regions:
- cn-beijing (北京)
- cn-shanghai (上海)
- cn-guangzhou (广州)
- cn-hongkong (香港)
- ap-southeast-1 (新加坡)

## Related Resources

| Resource | Relationship |
|----------|--------------|
| VPC | VPN Gateway is associated with a VPC |
| Subnet | VPN Gateway is created in a specific subnet |
| EIP | Can be used for Customer Gateway public IP |
| Security Group | Controls traffic to/from VPN Gateway |
| Route Table | Routes traffic to VPN destinations |
