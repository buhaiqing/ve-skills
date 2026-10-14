---
name: ve-ark-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) Ark (方舟大模型平台) — model inference endpoints, model training,
  dataset management, evaluation, and model marketplace. User mentions Ark, 方舟,
  大模型, LLM, model inference, model training, fine-tuning, endpoint, or describes
  AI/LLM deployment scenarios even without naming the product directly.
  Not for ECS, VKE, or traditional compute services.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine endpoints (open.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_script_syntax_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "Ark OpenAPI — https://www.volcengine.com/docs/82379"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve ark --help` — Ark is supported by the ve CLI.
    Service ID: ark.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine Ark (方舟大模型平台) Operations Skill

## Overview

Volcengine Ark (火山引擎方舟大模型平台) provides LLM model inference, fine-tuning, dataset management, and model marketplace services. Core resources include **Inference Endpoints** (推理端点), **Model Training/Finetuning Jobs** (模型精调), **Datasets** (数据集), **Model Evaluation** (模型评估), and **Model Marketplace** (模型市场). This skill is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and the official **`ve` CLI** flows), response validation, and failure recovery. **Do not use the web console as the primary agent execution path.**

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports Ark operations. **MUST** document both the **SDK** and **`ve` CLI** steps. CLI prefix: `ve ark`. Go SDK: `github.com/volcengine/volc-sdk-golang/service/ark`.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It | Concrete Validation Criteria |
|---|----------|---------------------------|------------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise Ark triggers; compute/network delegated to other skills | ≥ 3 SHOULD entries with specific triggers; ≥ 3 SHOULD NOT entries with named delegation targets |
| 2 | **Structured I/O** | `{{env.*}}` for credentials, `{{user.*}}` for endpoint/model/training params, `{{output.*}}` for API responses | Zero bare variable names; every input uses a typed placeholder; every output maps to a JSON path |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover, with numbered imperative steps | All 8+ operations have all 4 phases; steps numbered and imperative |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 12 Ark-specific codes; HALT vs retry per error type | Error table has ≥ 12 rows; each row has: code, max retries, backoff, agent action, UX template |
| 5 | **Absolute Single Responsibility** | Ark only; cross-product delegation (VKE, ECS, IAM, TOS) to other skills | SKILL.md covers exactly 1 product; cross-product ops delegate; naming follows `ve-ark-ops` |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine Ark", "火山引擎方舟", "大模型平台", "方舟", "LLM inference"
- Task involves **inference endpoint** CRUD: ListEndpoints, CreateEndpoint, DeleteEndpoint, ModifyEndpoint
- Task involves **model training/fine-tuning**: CreateTrainingJob, ListTrainingJobs, StopTrainingJob, DescribeTrainingJob
- Task involves **dataset management**: ListDatasets, CreateDataset, DescribeDataset, DeleteDataset
- Task involves **model evaluation**: CreateEvaluationJob, ListEvaluationJobs, DescribeEvaluationJob
- Task involves **model marketplace**: ListModels, DescribeModel, PurchaseModel
- Task keywords: endpoint, 推理端点, 精调, fine-tuning, training, dataset, 数据集, 评估, evaluation, 模型, model, inference
- User asks to deploy, configure, troubleshoot, or monitor Ark resources **via API, SDK, CLI, or automation**

### SHOULD NOT Use This Skill When

- Task is purely ECS instance management → delegate to: `ve-ecs-ops`
- Task is purely container service (VKE) → delegate to: `ve-vke-ops`
- Task is IAM / permission model only → delegate to: `ve-iam-ops`
- Task is purely object storage (TOS) → delegate to: `ve-tos-ops`
- Task is purely billing / account management → delegate to: `ve-billing-ops` (when present)
- User insists on **console-only** flows with no API → state limitation; do not invent undocumented HTTP steps

### Delegation Rules

- Ark endpoints can use VPC for private network access: verify VPC exists via `ve-vpc-ops` before configuring VPC for endpoints
- Training jobs may reference TOS buckets for datasets: verify TOS bucket exists via `ve-tos-ops`
- IAM policies control access to Ark resources: use `ve-iam-ops` for permission configuration
- Multi-product requests: handle each product with its skill; do not merge unrelated APIs into one ambiguous flow

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask the user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | From runtime environment | Use documented default only if skill explicitly allows |
| `{{user.region}}` | User-supplied region | Ask once; reuse |
| `{{user.endpoint_name}}` | Inference endpoint name | Ask once; reuse; pattern `^[a-zA-Z][a-zA-Z0-9-_]{1,63}$` |
| `{{user.model_id}}` | Model ID from marketplace | Ask once; reuse |
| `{{user.training_job_name}}` | Training job name | Ask once; reuse |
| `{{user.dataset_name}}` | Dataset name | Ask once; reuse |
| `{{user.dataset_type}}` | Dataset type (e.g., Text, QAPair) | Ask once; reuse; validate against supported list |
| `{{user.evaluation_job_name}}` | Evaluation job name | Ask once; reuse |
| `{{user.model_version_id}}` | Specific model version | Ask once; reuse |
| `{{output.endpoint_id}}` | From CreateEndpoint response `$.Result.EndpointId` | Parse from response |
| `{{output.endpoint_status}}` | Endpoint status from DescribeEndpoint | Parse from response |
| `{{output.training_job_id}}` | From CreateTrainingJob response `$.Result.TrainingJobId` | Parse from response |
| `{{output.dataset_id}}` | From CreateDataset response `$.Result.DatasetId` | Parse from response |
| `{{output.evaluation_job_id}}` | From CreateEvaluationJob response `$.Result.EvaluationJobId` | Parse from response |
| `{{output.endpoint_url}}` | Inference endpoint URL for model invocation | Parse from response |

> **`{{env.*}}` MUST NOT** be collected from the user. **`{{user.*}}`** MUST be collected interactively when missing.

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY`, `secret_key`, `SecretKey`, or any credential field value in console output, debug messages, error messages, or logs.
>
> **Credential verification MUST check existence only**, never echo the value:
> - Bash: `test -n "$VOLCENGINE_SECRET_KEY"` ✅ | `echo $VOLCENGINE_SECRET_KEY` ❌
> - Go: `if os.Getenv("VOLCENGINE_SECRET_KEY") == ""` ✅ | `fmt.Println(os.Getenv("VOLCENGINE_SECRET_KEY"))` ❌

## API and Response Conventions (Agent-Readable)

- **Volcengine Ark OpenAPI** is canonical for path, query, body fields, enums, and response shapes.
- **Endpoint:** `open.volcengineapi.com` (service: `ark`)
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/ark`
- **Errors:** Standard Volcengine response format with `ResponseMetadata.Error` containing `Code` and `Message`.
- **Timestamps:** ISO 8601 format (e.g. `2026-04-28T10:00:00Z`).

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| ListEndpoints | `$.Result.Endpoints[].EndpointId` | array | Endpoint IDs |
| ListEndpoints | `$.Result.Endpoints[].EndpointName` | array | Endpoint names |
| ListEndpoints | `$.Result.Endpoints[].Status` | array | Endpoint status (`Running`, `Creating`, `Failed`, `Stopped`) |
| CreateEndpoint | `$.Result.EndpointId` | string | New endpoint ID |
| CreateEndpoint | `$.Result.EndpointName` | string | Endpoint name |
| CreateEndpoint | `$.Result.Status` | string | Initial status (`Creating`) |
| DescribeEndpoint | `$.Result.Status` | string | Lifecycle state |
| DescribeEndpoint | `$.Result.ModelVersionId` | string | Deployed model version |
| ListModels | `$.Result.Models[].ModelId` | array | Marketplace model IDs |
| CreateTrainingJob | `$.Result.TrainingJobId` | string | Training job ID |
| ListTrainingJobs | `$.Result.TrainingJobs[].TrainingJobId` | array | Training job IDs |
| ListTrainingJobs | `$.Result.TrainingJobs[].Status` | array | Job status (`Running`, `Succeeded`, `Failed`) |
| CreateDataset | `$.Result.DatasetId` | string | Dataset ID |
| ListDatasets | `$.Result.Datasets[].DatasetId` | array | Dataset IDs |
| CreateEvaluationJob | `$.Result.EvaluationJobId` | string | Evaluation job ID |

### Expected State Transitions

| Operation | Initial State | Target State | Poll Interval | Max Wait |
|-----------|---------------|--------------|---------------|----------|
| CreateEndpoint | `Creating` | `Running` | 10s | 600s |
| DeleteEndpoint | `Running` | absent (404) | 10s | 300s |
| CreateTrainingJob | `Pending` | `Running` | 10s | 120s |
| CreateTrainingJob (complete) | `Running` | `Succeeded` | 30s | 86400s (24h) |
| CreateDataset | — | `Available` | 5s | 60s |
| CreateEvaluationJob | `Pending` | `Running` | 10s | 120s |

## Quick Start

### What This Skill Does
This skill enables you to deploy, configure, troubleshoot, and monitor Ark (方舟大模型平台) resources on Volcengine — inference endpoints, model training, datasets, and evaluations using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
ve ark ListEndpoints --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all inference endpoints
ve ark ListEndpoints --Region {{env.VOLCENGINE_REGION}}

# List available marketplace models
ve ark ListModels --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand Ark architecture
- [Common Operations](#execution-flows) — Create and manage endpoints, training, datasets
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| ListEndpoints | List all inference endpoints | Low | None |
| CreateEndpoint | Create a new inference endpoint | Medium | Low |
| DescribeEndpoint | View endpoint details | Low | None |
| ModifyEndpoint | Modify endpoint configuration | Medium | Medium |
| DeleteEndpoint | Delete an inference endpoint | Low | **High** — irreversible, disrupts live inference |
| ListModels | List marketplace models | Low | None |
| DescribeModel | View model details | Low | None |
| CreateTrainingJob | Create a model training/finetuning job | High | Low |
| ListTrainingJobs | List all training jobs | Low | None |
| DescribeTrainingJob | View training job details | Low | None |
| StopTrainingJob | Stop a running training job | Low | Medium — loses partial progress |
| ListDatasets | List all datasets | Low | None |
| CreateDataset | Create a new dataset | Medium | Low |
| DescribeDataset | View dataset details | Low | None |
| DeleteDataset | Delete a dataset | Low | **High** — irreversible data loss |
| CreateEvaluationJob | Create a model evaluation job | High | Low |
| ListEvaluationJobs | List all evaluation jobs | Low | None |

# See Execution Flows (CLI detail) in references/execution-flows.md

### Bootstrap Go Runtime (for JIT SDK fallback)

```bash
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/{{env.GO_VERSION}}.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
fi
```

### Configure Credentials

```bash
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="{{env.VOLCENGINE_REGION}}"
```

## Reference Directory

- [Core Concepts](references/core-concepts.md) — Understand Ark architecture
- [API & SDK Usage](references/api-sdk-usage.md) — SDK reference and API conventions
- [CLI Usage](references/cli-usage.md) — `ve ark` CLI command reference
- [Troubleshooting Guide](references/troubleshooting.md) — Common issues and fixes
- [Monitoring & Alerts](references/monitoring.md) — Endpoint and training monitoring
- [Integration](references/integration.md) — Environment setup and JIT SDK workflow
- [Knowledge Base](references/knowledge-base.md) — FAQ and best practices
- [GCL Rubric](references/rubric.md)
- [GCL Prompt Templates](references/prompt-templates.md)

## Diagnostic Commands

```bash
# List all endpoints
ve ark ListEndpoints --Region "{{env.VOLCENGINE_REGION}}"

# View endpoint details
ve ark DescribeEndpoint --EndpointId "{{output.endpoint_id}}" --Region "{{env.VOLCENGINE_REGION}}"

# List marketplace models
ve ark ListModels --Region "{{env.VOLCENGINE_REGION}}"

# List training jobs
ve ark ListTrainingJobs --Region "{{env.VOLCENGINE_REGION}}"

# List datasets
ve ark ListDatasets --Region "{{env.VOLCENGINE_REGION}}"

# List evaluation jobs
ve ark ListEvaluationJobs --Region "{{env.VOLCENGINE_REGION}}"
```

## Related Skills

- **ve-iam-ops** — Configure IAM permissions for Ark resources
- **ve-vpc-ops** — Configure VPC for private endpoint access
- **ve-tos-ops** — Manage TOS buckets for dataset storage
- **ve-cms-ops** — Configure monitoring and alerts for endpoints
- **ve-ecs-ops** — Manage compute resources (if using custom model hosting)
- **ve-billing-ops** — View Ark service billing and cost

## Operational Best Practices

- **Endpoints:** Use descriptive endpoint names with environment suffixes (`prod`, `staging`). Configure auto-scaling (MinReplicas/MaxReplicas) for production workloads.
- **Training:** Start with small epochs (2-3) to validate dataset format and training pipeline before full training. Monitor loss metrics during training.
- **Datasets:** Validate dataset format (JSONL for SFT, preference pairs for DPO) before uploading. Use TOS for large datasets (>100MB).
- **Cost:** Pause unused endpoints to avoid idle compute charges. Use batch inference for non-real-time workloads.
- **Security:** Use VPC isolation for production endpoints. Rotate API keys regularly. Grant least-privilege IAM policies.
- **Monitoring:** Set up endpoint latency and error rate alerts via CMS. Monitor training job status via Ark evaluation.

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-31 | Initial release with Ark endpoint, model, training, dataset, and evaluation management |
| 1.1.0 | 2026-06-04 | GCL rollout: added ## Quality Gate (GCL), references/rubric.md, references/prompt-templates.md |

## Quality Gate (GCL)

> Mandatory. max_iter=3.

| Tier | Operations | Safety |
|---|---|---|
| **Destructive** | DeleteEndpoint, DeleteDataset | 1.0 |
| **State-changing** | CreateEndpoint, StopEndpoint, CreateTrainingJob, StopTrainingJob, CreateEvaluationJob | 1.0 |
| **Mutating** | — (all state-changing) | ≥0.5 |
| **Read-only** | ListEndpoints, DescribeEndpoint, ListModels, ListTrainingJobs, ListDatasets | ≥0 |

Safety: DeleteEndpoint model inference stops. StopTrainingJob training progress lost. VOLCENGINE_SECRET_KEY never.

### Cross-skill: Billing→ve-billing-ops
