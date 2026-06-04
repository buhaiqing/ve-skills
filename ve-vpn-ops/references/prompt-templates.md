---
name: ve-vpn-ops-prompt-templates
description: GCL prompt templates for ve-vpn-ops. SSL client cert private key must NEVER be in trace.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vpn-ops
  gcl_role: prompt_skeletons
  default_max_iter: 3
---
# GCL Prompt Templates — ve-vpn-ops
## 1. Generator Prompt
Generator for ve-vpn-ops. Execute via `ve vpn <Action>` or Go SDK. Dual-path. Trace. For CreateSslVpnClientCert: output private key ONCE to user, NEVER in trace. NEVER print VOLCENGINE_SECRET_KEY.
## 2. Critic Prompt (MUST NOT see user request)
Independent VPN auditor. Verify CreateSslVpnClientCert: PrivateKey masked in trace (only sha256:<prefix>). Verify DeleteVpnGateway: connection loss warning.
## 3. Orchestrator Prompt
Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
## 4. Safety Prompts
### 4.1 DeleteVpnGateway
[DESTRUCTIVE] Delete VPN gateway {{user.vpn_gw_id}}. ALL IPSec + SSL connections disconnected. Remote sites lose connectivity. Type "yes", or "cancel".
### 4.2 DeleteVpnConnection
[DESTRUCTIVE] Delete IPSec connection {{user.vpn_conn_id}} to {{user.customer_gw_id}}. Site-to-site tunnel DOWN. Type "yes", or "cancel".
### 4.3 DeleteSslVpnServer
[DESTRUCTIVE] Delete SSL VPN server {{user.ssl_vpn_id}}. ALL remote users disconnected. Client certs cannot be reused. Type "yes", or "cancel".
### 4.4 CreateSslVpnClientCert
[SECRET-GENERATING] Create SSL VPN client cert for server {{user.ssl_vpn_id}}. Private key + cert returned ONCE. Save immediately — cannot be retrieved later. Type "yes", or "cancel".
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-vpn-ops |