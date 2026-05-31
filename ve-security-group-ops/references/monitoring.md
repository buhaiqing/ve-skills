# Monitoring Security Groups

## Key Metrics

Security Group monitoring is primarily about auditing, not real-time metrics. Use CMS for ECS-level network metrics.

## Audit Recommendations

| Check | Frequency | Method |
|-------|-----------|--------|
| Unused security groups | Monthly | Check instance association count |
| Overly permissive rules | Weekly | Scan for 0.0.0.0/0 rules |
| Rule count per SG | Monthly | Check against quota limit |
| Default SG usage | Monthly | Verify default SG rules |

## Alert Recommendations

| Alert Name | Condition | Severity |
|------------|-----------|----------|
| New 0.0.0.0/0 Rule | Any inbound rule with `CidrIp=0.0.0.0/0` | Warning |
| Sensitive Port Exposed | Port 22/3389/3306/6379 open to internet | Critical |
| Unused SG Detected | SG with 0 attached instances > 30 days | Warning |
