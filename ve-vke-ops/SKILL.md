---
name: ve-vke-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or monitor Volcengine
  VKE (容器服务 / Kubernetes Engine) — Cluster lifecycle, NodePool management, node
  operations, and diagnostics. User mentions VKE, 容器服务, Kubernetes Engine, K8s
  cluster, NodePool, or describes product-specific scenarios (e.g., cluster creation
  failures, node pool scaling issues, node join failures) even without naming the
  product directly. Not for billing, IAM, VPC networking, or container image registry.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints.
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_version_jit: "1.21+"
  api_profile: "VKE OpenAPI — https://www.volcengine.com/docs/6460"
  cli_applicability: dual-path
  cli_support_evidence: >-
    VKE is a core Volcengine service; `ve vke --help` lists cluster and node pool operations.
    OpenAPI service ID: vke. Go SDK: github.com/volcengine/volc-sdk-golang/service/vke.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This template follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine VKE Operations Skill

## Overview

Volcengine VKE (Volcengine Kubernetes Engine / 火山引擎容器服务) provides managed Kubernetes clusters for containerized workloads. This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and official **`ve` CLI** flows), response validation, and failure recovery for VKE Cluster and NodePool lifecycle management. **Do not use the web console as the primary agent execution path.**

### CLI applicability (repository policy)

- **`cli_applicability: dual-path`:** Official `ve` CLI supports VKE operations. Document **both** the SDK step **and** the `ve` CLI step for every operation the CLI exposes.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use with VKE triggers, delegation to ve-vpc-ops for networking |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for cluster/node pool names, `{{output.*}}` for API responses |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute (CLI + SDK) → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with 12+ VKE-specific codes, HALT vs retry per error type |
| 5 | **Absolute Single Responsibility** | VKE only (Cluster, NodePool, Node); VPC/ECS delegated to other skills |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "VKE", "容器服务", "Volcengine Kubernetes Engine", "K8s cluster on Volcengine"
- Task involves VKE Cluster or NodePool lifecycle (create, describe, modify, delete, list)
- Task keywords: cluster, nodepool, node, kubernetes, k8s, kubeconfig, 容器, 集群, 节点池
- User asks to deploy, configure, troubleshoot, or monitor VKE clusters via API, SDK, CLI, or automation

### SHOULD NOT Use This Skill When

- Task is billing / account management → delegate to: `ve-billing-ops` (if not available, use Volcengine Billing API directly)
- Task is IAM / permission model → delegate to: `ve-iam-ops` (if not available, use Volcengine IAM API directly)
- Task is VPC / subnet / security group → delegate to: `ve-vpc-ops` (if not available, use Volcengine VPC API directly)
- Task is ECS instance creation directly → delegate to: `ve-ecs-ops` (if not available, use Volcengine ECS API directly)
- User insists on **console-only** flows → state limitation; do not invent undocumented steps

### Delegation Rules

- VKE clusters require VPC and subnets: verify VPC exists via `ve-vpc-ops` before cluster creation
- VKE nodes are ECS instances: ECS-related diagnostics delegate to `ve-ecs-ops`
- Multi-product requests: handle each product with its skill; do not merge unrelated APIs

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill allows |
| `{{user.cluster_name}}` | User-supplied cluster name | Ask once; reuse |
| `{{user.node_pool_name}}` | User-supplied node pool name | Ask once; reuse |
| `{{user.k8s_version}}` | Kubernetes version (e.g., "v1.28") | Ask with supported version list |
| `{{output.cluster_id}}` | From CreateCluster API response | Parse `$.Result.ClusterId` |
| `{{output.node_pool_id}}` | From CreateNodePool API response | Parse `$.Result.NodePoolId` |
| `{{user.cluster_id}}` | Cluster ID | Format `cls-xxxxxxxxx`; ask once or parse from `{{output.cluster_id}}` |
| `{{user.subnet_id}}` | Cluster deployment subnet | Format `subnet-xxxxxxxxx`; ask once |
| `{{user.pod_subnet_id}}` | Pod network subnet | Format `subnet-xxxxxxxxx`; ask once |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

## API and Response Conventions

