# KMS Monitoring & Alerts

## Overview

Monitoring KMS operations is critical for security and compliance. This guide covers key metrics, CloudTrail integration, and alerting strategies.

## Key Metrics

### API Call Metrics

| Metric | Description | Source |
|--------|-------------|--------|
| `KMS.API.Requests` | Total API requests | CloudTrail |
| `KMS.API.Errors` | Failed API requests | CloudTrail |
| `KMS.API.Latency` | API response time | CloudTrail |

### Key Usage Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `KMS.Key.EncryptCalls` | Encryption operations per key | Sudden spike |
| `KMS.Key.DecryptCalls` | Decryption operations per key | Sudden spike |
| `KMS.Key.GenerateDataKeyCalls` | Data key generation | Baseline deviation |
| `KMS.Key.Disabled` | Key state changes to disabled | Immediate |
| `KMS.Key.PendingDeletion` | Key scheduled for deletion | Immediate |

### Security Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `KMS.Grant.Created` | New grants created | Review weekly |
| `KMS.Grant.Revoked` | Grants revoked | Review weekly |
| `KMS.Policy.Changed` | Key policy modifications | Immediate |
| `KMS.Key.AccessDenied` | Access denied errors | > 10 in 5 min |

## CloudTrail Integration

### Enable CloudTrail for KMS

KMS API calls are automatically logged to CloudTrail. Key events to monitor:

### Critical Events

| Event Name | Severity | Description |
|------------|----------|-------------|
| `ScheduleKeyDeletion` | **Critical** | Key scheduled for deletion |
| `CancelKeyDeletion` | High | Pending deletion canceled |
| `DisableKey` | **Critical** | Key disabled |
| `PutKeyPolicy` | **Critical** | Key access policy changed |
| `CreateGrant` | High | New grant created |
| `RevokeGrant` | Medium | Grant revoked |
| `Decrypt` | Medium | Data decryption |
| `GenerateDataKey` | Medium | Data key generated |

### CloudTrail Event Examples

**ScheduleKeyDeletion Event:**
```json
{
  "eventVersion": "1.0",
  "eventSource": "kms.volcengineapi.com",
  "eventName": "ScheduleKeyDeletion",
  "eventTime": "2026-05-27T10:00:00Z",
  "requestParameters": {
    "KeyId": "key-1234567890abcdef",
    "PendingWindowInDays": 7
  },
  "responseElements": null,
  "requestID": "req-1234567890",
  "eventID": "evt-1234567890",
  "readOnly": false,
  "eventType": "AwsApiCall",
  "recipientAccountId": "123456789"
}
```

**Decrypt Event:**
```json
{
  "eventVersion": "1.0",
  "eventSource": "kms.volcengineapi.com",
  "eventName": "Decrypt",
  "eventTime": "2026-05-27T10:05:00Z",
  "requestParameters": {
    "EncryptionContext": {
      "Purpose": "ApplicationData"
    }
  },
  "responseElements": {
    "KeyId": "arn:volc:kms:cn-beijing:123456789:key/key-1234567890abcdef"
  },
  "requestID": "req-2345678901",
  "eventID": "evt-2345678901",
  "readOnly": true,
  "eventType": "AwsApiCall",
  "recipientAccountId": "123456789"
}
```

## Alert Configuration

### Critical Alerts (Immediate Response)

```yaml
# Key Deletion Scheduled
Alert: KMS_Key_Deletion_Scheduled
Condition: eventName = "ScheduleKeyDeletion"
Severity: P1-Critical
Action: Notify security team immediately
Auto-Response: Create incident ticket

# Key Disabled
Alert: KMS_Key_Disabled
Condition: eventName = "DisableKey"
Severity: P1-Critical
Action: Notify security and operations teams

# Key Policy Changed
Alert: KMS_Policy_Changed
Condition: eventName = "PutKeyPolicy"
Severity: P1-Critical
Action: Require approval workflow
```

### High Priority Alerts

```yaml
# Excessive Failed Decrypt Attempts
Alert: KMS_High_Decrypt_Failures
Condition: Decrypt errors > 50 in 5 minutes
Severity: P2-High
Action: Investigate potential attack

# Unusual Encryption Volume
Alert: KMS_Unusual_Encrypt_Volume
Condition: Encrypt calls > 3x baseline
Severity: P2-High
Action: Review for crypto mining or ransomware

# Grant Created Outside Business Hours
Alert: KMS_Grant_Off_Hours
Condition: eventName = "CreateGrant" AND hour NOT IN (9-18)
Severity: P2-High
Action: Review grant necessity
```

### Medium Priority Alerts

```yaml
# Key Rotation Status
Alert: KMS_Rotation_Disabled
Condition: Automatic rotation disabled
Severity: P3-Medium
Action: Review and re-enable if needed

# Pending Deletion Keys
Alert: KMS_Pending_Deletion
Condition: KeyState = "PendingDeletion"
Severity: P3-Medium
Action: Weekly review
```

## Monitoring Dashboard

### Key Metrics Panel

