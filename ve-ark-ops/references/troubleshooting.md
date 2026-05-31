# Troubleshooting Ark

## Common API Error Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| InvalidParameter | Bad request | Fix parameters |
| EndpointNotFound | Endpoint does not exist | Check endpoint name/ID |
| ModelNotFound | Model not available | Check model name |
| InsufficientQuota | Endpoint quota exceeded | Request quota increase |
| Forbidden.RAM | Insufficient permissions | Add IAM policy |
| InternalError | Server error | Retry; then HALT |

## Diagnostic Order

1. Check endpoint status: `ve ark DescribeEndpoint`
2. Verify model availability: `ve ark ListModels`
3. Check training job status: `ve ark DescribeTrainingJob`
4. Review dataset format for training errors