- **OpenAPI is canonical** for all field names, JSON paths, enums, and response shapes
- **Errors:** Map SDK/HTTP errors to `code` / message fields per spec
- **Timestamps:** ISO 8601 (e.g. `2026-05-16T10:00:00Z`)
- **Idempotency:** Use `ClientToken` for create operations to prevent duplicates

### Key Response Field Table

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateCluster | `$.Result.ClusterId` | string | New cluster ID |
| DescribeCluster | `$.Result.Status` | string | Cluster lifecycle state |
| DescribeCluster | `$.Result.KubernetesVersion` | string | K8s version |
| ListClusters | `$.Result.Items[].ClusterId` | array | Cluster IDs |
| ListClusters | `$.Result.Items[].Name` | array | Cluster names |
| ListClusters | `$.Result.Items[].Status` | array | Cluster states |
| CreateNodePool | `$.Result.NodePoolId` | string | New node pool ID |
| DescribeNodePool | `$.Result.Status` | string | Node pool state |
| DeleteCluster | `$.Metadata.RequestId` | string | Request correlation ID |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| Create Cluster | — | ✅ `Running` | 5s | 600s |
| Delete Cluster | any | ✅ deleted (404) | 10s | 600s |
| Create NodePool | — | ✅ `Running` | 5s | 300s |
| Delete NodePool | any | ✅ deleted | 5s | 300s |
| Update Cluster | ✅ `Running` | ✅ `Running` | 5s | 300s |

## Quick Start

### What This Skill Does
Deploy, configure, and manage VKE clusters and node pools on Volcengine using `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- ✅ `ve` CLI installed
- ✅ Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- ✅ Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve vke ListSupportedVersions --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
ve vke ListClusters --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand VKE architecture
- [Common Operations](#execution-flows) — Create, manage, and manage container clusters
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Level |
|-----------|-------------|-------|
| CreateCluster | Create managed K8s cluster | High/Med |
| DescribeCluster | Get cluster details | Low/None |
| ListClusters | List all clusters | Low/None |
| UpdateClusterConfig | Modify cluster config | Med/Med |
| DeleteCluster | Delete cluster | Med/**High** |
| CreateNodePool | Create node pool | Med/Low |
| DescribeNodePool | Get node pool details | Low/None |
| UpdateNodePool | Modify node pool | Med/Med |
| DeleteNodePool | Delete node pool | Med/**High** |
| AddNodes | Add nodes to pool | Med/Low |
| RemoveNodes | Remove nodes from pool | Med/Med |
| DeleteNodes | Delete nodes | Med/**High** |
| ListSupportedVersions | List K8s versions | Low/None |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-16 | Initial release with VKE cluster and node pool lifecycle |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteCluster, DeleteNodePool | 1.0 |
| **State-changing** | UpdateCluster, UpdateNodePool | 1.0 |
| **Mutating** | CreateCluster, CreateNodePool | ≥0.5 |
| **Read-only** | ListClusters, DescribeCluster, ListNodePools, DescribeNodePool | ≥0 |

Safety: DeleteCluster ALL k8s workloads lost. DeleteNodePool pods rescheduled.

### Cross-skill: VPC→ve-vpc-ops, ECS→ve-ecs-ops, Billing→ve-billing-ops

## Execution Flows (Agent-Readable)

### Operation: Create Cluster

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI available | `ve version` | Exit code 0 | Go to JIT Go SDK fallback |
| Credentials | Verify `VOLCENGINE_ACCESS_KEY` and `VOLCENGINE_SECRET_KEY` | Non-empty | HALT; user configures env |
| Region | Verify `{{env.VOLCENGINE_REGION}}` is valid VKE region | Supported region | HALT; suggest valid regions |
| VPC/Subnet | Verify VPC and subnet IDs exist and are in target region | Valid network | HALT; use ve-vpc-ops to verify |
| Quota | Check cluster quota via API | Sufficient | HALT; user raises quota |

#### Execution — CLI (`ve`) (Primary Path)

```bash
ve vke CreateCluster \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --Name "{{user.cluster_name}}" \
  --KubernetesVersion "{{user.k8s_version}}" \
  --ClusterConfig.SubnetIds '["{{user.subnet_id}}"]' \
  --ClusterConfig.ApiServerPublicAccessEnabled true \
  --PodsConfig.PodNetworkMode "VpcCniShared" \
  --PodsConfig.VpcCniConfig.SubnetIds '["{{user.pod_subnet_id}}"]' \
  --ServicesConfig.ServiceCidrsv4s '["172.30.0.0/18"]'
