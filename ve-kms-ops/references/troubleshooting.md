# KMS Troubleshooting Guide

## Common API Error Codes

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `InvalidParameter` / 400 | **HALT** — parameter validation failed | Check parameter values against API docs |
| `InvalidArn` / 400 | **HALT** — ARN format invalid | Verify ARN follows correct pattern |
| `NotFound` / 404 | **HALT** — key/resource not found | Verify KeyId exists and is correct |
| `Disabled` / 409 | **HALT** — key is disabled | Enable key using EnableKey operation |
| `KMSInvalidState` / 409 | **HALT** — key in wrong state | Check key state and ensure compatible |
| `InvalidKeyUsage` / 400 | **HALT** — key usage doesn't support operation | Use compatible key (e.g., ENCRYPT_DECRYPT for encryption) |
| `DependencyViolation` / 409 | **HALT** — key in use by other resources | Remove dependencies (grants) before deletion |
| `UnsupportedOperation` / 400 | **HALT** — operation not supported for key type | Use different key or operation |
| `IncorrectEncryptionContext` / 400 | **HALT** — encryption context mismatch | Use same encryption context as encryption |
| `InvalidCiphertext` / 400 | **HALT** — ciphertext malformed | Verify ciphertext is complete and correct |
| `KeyUnavailable` / 503 | Retry with exponential backoff | Key temporarily unavailable |
| `AccessDenied` / 403 | **HALT** — insufficient IAM permissions | Add required IAM policy |
| `QuotaExceeded` / 429 | **HALT** — resource quota limit reached | Delete unused keys or request quota increase |
| `Throttling` / 429 | Back off and retry | Exponential backoff, max 3 retries |
| `InternalError` / 500 | Retry; then escalate with RequestId | Retry with backoff; **HALT** after 3 |

## Diagnostic Order

### 1. Verify Key Exists and State

```bash
# Check key metadata and state
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Expected output shows KeyState
```

### 2. Verify IAM Permissions

```bash
# Test basic access
ve kms DescribeKeys --Region {{env.VOLCENGINE_REGION}}

# If this fails, check IAM permissions
```

### 3. Check Key Usage Compatibility

```bash
# Get key metadata
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq '.Result.KeyMetadata | {KeySpec, KeyUsage, KeyState}'
```

### 4. Verify Encryption Context (for decrypt)

```bash
# Decrypt without encryption context first
# If that fails, you need the exact context used during encryption
```

## Common Issues

### Issue: "AccessDenied" Error

**Symptoms:** API returns `AccessDenied` error

**Diagnosis:**
```bash
# Check current IAM policies attached to user/role
# Verify the policy includes required KMS actions
```

**Solution:**
Add IAM policy with required permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:volc:kms:*:*:key/{{user.key_id}}"
    }
  ]
}
```

### Issue: "Disabled" Error

**Symptoms:** Encryption/decryption fails with `Disabled` error

**Diagnosis:**
```bash
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyState'
# Returns: Disabled
```

**Solution:**
```bash
# Enable the key
ve kms EnableKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

### Issue: "IncorrectEncryptionContext" Error

**Symptoms:** Decrypt fails even with correct ciphertext

**Diagnosis:**
The encryption context used during decryption doesn't match encryption.

**Solution:**
Use the exact same encryption context key-value pairs:

```bash
# If encrypted with:
ve kms Encrypt --KeyId "key-xxx" --Plaintext "..." --EncryptionContext '{"App":"MyApp","Env":"Prod"}'

# Must decrypt with:
ve kms Decrypt --CiphertextBlob "..." --EncryptionContext '{"App":"MyApp","Env":"Prod"}'
```

### Issue: Key Deletion Failed - "DependencyViolation"

**Symptoms:** ScheduleKeyDeletion fails with dependency error

**Diagnosis:**
```bash
# Check if key has grants
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Check what resources use this key (via CloudTrail or service-specific APIs)
```

