# Troubleshooting DNS

## Common API Error Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| InvalidParameter | Bad request | Fix parameters |
| DomainNotFound | Domain does not exist | Check domain name |
| DomainAlreadyExists | Domain already added | Skip or check spelling |
| Forbidden.RAM | Insufficient permissions | Add IAM policy |
| InternalError | Server error | Retry; then HALT |

## Diagnostic Order

1. Verify domain exists: `ve dns ListDomains`
2. Check DNS resolution: `dig example.com @ns1.volcengine.com`
3. Verify record sets: `ve dns DescribeRecordSets --ZoneName <domain>`
4. Check TTL propagation: Wait for TTL expiry after changes