```

#### Execution — JIT Go SDK (Fallback Path)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/vke"
)

func main() {
    instance := vke.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region":            os.Getenv("VOLCENGINE_REGION"),
        "Name":              os.Getenv("CLUSTER_NAME"),
        "KubernetesVersion": os.Getenv("K8S_VERSION"),
        "ClusterConfig": map[string]interface{}{
            "SubnetIds":                   []string{os.Getenv("SUBNET_ID")},
            "ApiServerPublicAccessEnabled": true,
        },
        "PodsConfig": map[string]interface{}{
            "PodNetworkMode": "VpcCniShared",
        },
        "ServicesConfig": map[string]interface{}{
            "ServiceCidrsv4s": []string{"172.30.0.0/18"},
        },
    }

    resp, err := instance.Client.Request("vke", "CreateCluster", params)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp))
}
```

Execute:
```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
go run ./main.go
```

#### Post-execution Validation

```bash
# Poll until cluster reaches Running state
for i in $(seq 1 120); do
  STATUS=$(ve vke DescribeCluster --ClusterId "{{output.cluster_id}}" 2>/dev/null | jq -r '.Result.Status // ""')
  [ "$STATUS" = "Running" ] && break
  [ "$STATUS" = "Error" ] && echo "Cluster creation failed" && exit 1
  sleep 5
done
```

#### Failure Recovery

| Error Pattern | Action | Recovery |
|---------------|--------|----------|
| `InvalidParameter.ClusterName` | HALT | 1-64 chars, lowercase, digits, hyphens |
| `QuotaExceeded.ClusterCount` | HALT | Delete unused or request increase |
| `InvalidParameter.VpcConfig` | HALT | Verify VPC/subnet via ve-vpc-ops |
| `ResourceAlreadyExists` | HALT | Use unique name |
| `Throttling` | Retry ×3, exponential | Back off |
| `InternalError` | Retry ×3, 2s/4s/8s | HALT with RequestId |
| `InsufficientBalance` | HALT | Recharge account |

### Operation: Describe Cluster

#### Execution

```bash
ve vke DescribeCluster --ClusterId "{{user.cluster_id}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| Cluster ID | `$.Result.ClusterId` | Primary identifier |
| Name | `$.Result.Name` | Cluster name |
| Status | `$.Result.Status` | Running/Creating/Deleting/Error |
| K8s Version | `$.Result.KubernetesVersion` | e.g., v1.28 |
| VPC ID | `$.Result.ClusterConfig.VpcId` | Network context |
| Endpoints | `$.Result.Endpoints` | API server address |

### Operation: List Clusters

#### Execution

```bash
ve vke ListClusters \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PageNumber 1 \
  --PageSize 100
```

### Operation: Update Cluster Config

#### Execution

```bash
ve vke UpdateClusterConfig --ClusterId "{{user.cluster_id}}" --body '{"DeleteProtectionEnabled": true}'
```

### Operation: Delete Cluster

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit user confirmation for irreversible delete of `{{user.cluster_name}}` (`{{user.cluster_id}}`)
- **MUST** verify delete protection is disabled or user confirms override
- **MUST NOT** proceed without clear user assent

#### Execution

```bash
ve vke DeleteCluster --ClusterId "{{user.cluster_id}}"
```

#### Post-execution Validation

```bash
# Poll until cluster is gone
for i in $(seq 1 60); do
  HTTP_CODE=$(ve vke DescribeCluster --ClusterId "{{user.cluster_id}}" 2>&1 | grep -o '"ResponseMessage"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1)
  [ -n "$HTTP_CODE" ] && break
  sleep 10
done
```

| Error Pattern | Action | Recovery |
|---------------|--------|----------|
| `OperationDenied.DeleteProtection` | HALT | Disable protection first |
| `ResourceNotFound.Cluster` | Skip | Already deleted → ✅ |

### Operation: Create NodePool

#### Execution

```bash
ve vke CreateNodePool \
  --ClusterId "{{user.cluster_id}}" \
  --Name "{{user.node_pool_name}}" \
  --NodeConfig.InstanceTypeIds '["ecs.g1ie.xlarge"]' \
  --NodeConfig.SubnetIds '["{{user.subnet_id}}"]' \
  --NodeConfig.SystemVolume.Type "ESSD_PL0" \
  --NodeConfig.SystemVolume.Size 50 \
  --AutoScaling.Enabled true \
  --AutoScaling.MinReplicas 1 \
  --AutoScaling.MaxReplicas 5
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve vke DescribeNodePool --ClusterId "{{user.cluster_id}}" --NodePoolId "{{output.node_pool_id}}" | jq -r '.Result.Status // ""')
  [ "$STATUS" = "Running" ] && break
  sleep 5
