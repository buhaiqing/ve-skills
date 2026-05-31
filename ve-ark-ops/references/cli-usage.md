# CLI — Ark (方舟大模型平台) (`ve ark`)

## Install and Config

- **Install:** Download `ve` binary from [Volcengine CLI Releases](https://github.com/volcengine/volcengine-cli/releases)
- **Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json` (JSON format).
- **Verify:** `ve ark --help`

## Conventions (Agent Execution)

- Output is **JSON by default**
- Service prefix: `ve ark`
- Invocation: `ve ark <Action> --Parameter value`
- For JSON body parameters: use `--body '{"key":"value"}'`
- All operations require `--Region` parameter

## Command Map

### Endpoint Operations

| Goal | `ve` Invocation | Notes |
|------|----------------|-------|
| List endpoints | `ve ark ListEndpoints --Region cn-beijing` | Default output is JSON |
| Get endpoint | `ve ark DescribeEndpoint --EndpointId "ep-xxx" --Region cn-beijing` | Check status field |
| Create endpoint | `ve ark CreateEndpoint --EndpointName "prod-llm" --ModelVersionId "mv-xxx" --Region cn-beijing` | Poll for Running state |
| Modify endpoint | `ve ark ModifyEndpoint --EndpointId "ep-xxx" --MinReplicas 2 --MaxReplicas 10 --Region cn-beijing` | Updates config only |
| Delete endpoint | `ve ark DeleteEndpoint --EndpointId "ep-xxx" --Region cn-beijing` | Requires confirmation |

### Model Operations

| Goal | `ve` Invocation | Notes |
|------|----------------|-------|
| List models | `ve ark ListModels --Region cn-beijing` | Filter by vendor/type |
| Get model | `ve ark DescribeModel --ModelId "model-xxx" --Region cn-beijing` | Shows versions |
| List by vendor | `ve ark ListModels --ModelVendor BYTEDANCE --Region cn-beijing` | BYTEDANCE / OPEN_SOURCE / THIRD_PARTY |

### Training Operations

| Goal | `ve` Invocation | Notes |
|------|----------------|-------|
| List training jobs | `ve ark ListTrainingJobs --Region cn-beijing` | Filter by status |
| Get training job | `ve ark DescribeTrainingJob --TrainingJobId "tj-xxx" --Region cn-beijing` | Shows progress |
| Create training job | `ve ark CreateTrainingJob --TrainingJobName "my-sft" --ModelId "model-xxx" --DatasetId "ds-xxx" --HyperParameters '{"epochs":3}' --Region cn-beijing` | JSON hyperparameters |
| Stop training job | `ve ark StopTrainingJob --TrainingJobId "tj-xxx" --Region cn-beijing` | Irreversible |
| Filter by status | `ve ark ListTrainingJobs --Status Running --Region cn-beijing` | Pending/Running/Succeeded/Failed |

### Dataset Operations

| Goal | `ve` Invocation | Notes |
|------|----------------|-------|
| List datasets | `ve ark ListDatasets --Region cn-beijing` | Filter by type |
| Get dataset | `ve ark DescribeDataset --DatasetId "ds-xxx" --Region cn-beijing` | Check availability |
| Create dataset | `ve ark CreateDataset --DatasetName "train-data" --DatasetType QAPair --DataSourceType TOS --TosPath "tos://bucket/prefix/" --Region cn-beijing` | TOS or Upload |
| Delete dataset | `ve ark DeleteDataset --DatasetId "ds-xxx" --Region cn-beijing` | Requires confirmation |

### Evaluation Operations

| Goal | `ve` Invocation | Notes |
|------|----------------|-------|
| List evaluations | `ve ark ListEvaluationJobs --Region cn-beijing` | Filter by status |
| Get evaluation | `ve ark DescribeEvaluationJob --EvaluationJobId "ej-xxx" --Region cn-beijing` | Shows metrics |
| Create evaluation | `ve ark CreateEvaluationJob --EvaluationJobName "eval-1" --ModelVersionId "mv-xxx" --DatasetId "ds-xxx" --Region cn-beijing` | Evaluates model on dataset |

## CLI Output Examples

### ListEndpoints Response
```json
{
  "Result": {
    "Endpoints": [
      {
        "EndpointId": "ep-20260428100000",
        "EndpointName": "prod-llm",
        "ModelVersionId": "mv-20260428090000",
        "Status": "Running",
        "CreateTime": "2026-04-28T10:00:00Z"
      }
    ],
    "Total": 1
  },
  "ResponseMetadata": {
    "RequestId": "req-xxx",
    "Action": "ListEndpoints",
    "Service": "ark",
    "Region": "cn-beijing"
  }
}
```

### CreateEndpoint Response
```json
{
  "Result": {
    "EndpointId": "ep-20260428100000",
    "EndpointName": "prod-llm",
    "Status": "Creating"
  },
  "ResponseMetadata": {
    "RequestId": "req-xxx",
    "Action": "CreateEndpoint",
    "Service": "ark"
  }
}
```

## Useful jq Filters

```bash
# Extract endpoint IDs and status
ve ark ListEndpoints --Region cn-beijing | jq '.Result.Endpoints[] | {id: .EndpointId, name: .EndpointName, status: .Status}'

# Get total count
ve ark ListEndpoints --Region cn-beijing | jq '.Result.Total'

# Filter running endpoints
ve ark ListEndpoints --Region cn-beijing | jq '.Result.Endpoints[] | select(.Status == "Running")'

# Extract model IDs
ve ark ListModels --Region cn-beijing | jq '.Result.Models[].ModelId'

# Training job status as table (TSV)
ve ark ListTrainingJobs --Region cn-beijing | jq -r '.Result.TrainingJobs[] | [.TrainingJobId, .TrainingJobName, .Status] | @tsv'
```

## CLI Coverage Gap

The `ve ark` CLI covers all Ark operations documented in this skill. No SDK-only operations identified at this time. If an operation is missing from CLI, use the JIT Go SDK fallback path described in the main SKILL.md.