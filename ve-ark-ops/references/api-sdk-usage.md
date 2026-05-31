# API & SDK — Ark (方舟大模型平台)

## OpenAPI

- **Documentation:** https://www.volcengine.com/docs/82379
- **Service:** `ark`
- **Base endpoint:** `open.volcengineapi.com`
- **SDK Package:** `github.com/volcengine/volc-sdk-golang/service/ark`

## SDK Operations Map

| Goal | API Action | SDK Method |
|------|-----------|------------|
| List endpoints | `ListEndpoints` | `instance.Client.Request("ark", "ListEndpoints", params)` |
| Create endpoint | `CreateEndpoint` | `instance.Client.Request("ark", "CreateEndpoint", params)` |
| Describe endpoint | `DescribeEndpoint` | `instance.Client.Request("ark", "DescribeEndpoint", params)` |
| Modify endpoint | `ModifyEndpoint` | `instance.Client.Request("ark", "ModifyEndpoint", params)` |
| Delete endpoint | `DeleteEndpoint` | `instance.Client.Request("ark", "DeleteEndpoint", params)` |
| List models | `ListModels` | `instance.Client.Request("ark", "ListModels", params)` |
| Describe model | `DescribeModel` | `instance.Client.Request("ark", "DescribeModel", params)` |
| Create training job | `CreateTrainingJob` | `instance.Client.Request("ark", "CreateTrainingJob", params)` |
| List training jobs | `ListTrainingJobs` | `instance.Client.Request("ark", "ListTrainingJobs", params)` |
| Describe training job | `DescribeTrainingJob` | `instance.Client.Request("ark", "DescribeTrainingJob", params)` |
| Stop training job | `StopTrainingJob` | `instance.Client.Request("ark", "StopTrainingJob", params)` |
| List datasets | `ListDatasets` | `instance.Client.Request("ark", "ListDatasets", params)` |
| Create dataset | `CreateDataset` | `instance.Client.Request("ark", "CreateDataset", params)` |
| Describe dataset | `DescribeDataset` | `instance.Client.Request("ark", "DescribeDataset", params)` |
| Delete dataset | `DeleteDataset` | `instance.Client.Request("ark", "DeleteDataset", params)` |
| Create evaluation job | `CreateEvaluationJob` | `instance.Client.Request("ark", "CreateEvaluationJob", params)` |
| List evaluation jobs | `ListEvaluationJobs` | `instance.Client.Request("ark", "ListEvaluationJobs", params)` |
| Describe evaluation job | `DescribeEvaluationJob` | `instance.Client.Request("ark", "DescribeEvaluationJob", params)` |

## Request / Response Notes

### Common Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `Region` | string | Y | Region (e.g., `cn-beijing`) |
| `EndpointName` | string | Y* | Endpoint name (for Endpoint ops) |
| `ModelVersionId` | string | Y* | Model version to deploy |
| `TrainingJobName` | string | Y* | Training job name |
| `DatasetName` | string | Y* | Dataset name |
| `DatasetType` | string | Y* | Dataset type (`Text`, `QAPair`, `MultiTurn`) |
| `DataSourceType` | string | Y* | Data source (`TOS`, `Upload`) |
| `PageSize` / `PageNumber` | int | N | Pagination |

### Pagination
List operations support pagination via `PageNumber` (1-based) and `PageSize` (default 10, max 100).

### Idempotency
Create operations (`CreateEndpoint`, `CreateTrainingJob`, `CreateDataset`) are **not** idempotent by name — duplicate names return `ResourceAlreadyExists`. Use unique names per resource.

### Timestamps
All timestamps are ISO 8601 format: `2026-04-28T10:00:00Z`.

## Standard Response Format

```json
{
  "ResponseMetadata": {
    "RequestId": "20260428100000ABC123",
    "Action": "CreateEndpoint",
    "Version": "2024-01-01",
    "Service": "ark",
    "Region": "cn-beijing"
  },
  "Result": {
    "EndpointId": "ep-xxxxx",
    "EndpointName": "my-endpoint",
    "Status": "Creating"
  }
}
```

## Error Response Format

```json
{
  "ResponseMetadata": {
    "RequestId": "20260428100000ABC123",
    "Action": "CreateEndpoint",
    "Version": "2024-01-01",
    "Service": "ark",
    "Region": "cn-beijing",
    "Error": {
      "Code": "EndpointAlreadyExists",
      "Message": "The specified endpoint name already exists."
    }
  }
}
```

## Go SDK Setup

```bash
go get -u github.com/volcengine/volc-sdk-golang
```

Then import:
```go
import "github.com/volcengine/volc-sdk-golang/service/ark"
```

## CLI vs SDK Coverage

| Operation | ve CLI | Go SDK | Notes |
|-----------|--------|--------|-------|
| ListEndpoints | ✅ | ✅ | Primary CLI path |
| CreateEndpoint | ✅ | ✅ | CLI for simple cases |
| DescribeEndpoint | ✅ | ✅ | CLI for status checks |
| ModifyEndpoint | ✅ | ✅ | CLI for config changes |
| DeleteEndpoint | ✅ | ✅ | CLI with safety gate |
| ListModels | ✅ | ✅ | CLI for model discovery |
| DescribeModel | ✅ | ✅ | CLI for model details |
| CreateTrainingJob | ✅ | ✅ | CLI with hyperparams JSON |
| ListTrainingJobs | ✅ | ✅ | CLI for status checks |
| DescribeTrainingJob | ✅ | ✅ | CLI for job details |
| StopTrainingJob | ✅ | ✅ | CLI for stopping jobs |
| ListDatasets | ✅ | ✅ | CLI for dataset listing |
| CreateDataset | ✅ | ✅ | CLI for data source config |
| DescribeDataset | ✅ | ✅ | CLI for status checks |
| DeleteDataset | ✅ | ✅ | CLI with safety gate |
| CreateEvaluationJob | ✅ | ✅ | CLI for evaluation config |
| ListEvaluationJobs | ✅ | ✅ | CLI for status checks |
| DescribeEvaluationJob | ✅ | ✅ | CLI for evaluation details |
