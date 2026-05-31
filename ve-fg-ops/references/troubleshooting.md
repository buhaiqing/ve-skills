# Troubleshooting FunctionGraph

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| InvalidParameter / 400 | Request failed validation | Align body with OpenAPI |
| FunctionNotFound / 404 | Function does not exist | Check function name; suggest creation |
| FunctionAlreadyExists / 409 | Function name in use | Use UpdateFunction or different name |
| Forbidden.RAM | Insufficient IAM permissions | User adds IAM policy |
| InternalError / 500 | Server-side error | Retry with backoff; then HALT |
| Throttling / 429 | Rate limit exceeded | Retry with backoff |
| InvalidRuntime / 400 | Unsupported runtime | Check supported runtime list |
| ResourceLimitExceeded / 403 | Quota limit reached | Delete unused functions or request increase |

## Diagnostic Order

1. Check function exists: `ve functiongraph GetFunction --FunctionName <name>`
2. Check function status: Should be `Active`
3. View function logs: `ve functiongraph GetFunctionLogs --FunctionName <name>`
4. Check function metrics: `ve functiongraph GetFunctionMetrics --FunctionName <name>`
5. Verify trigger configuration: `ve functiongraph ListTriggers --FunctionName <name>`
6. Check IAM permissions if access denied
7. Verify network access (VPC if configured)
8. Review function code for errors (timeout, memory, exceptions)
