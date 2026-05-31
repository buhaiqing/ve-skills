# Troubleshooting Ark (方舟大模型平台)

## Diagnostic Order

1. **Verify credentials** — Check `VOLCENGINE_ACCESS_KEY` and `VOLCENGINE_SECRET_KEY` are set
2. **Verify region** — Confirm `VOLCENGINE_REGION` is a supported Ark region
3. **Check resource existence** — Use `DescribeEndpoint`/`DescribeTrainingJob`/`DescribeDataset`
4. **Inspect error response** — Parse `ResponseMetadata.Error.Code` and `Message`
5. **Check regional availability** — Some models/features are region-specific
6. **Verify CLI coverage** — `ve ark --help`
7. **Escalate with RequestId** — Include `RequestId` when contacting support

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| `EndpointNotFound` / 404 | Endpoint ID does not exist | Check endpoint ID; list existing endpoints |
| `EndpointAlreadyExists` / 409 | Endpoint name conflict | Use a unique endpoint name |
| `ModelNotFound` / 404 | Model or version not found | List marketplace models for valid IDs |
| `ModelNotTrainable` / 400 | Model doesn't support fine-tuning | Choose a trainable model (check model listing) |
| `InvalidParameter` / 400 | Request failed validation | Align body with OpenAPI schema |
| `InvalidHyperParameters` / 400 | Training hyperparams invalid | Check supported ranges and format |
| `InvalidDatasetType` / 400 | Dataset type not supported | Use Text, QAPair, or MultiTurn |
| `DatasetTooLarge` / 413 | Dataset exceeds size limit | Reduce dataset or split into chunks |
| `TosPathNotFound` / 400 | TOS path unreachable | Verify bucket and object prefix |
| `TrainingJobAlreadyExists` / 409 | Training job name conflict | Use a different name |
| `QuotaExceeded` / 403 | Resource quota limit | Delete unused resources or request increase |
| `InsufficientBalance` / 403 | Account balance insufficient | Recharge account |
| `InvalidVpcConfig` / 400 | VPC configuration invalid | Verify VPC/subnet exist in same region |
| `AccessDenied` / 403 | IAM permission denied | Check IAM policies for Ark actions |
| `EndpointInUse` / 409 | Endpoint has active traffic | Stop inference before deletion |
| `ResourceLimitExceeded` / 403 | GPU/quota limit reached | Request quota increase via support |
| `Throttling` / 429 | Rate limit exceeded | Back off with exponential delay |
| `InternalError` / 500 | Server-side error | Retry with backoff; escalate with RequestId |

## Common Failure Scenarios

### Scenario 1: Endpoint Creation Fails with QuotaExceeded
```
[ERROR] QuotaExceeded: Endpoint quota limit reached.
Causes: Maximum concurrent endpoints reached (default varies by account type).
Fix: Delete unused endpoints or request quota increase from support.
```

### Scenario 2: Training Job Stays in Pending
```
[ERROR] Training job "my-job" has been in Pending state for > 5 minutes.
Causes: Resource contention for GPU nodes.
Fix: Wait for resources to become available. Consider using a different model size.
```

### Scenario 3: CreateEndpoint Fails with ModelNotFound
```
[ERROR] ModelNotFound: Model version "mv-xxx" not found.
Causes: Model version ID is incorrect or the model is not available in current region.
Fix: Run `ve ark ListModels --Region cn-beijing` to find valid model version IDs.
```

### Scenario 4: Dataset Creation Fails with TosPathNotFound
```
[ERROR] TosPathNotFound: The specified TOS path does not exist.
Causes: Bucket or object prefix is incorrect, or TOS credentials insufficient.
Fix: Validate `tos://bucket/prefix/` exists. Check TOS bucket is in same region.
```

### Scenario 5: DeleteEndpoint Fails with EndpointInUse
```
[ERROR] EndpointInUse: Endpoint is currently serving inference traffic.
Causes: Active API calls to the endpoint.
Fix: Stop all inference traffic to the endpoint before deletion. Contact application owners.
```

## Validation Commands

```bash
# 1. Check credentials
test -n "$VOLCENGINE_ACCESS_KEY" || echo "ACCESS_KEY missing"
test -n "$VOLCENGINE_SECRET_KEY" || echo "SECRET_KEY missing"

# 2. Verify CLI works
ve version
ve ark --help

# 3. List resources to verify connectivity
ve ark ListEndpoints --Region "$VOLCENGINE_REGION"
ve ark ListModels --Region "$VOLCENGINE_REGION"

# 4. Check specific resource status
ve ark DescribeEndpoint --EndpointId "ep-xxx" --Region "$VOLCENGINE_REGION"

# 5. Check RequestId in error responses (include when escalating)
echo "RequestId: req-xxx"
```

## Support Escalation

When escalating to Volcengine support, include:
- **RequestId** from the failed API response
- **Action** that failed
- **Region** used
- **Timestamp** of failure
- **Full error response** (mask credentials)