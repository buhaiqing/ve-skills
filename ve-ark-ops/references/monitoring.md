# Monitoring Ark (方舟大模型平台)

## Key Metrics

### Inference Endpoint Metrics

| Metric | Namespace | Description |
|--------|-----------|-------------|
| `ark_endpoint_invocation_count` | Ark | Number of invocations per endpoint |
| `ark_endpoint_invocation_latency` | Ark | Average invocation latency (ms) |
| `ark_endpoint_invocation_error_rate` | Ark | Error rate of invocations (%) |
| `ark_endpoint_invocation_token_count` | Ark | Total tokens processed (input + output) |
| `ark_endpoint_cpu_utilization` | Ark | CPU usage (%): 0-100 |
| `ark_endpoint_memory_utilization` | Ark | Memory usage (%): 0-100 |
| `ark_endpoint_replica_count` | Ark | Number of running replicas |
| `ark_endpoint_status` | Ark | Endpoint status: Running, Creating, Failed, Stopped |

### Training Job Metrics

| Metric | Namespace | Description |
|--------|-----------|-------------|
| `ark_training_job_loss` | Ark | Training loss value |
| `ark_training_job_accuracy` | Ark | Validation accuracy |
| `ark_training_job_progress` | Ark | Training progress (%) |
| `ark_training_job_duration` | Ark | Elapsed training time (seconds) |
| `ark_training_job_gpu_utilization` | Ark | GPU utilization (%): 0-100 |

### Dataset Metrics

| Metric | Namespace | Description |
|--------|-----------|-------------|
| `ark_dataset_size_bytes` | Ark | Dataset size in bytes |
| `ark_dataset_record_count` | Ark | Number of records in dataset |
| `ark_dataset_status` | Ark | Dataset status: Available, Processing, Failed |

## CMS Integration

Ark metrics can be monitored through Volcengine Cloud Monitoring Service (CMS).

### Setting Up Alerts via CLI

```bash
# Create alarm rule for endpoint error rate (via ve-cms-ops)
ve cms CreateAlarmRule \
  --AlarmRuleName "ark-endpoint-error-rate" \
  --Namespace "ark" \
  --MetricName "ark_endpoint_invocation_error_rate" \
  --Condition '{"ComparisonOperator":"GreaterThanThreshold","Threshold":5,"Period":300}' \
  --SilenceTime 600 \
  --Region "{{env.VOLCENGINE_REGION}}"

# Alert on endpoint failure
ve cms CreateAlarmRule \
  --AlarmRuleName "ark-endpoint-failed" \
  --Namespace "ark" \
  --MetricName "ark_endpoint_status" \
  --Condition '{"ComparisonOperator":"GreaterThanThreshold","Threshold":0,"Period":60}' \
  --SilenceTime 300 \
  --Region "{{env.VOLCENGINE_REGION}}"
```

## Logs

### Endpoint Logs
Endpoint inference logs are available via Ark console log viewer. For custom logging, integrate with your application's logging framework.

### Training Logs
Training job logs track:
- Training/validation loss per epoch
- Learning rate schedule
- GPU memory utilization
- Hyperparameter configurations

Check training logs:
```bash
ve ark DescribeTrainingJob \
  --TrainingJobId "{{output.training_job_id}}" \
  --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result.TrainingMetrics'
```

## Recommended Alerts

| Alert | Metric | Threshold | Severity |
|-------|--------|-----------|----------|
| High endpoint error rate | `ark_endpoint_invocation_error_rate` | > 5% over 5 min | Critical |
| High endpoint latency | `ark_endpoint_invocation_latency` | > 5000ms p95 | Warning |
| Endpoint failed | `ark_endpoint_status` | == Failed | Critical |
| Endpoint replica drop | `ark_endpoint_replica_count` | < MinReplicas | Warning |
| Training job failure | `ark_training_job_progress` | No progress > 1h | Warning |
| High GPU utilization | `ark_training_job_gpu_utilization` | > 95% sustained | Info |

## Dashboard Setup

Volcengine CMS dashboards can visualize Ark metrics. Configure a dashboard with:
- Endpoint invocation rate and latency graphs
- Token consumption over time
- Training job progress and loss curves
- Active endpoint count and status breakdown