# Ark (方舟大模型平台) Core Concepts

## Platform Overview

Volcengine Ark (方舟大模型平台) is a full-lifecycle LLM platform providing model inference, fine-tuning, evaluation, and marketplace services. It hosts Doubao (豆包) series models and third-party open-source models.

## Key Resources

### 1. Inference Endpoints (推理端点)
Inference endpoints are deployed model instances that serve inference requests. Each endpoint runs a specific model version with configurable compute resources, scaling, and networking.

**States:** `Creating` → `Running` → `Stopping`/`Stopped` → `Failed`

**Types:**
- **Standard Inference:** General-purpose endpoint for Chat/Completion APIs
- **Low-latency Inference:** Optimized for real-time use cases
- **TPM-Guaranteed:** Reserved throughput (Tokens Per Minute) for production workloads

### 2. Models (模型)
Marketplace models available for deployment and fine-tuning.

**Vendors:**
- **BYTEDANCE:** Doubao series (豆包大模型)
- **OPEN_SOURCE:** Llama, Qwen, ChatGLM, etc.
- **THIRD_PARTY:** Partner-provided models

**Model Types:**
- **Chat:** Text generation / conversation
- **Embedding:** Vector embedding generation
- **Image:** Image generation and understanding
- **Video:** Video generation

### 3. Training Jobs (模型精调)
Fine-tuning jobs that adapt base models to custom datasets using SFT (Supervised Fine-Tuning), DPO (Direct Preference Optimization), or RL (Reinforcement Learning).

**States:** `Pending` → `Running` → `Succeeded`/`Failed`/`Stopped`

### 4. Datasets (数据集)
Structured data used for model training and evaluation.

**Types:**
- **Text:** Raw text data
- **QAPair:** Question-answer pairs for SFT
- **MultiTurn:** Multi-turn dialogue data
- **Preference:** Preference pairs for DPO/RL

**Sources:**
- **TOS:** Import from Volcengine TOS buckets
- **Upload:** Direct file upload

### 5. Evaluation Jobs (模型评估)
Automated evaluation of model performance on benchmark datasets and custom test sets.

## Region & Endpoint

- Ark resources are region-scoped
- Available regions: `cn-beijing`, `cn-shanghai`, `cn-guangzhou`
- API endpoint: `open.volcengineapi.com` (service: `ark`)
- Console: `https://console.volcengine.com/ark`

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Ark Platform                    │
│  ┌───────────┐  ┌──────────┐  ┌──────────────┐ │
│  │ Inference │  │ Training │  │   Evaluation  │ │
│  │ Endpoints │  │   Jobs   │  │     Jobs      │ │
│  └───────────┘  └──────────┘  └──────────────┘ │
│  ┌───────────┐  ┌──────────┐  ┌──────────────┐ │
│  │   Model   │  │ Datasets │  │    Model      │ │
│  │Marketplace│  │          │  │   Registry   │ │
│  └───────────┘  └──────────┘  └──────────────┘ │
└─────────────────────────────────────────────────┘
         ▲              ▲              ▲
         │              │              │
    ve ark CLI     Go SDK        OpenAPI
```

## IAM Permissions

Ark resources are governed by IAM policies. Common action prefixes:
- `ark:ListEndpoints`, `ark:CreateEndpoint`, `ark:DeleteEndpoint`
- `ark:ListTrainingJobs`, `ark:CreateTrainingJob`
- `ark:ListDatasets`, `ark:CreateDataset`
- `ark:ListModels`, `ark:DescribeModel`
- `ark:CreateEvaluationJob`, `ark:ListEvaluationJobs`

## Related Services

| Service | Relation |
|---------|----------|
| TOS | Dataset storage (model training data) |
| VPC | Private network access for endpoints |
| IAM | Access control and permissions |
| CMS | Endpoint monitoring and alerting |
| Billing | Usage tracking and cost management |
