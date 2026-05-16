# Troubleshooting RDS MySQL

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `InvalidParameter.InstanceName` | Name format invalid | Use valid name per naming rules |
| `InvalidParameter.NodeSpec` | Spec invalid | Check available specs via DescribeDBInstanceSpecs |
| `InvalidParameter.StorageSpace` | Storage out of range [20-3000]GB | Provide valid storage size |
| `InvalidParameter.Parameter` | Value outside CheckingCode | Fix value to match CheckingCode range |
| `InvalidParameter.NetworkConfig` | VPC/subnet invalid | Verify network exists in region |
| `ResourceNotFound.Instance` | Instance doesn't exist | Verify InstanceId |
| `ResourceNotFound.Account` | Account doesn't exist | Verify AccountName via DescribeDBAccounts |
| `OperationDenied.InstanceStatus` | Invalid state | Wait for current operation to complete |
| `QuotaExceeded.InstanceCount` | Max instances reached | Delete unused or raise quota |
| `QuotaExceeded.AccountCount` | Max accounts reached | Delete unused accounts |
| `InsufficientBalance` | No funds | Recharge account |
| `InternalError` | Server error | Retry with backoff; HALT after 3 with RequestId |
| `Throttling` | Rate limit | Exponential backoff |
| `ResourceInUse` | In use by another operation | Wait for operation to complete |
| `Forbidden.RAM` | Insufficient permissions | Add RAM policy for RDS |

## Diagnostic Order

1. Check instance status: `ve rds_mysql DescribeDBInstanceDetail --InstanceId <id>`
2. Verify VPC/subnet exists in target region
3. Check storage type and space compatibility with node spec
4. Verify engine version (MySQL_5_7 / MySQL_8_0)
5. Check parameter values against CheckingCode
6. Review parameter modification log: `DescribeDBInstanceParametersLog`

## Common Patterns

### Instance Creation Stuck
- Cause: Storage provisioning delay, VPC misconfiguration
- Fix: Verify VPC/subnet; check instance status for error details

### Parameter Change Not Applied
- Cause: ForceRestart=true but instance not restarted
- Fix: Restart instance after modifying parameters that require restart

### Connection Timeout
- Cause: IP whitelist not configured, security group blocking
- Fix: Add client IP to IP whitelist; verify security group allows port 3306

### Slow Spec Modification
- Cause: Storage resize or node spec change is async
- Fix: Monitor InstanceStatus until RUNNING; may take up to 900s