```
┌─────────────────────────────────────────────────────────┐
│ KMS Overview                                            │
├─────────────────────────────────────────────────────────┤
│ Active Keys:        45    │ Disabled Keys:     3       │
│ Pending Deletion:   2     │ Total Grants:      127     │
│ API Calls (24h):    15.2K │ Error Rate:        0.1%    │
└─────────────────────────────────────────────────────────┘
```

### Security Panel

```
┌─────────────────────────────────────────────────────────┐
│ Security Events (Last 24h)                              │
├─────────────────────────────────────────────────────────┤
│ Key Deletions Scheduled:     0                          │
│ Policy Changes:              2  (Review)                │
│ Failed Auth Attempts:        15 (Investigate)           │
│ Off-Hours Access:            3  (Review)                │
└─────────────────────────────────────────────────────────┘
```

## Query Examples

### CloudTrail Queries

```sql
-- Find all key deletions in the last 7 days
SELECT eventTime, eventName, requestParameters.KeyId, userIdentity.arn
FROM cloudtrail_logs
WHERE eventSource = 'kms.volcengineapi.com'
  AND eventName = 'ScheduleKeyDeletion'
  AND eventTime > now() - interval '7' day
ORDER BY eventTime DESC;

-- Find failed decryption attempts
SELECT eventTime, userIdentity.arn, errorCode, errorMessage
FROM cloudtrail_logs
WHERE eventSource = 'kms.volcengineapi.com'
  AND eventName = 'Decrypt'
  AND errorCode IS NOT NULL
  AND eventTime > now() - interval '1' hour
ORDER BY eventTime DESC;

-- Find policy changes
SELECT eventTime, userIdentity.arn, requestParameters.KeyId
FROM cloudtrail_logs
WHERE eventSource = 'kms.volcengineapi.com'
  AND eventName = 'PutKeyPolicy'
  AND eventTime > now() - interval '24' hour
ORDER BY eventTime DESC;
```

### Log Analysis

```bash
# Count API calls by operation
ve cloudtrail LookupEvents \
  --LookupAttributes AttributeKey=EventSource,AttributeValue=kms.volcengineapi.com \
  --Region cn-beijing | jq -r '.Events[].CloudTrailEvent' | jq -r '.eventName' | sort | uniq -c | sort -rn

# Find specific key usage
ve cloudtrail LookupEvents \
  --LookupAttributes AttributeKey=ResourceName,AttributeValue=key-1234567890abcdef \
  --Region cn-beijing
```

## Compliance Monitoring

### Rotation Compliance

```bash
# Check keys without rotation enabled
for key in $(ve kms DescribeKeys --Region cn-beijing | jq -r '.Result.Keys[].KeyId'); do
  ROTATION=$(ve kms DescribeKeyRotation --KeyId "$key" --Region cn-beijing | jq -r '.Result.KeyRotationEnabled')
  if [ "$ROTATION" = "false" ]; then
    echo "Key without rotation: $key"
  fi
done
```

### Access Review

```bash
# List all grants for audit
for key in $(ve kms DescribeKeys --Region cn-beijing | jq -r '.Result.Keys[].KeyId'); do
  echo "=== Key: $key ==="
  ve kms ListGrants --KeyId "$key" --Region cn-beijing | jq -r '.Result.Grants[] | [.GranteePrincipal, .Operations, .CreationDate] | @tsv'
done
```

## Incident Response

### Key Compromise Response

1. **Immediate Actions:**
   ```bash
   # Disable the compromised key
   ve kms DisableKey --KeyId "{{compromised_key}}" --Region {{env.VOLCENGINE_REGION}}

   # Create new key
   NEW_KEY=$(ve kms CreateKey --KeySpec AES_256 --KeyUsage ENCRYPT_DECRYPT --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyId')

   # Rotate all data keys (re-encrypt data)
   # [Application-specific re-encryption logic]

   # Schedule old key deletion after re-encryption
   ve kms ScheduleKeyDeletion --KeyId "{{compromised_key}}" --PendingWindowInDays 7 --Region {{env.VOLCENGINE_REGION}}
   ```

2. **Investigation:**
   - Review CloudTrail for unauthorized access
   - Identify affected resources
   - Document timeline

3. **Notification:**
   - Alert security team
   - Notify affected stakeholders
   - Update incident response log

### Unauthorized Access Response

```bash
# Immediately revoke suspicious grants
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | \
  jq -r '.Result.Grants[] | select(.GranteePrincipal == "suspicious-principal") | .GrantId' | \
  while read grant; do
    ve kms RevokeGrant --KeyId "{{user.key_id}}" --GrantId "$grant" --Region {{env.VOLCENGINE_REGION}}
  done

# Update key policy to deny suspicious principal
# [Policy update logic]
```

## Best Practices

1. **Enable CloudTrail** for all regions with KMS usage
2. **Set up real-time alerts** for critical events (deletion, disable, policy change)
3. **Regular access reviews** - monthly grant and policy audits
4. **Baseline metrics** - establish normal usage patterns
5. **Incident response plan** - documented procedures for key compromise
6. **Retention policies** - retain CloudTrail logs for compliance period
