# Troubleshooting SLS

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| InvalidParameter / 400 | Request failed validation | Align body with OpenAPI |
| ProjectNotFound / 404 | Project does not exist | Check project ID |
| TopicNotFound / 404 | Topic does not exist | Check topic ID |
| Forbidden.RAM | Insufficient IAM permissions | User adds IAM policy |
| InternalError / 500 | Server-side error | Retry with backoff; then HALT |
| Throttling / 429 | Rate limit exceeded | Retry with backoff |

## Diagnostic Order

1. Verify project exists: `ve tls DescribeProjects --ProjectId <id>`
2. Verify topic exists: `ve tls DescribeTopics --ProjectId <id>`
3. Check index configuration: `ve tls DescribeIndex --ProjectId <id> --TopicId <id>`
4. Search for recent logs: `ve tls SearchLogs --ProjectId <id> --TopicId <id> --Query "*"`
5. Check LogShipper status: `ve tls DescribeShippers --ProjectId <id> --TopicId <id>`
6. Verify Logtail agent on ECS if collecting from instances
