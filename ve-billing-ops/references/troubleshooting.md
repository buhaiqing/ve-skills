# Troubleshooting Billing

## Common API Error Codes

| Code | Meaning | Agent Action |
|------|---------|--------------|
| InvalidParameter | Bad request | Fix parameters |
| Forbidden.RAM | Insufficient permissions | Add IAM policy |
| InsufficientBalance | Not enough balance | Recharge account |
| InternalError | Server error | Retry; then HALT |
| Throttling | Rate limit | Backoff and retry |