done
```

### Operation: DescribeNodePool — Get NodePool Details

#### Execution

```bash
ve vke DescribeNodePool \
  --ClusterId "{{user.cluster_id}}" \
  --NodePoolId "{{user.node_pool_id}}"
```

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| NodePool ID | `$.Result.NodePoolId` | Primary identifier |
| Name | `$.Result.Name` | Node pool name |
| Status | `$.Result.Status` | Running/Creating/Deleting/Error |
| Desired Replicas | `$.Result.AutoScaling.DesiredReplicas` | Current node count |
| Min/Max Replicas | `$.Result.AutoScaling.MinReplicas` / `MaxReplicas` | Scaling bounds |
| Instance Types | `$.Result.NodeConfig.InstanceTypeIds` | ECS types in pool |

---

### Operation: DeleteNodePool — Delete NodePool

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit user confirmation for irreversible delete of `{{user.node_pool_name}}` (`{{user.node_pool_id}}`)
- **MUST** warn: all nodes in pool will be terminated
- **MUST NOT** proceed without clear user assent

#### Execution

```bash
ve vke DeleteNodePool --ClusterId "{{user.cluster_id}}" --NodePoolId "{{user.node_pool_id}}"
```

#### Post-execution Validation

```bash
for i in $(seq 1 60); do
  STATUS=$(ve vke DescribeNodePool --ClusterId "{{user.cluster_id}}" --NodePoolId "{{user.node_pool_id}}" 2>/dev/null | jq -r '.Result.Status // "deleted"')
  [ "$STATUS" = "deleted" ] || [ -z "$STATUS" ] && break
  sleep 5
done
```

---

### Operation: Add/Remove/Delete Nodes

#### AddNodes

```bash
ve vke AddNodes \
  --ClusterId "{{user.cluster_id}}" \
  --NodePoolId "{{user.node_pool_id}}" \
  --InstanceIds '["{{user.ecs_instance_id}}"]'
```

#### RemoveNodes (without deleting ECS)

```bash
ve vke RemoveNodes \
  --ClusterId "{{user.cluster_id}}" \
  --NodePoolId "{{user.node_pool_id}}" \
  --InstanceIds '["{{user.ecs_instance_id}}"]'
```

#### DeleteNodes (destroys ECS instances)

**Safety Gate:** Explicit confirmation required; warn about pod eviction and data loss.

```bash
ve vke DeleteNodes \
  --ClusterId "{{user.cluster_id}}" \
  --NodePoolId "{{user.node_pool_id}}" \
  --InstanceIds '["{{user.ecs_instance_id}}"]'
