# Troubleshooting VKE

## Common Error Codes

| Error Code | Agent Action | Recovery |
|-----------|--------------|----------|
| `QuotaExceeded.ClusterCount` | Max clusters reached → delete unused or request quota increase | Escalate to account admin for limit increase |
| `QuotaExceeded.NodeCount` | Max nodes in pool reached → remove nodes or raise quota | Escalate to account admin for limit increase |
| `InvalidParameter.ClusterName` | Name format invalid → use 1-64 chars, lowercase, digits, hyphens | HALT; prompt user for valid name |
| `InvalidParameter.NodePoolName` | Pool name invalid → fix naming per rules | HALT; prompt user for valid name |
| `ResourceNotFound.Cluster` | Cluster doesn't exist → verify ClusterId via ListClusters | HALT; prompt user for correct ClusterId |
| `ResourceNotFound.NodePool` | Node pool doesn't exist → verify NodePoolId via DescribeNodePool | HALT; prompt user for correct NodePoolId |
| `OperationDenied.ClusterStatus` | Invalid state for operation → wait for state change; check current status | Poll until `Running` then retry |
| `OperationDenied.DeleteProtection` | Delete protection enabled → disable delete protection first | HALT; guide user to disable protection |
| `InsufficientBalance` | No funds → recharge account | HALT; direct user to billing console |
| `InternalError` | Server error → retry with backoff; HALT after 3 retries with RequestId | Escalate with RequestId to support |
| `Throttling` | Rate limit hit → exponential backoff; respect Retry-After | Retry with backoff; HALT after 3 retries |
| `ResourceInUse` | Resource being used → wait for operation to complete | Poll and retry

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
