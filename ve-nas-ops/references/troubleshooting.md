# Troubleshooting NAS

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| InvalidParameter / 400 | Request failed validation | Align body with OpenAPI |
| FileSystemNotFound / 404 | File system does not exist | Check FS ID |
| FileSystemInUse / 409 | FS has active mount targets | Delete mount targets first |
| Forbidden.RAM | Insufficient IAM permissions | User adds IAM policy |
| InternalError / 500 | Server-side error | Retry with backoff; then HALT |
| Throttling / 429 | Rate limit exceeded | Retry with backoff |
| QuotaExceeded / 403 | Resource quota limit reached | Delete unused resources or request increase |

## Diagnostic Order

1. Verify FS exists: `ve nas DescribeFileSystems --FileSystemId <id>`
2. Check mount targets: `ve nas DescribeMountTargets --FileSystemId <id>`
3. Verify network: Check VPC/subnet existence via `ve-vpc-ops`
4. Check permission groups: `ve nas DescribePermissionGroups`
5. Test mount: Mount on ECS instance and verify connectivity
6. Check snapshot status: `ve nas DescribeSnapshots --FileSystemId <id>`
