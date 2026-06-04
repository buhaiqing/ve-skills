---
name: ve-vpn-ops-rubric
description: GCL rubric for ve-vpn-ops. SSL cert keys must be masked.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-vpn-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-vpn-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteVpnGateway, DeleteCustomerGateway, DeleteVpnConnection, DeleteSslVpnServer, DeleteSslVpnClientCert | 3 | 1.0 |
| **State-changing** | CreateVpnConnection, ModifyVpnConnectionAttribute, CreateSslVpnClientCert | 3 | 1.0 |
| **Mutating** | CreateVpnGateway, CreateCustomerGateway, CreateSslVpnServer | 3 | ≥0.5 |
| **Read-only** | DescribeVpnGateways, DescribeCustomerGateways, DescribeVpnConnections, DescribeSslVpnServers, DescribeSslVpnClientCerts, DownloadSslVpnClientConfig | 3 | ≥0 |
## 1-5. Dimensions (standard)
Safety: DeleteVpnGateway warn about ALL connections disconnected. CreateSslVpnClientCert: private key returned ONCE — NEVER in trace. VOLCENGINE_SECRET_KEY never in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-vpn-ops |