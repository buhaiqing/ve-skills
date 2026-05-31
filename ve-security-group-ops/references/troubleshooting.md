# Troubleshooting Security Groups

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| InvalidParameter / 400 | Request failed validation | Align body with OpenAPI |
| SecurityGroupNotFound / 404 | SG does not exist | Check SG ID |
| SecurityGroupInUse / 409 | SG has attached instances | Detach before delete |
| Forbidden.RAM | Insufficient IAM permissions | User adds IAM policy |
| InternalError / 500 | Server-side error | Retry with backoff; then HALT |
| Throttling / 429 | Rate limit exceeded | Retry with backoff |

## Diagnostic Order

1. List SGs: `ve vpc DescribeSecurityGroups --Region <region>`
2. View SG rules: `ve vpc DescribeSecurityGroupAttributes --SecurityGroupId <id>`
3. Check attached instances: `ve ecs DescribeInstances --SecurityGroupIds '["<id>"]'`
4. Verify rule syntax: Protocol, PortRange, CidrIp format
5. Check enterprise SG priority (lower number = higher priority)
