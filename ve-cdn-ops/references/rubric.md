---
name: ve-cdn-ops-rubric
description: GCL rubric for ve-cdn-ops. Optional tier, max_iter=5. Low risk — cache purging and domain config.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-cdn-ops
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 5
---
# GCL Rubric — ve-cdn-ops
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | DeleteDomain | 5 | 1.0 |
| **State-changing** | StartDomain, StopDomain, CreateDomain, ModifyDomainConfig, SetOrigin, SubmitRefreshTask, SubmitDirRefreshTask | 5 | 1.0 |
| **Mutating** | PreLoadCache | 5 | ≥0.5 |
| **Read-only** | DescribeDomains, DescribeDomainDetail, DescribeCdnData, DescribeCdnOrigin, DescribeCdnTopData, DescribeContentQuota | 5 | ≥0 |
Safety: DeleteDomain — CDN domain config + DNS mapping LOST. StopDomain — traffic stops serving. SubmitRefreshTask — purge irreversible until TTL expires. VOLCENGINE_SECRET_KEY never in trace.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-cdn-ops |