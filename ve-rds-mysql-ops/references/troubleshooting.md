# Troubleshooting RDS MySQL

## Common Error Codes

> **TE-6:** Per-operation error codes in SKILL.md §Failure Recovery (detailed). Summary below for quick reference.

| Error Pattern | Action | Recovery |
|---|--------|----------|
| `InvalidParameter.*` | **HALT** | Fix param per API docs |
| `ResourceNotFound.*` | **HALT** | Verify ID via Describe* |
| `OperationDenied.*` | **HALT** | Wait for current op to finish |
| `QuotaExceeded.*` | **HALT** | Clean up or raise quota |
| `InsufficientBalance` | **HALT** | Recharge account |
| `InternalError` | RETRY 3x → **HALT** | Capture RequestId |
| `Throttling` | RETRY — backoff | Rate limit |
| `ResourceInUse` | WAIT | In use by another operation |
| `Forbidden.RAM` | **HALT** | Add RAM policy for RDS |

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