```

### Operation: List Supported Versions

```bash
ve vke ListSupportedVersions --Region "{{env.VOLCENGINE_REGION}}"
```

## Error Taxonomy

| `code` | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidClusterId` | 集群 ID 不存在或格式错误 | 0 retries; **HALT** — 检查 ClusterId 格式 (cls-xxxxxxxxx) |
| `IncorrectClusterState` | 集群状态不允许当前操作 | 0 retries; **HALT** — 等待集群状态变为 Running 后重试 |
| `QuotaExceeded.Cluster` | 集群数量超出配额限制 | 0 retries; **HALT** — 删除未使用集群或提额 |
| `InvalidNodePoolId` | 节点池 ID 不存在或格式错误 | 0 retries; **HALT** — 检查 NodePoolId 格式 |
| `InvalidVpcId` | VPC ID 不存在或不属于当前账号 | 0 retries; **HALT** — 通过 ve-vpc-ops 验证 VPC |
| `InvalidSecurityGroupId` | 安全组 ID 不存在或不属于当前 VPC | 0 retries; **HALT** — 通过 ve-vpc-ops 验证安全组 |
| `NodeInsufficientResource` | 节点资源不足 (CPU/内存) | 2 retries/exponential/30s/60s/120s; **RETRY** — 更换实例规格或扩容节点池 |
| `InvalidFlavorId` | 实例规格 ID 不存在或不可用 | 0 retries; **HALT** — 检查实例规格 ID 是否在当前可用区可用 |
| `Throttling` | 请求频率过高触发限流 | 3 retries/exponential/2s/4s/8s; **RETRY** — 背退等待后重试 |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 超过重试次数后 HALT 并记录 RequestId |
| `OperationDenied.DeleteProtection` | 删除保护已开启，禁止删除集群 | 0 retries; **HALT** — 先关闭删除保护后再操作 |
| `ResourceNotFound.Cluster` | 集群未找到 | 0 retries; **HALT** — 检查集群 ID 是否已删除 |

## Prerequisites

1. **Install `ve` CLI:** See [Execution Environment](../ve-skill-generator/references/execution-environment.md) for details.
2. **Bootstrap Go runtime** (for JIT SDK fallback): See references/integration.md
3. **Configure Credentials:** Environment variables `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
4. **Verify:** `ve vke ListSupportedVersions --Region "{{env.VOLCENGINE_REGION}}"`

> **Security:** Never commit credentials. See credential masking rules in Variable Convention section.

## Operational Best Practices

- **Cluster networking (CNI):** Use `VpcCniShared` mode for pod networking to leverage VPC ENI (Elastic Network Interface). Plan pod and service CIDR ranges before cluster creation, ensuring no overlap with existing VPC CIDR. Leave sufficient IP headroom in subnets for cluster scaling.
- **Node pool sizing:** Configure `AutoScaling` with `MinReplicas` >= 2 for HA. Set `MaxReplicas` based on workload burst capacity. Use `ecs.g3i` or `ecs.g1ie` instance families for general workloads. Right-size instance types to avoid resource fragmentation.
- **RBAC and IAM integration:** Bind IAM roles to VKE cluster RBAC via `ClusterRole` + `RoleBinding` for human operators. Use service accounts with mapped IAM roles for pod-level AWS-compatible IRSA-like access. Audit `kubectl describe clusterrolebinding` regularly.
- **Backup and disaster recovery (etcd):** Enable VKE auto-backup for cluster control plane state. Schedule periodic etcd snapshots for critical namespaces. Use `Velero` (or Volcengine backup service) for persistent volume backup of stateful workloads.
- **Monitoring and observability:** Deploy `metrics-server` for `kubectl top` and HPA. Install Prometheus + Grafana stack for cluster-wide metrics. Monitor control plane metrics via `ve-cms-ops`: API server latency, etcd leader changes, node `NotReady` count.
- **Security scanning:** Enable image scanning on container registry for known CVEs before deployment. Apply `PodSecurityAdmission` (baseline or restricted) via VKE admission configuration. Rotate cluster certificate and admin kubeconfig periodically.

## Reference Directory

- [Core Concepts](references/core-concepts.md) — VKE architecture, cluster types, node pools
- [API & SDK Usage](references/api-sdk-usage.md) — Operation map, fields, JSON paths
- [CLI Usage](references/cli-usage.md) — `ve vke` command reference
- [Knowledge Base](references/knowledge-base.md) — fault pattern library (AIOps diagnosis)
- [Troubleshooting Guide](references/troubleshooting.md) — Error codes, diagnostics
- [Monitoring & Alerts](references/monitoring.md) — VKE monitoring metrics
- [Integration](references/integration.md) — Go SDK setup, JIT workflow
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)
