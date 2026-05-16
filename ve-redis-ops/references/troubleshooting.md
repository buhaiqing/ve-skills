# Troubleshooting Redis

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `InvalidParameter.InstanceName` | Name format invalid | Use valid name (1-128 chars, alphanumeric, hyphens) |
| `InvalidParameter.Parameter` | Parameter value invalid | Check parameter against allowed values |
| `InvalidParameter.NetworkConfig` | VPC/subnet invalid | Verify network exists in region |
| `InvalidParameter.Password` | Password format invalid | 8-32 chars, mix of letters, digits, special chars |
| `ResourceNotFound.Instance` | Instance doesn't exist | Verify InstanceId via DescribeDBInstances |
| `ResourceNotFound.AllowList` | Allowlist doesn't exist | Verify AllowListId |
| `OperationDenied.InstanceStatus` | Invalid state | Wait for current operation to complete |
| `OperationDenied.DeletionProtection` | Delete protection enabled | Disable protection first |
| `QuotaExceeded.InstanceCount` | Max instances reached | Delete unused or raise quota |
| `QuotaExceeded.AllowListCount` | Max allowlists reached | Delete unused allowlists |
| `InsufficientBalance` | No funds | Recharge account |
| `InternalError` | Server error | Retry with backoff; HALT after 3 with RequestId |
| `Throttling` | Rate limit | Exponential backoff |
| `ResourceInUse` | Being used | Wait for operation to complete |
| `Forbidden.RAM` | Insufficient permissions | Add RAM policy for Redis |

## Diagnostic Order

1. Check instance status: `ve redis DescribeDBInstanceDetail --InstanceId <id>`
2. Verify VPC/subnet exists in target region
3. Check allowlist configuration (new instances start with no allowed IPs)
4. Verify engine version compatibility
5. Check quota limits

## Common Patterns

### Connection Refused
- Cause: AllowList not configured (default deny all)
- Fix: Add client IP to allowlist via ModifyAllowList or CreateAllowList

### Instance Creation Timeout
- Cause: VPC/subnet issues, capacity unavailable
- Fix: Verify network config; try different zone

### Parameter Change Not Applied
- Cause: Some parameters require restart (check ForceRestart flag)
- Fix: Restart instance after parameter modification
