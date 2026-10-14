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
| `EndpointNotFound` | Endpoint does not exist | 0 retries; **HALT** — Check endpoint ID, suggest creation |
| `EndpointAlreadyExists` | Endpoint name conflict | 0 retries; **RETRY** — Suggest different name |
| `ModelNotFound` | Model/version not found | 0 retries; **HALT** — List marketplace models for valid IDs |
| `ModelNotTrainable` | Model does not support fine-tuning | 0 retries; **HALT** — Suggest trainable models |
| `InvalidParameter` | Request validation failed | 1 retry; **RETRY** — Align with OpenAPI schema |
| `InvalidHyperParameters` | Training hyperparameters invalid | 0 retries; **HALT** — Check supported ranges |
| `InvalidDatasetType` | Dataset type unsupported | 0 retries; **HALT** — Suggest supported types |
| `DatasetAlreadyExists` | Dataset name conflict | 0 retries; **RETRY** — Use different name |
| `DatasetTooLarge` | Dataset exceeds size limit | 0 retries; **HALT** — Reduce dataset size |
| `TosPathNotFound` | TOS data source unreachable | 0 retries; **HALT** — Verify TOS bucket/path |
| `TrainingJobAlreadyExists` | Training job name conflict | 0 retries; **RETRY** — Use different name |
| `QuotaExceeded` | Resource quota reached | 0 retries; **HALT** — request quota increase |
| `InsufficientBalance` | Account not funded | 0 retries; **HALT** — recharge required |
| `InvalidVpcConfig` | VPC configuration invalid | 0 retries; **HALT** — Verify VPC/subnet in region |
| `AccessDenied` | IAM permission denied | 0 retries; **HALT** — check IAM policies |
| `EndpointInUse` | Endpoint has active traffic | 0 retries; **HALT** — Stop inference before delete |
| `ResourceLimitExceeded` | GPU/resource limit hit | 0 retries; **HALT** — request quota increase |
| `Throttling` | Rate limit exceeded | 3 retries/exponential; **RETRY** — Backoff with delay |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; **RETRY** — Retry, escalate with RequestId |
| `InvalidJobStatus` | Job status invalid for action | 0 retries; **HALT** — check current status |
| `InvalidModelVersion` | Model version doesn't support training/evaluation | 0 retries; **HALT** — select compatible version |
| `TrainingJobNotFound` | Training job does not exist | 0 retries; **HALT** — verify job ID |
| `DatasetNotFound` | Dataset does not exist | 0 retries; **HALT** — verify dataset ID |
| `DependencyViolation` | Resource has dependencies | 0 retries; **HALT** — remove dependencies first |
| `EvaluationJobNotFound` | Evaluation job does not exist | 0 retries; **HALT** — verify job ID |
| `DatasetInUse` | Dataset referenced by active jobs | 0 retries; **HALT** — stop dependent jobs first |

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