**Solution:**
1. Revoke all grants:
```bash
# List and revoke each grant
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.Grants[].GrantId' | while read grant; do
  ve kms RevokeGrant --KeyId "{{user.key_id}}" --GrantId "$grant" --Region {{env.VOLCENGINE_REGION}}
done
```

2. Update dependent resources to use different key

3. Retry deletion scheduling

### Issue: "InvalidKeyUsage" Error

**Symptoms:** Cannot encrypt with RSA key

**Diagnosis:**
```bash
# Check key usage
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyUsage'
# Returns: SIGN_VERIFY
```

**Solution:**
Create a new key with `ENCRYPT_DECRYPT` usage:

```bash
ve kms CreateKey --KeySpec AES_256 --KeyUsage ENCRYPT_DECRYPT --Region {{env.VOLCENGINE_REGION}}
```

### Issue: Decryption Returns Wrong Data

**Symptoms:** Decrypt succeeds but output doesn't match original

**Possible Causes:**
1. Ciphertext was truncated or modified
2. Wrong encoding (base64)
3. Wrong key was used

**Diagnosis:**
```bash
# Verify ciphertext encoding
echo "{{user.ciphertext}}" | base64 -d > /dev/null && echo "Valid base64" || echo "Invalid base64"

# Check if ciphertext contains correct key ID info
```

**Solution:**
- Ensure ciphertext is transmitted/stored without modification
- Verify base64 encoding/decoding steps

### Issue: "QuotaExceeded" Error

**Symptoms:** Cannot create new key

**Diagnosis:**
```bash
# Count existing keys
ve kms DescribeKeys --Region {{env.VOLCENGINE_REGION}} | jq '.Result.Keys | length'
```

**Solution:**
1. Delete unused keys (with waiting period)
2. Request quota increase from Volcengine support

### Issue: Grant Not Working

**Symptoms:** Grantee principal cannot use key despite grant

**Diagnosis:**
```bash
# Verify grant exists
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq '.Result.Grants[] | select(.GranteePrincipal == "{{user.grantee_principal}}")'

# Check grant constraints
```

**Solution:**
- Verify GranteePrincipal ARN is correct
- Check if grant has constraints that limit operations
- Ensure grant hasn't been revoked

## Debugging Commands

### Get Full Key Metadata

```bash
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq '.Result.KeyMetadata'
```

### List All Keys with States

```bash
ve kms DescribeKeys --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.Keys[] | [.KeyId, .KeyState, .KeyUsage, .KeySpec] | @tsv'
```

### Test Encryption/Decryption Roundtrip

```bash
# Test data
TEST_DATA="$(echo 'Hello, KMS!' | base64)"

# Encrypt
CIPHER=$(ve kms Encrypt --KeyId "{{user.key_id}}" --Plaintext "$TEST_DATA" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.CiphertextBlob')
echo "Encrypted: $CIPHER"

# Decrypt
PLAIN=$(ve kms Decrypt --CiphertextBlob "$CIPHER" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.Plaintext')
echo "Decrypted: $(echo "$PLAIN" | base64 -d)"
```

### Check Key Rotation Status

```bash
ve kms DescribeKeyRotation --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

### View Recent API Calls (via CloudTrail)

```bash
# List recent KMS events
# Note: Requires CloudTrail integration
ve cloudtrail LookupEvents --LookupAttributes AttributeKey=EventSource,AttributeValue=kms.volcengineapi.com --Region {{env.VOLCENGINE_REGION}}
```

## Recovery Procedures

### Recover from Accidental Key Disable

```bash
# Immediate recovery
ve kms EnableKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Verify
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyState'
```

### Cancel Accidental Deletion

```bash
# Must be done before waiting period expires
ve kms CancelKeyDeletion --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Enable key if needed
ve kms EnableKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

### Restore Key with Deleted Material (External Keys)

```bash
# For keys with Origin=EXTERNAL and deleted material
# Must re-import key material
ve kms GetParametersForImport --KeyId "{{user.key_id}}" --WrappingAlgorithm RSAES_OAEP_SHA_256 --Region {{env.VOLCENGINE_REGION}}

# Then use ImportKeyMaterial with new wrapped key material
```
