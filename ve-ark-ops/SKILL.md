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

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute (CLI + SDK) → Validate → Recover**. Do not skip phases.

**Preference hint:** CLI (`ve ark`) is the primary execution path. When CLI does not support a specific operation, JIT build a Go SDK script.

### Operation: ListEndpoints — List Inference Endpoints

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI / deps | `ve version` | Exit code 0 | Install CLI |
| Credentials | Verify env vars set | Non-empty keys | HALT; configure credentials |
| Region | `{{env.VOLCENGINE_REGION}}` | Region supported | Suggest valid region |

#### Execution — CLI (`ve`) (Primary Path)

```bash
ve ark ListEndpoints --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback Path)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    params["PageSize"] = 100

    resp, err := instance.Client.Request("ark", "ListEndpoints", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

Execute:
```bash
cd /tmp/ve-sdk-workspace && go run main.go
```

#### Post-execution Validation

1. Parse the JSON response for `$.Result.Endpoints` array.
2. If empty, report: "No inference endpoints found in region {{env.VOLCENGINE_REGION}}."
3. If non-empty, present as a table:

```bash
ve ark ListEndpoints --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Endpoints[] | [.EndpointId, .EndpointName, .Status] | @tsv'
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidParameter` / 400 | 0–1 | Fix args; retry once — `[ERROR] InvalidParameter: The request parameter is invalid. Check parameter format.` |
| `AccessDenied` / 403 | 0 | HALT; check IAM permissions — `[ERROR] AccessDenied: Your account does not have permission for this operation.` |
| `InternalError` / 5xx | 3 | 2s, 4s, 8s backoff; retry, then HALT with RequestId |
| Throttling / 429 | 3 | Exponential backoff — `⚠️ Rate limit reached. Retrying in {backoff}s...` |

---

### Operation: CreateEndpoint — Create an Inference Endpoint

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| CLI / deps | `ve version` | Exit code 0 | Install CLI |
| Credentials | Verify env vars set | Non-empty keys | HALT |
| Region | `{{user.region}}` | Region supported | Suggest valid region |
| Endpoint name | Validate pattern `^[a-zA-Z][a-zA-Z0-9-_]{1,63}$` | Valid name | HALT; provide naming rules |
| Model ID | `ve ark ListModels --Region "{{env.VOLCENGINE_REGION}}"` | Model exists | HALT; show available models |
| Resource quota | Query existing endpoints | Under quota limit | HALT; delete unused endpoints |

#### Execution — CLI (`ve`) (Primary Path)

```bash
ve ark CreateEndpoint \
  --EndpointName "{{user.endpoint_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --EndpointType "Inference" \
  --Description "{{user.description}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

Optional parameters:
```bash
# With VPC configuration
ve ark CreateEndpoint \
  --EndpointName "{{user.endpoint_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --EndpointType "Inference" \
  --VpcId "vpc-xxx" \
  --SubnetIds '["subnet-xxx","subnet-yyy"]' \
  --Region "{{env.VOLCENGINE_REGION}}"

# With scaling configuration (min/max replicas)
ve ark CreateEndpoint \
  --EndpointName "{{user.endpoint_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --EndpointType "Inference" \
  --MinReplicas 1 \
  --MaxReplicas 5 \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback Path)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "EndpointName":  os.Args[1],
        "ModelVersionId": os.Args[2],
        "EndpointType": "Inference",
        "Description":  os.Args[3],
        "Region":       os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("ark", "CreateEndpoint", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

Execute:
```bash
cd /tmp/ve-sdk-workspace && go run main.go \
  "{{user.endpoint_name}}" "{{user.model_version_id}}" "{{user.description}}"
```

#### Post-execution Validation

1. Read `{{output.endpoint_id}}` from `$.Result.EndpointId`.
2. Poll until `Status == "Running"` or timeout:

```bash
for i in $(seq 1 60); do
  STATUS=$(ve ark DescribeEndpoint \
    --EndpointId "{{output.endpoint_id}}" \
    --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  [ "$STATUS" = "Running" ] && break
  sleep 10
done
```

3. On success: `✅ Inference endpoint {{user.endpoint_name}} (ID: {{output.endpoint_id}}) is now Running.`
4. On timeout: HALT and report endpoint stuck in `Creating` state.

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidParameter` / 400 | 0–1 | Fix args from OpenAPI; retry once — `[ERROR] InvalidParameter: Check parameters against OpenAPI docs` |
| `EndpointAlreadyExists` | 0 | HALT; ask user for different name — `[ERROR] Endpoint "{{user.endpoint_name}}" already exists.` |
| `ModelNotFound` | 0 | HALT; show available models — `[ERROR] Model version "{{user.model_version_id}}" not found.` |
| `QuotaExceeded` | 0 | HALT; delete unused endpoints — `[ERROR] Endpoint quota exceeded.` |
| `InsufficientBalance` | 0 | HALT — `[ERROR] Account balance insufficient. Recharge before proceeding.` |
| `InvalidVpcConfig` | 0 | HALT; verify VPC/subnet — `[ERROR] VPC configuration invalid.` |
| `ResourceLimitExceeded` | 0 | HALT — `[ERROR] Resource limit reached (e.g., GPU quota). Request quota increase.` |
| Throttling / 429 | 3 | Exponential backoff — `⚠️ Rate limit reached. Retrying in {backoff}s...` |
| `InternalError` / 5xx | 3 | 2s, 4s, 8s backoff; retry, then HALT with RequestId |

---

### Operation: DeleteEndpoint — Delete an Inference Endpoint

#### Pre-flight (Safety Gate)

- **⚠️ MUST obtain explicit confirmation**: irreversible delete of endpoint `{{user.endpoint_name}}` (`{{output.endpoint_id}}`). This disrupts all live inference traffic.
- **MUST NOT** proceed without clear user assent.
- **MUST** verify endpoint exists before attempting deletion.

```bash
# Verify endpoint exists
ve ark DescribeEndpoint \
  --EndpointId "{{output.endpoint_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — CLI (`ve`)

```bash
ve ark DeleteEndpoint \
  --EndpointId "{{output.endpoint_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "EndpointId": os.Args[1],
        "Region":     os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("ark", "DeleteEndpoint", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

```bash
ve ark DescribeEndpoint \
  --EndpointId "{{output.endpoint_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}" 2>&1 | grep -q "EndpointNotFound"
```

Expected: `EndpointNotFound` error confirms deletion. If endpoint still exists, poll up to 300s.

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `EndpointNotFound` | 0 | Already deleted; skip | `ℹ️ Endpoint "{{user.endpoint_name}}" does not exist — already deleted.` |
| `EndpointInUse` | 0 | HALT; check for active connections | `[ERROR] Endpoint is currently in use. Stop inference traffic before deleting.` |
| `AccessDenied` | 0 | HALT; check IAM permissions | `[ERROR] AccessDenied: No permission to delete endpoints. Check IAM.` |

---

### Operation: ListModels — List Marketplace Models

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Verify env vars | Non-empty | HALT |

#### Execution — CLI (`ve`)

```bash
# List all available models
ve ark ListModels --Region "{{env.VOLCENGINE_REGION}}"

# Filter by vendor (Bytedance, Open-source, Third-party)
ve ark ListModels --ModelVendor "BYTEDANCE" --Region "{{env.VOLCENGINE_REGION}}"

# Filter by model type (Chat, Embedding, Image)
ve ark ListModels --ModelType "Chat" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("ark", "ListModels", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| Model ID | `$.Result.Models[].ModelId` | Use for CreateEndpoint |
| Model Name | `$.Result.Models[].ModelName` | Display name |
| Vendor | `$.Result.Models[].ModelVendor` | BYTEDANCE / OPEN_SOURCE / THIRD_PARTY |
| Model Type | `$.Result.Models[].ModelType` | Chat / Embedding / Image |
| Model Versions | `$.Result.Models[].Versions` | Version list for deployment |

---

### Operation: CreateTrainingJob — Create a Model Training/Finetuning Job

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | Verify env vars | Non-empty | HALT |
| Training name | Validate name pattern | Valid | HALT; fix name |
| Base model | `ve ark ListModels` | Model supports training | HALT; suggest trainable models |
| Dataset | `ve ark DescribeDataset --DatasetId` | Dataset exists + Available | HALT; create dataset first |
| Hyperparameters | Validate JSON/format | Valid config | HALT; fix hyperparameter format |

#### Execution — CLI (`ve`)

```bash
ve ark CreateTrainingJob \
  --TrainingJobName "{{user.training_job_name}}" \
  --ModelId "{{user.model_id}}" \
  --DatasetId "{{output.dataset_id}}" \
  --HyperParameters '{"learning_rate": 1e-5, "epochs": 3, "batch_size": 4}' \
  --Description "{{user.description}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

With output model name:
```bash
ve ark CreateTrainingJob \
  --TrainingJobName "{{user.training_job_name}}" \
  --ModelId "{{user.model_id}}" \
  --DatasetId "{{output.dataset_id}}" \
  --OutputModelName "{{user.output_model_name}}" \
  --HyperParameters '{"learning_rate": 2e-5, "epochs": 5, "batch_size": 8}' \
  --FinetuningMethod "SFT" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/ark"
)

func main() {
    instance := ark.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "TrainingJobName":  os.Args[1],
        "ModelId":          os.Args[2],
        "DatasetId":        os.Args[3],
        "HyperParameters":  os.Args[4],
        "Region":           os.Getenv("VOLCENGINE_REGION"),
    }

    resp, err := instance.Client.Request("ark", "CreateTrainingJob", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
```

#### Post-execution Validation

1. Read `{{output.training_job_id}}` from `$.Result.TrainingJobId`.
2. Poll until `Status` transitions to `Running`:

```bash
for i in $(seq 1 12); do
  STATUS=$(ve ark DescribeTrainingJob \
    --TrainingJobId "{{output.training_job_id}}" \
    --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  [ "$STATUS" = "Running" ] && break
  sleep 10
done
```

3. For long-running jobs, report periodically.
4. On `Succeeded`: `✅ Training job {{user.training_job_name}} completed successfully. Output model created.`

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidDataset` | 0 | HALT; verify dataset format | `[ERROR] Dataset format is not valid for training. Check dataset type matches model requirements.` |
| `ModelNotTrainable` | 0 | HALT; model does not support fine-tuning | `[ERROR] Model "{{user.model_id}}" does not support fine-tuning. Choose a trainable model.` |
| `TrainingJobAlreadyExists` | 0 | HALT; use different name | `[ERROR] Training job name already exists. Use a different name.` |
| `InvalidHyperParameters` | 0 | HALT; check hyperparameter format | `[ERROR] HyperParameters format or values are invalid. Check supported ranges.` |
| `QuotaExceeded` | 0 | HALT | `[ERROR] Training job quota exceeded.` |
| `InsufficientBalance` | 0 | HALT | `[ERROR] Insufficient balance for training. Recharge account.` |
| Throttling / 429 | 3 | Back off | `⚠️ Rate limit reached.` |
| `InternalError` / 5xx | 3 | Retry; HALT with RequestId | `[ERROR] Internal server error. RequestId: {id}.` |

---

### Operation: ListTrainingJobs — List Training Jobs

#### Execution — CLI (`ve`)

```bash
# List all training jobs
ve ark ListTrainingJobs --Region "{{env.VOLCENGINE_REGION}}"

# Filter by status
ve ark ListTrainingJobs --Status "Running" --Region "{{env.VOLCENGINE_REGION}}"

# Filter by model
ve ark ListTrainingJobs --ModelId "{{user.model_id}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

```bash
ve ark ListTrainingJobs --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.TrainingJobs[] | [.TrainingJobId, .TrainingJobName, .Status, .CreateTime] | @tsv'
```

---

### Operation: ListDatasets — List Datasets

#### Execution — CLI (`ve`)

```bash
# List all datasets
ve ark ListDatasets --Region "{{env.VOLCENGINE_REGION}}"

# Filter by type
ve ark ListDatasets --DatasetType "Text" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | Path | Notes |
|-------|------|-------|
| Dataset ID | `$.Result.Datasets[].DatasetId` | Use for training jobs |
| Dataset Name | `$.Result.Datasets[].DatasetName` | Display name |
| Dataset Type | `$.Result.Datasets[].DatasetType` | Text / QAPair / MultiTurn |
| Status | `$.Result.Datasets[].Status` | Available / Processing / Failed |
| Size | `$.Result.Datasets[].SizeBytes` | File size in bytes |

---

### Operation: CreateDataset — Create a Dataset

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Dataset name | Validate pattern | Valid | HALT; fix name |
| Dataset type | Supported types | Valid | HALT; show supported types |
| File / data source | Check accessible | Accessible | HALT; provide valid source |

#### Execution — CLI (`ve`)

```bash
# Create dataset from TOS file
ve ark CreateDataset \
  --DatasetName "{{user.dataset_name}}" \
  --DatasetType "{{user.dataset_type}}" \
  --DataSourceType "TOS" \
  --TosPath "{{user.tos_path}}" \
  --Description "{{user.description}}" \
  --Region "{{env.VOLCENGINE_REGION}}"

# Create dataset from local file upload
ve ark CreateDataset \
  --DatasetName "{{user.dataset_name}}" \
  --DatasetType "{{user.dataset_type}}" \
  --DataSourceType "Upload" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Post-execution Validation

1. Read `{{output.dataset_id}}` from `$.Result.DatasetId`.
2. Poll until `Status == "Available"`:

```bash
for i in $(seq 1 12); do
  STATUS=$(ve ark DescribeDataset \
    --DatasetId "{{output.dataset_id}}" \
    --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  [ "$STATUS" = "Available" ] && break
  sleep 5
done
```

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `DatasetAlreadyExists` | 0 | HALT; use different name | `[ERROR] Dataset name already exists.` |
| `InvalidDatasetType` | 0 | HALT; show supported types | `[ERROR] Dataset type not supported. Supported: Text, QAPair, MultiTurn.` |
| `TosPathNotFound` | 0 | HALT; verify TOS path | `[ERROR] TOS path not found. Verify bucket and prefix exist.` |
| `DatasetTooLarge` | 0 | HALT; reduce dataset size | `[ERROR] Dataset exceeds maximum allowed size.` |

---

### Operation: DescribeEndpoint — View Endpoint Details

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Endpoint exists | `ve ark ListEndpoints` | Endpoint in list | HALT; suggest creation |

#### Execution — CLI (`ve`)

```bash
# Get endpoint details
ve ark DescribeEndpoint \
  --EndpointId "{{user.endpoint_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Endpoint ID | `$.Result.EndpointId` | Unique identifier |
| Name | `$.Result.EndpointName` | User-defined name |
| Model | `$.Result.ModelVersionId` | Deployed model version |
| Status | `$.Result.Status` | Running, Creating, Updating, Deleting |
| Spec | `$.Result.Specification` | Instance type |
| Replicas | `$.Result.Replicas` | Number of replicas |
| CreateTime | `$.Result.CreateTime` | Creation timestamp |

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `EndpointNotFound` | 0 | HALT; verify endpoint ID | `[ERROR] Endpoint not found. Verify the endpoint ID.` |

---

### Operation: ModifyEndpoint — Modify Endpoint Configuration

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Endpoint exists | `ve ark DescribeEndpoint` | Status Running | HALT |
| No active traffic | Confirm with user | User acknowledges | Proceed with caution |

#### Execution — CLI (`ve`)

```bash
# Modify endpoint replicas (scaling)
ve ark ModifyEndpoint \
  --EndpointId "{{user.endpoint_id}}" \
  --Replicas {{user.replica_count}} \
  --Region "{{env.VOLCENGINE_REGION}}"

# Modify endpoint specification (upgrade/downgrade)
ve ark ModifyEndpoint \
  --EndpointId "{{user.endpoint_id}}" \
  --Specification "{{user.new_spec}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

```bash
# Poll until endpoint is Running
for i in $(seq 1 30); do
  STATUS=$(ve ark DescribeEndpoint --EndpointId "{{user.endpoint_id}}" --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  echo "Status: $STATUS (attempt $i/30)"
  [ "$STATUS" = "Running" ] && break
  sleep 10
done

# Verify new configuration
ve ark DescribeEndpoint --EndpointId "{{user.endpoint_id}}" --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result | {Specification, Replicas}'
```

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `EndpointNotFound` | 0 | HALT; verify endpoint ID | `[ERROR] Endpoint not found.` |
| `InvalidParameter` | 1 | Fix parameter | `[ERROR] Invalid parameter: check spec/replicas.` |
| `ResourceLimitExceeded` | 0 | HALT; request quota increase | `[ERROR] Resource limit exceeded.` |

---

### Operation: DescribeModel — View Marketplace Model Details

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Model ID valid | `ve ark ListModels` | Model in list | HALT; list available models |

#### Execution — CLI (`ve`)

```bash
# Get model details
ve ark DescribeModel \
  --ModelId "{{user.model_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"

# Get specific model version details
ve ark DescribeModel \
  --ModelId "{{user.model_id}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Model ID | `$.Result.ModelId` | Unique identifier |
| Name | `$.Result.ModelName` | Model display name |
| Provider | `$.Result.Provider` | Model provider |
| Versions | `$.Result.Versions` | Available versions array |
| Trainable | `$.Result.Trainable` | Whether fine-tuning is supported |

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `ModelNotFound` | 0 | HALT; verify model ID | `[ERROR] Model not found. List marketplace models: ve ark ListModels.` |

---

### Operation: DescribeTrainingJob — View Training Job Details

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Training job exists | `ve ark ListTrainingJobs` | Job in list | HALT |

#### Execution — CLI (`ve`)

```bash
# Get training job details
ve ark DescribeTrainingJob \
  --TrainingJobId "{{user.training_job_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Job ID | `$.Result.JobId` | Unique identifier |
| Name | `$.Result.JobName` | Job name |
| Status | `$.Result.Status` | Running, Succeeded, Failed, Stopped |
| Model | `$.Result.ModelVersionId` | Model being fine-tuned |
| Dataset | `$.Result.DatasetId` | Training dataset |
| Progress | `$.Result.Progress` | Training progress (0-100) |
| Duration | `$.Result.Duration` | Training duration |
| HyperParams | `$.Result.HyperParameters` | Training parameters |

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `TrainingJobNotFound` | 0 | HALT; verify job ID | `[ERROR] Training job not found.` |

---

### Operation: StopTrainingJob — Stop a Running Training Job

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before stopping a training job
- **MUST** warn that partial training progress will be lost

```bash
# Verify job status before stopping
ve ark DescribeTrainingJob --TrainingJobId "{{user.training_job_id}}" --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result | {JobName, Status, Progress}'
```

#### Execution — CLI (`ve`)

```bash
# Stop the training job
ve ark StopTrainingJob \
  --TrainingJobId "{{user.training_job_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

```bash
# Verify job is stopped
for i in $(seq 1 10); do
  STATUS=$(ve ark DescribeTrainingJob --TrainingJobId "{{user.training_job_id}}" --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Status')
  echo "Status: $STATUS (attempt $i/10)"
  [ "$STATUS" = "Stopped" ] || [ "$STATUS" = "Failed" ] && break
  sleep 5
done
```

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `TrainingJobNotFound` | 0 | HALT; verify job ID | `[ERROR] Training job not found.` |
| `InvalidJobStatus` | 0 | HALT; job already completed | `[ERROR] Job already completed. Cannot stop a finished job.` |

---

### Operation: DescribeDataset — View Dataset Details

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Dataset exists | `ve ark ListDatasets` | Dataset in list | HALT |

#### Execution — CLI (`ve`)

```bash
# Get dataset details
ve ark DescribeDataset \
  --DatasetId "{{user.dataset_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Dataset ID | `$.Result.DatasetId` | Unique identifier |
| Name | `$.Result.DatasetName` | Dataset name |
| Type | `$.Result.DatasetType` | Text, QAPair, MultiTurn |
| Status | `$.Result.Status` | Available, Processing, Failed |
| Size | `$.Result.Size` | Dataset size |
| Source | `$.Result.DataSource` | Upload or TOS path |
| CreateTime | `$.Result.CreateTime` | Creation timestamp |

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `DatasetNotFound` | 0 | HALT; verify dataset ID | `[ERROR] Dataset not found.` |

---

### Operation: DeleteDataset — Delete a Dataset

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: delete dataset `{{user.dataset_name}}`
- **MUST** warn that data is permanently lost
- **MUST** check if any training jobs reference this dataset

```bash
# Check if any training jobs use this dataset
ve ark ListTrainingJobs --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.Jobs[] | select(.DatasetId == "{{user.dataset_id}}") | .JobName'
```

#### Execution — CLI (`ve`)

```bash
# Delete the dataset
ve ark DeleteDataset \
  --DatasetId "{{user.dataset_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

```bash
# Verify dataset is deleted
if ve ark DescribeDataset --DatasetId "{{user.dataset_id}}" --Region "{{env.VOLCENGINE_REGION}}" 2>&1 | grep -q "DatasetNotFound"; then
  echo "Dataset deleted successfully"
fi
```

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `DatasetNotFound` | 0 | Already deleted; skip | `[INFO] Dataset already deleted.` |
| `DependencyViolation` | 0 | HALT; training jobs reference this dataset | `[ERROR] Dataset in use by training jobs. Remove dependencies first.` |

---

### Operation: CreateEvaluationJob — Create a Model Evaluation Job

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Model exists | `ve ark ListModels` | Model available | HALT |
| Dataset exists | `ve ark DescribeDataset` | Dataset Available | HALT; create dataset first |
| Quota | Check evaluation job quota | Sufficient | HALT; request quota increase |

#### Execution — CLI (`ve`)

```bash
# Create evaluation job
ve ark CreateEvaluationJob \
  --EvaluationJobName "{{user.evaluation_job_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --DatasetId "{{user.dataset_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}"

# Create evaluation with custom parameters
ve ark CreateEvaluationJob \
  --EvaluationJobName "{{user.evaluation_job_name}}" \
  --ModelVersionId "{{user.model_version_id}}" \
  --DatasetId "{{user.dataset_id}}" \
  --EvaluationConfig '{"metrics": ["accuracy", "f1", "rouge"]}' \
  --Region "{{env.VOLCENGINE_REGION}}"
```

#### Validation

```bash
# Poll until job completes
for i in $(seq 1 60); do
  STATUS=$(ve ark DescribeEvaluationJob --EvaluationJobId "{{output.evaluation_job_id}}" --Region "{{env.VOLCENGINE_REGION}}" 2>/dev/null | jq -r '.Result.Status // "pending"')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "Succeeded" ] && break
  [ "$STATUS" = "Failed" ] && echo "Evaluation failed" && break
  sleep 10
done
```

#### Failure Recovery

| Error pattern | Max retries | Agent Action | UX Feedback |
|--------------|-------------|--------------|-------------|
| `InvalidModelVersion` | 0 | HALT; model version not evaluable | `[ERROR] Model version does not support evaluation.` |
| `InvalidDatasetType` | 0 | HALT; dataset type incompatible | `[ERROR] Dataset type not compatible with evaluation.` |
| `QuotaExceeded` | 0 | HALT; request quota increase | `[ERROR] Evaluation job quota exceeded.` |

---

### Operation: ListEvaluationJobs — List Evaluation Jobs

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Model access | Verify user has access | Access granted | HALT |

#### Execution — CLI (`ve`)

```bash
# List all evaluation jobs
ve ark ListEvaluationJobs --Region "{{env.VOLCENGINE_REGION}}"

# Filter by model version
ve ark ListEvaluationJobs --Region "{{env.VOLCENGINE_REGION}}" --ModelVersionId "{{user.model_version_id}}"

# Filter by status
ve ark ListEvaluationJobs --Region "{{env.VOLCENGINE_REGION}}" --Status "Succeeded"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Job ID | `$.Result.Jobs[].JobId` | Unique identifier |
| Name | `$.Result.Jobs[].JobName` | Job name |
| Model | `$.Result.Jobs[].ModelVersionId` | Evaluated model |
| Status | `$.Result.Jobs[].Status` | Running, Succeeded, Failed |
| Score | `$.Result.Jobs[].Score` | Evaluation score |
| CreateTime | `$.Result.Jobs[].CreateTime` | Creation timestamp |

---

## Error Taxonomy

| Error Code | Meaning | Resolution |
|-----------|---------|-----------|
| `EndpointNotFound` | Endpoint does not exist | 0 retries; Check endpoint ID, suggest creation |
| `EndpointAlreadyExists` | Endpoint name conflict | 0 retries; Suggest different name |
| `ModelNotFound` | Model/version not found | 0 retries; List marketplace models for valid IDs |
| `ModelNotTrainable` | Model does not support fine-tuning | 0 retries; Suggest trainable models |
| `InvalidParameter` | Request validation failed | 1 retry; Align with OpenAPI schema |
| `InvalidHyperParameters` | Training hyperparameters invalid | 0 retries; Check supported ranges |
| `InvalidDatasetType` | Dataset type unsupported | 0 retries; Suggest supported types |
| `DatasetAlreadyExists` | Dataset name conflict | 0 retries; Use different name |
| `DatasetTooLarge` | Dataset exceeds size limit | 0 retries; Reduce dataset size |
| `TosPathNotFound` | TOS data source unreachable | 0 retries; Verify TOS bucket/path |
| `TrainingJobAlreadyExists` | Training job name conflict | 0 retries; Use different name |
| `QuotaExceeded` | Resource quota reached | 0 retries; HALT — request quota increase |
| `InsufficientBalance` | Account not funded | 0 retries; HALT — recharge required |
| `InvalidVpcConfig` | VPC configuration invalid | 0 retries; Verify VPC/subnet in region |
| `AccessDenied` | IAM permission denied | 0 retries; HALT — check IAM policies |
| `EndpointInUse` | Endpoint has active traffic | 0 retries; Stop inference before delete |
| `ResourceLimitExceeded` | GPU/resource limit hit | 0 retries; HALT — request quota increase |
| Throttling | Rate limit exceeded | 3 retries/exponential; Backoff with delay |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; Retry, escalate with RequestId |
| `InvalidJobStatus` | Job status invalid for action | 0 retries; HALT — check current status |
| `InvalidModelVersion` | Model version doesn't support training/evaluation | 0 retries; HALT — select compatible version |
| `TrainingJobNotFound` | Training job does not exist | 0 retries; HALT — verify job ID |
| `DatasetNotFound` | Dataset does not exist | 0 retries; HALT — verify dataset ID |
| `DependencyViolation` | Resource has dependencies | 0 retries; HALT — remove dependencies first |
| `EvaluationJobNotFound` | Evaluation job does not exist | 0 retries; HALT — verify job ID |
| `DatasetInUse` | Dataset referenced by active jobs | 0 retries; HALT — stop dependent jobs first |

## Prerequisites

### Install CLI

```bash
# macOS ARM64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-darwin-arm64 -o /usr/local/bin/ve && chmod +x /usr/local/bin/ve

# Linux x86_64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-linux-amd64 -o /usr/local/bin/ve && chmod +x /usr/local/bin/ve

# Verify
ve version
```

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
