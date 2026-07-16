---
name: ve-vpn-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) VPN (虚拟专用网络) — VPN gateway lifecycle, IPSec/SSL connections,
  customer gateways, site-to-site VPN tunnels, SSL VPN server/client configuration.
  User mentions VPN, 虚拟专用网络, IPSec, SSL VPN, VPN Gateway, 网关,
  Customer Gateway, or describes secure site-to-site connections, remote access VPN,
  encrypted tunnel scenarios. Not for billing, IAM, or related products that have
  their own ops skills.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "VPN API 2020-04-01 (https://www.volcengine.com/docs/6491/130519)"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve vpn --help` — VPN is supported by the ve CLI.
    See: https://github.com/volcengine/volcengine-cli
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
    - VPN_PORT
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine VPN Operations Skill

## Overview

VPN (虚拟专用网络) on Volcengine (火山引擎) provides secure encrypted network connections including VPN Gateways, IPSec connections (site-to-site VPN), SSL VPN servers/clients, and Customer Gateways. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

> **UX Compliance:** This skill follows the [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md). All operations include onboarding guidance, minimal prompts, smart defaults, clear feedback, and user-friendly error handling.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports VPN. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 VPN-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (VPN), cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine VPN", "火山引擎 VPN", "虚拟专用网络", "IPSec", "SSL VPN", "VPN Gateway", "VPN网关", "Customer Gateway", "客户网关"
- Task involves CRUD or lifecycle operations on **VPN Gateways**: CreateVpnGateway, DescribeVpnGateways, DeleteVpnGateway, ModifyVpnGatewayAttribute
- Task involves **IPSec Connections**: CreateVpnConnection, DescribeVpnConnections, DeleteVpnConnection, ModifyVpnConnectionAttribute
- Task involves **SSL VPN Servers**: CreateSslVpnServer, DescribeSslVpnServers, DeleteSslVpnServer, ModifySslVpnServer
- Task involves **SSL VPN Clients**: CreateSslVpnClientCert, DescribeSslVpnClientCerts, DeleteSslVpnClientCert
- Task involves **Customer Gateways**: CreateCustomerGateway, DescribeCustomerGateways, DeleteCustomerGateway
- Task involves **VPN connections**: site-to-site VPN, IPSec tunnel, IKE configuration, IPsec configuration
- Task involves **remote access VPN**: SSL VPN, client certificate, VPN client configuration
- Task keywords: tunnel, 隧道, encryption, 加密, IKE, IPsec, pre-shared key, certificate

### SHOULD NOT Use This Skill When

- Task is purely billing / account management → delegate to billing ops
- Task is IAM / permission model only → delegate to: `ve-iam-ops` (when present)
- Task is about **VPC networking** (routing, subnets) → delegate to: `ve-vpc-ops`
- Task is about **EIP** (Elastic IP) → delegate to: `ve-eip-ops`
- Task is about **NAT Gateway** → delegate to: `ve-nat-ops`
- Task is about **Load Balancer** → delegate to: `ve-clb-ops`
- User insists on **console-only** flows with no API → state limitation

### Delegation Rules

- VPN Gateway requires VPC → verify VPC exists via `ve-vpc-ops`
- IPSec connection requires Customer Gateway and VPN Gateway → verify both exist
- SSL VPN requires VPN Gateway with SSL capability → verify gateway type
- SSL VPN clients require SSL VPN Server → verify server exists first

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.vpn_gateway_name}}` | User-supplied VPN Gateway name | Ask once; reuse |
| `{{user.vpn_gateway_id}}` | User-supplied VPN Gateway ID | Format `vgw-xxxxxxxxx` |
| `{{user.customer_gateway_name}}` | User-supplied Customer Gateway name | Ask once; reuse |
| `{{user.customer_gateway_id}}` | User-supplied Customer Gateway ID | Format `cgw-xxxxxxxxx` |
| `{{user.vpn_connection_name}}` | User-supplied IPSec connection name | Ask once; reuse |
| `{{user.vpn_connection_id}}` | User-supplied IPSec connection ID | Format `vpn-xxxxxxxxx` |
| `{{user.ssl_vpn_server_name}}` | User-supplied SSL VPN Server name | Ask once; reuse |
| `{{user.ssl_vpn_server_id}}` | User-supplied SSL VPN Server ID | Format `ssl-xxxxxxxxx` |
| `{{user.ssl_vpn_client_cert_name}}` | User-supplied SSL VPN client cert name | Ask once; reuse |
| `{{user.vpc_id}}` | User-supplied VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | User-supplied Subnet ID | Format `subnet-xxxxxxxxx` |
| `{{user.bandwidth}}` | VPN Gateway bandwidth in Mbps | Ask once; reuse; valid range 1-1000 |
| `{{user.customer_ip}}` | Customer Gateway public IPv4 address | Ask once; reuse; must be public routable |
| `{{user.pre_shared_key}}` | IPSec pre-shared key (PSK) | Ask once; reuse; **NEVER log or echo** |
| `{{user.local_cidr}}` | Local subnet CIDR(s) for IPSec | Ask once; reuse; non-overlapping |
| `{{user.remote_cidr}}` | Remote subnet CIDR(s) for IPSec | Ask once; reuse; non-overlapping |
| `{{user.local_subnet}}` | Local subnets accessible via SSL VPN | Ask once; reuse; valid CIDR |
| `{{user.client_ip_pool}}` | Client IP pool for SSL VPN clients | Ask once; reuse; valid CIDR |
| `{{env.VPN_PORT}}` | SSL VPN server port (from environment) | NEVER ask the user; fail if unset |
| `{{output.vpn_gateway_id}}` | From CreateVpnGateway response | Parse from `$.Result.VpnGatewayId` |
| `{{output.customer_gateway_id}}` | From CreateCustomerGateway response | Parse from `$.Result.CustomerGatewayId` |
| `{{output.vpn_connection_id}}` | From CreateVpnConnection response | Parse from `$.Result.VpnConnectionId` |
| `{{output.ssl_vpn_server_id}}` | From CreateSslVpnServer response | Parse from `$.Result.SslVpnServerId` |
| `{{output.ssl_vpn_client_cert_id}}` | From CreateSslVpnClientCert response | Parse from `$.Result.SslVpnClientCertId` |
| `{{output.certificate}}` | Client certificate (PEM) from CreateSslVpnClientCert | Parse from `$.Result.Certificate`; save immediately |
| `{{output.private_key}}` | Client private key (PEM) from CreateSslVpnClientCert | Parse from `$.Result.PrivateKey`; save securely; returned only once |
| `{{output.ca_cert}}` | CA certificate (PEM) from CreateSslVpnClientCert | Parse from `$.Result.CaCert`; save immediately |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, `PreSharedKey`, `pre_shared_key`, `Certificate`, `PrivateKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌
> - Never log pre-shared keys: use `<masked>` for PSK in all output

## API and Response Conventions (Agent-Readable)

- **Volcengine VPN OpenAPI (2020-04-01)** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `vpc.volcengineapi.com` (default: `open.volcengineapi.com`)
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format (e.g. `2026-04-28T10:00:00Z`).

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateVpnGateway | `$.Result.VpnGatewayId` | string | Created VPN Gateway ID |
| DescribeVpnGateways | `$.Result.VpnGateways` | array | VPN Gateway list |
| DescribeVpnGateways | `$.Result.VpnGateways[].VpnGatewayId` | string | VPN Gateway ID |
| DescribeVpnGateways | `$.Result.VpnGateways[].VpnGatewayName` | string | VPN Gateway name |
| DescribeVpnGateways | `$.Result.VpnGateways[].VpcId` | string | Associated VPC ID |
| DescribeVpnGateways | `$.Result.VpnGateways[].Status` | string | Gateway status |
| CreateCustomerGateway | `$.Result.CustomerGatewayId` | string | Created Customer Gateway ID |
| DescribeCustomerGateways | `$.Result.CustomerGateways` | array | Customer Gateway list |
| CreateVpnConnection | `$.Result.VpnConnectionId` | string | Created IPSec connection ID |
| DescribeVpnConnections | `$.Result.VpnConnections` | array | IPSec connection list |
| DescribeVpnConnections | `$.Result.VpnConnections[].Status` | string | Connection status |
| CreateSslVpnServer | `$.Result.SslVpnServerId` | string | Created SSL VPN Server ID |
| DescribeSslVpnServers | `$.Result.SslVpnServers` | array | SSL VPN Server list |
| CreateSslVpnClientCert | `$.Result.SslVpnClientCertId` | string | Created SSL VPN client cert ID |
| CreateSslVpnClientCert | `$.Result.Certificate` | string | Client certificate (PEM) — parse to `{{output.certificate}}` |
| CreateSslVpnClientCert | `$.Result.PrivateKey` | string | Client private key (PEM) — parse to `{{output.private_key}}`; **SAVE IMMEDIATELY**, returned only once |
| CreateSslVpnClientCert | `$.Result.CaCert` | string | CA certificate (PEM) — parse to `{{output.ca_cert}}` |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| CreateVpnGateway | — | `Available` | 5s | 300s |
| DeleteVpnGateway | `Available` | absent | 5s | 300s |
| CreateCustomerGateway | — | `Available` | 5s | 60s |
| CreateVpnConnection | — | `Available` | 5s | 300s |
| DeleteVpnConnection | `Available` | absent | 5s | 300s |
| CreateSslVpnServer | — | `Available` | 5s | 300s |
| DeleteSslVpnServer | `Available` | absent | 5s | 300s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and manage Volcengine (火山引擎) VPN resources including VPN Gateways, IPSec connections, SSL VPN servers/clients, and Customer Gateways using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY` (masked as `<masked>`)
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
# Verify credentials (existence only, never echo values)
test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ Credentials configured"

