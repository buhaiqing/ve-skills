# Troubleshooting VKE

## Common Error Codes

| Error Code | Meaning | Agent Action |
|-----------|---------|--------------|
| `QuotaExceeded.ClusterCount` | Max clusters reached | Delete unused or request quota increase |
| `QuotaExceeded.NodeCount` | Max nodes in pool reached | Remove nodes or raise quota |
| `InvalidParameter.ClusterName` | Name format invalid | Use 1-64 chars, lowercase, digits, hyphens |
| `InvalidParameter.NodePoolName` | Pool name invalid | Fix naming per rules |
| `ResourceNotFound.Cluster` | Cluster doesn't exist | Verify ClusterId via ListClusters |
| `ResourceNotFound.NodePool` | Node pool doesn't exist | Verify NodePoolId via DescribeNodePool |
| `OperationDenied.ClusterStatus` | Invalid state for operation | Wait for state change; check current status |
| `OperationDenied.DeleteProtection` | Delete protection on | Disable delete protection first |
| `InsufficientBalance` | No funds | Recharge account |
| `InternalError` | Server error | Retry with backoff; HALT after 3 retries with RequestId |
| `Throttling` | Rate limit | Exponential backoff; respect Retry-After |
| `ResourceInUse` | Resource being used | Wait for operation to complete |

## Diagnostic Order

1. Check cluster status: `ve vke DescribeCluster --ClusterId <id>`
2. Check node pool status: `ve vke DescribeNodePool --ClusterId <id> --NodePoolId <id>`
3. Check node events: `ve vke ListNodes --ClusterId <id> --NodePoolId <id>`
4. Verify VPC/subnet exists in target region
5. Check quota: compare against limits in core-concepts.md
6. Review API logs for RequestId correlation

## Common Patterns

### Cluster Stuck in Creating
- Cause: VPC/subnet issues, quota exhausted, insufficient balance
- Fix: Verify prerequisites, check API response for specific error

### Node Not Joining Cluster
- Cause: Security group misconfiguration, network unreachable
- Fix: Verify security group allows K8s communication (6443, 10250, etc.)

### Node Pool Scaling Failing
- Cause: ECS capacity unavailable in zone, quota limit
- Fix: Try different zone or instance type; check ECS quota
