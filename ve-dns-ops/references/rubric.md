---
name: ve-dns-ops-rubric
description: GCL rubric for ve-dns-ops. Optional tier, max_iter=5.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-dns-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---
# GCL Rubric — ve-dns-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteDNSRecord, DeleteDNSDomain | 5 | 1.0 |
| **State-changing** | CreateDNSRecord, ModifyDNSRecord, ModifyDNSDomain | 5 | 1.0 |
| **Mutating** | CreateDNSDomain | 5 | ≥0.5 |
| **Read-only** | DescribeDNSDomains, DescribeDNSRecords, DescribeDomainStatistics | 5 | ≥0 |
Safety: DeleteDNSDomain ALL records under domain LOST — DNS resolution breaks. DeleteDNSRecord — specific record stops resolving. ModifyDNSRecord with wrong value causes resolution failures. VOLCENGINE_SECRET_KEY never.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-dns-ops |