# List VPN Gateways
ve vpn DescribeVpnGateways --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all VPN Gateways in the configured region
ve vpn DescribeVpnGateways --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand VPN architecture
- [Common Operations](#execution-flows) — Create, manage, and manage VPN connections
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk |
|-----------|-------------|------------|------|
| CreateVpnGateway | Create a VPN Gateway | Medium | Low |
| DescribeVpnGateways | Query VPN Gateway list | Low | None |
| DeleteVpnGateway | Delete a VPN Gateway | Low | **High** — irreversible |
| ModifyVpnGatewayAttribute | Modify VPN Gateway attributes | Low | Low |
| CreateCustomerGateway | Create a Customer Gateway | Medium | Low |
| DescribeCustomerGateways | Query Customer Gateway list | Low | None |
| DeleteCustomerGateway | Delete a Customer Gateway | Low | **High** — irreversible |
| CreateVpnConnection | Create an IPSec connection | High | Medium |
| DescribeVpnConnections | Query IPSec connection list | Low | None |
| DeleteVpnConnection | Delete an IPSec connection | Low | **High** — irreversible |
| ModifyVpnConnectionAttribute | Modify IPSec connection | Medium | Medium |
| CreateSslVpnServer | Create an SSL VPN Server | Medium | Low |
| DescribeSslVpnServers | Query SSL VPN Server list | Low | None |
| DeleteSslVpnServer | Delete an SSL VPN Server | Low | **High** — irreversible |
| CreateSslVpnClientCert | Create SSL VPN client certificate | Medium | Low |
| DescribeSslVpnClientCerts | Query SSL VPN client certs | Low | None |
| DeleteSslVpnClientCert | Delete SSL VPN client cert | Low | Medium |
| DownloadSslVpnClientConfig | Download SSL VPN client config | Low | None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-27 | Initial release with VPN Gateway, IPSec, SSL VPN, and Customer Gateway management |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteVpnGateway, DeleteCustomerGateway, DeleteVpnConnection, DeleteSslVpnServer, DeleteSslVpnClientCert | 1.0 |
| **State-changing** | CreateVpnConnection, ModifyVpnConnectionAttribute, CreateSslVpnClientCert | 1.0 |
| **Mutating** | CreateVpnGateway, CreateCustomerGateway, CreateSslVpnServer | ≥0.5 |
| **Read-only** | DescribeVpnGateways, DescribeCustomerGateways, DescribeVpnConnections, DescribeSslVpnServers, DescribeSslVpnClientCerts, DownloadSslVpnClientConfig | ≥0 |

Safety: DeleteVpnGateway ALL connections disconnected. CreateSslVpnClientCert: PrivateKey never in trace (masked). VOLCENGINE_SECRET_KEY never.

### Cross-skill: VPC→ve-vpc-ops, Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (ve CLI primary + JIT Go SDK fallback) → Validate → Recover**.

### Operation: DescribeVpnGateways — Query VPN Gateway List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` or `{{env.VOLCENGINE_REGION}}` is set | Valid region | Suggest: `ve vpc DescribeRegions` |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution — CLI (`ve`)

```bash
# List all VPN Gateways (JSON output by default)
ve vpn DescribeVpnGateways --Region "{{user.region}}"

# Filter by VPN Gateway ID
ve vpn DescribeVpnGateways --Region "{{user.region}}" --VpnGatewayIds '["{{user.vpn_gateway_id}}"]'

# Filter by VPC ID
ve vpn DescribeVpnGateways --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/vpn"
)

func main() {
    instance := vpn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":     os.Getenv("VOLCENGINE_REGION"),
        "MaxResults": 50,
    }

    resp, err := instance.Client.Request("DescribeVpnGateways", nil, params)
    if err != nil {
        log.Fatalf("Failed to describe VPN gateways: %v", err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

1. Check `$.Result.TotalCount` for total matching gateways
2. Parse `$.Result.VpnGateways[]` for gateway details
3. Report gateway count, IDs, names, VPC associations, and statuses

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action |
|--------------|-------------|--------------|
| `InvalidRegion.NotFound` | 0 | List valid regions via `DescribeRegions`; HALT |
| `Unauthorized` | 0 | HALT; check IAM permissions |
| `InternalError` | 3 | Retry with exponential backoff; HALT after 3 |
| Throttling / 429 | 3 | Back off (2s, 4s, 8s); retry |

---

### Operation: CreateVpnGateway — Create VPN Gateway

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| VPC exists | `ve vpc DescribeVpcs --VpcIds '["{{user.vpc_id}}"]'` | VPC found | HALT; create VPC first |
| Subnet exists | `ve vpc DescribeSubnets --SubnetIds '["{{user.subnet_id}}"]'` | Subnet found | HALT; create subnet first |
| Bandwidth valid | Verify bandwidth within range (1-1000 Mbps) | Valid | HALT; adjust bandwidth |
| Quota | Check VPN Gateway quota per region | Sufficient | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create VPN Gateway with minimal parameters
ve vpn CreateVpnGateway \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --Bandwidth "{{user.bandwidth}}" \
  --VpnGatewayName "{{user.vpn_gateway_name}}"

# Create with description
ve vpn CreateVpnGateway \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --Bandwidth 10 \
  --VpnGatewayName "prod-vpn-gateway" \
  --Description "Production VPN Gateway"
```

**Common parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `VpcId` | VPC ID | `vpc-xxxxx` |
| `SubnetId` | Subnet ID | `subnet-xxxxx` |
| `Bandwidth` | Bandwidth in Mbps (1-1000) | `10` |
| `VpnGatewayName` | Gateway name | `my-vpn-gateway` |
| `Description` | Description | `Production VPN` |
| `ChargeType` | Billing: PostPaid/PrePaid | `PostPaid` |
| `Period` | PrePaid period in months | `1` |

#### Post-execution Validation

1. Parse `{{output.vpn_gateway_id}}` from `$.Result.VpnGatewayId`
2. Poll DescribeVpnGateways until status is `Available`:

```bash
for i in $(seq 1 60); do
  STATUS=$(ve vpn DescribeVpnGateways --Region "{{user.region}}" --VpnGatewayIds '["{{output.vpn_gateway_id}}"]' | jq -r '.Result.VpnGateways[0].Status')
  [ "$STATUS" = "Available" ] && break
  echo "Current: $STATUS (poll $i/60)"
  sleep 5
done
[ "$STATUS" != "Available" ] && echo "[ERROR] Gateway failed to become Available (current: $STATUS)" && exit 1
```

3. On success, report gateway ID, associated VPC, bandwidth, and status
4. On terminal failure, go to Failure Recovery

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidVpcId.NotFound` | 0 | HALT; verify VPC ID | `[ERROR] InvalidVpcId: VPC not found. Use DescribeVpcs to find valid IDs.` |
| `InvalidSubnetId.NotFound` | 0 | HALT; verify subnet ID | `[ERROR] InvalidSubnetId: Subnet not found. Use DescribeSubnets to find valid IDs.` |
| `QuotaExceeded.VpnGateway` | 0 | HALT | `[ERROR] QuotaExceeded: VPN Gateway quota reached. Request increase or delete unused.` |
| `InvalidBandwidth.ValueNotSupported` | 0 | HALT | `[ERROR] InvalidBandwidth: Use 1-1000 Mbps.` |
| `InvalidVpnGatewayName.Duplicate` | 0 | HALT | `[ERROR] DuplicateName: Use a unique name.` |
| `InsufficientBalance` | 0 | HALT | `[ERROR] InsufficientBalance: Recharge your account.` |
| `InternalError` | 3 | Retry with backoff | `[ERROR] InternalError: Server-side error. Will retry.` |
| `Throttling` | 3 | Exponential backoff | `WARNING Rate limit reached. Retrying...` |

---

### Operation: DeleteVpnGateway — Delete VPN Gateway

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** VPN Gateway `{{user.vpn_gateway_id}}` (name: `{{user.vpn_gateway_name}}`).
> This is **IRREVERSIBLE** — all IPSec connections and SSL VPN Servers attached to this gateway are destroyed, and all VPN tunnels are terminated.
> Type the gateway ID `{{user.vpn_gateway_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation (type-to-confirm above) — **MUST NOT** proceed without clear assent.
2. **MUST** verify the gateway has no dependent resources before deletion:

```bash
# Check for IPSec connections — MUST be empty (TotalCount = 0)
ve vpn DescribeVpnConnections --Region "{{user.region}}" --VpnGatewayId "{{user.vpn_gateway_id}}"

# Check for SSL VPN Servers — MUST be empty (TotalCount = 0)
ve vpn DescribeSslVpnServers --Region "{{user.region}}" --VpnGatewayId "{{user.vpn_gateway_id}}"
```

3. If either returns non-zero resources, **HALT** and warn the user to delete the dependent resources first.

#### Execution — CLI (`ve`)

```bash
ve vpn DeleteVpnGateway --Region "{{user.region}}" --VpnGatewayId "{{user.vpn_gateway_id}}"
```

#### Post-execution Validation

Poll DescribeVpnGateways until gateway not found (404) or not in result list (max 300s).

---

### Operation: CreateCustomerGateway — Create Customer Gateway

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| IP address valid | Validate IP format | Valid IPv4 | HALT; provide valid IP |
| IP is public | Verify IP is public routable | Public IP | HALT; use public IP |

#### Execution — CLI (`ve`)

```bash
# Create Customer Gateway with static IP
ve vpn CreateCustomerGateway \
  --Region "{{user.region}}" \
  --IpAddress "{{user.customer_ip}}" \
  --CustomerGatewayName "{{user.customer_gateway_name}}"

# Create with description
ve vpn CreateCustomerGateway \
  --Region "{{user.region}}" \
  --IpAddress "203.0.113.1" \
  --CustomerGatewayName "hq-office-gateway" \
  --Description "HQ Office Customer Gateway"
```

#### Post-execution Validation

1. Parse `{{output.customer_gateway_id}}` from response
2. Verify gateway is `Available`

---

### Operation: DeleteCustomerGateway — Delete Customer Gateway

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** Customer Gateway `{{user.customer_gateway_id}}` (name: `{{user.customer_gateway_name}}`).
> This is **IRREVERSIBLE** — any IPSec connections referencing this gateway are severed.
> Type the gateway ID `{{user.customer_gateway_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation (type-to-confirm above) — **MUST NOT** proceed without clear assent.
2. **MUST** verify the Customer Gateway is not associated with any IPSec connections:

```bash
# Check for IPSec connections — MUST be empty (TotalCount = 0)
ve vpn DescribeVpnConnections --Region "{{user.region}}" --CustomerGatewayId "{{user.customer_gateway_id}}"
```

3. If non-zero connections are returned, **HALT** and warn the user to delete the IPSec connections first.

#### Execution

```bash
ve vpn DeleteCustomerGateway --Region "{{user.region}}" --CustomerGatewayId "{{user.customer_gateway_id}}"
```

---

### Operation: CreateVpnConnection — Create IPSec Connection

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| VPN Gateway exists | DescribeVpnGateways with ID | Gateway `Available` | HALT |
| Customer Gateway exists | DescribeCustomerGateways with ID | Gateway `Available` | HALT |
| PSK provided | `{{user.pre_shared_key}}` is set | Non-empty | HALT |
| IKE config valid | Verify IKE version, encryption, DH group | Valid config | HALT |
| IPsec config valid | Verify encryption, PFS | Valid config | HALT |

#### Execution — CLI (`ve`)

```bash
# Create IPSec connection with IKEv2
ve vpn CreateVpnConnection \
  --Region "{{user.region}}" \
  --VpnGatewayId "{{user.vpn_gateway_id}}" \
  --CustomerGatewayId "{{user.customer_gateway_id}}" \
  --VpnConnectionName "{{user.vpn_connection_name}}" \
  --PreSharedKey "{{user.pre_shared_key}}" \
  --LocalSubnet '["{{user.local_cidr}}"]' \
  --RemoteSubnet '["{{user.remote_cidr}}"]' \
  --IKEConfig.Version "ikev2" \
  --IKEConfig.Mode "main" \
  --IKEConfig.EncAlg "aes256" \
  --IKEConfig.AuthAlg "sha256" \
  --IKEConfig.DhGroup "group14" \
  --IKEConfig.Lifetime "86400" \
  --IPsecConfig.EncAlg "aes256" \
  --IPsecConfig.AuthAlg "sha256" \
  --IPsecConfig.Pfs "group14" \
  --IPsecConfig.Lifetime "3600"
```

**Important parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `VpnGatewayId` | VPN Gateway ID | `vgw-xxxxx` |
| `CustomerGatewayId` | Customer Gateway ID | `cgw-xxxxx` |
| `PreSharedKey` | Pre-shared key (PSK) | `<masked>` — never log this |
| `LocalSubnet` | Local CIDR(s) | `["10.0.0.0/16"]` |
| `RemoteSubnet` | Remote CIDR(s) | `["192.168.0.0/16"]` |
| `IKEConfig.Version` | IKE version (ikev1/ikev2) | `ikev2` |
| `IKEConfig.EncAlg` | Encryption (aes128/aes192/aes256) | `aes256` |
| `IKEConfig.AuthAlg` | Authentication (md5/sha1/sha256/sha384) | `sha256` |
| `IPsecConfig.EncAlg` | IPsec encryption | `aes256` |
| `IPsecConfig.Pfs` | PFS group (disabled/group2/group5/group14) | `group14` |

> **Security:** Never log or display `PreSharedKey`. Use `<masked>` in all output.

#### Post-execution Validation

1. Parse `{{output.vpn_connection_id}}` from `$.Result.VpnConnectionId`
2. Poll DescribeVpnConnections until status is `Available`:

```bash
for i in $(seq 1 60); do
  STATUS=$(ve vpn DescribeVpnConnections --Region "{{user.region}}" --VpnConnectionIds '["{{output.vpn_connection_id}}"]' | jq -r '.Result.VpnConnections[0].Status')
  [ "$STATUS" = "Available" ] && break
  echo "Current: $STATUS (poll $i/60)"
  sleep 5
done
```

3. Report connection ID, status, and tunnel information

#### Failure Recovery

| Error Pattern | Max Retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidVpnGatewayId.NotFound` | 0 | HALT; verify gateway | `[ERROR] InvalidVpnGatewayId: VPN Gateway not found.` |
| `InvalidCustomerGatewayId.NotFound` | 0 | HALT; verify customer gateway | `[ERROR] InvalidCustomerGatewayId: Customer Gateway not found.` |
| `InvalidPreSharedKey.Format` | 0 | HALT | `[ERROR] InvalidPreSharedKey: PSK format invalid. Length: 1-128 chars.` |
| `VpnGateway.NotAvailable` | 0 | Wait and retry | `[ERROR] VpnGateway.NotAvailable: VPN Gateway not in Available state.` |
| `CustomerGateway.NotAvailable` | 0 | Wait and retry | `[ERROR] CustomerGateway.NotAvailable: Customer Gateway not in Available state.` |
| `QuotaExceeded.VpnConnection` | 0 | HALT | `[ERROR] QuotaExceeded: IPSec connection quota reached.` |
| `InvalidLocalSubnet.Malformed` | 0 | HALT | `[ERROR] InvalidLocalSubnet: CIDR format invalid.` |
| `InvalidRemoteSubnet.Malformed` | 0 | HALT | `[ERROR] InvalidRemoteSubnet: CIDR format invalid.` |
| `SubnetConflict` | 0 | HALT | `[ERROR] SubnetConflict: CIDRs overlap.` |
| `InternalError` | 3 | Retry with backoff | `[ERROR] InternalError: Server-side error. Will retry.` |
| `Throttling` | 3 | Exponential backoff | `WARNING Rate limit reached. Retrying...` |

---

### Operation: DeleteVpnConnection — Delete IPSec Connection

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** IPSec connection `{{user.vpn_connection_id}}` (name: `{{user.vpn_connection_name}}`).
> This is **IRREVERSIBLE** — the VPN tunnel is terminated immediately and all encrypted traffic between the two sites stops.
> Type the connection ID `{{user.vpn_connection_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation (type-to-confirm above) — **MUST NOT** proceed without clear assent.
2. **MUST** warn: this will terminate the VPN tunnel immediately.
3. **SHOULD** verify the connection exists and is in a deletable state before deletion:

```bash
# Verify the IPSec connection is present and check its status
ve vpn DescribeVpnConnections --Region "{{user.region}}" --VpnConnectionIds '["{{user.vpn_connection_id}}"]'
```

4. If the connection is not found, **HALT** and report — nothing to delete.

#### Execution

```bash
ve vpn DeleteVpnConnection --Region "{{user.region}}" --VpnConnectionId "{{user.vpn_connection_id}}"
```

#### Validation

Poll DescribeVpnConnections until connection not found.

---

### Operation: CreateSslVpnServer — Create SSL VPN Server

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Env vars set | Non-empty | HALT |
| VPN Gateway exists | DescribeVpnGateways | Gateway `Available` | HALT |
| VPN Gateway supports SSL | Verify gateway type | SSL-enabled | HALT; create new gateway |
| Local subnet valid | CIDR format | Valid CIDR | HALT |
| Client IP pool valid | CIDR format | Valid CIDR | HALT |

#### Execution — CLI (`ve`)

```bash
# Create SSL VPN Server
ve vpn CreateSslVpnServer \
  --Region "{{user.region}}" \
  --VpnGatewayId "{{user.vpn_gateway_id}}" \
  --SslVpnServerName "{{user.ssl_vpn_server_name}}" \
  --LocalSubnets '["{{user.local_subnet}}"]' \
  --ClientIpPool "{{user.client_ip_pool}}" \
  --SslVpnServerProtocol "UDP" \
  --SslVpnServerPort "{{env.VPN_PORT}}" \
  --Cipher "AES-256-CBC" \
  --Auth "SHA256" \
  --Compress "false"
```

**Parameters:**
| Parameter | Description | Example |
|-----------|-------------|---------|
| `VpnGatewayId` | VPN Gateway ID | `vgw-xxxxx` |
| `LocalSubnets` | Subnets accessible via VPN | `["10.0.0.0/16"]` |
| `ClientIpPool` | IP pool for VPN clients | `10.0.100.0/24` |
| `SslVpnServerProtocol` | Protocol (UDP/TCP) | `UDP` |
| `SslVpnServerPort` | Port number | `{{env.VPN_PORT}}` |
| `Cipher` | Encryption cipher | `AES-256-CBC` |
| `Auth` | Auth digest | `SHA256` |

---

### Operation: DeleteSslVpnServer — Delete SSL VPN Server

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** SSL VPN Server `{{user.ssl_vpn_server_id}}` (name: `{{user.ssl_vpn_server_name}}`).
> This is **IRREVERSIBLE** — all client certificates under this server are invalidated and all remote-access VPN clients lose connectivity.
> Type the server ID `{{user.ssl_vpn_server_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation (type-to-confirm above) — **MUST NOT** proceed without clear assent.
2. **MUST** verify no client certificates are associated (or warn about them):

```bash
# Check for client certificates — MUST be empty (TotalCount = 0)
ve vpn DescribeSslVpnClientCerts --Region "{{user.region}}" --SslVpnServerId "{{user.ssl_vpn_server_id}}"
```

3. If non-zero certificates are returned, **HALT** and warn the user to revoke/delete the client certs first.

#### Execution

```bash
ve vpn DeleteSslVpnServer --Region "{{user.region}}" --SslVpnServerId "{{user.ssl_vpn_server_id}}"
```

---

### Operation: CreateSslVpnClientCert — Create SSL VPN Client Certificate

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| SSL VPN Server exists | DescribeSslVpnServers | Server `Available` | HALT |

#### Execution — CLI (`ve`)

```bash
ve vpn CreateSslVpnClientCert \
  --Region "{{user.region}}" \
  --SslVpnServerId "{{user.ssl_vpn_server_id}}" \
  --SslVpnClientCertName "{{user.ssl_vpn_client_cert_name}}"
```

#### Post-execution Validation

> **CRITICAL:** The API returns `Certificate` and `PrivateKey` **only once**. User MUST save these immediately.

1. Parse `{{output.ssl_vpn_client_cert_id}}` from response
2. Parse `Certificate` (PEM format) — save to file
3. Parse `PrivateKey` (PEM format) — save to file securely
4. Parse `CaCert` (CA certificate) — save to file

```bash
# Save certificates (example)
echo "{{output.certificate}}" > client.crt
echo "{{output.private_key}}" > client.key
echo "{{output.ca_cert}}" > ca.crt
chmod 600 client.key
```

**Warning displayed to user:**
> **IMPORTANT:** The private key is returned **only once** and cannot be retrieved again. Save it securely immediately. If lost, you must delete this certificate and create a new one.

---

### Operation: DeleteSslVpnClientCert — Delete SSL VPN Client Certificate

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** SSL VPN client certificate `{{user.ssl_vpn_client_cert_id}}` (name: `{{user.ssl_vpn_client_cert_name}}`).
> This is **IRREVERSIBLE** — the client will lose VPN access immediately and the certificate cannot be recovered (the private key was returned only once at creation).
> Type the cert ID `{{user.ssl_vpn_client_cert_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation (type-to-confirm above) — **MUST NOT** proceed without clear assent.
2. **MUST** warn: client will lose VPN access immediately.
3. **SHOULD** verify the certificate exists before deletion:

```bash
# Verify the client certificate is present
ve vpn DescribeSslVpnClientCerts --Region "{{user.region}}" --SslVpnServerId "{{user.ssl_vpn_server_id}}"
```

4. If the certificate is not found, **HALT** and report — nothing to delete.

#### Execution

```bash
ve vpn DeleteSslVpnClientCert \
  --Region "{{user.region}}" \
  --SslVpnServerId "{{user.ssl_vpn_server_id}}" \
  --SslVpnClientCertId "{{user.ssl_vpn_client_cert_id}}"
```

---

## Failure Feedback Format

When an operation fails, present the result to the user using this standardized block so failures are actionable and consistent:

```
❌ **Operation Failed: <OperationName>**
- **Error code**: `<code>` (from the table below)
- **What happened**: <one-line plain-language explanation>
- **Why it matters**: <impact on the user's resources / connectivity>
- **Action required**: <concrete next step — e.g. fix input, wait for state, or HALT and escalate>
- **Retry policy**: <0 retries; HALT> or <N retries with backoff> — state explicitly, never silent-retry
```

Rules:
- **MUST** surface the raw error code from the API — do not paraphrase into a generic "something went wrong".
- **MUST** state the retry policy (0 retries → HALT, or bounded retries) so the user knows whether the action auto-repeats.
- **MUST NOT** log or echo `{{user.pre_shared_key}}` / `{{output.private_key}}` in any failure output.
- On `**HALT**` conditions, stop the runbook and wait for explicit user direction — do not fall through to the next operation.

## Error Taxonomy

| 错误码 | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidVpnGatewayId` | VPN Gateway ID 不存在或格式错误 | 0 retries; **HALT** — 确认 VpnGatewayId 格式 `(vgw-xxx)` |
| `IncorrectVpnGatewayStatus` | VPN 网关状态不允许当前操作 | 0 retries; **HALT** — 等待状态变为 `Available` |
| `InvalidCustomerGatewayId` | Customer Gateway ID 不存在或格式错误 | 0 retries; **HALT** — 确认 CustomerGatewayId 格式 `(cgw-xxx)` |
| `InvalidVpnConnectionId` | VPN 连接 ID 不存在或格式错误 | 0 retries; **HALT** — 确认 VpnConnectionId 格式 `(vpn-xxx)` |
| `QuotaExceeded.VpnConnection` | IPSec 连接数量已达配额上限 | 0 retries; **HALT** — 删除未使用的连接或提额 |
| `IpsecConfig.Invalid` | IPsec 配置参数无效（加密/认证/生命周期） | 0 retries; **HALT** — 检查 EncAlg/AuthAlg/Pfs/Lifetime 取值 |
| `IkeConfig.Invalid` | IKE 配置参数无效（版本/模式/DH 分组） | 0 retries; **HALT** — 检查 Version/Mode/DhGroup/EncAlg |
| `TunnelStatus.Error` | VPN 隧道状态异常（对端不可达/协商失败） | 0 retries; **HALT** — 检查对端网关连通性和 PSK 配置 |
| `InvalidVpcId` | VPC ID 不存在或不在当前区域 | 0 retries; **HALT** — 通过 `ve-vpc-ops` 确认 VPC |
| `InvalidSubnetId` | Subnet ID 不存在或不匹配 VPC | 0 retries; **HALT** — 确认子网存在 |
| `Throttling` | 请求限流 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 记录 RequestId |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Operational Best Practices

- **Security:** Use IKEv2 with AES-256 encryption and SHA-256 authentication
- **PSK Management:** Store pre-shared keys securely; rotate periodically
- **Certificate Management:** Set expiration alerts for SSL VPN client certificates
- **Network Planning:** Use non-overlapping CIDR blocks for local and remote subnets
- **High Availability:** Deploy VPN Gateways in multi-zone configurations where available
- **Monitoring:** Enable VPN connection metrics and set up alerts for tunnel failures
- **Least privilege:** Use IAM policies scoped to VPN APIs only
