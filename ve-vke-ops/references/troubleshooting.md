# Troubleshooting VKE

## Common Error Codes

| Error Code | Action | Recovery |
|-----------|--------|----------|
| `QuotaExceeded.ClusterCount` | HALT | Delete unused or request increase |
| `QuotaExceeded.NodeCount` | HALT | Remove nodes or raise quota |
| `InvalidParameter.ClusterName` | HALT | 1-64 chars, lowercase, digits, hyphens |
| `InvalidParameter.NodePoolName` | HALT | Fix naming per rules |
| `ResourceNotFound.Cluster` | HALT | Verify ClusterId via ListClusters |
| `ResourceNotFound.NodePool` | HALT | Verify NodePoolId via DescribeNodePool |
| `OperationDenied.ClusterStatus` | Retry | Poll until `Running`, then retry |
| `OperationDenied.DeleteProtection` | HALT | Disable protection first |
| `InsufficientBalance` | HALT | Recharge account |
| `InternalError` | Retry ×3 | HALT with RequestId |
| `Throttling` | Retry ×3, exponential | Back off |
| `ResourceInUse` | Retry | Poll and retry

## Diagnostic Order

1. ✅ Check cluster status: `ve vke DescribeCluster --ClusterId <id>`
2. ✅ Check node pool status: `ve vke DescribeNodePool --ClusterId <id> --NodePoolId <id>`
3. ✅ Check node events: `ve vke ListNodes --ClusterId <id> --NodePoolId <id>`
4. ✅ Verify VPC/subnet exists in target region
5. ✅ Check quota: compare against limits in core-concepts.md
6. ✅ Review API logs for RequestId correlation

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
