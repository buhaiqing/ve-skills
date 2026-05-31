# Troubleshooting — IAM

## Common Error Codes (≥ 10 Codes)

| Error Code | HTTP Status | Meaning | Agent Action |
|-----------|-------------|---------|-------------|
| `EntityAlreadyExists` | 409 | User/role/policy/group already exists | HALT; use different name or use existing |
| `NoSuchEntity` | 404 | User/role/policy/group does not exist | HALT; verify name or create first |
| `DeleteConflict` | 409 | Cannot delete due to dependencies | HALT; remove all dependencies first |
| `LimitExceeded` | 400 | Resource limit reached (e.g., 1000 users) | HALT; delete unused resources or request limit increase |
| `MalformedPolicyDocument` | 400 | Policy JSON is invalid or has syntax errors | HALT; fix JSON syntax and validate |
| `InvalidUserName` | 400 | User name violates naming rules | HALT; use 1-64 chars with alphanumeric + `+=,.@_-` |
| `InvalidPolicyName` | 400 | Policy name violates naming rules | HALT; use 1-128 chars with alphanumeric + `_+=,.@-` |
| `InvalidRoleName` | 400 | Role name violates naming rules | HALT; use 1-64 chars with alphanumeric + `_+=,.@-` |
| `InvalidGroupName` | 400 | Group name violates naming rules | HALT; use 1-128 chars with alphanumeric + `_+=,.@-` |
| `InvalidParameter` | 400 | Request parameter is invalid | HALT; check parameter format and constraints |
| `EntityNotFound` | 404 | Referenced entity not found | HALT; verify entity exists |
| `PolicyNotAttached` | 404 | Policy not attached to identity | HALT; policy already detached |
| `DuplicatePolicyAttachment` | 409 | Policy already attached | HALT; policy already attached |
| `Unauthorized` | 403 | Insufficient IAM permissions | HALT; user needs IAM permissions to perform action |
| `AccessDenied` | 403 | Access denied by policy | HALT; check policy allows the action |
| `PasswordPolicyViolation` | 400 | Password does not meet complexity | HALT; use stronger password |
| `CredentialReportNotReady` | 400 | Credential report generation incomplete | Retry after delay |
| `TooManyAccessKeys` | 400 | User already has 2 access keys | HALT; delete existing key first |
| `InvalidPublicKey` | 400 | SSH public key is invalid | HALT; provide valid SSH key |
| `InvalidCertificate` | 400 | SAML certificate is invalid | HALT; provide valid X.509 certificate |
| `MalformedCertificate` | 400 | Certificate format is wrong | HALT; use PEM-encoded X.509 |
| `DuplicateCertificate` | 409 | Certificate already in use | HALT; use different certificate |
| `Throttling` | 429 | Rate limit exceeded | Exponential backoff (1s, 2s, 4s, 8s) |
| `InternalError` | 500 | Server-side error | Retry with backoff; HALT after 3 retries |
| `ServiceUnavailable` | 503 | Service temporarily unavailable | Retry with exponential backoff |

## Diagnostic Order

1. **Check credentials:**
   ```bash
   test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "OK"
   # Output: Secret key should NOT be displayed
   # Safe: VOLCENGINE_SECRET_KEY=<masked>
   ```

2. **Verify region:**
   ```bash
   echo "Region: $VOLCENGINE_REGION"
   # Expected: cn-beijing, cn-shanghai, etc.
   ```

3. **Test basic connectivity:**
   ```bash
   ve iam ListUsers --Region $VOLCENGINE_REGION
   ```

4. **Check specific entity:**
   ```bash
   ve iam GetUser --UserName $USER_NAME --Region $VOLCENGINE_REGION
   ```

5. **Verify IAM permissions:**
   ```bash
   # Check if caller has IAM access
   ve iam GetUser --UserName $(whoami) --Region $VOLCENGINE_REGION 2>&1 | grep -i "unauthorized\|accessdenied"
   ```

## Common Scenarios

### Scenario 1: Cannot Create User — `EntityAlreadyExists`

**Cause:** User name already exists in the account.

**Fix:**
```bash
# Check if user exists
ve iam GetUser --UserName existing-user --Region cn-beijing

# Use different name or reuse existing user
ve iam CreateUser --UserName new-user-name --Region cn-beijing
```

### Scenario 2: Cannot Delete User — `DeleteConflict`

**Cause:** User has dependencies (policies, groups, access keys, login profile).

**Fix:**
```bash
# 1. Detach all policies
for policy in $(ve iam ListAttachedUserPolicies --UserName $USER --Region cn-beijing | jq -r '.Result.AttachedPolicies[].PolicyName'); do
  ve iam DetachUserPolicy --UserName $USER --PolicyName $policy --Region cn-beijing
done

# 2. Remove from all groups
for group in $(ve iam ListGroupsForUser --UserName $USER --Region cn-beijing | jq -r '.Result.Groups[].GroupName'); do
  ve iam RemoveUserFromGroup --GroupName $group --UserName $USER --Region cn-beijing
done

# 3. Delete all access keys
for key in $(ve iam ListAccessKeys --UserName $USER --Region cn-beijing | jq -r '.Result.AccessKeyMetadata[].AccessKeyId'); do
  ve iam DeleteAccessKey --UserName $USER --AccessKeyId $key --Region cn-beijing
done

# 4. Delete login profile (if exists)
ve iam DeleteLoginProfile --UserName $USER --Region cn-beijing 2>/dev/null || true

# 5. Now delete user
ve iam DeleteUser --UserName $USER --Region cn-beijing
```

### Scenario 3: Policy Not Working — `AccessDenied`

**Cause:** Policy does not grant the specific action or resource.

**Fix:**
```bash
# 1. Get current policy
ve iam GetPolicy --PolicyName mypolicy --Region cn-beijing | jq '.Result.Policy.PolicyDocument'

# 2. Check for common issues:
# - Action spelling (case-sensitive)
# - Resource ARN format
# - Missing permissions for dependent actions

# 3. Test with IAM Policy Simulator (console) or add explicit deny check
```

### Scenario 4: AssumeRole Fails — `AccessDenied`

**Cause:**
- Trust policy does not allow caller
- Caller lacks `sts:AssumeRole` permission
- Session name format invalid

**Fix:**
```bash
# 1. Check role trust policy
ve iam GetRole --RoleName myrole --Region cn-beijing | jq '.Result.Role.AssumeRolePolicyDocument'

# 2. Verify caller is in trust policy principal
# 3. Check caller has sts:AssumeRole permission
# 4. Use valid session name (alphanumeric plus =,.@-)
```

### Scenario 5: InvalidPolicyDocument

**Cause:** JSON syntax error or invalid policy structure.

**Fix:**
```bash
# 1. Validate JSON syntax
jq '.' /tmp/policy.json

# 2. Check required fields:
# - Version: "2021-04-01"
# - Statement array with Effect, Action, Resource
# - Valid Effect: "Allow" or "Deny"

# 3. Common mistakes:
# - Missing quotes around keys/values
# - Trailing commas
# - Wrong bracket types
# - Unescaped quotes in strings
```

### Scenario 6: Cannot Assume Role — `Unauthorized`

**Cause:**
- Source account not in trust policy
- External ID required but not provided
- MFA required but not used

**Fix:**
```bash
# Check if MFA is required in trust policy condition
# If yes, use MFA token with AssumeRole

# Check if external ID is required
# If yes, provide --ExternalId parameter
```

### Scenario 7: Credential Report Generation Fails

**Cause:** Previous report still generating or account has many users.

**Fix:**
```bash
# Wait and retry
sleep 5
ve iam GenerateCredentialReport --Region cn-beijing

# Check status
ve iam GetCredentialReport --Region cn-beijing
```

### Scenario 8: Access Key Limit Exceeded

**Cause:** User already has 2 access keys (maximum).

**Fix:**
```bash
# List existing keys
ve iam ListAccessKeys --UserName $USER --Region cn-beijing

# Delete unused key before creating new one
ve iam UpdateAccessKey --UserName $USER --AccessKeyId AK... --Status Inactive --Region cn-beijing
ve iam DeleteAccessKey --UserName $USER --AccessKeyId AK... --Region cn-beijing

# Now create new key
ve iam CreateAccessKey --UserName $USER --Region cn-beijing
```

### Scenario 9: SAML Provider Certificate Error

**Cause:** Invalid or expired SAML certificate.

**Fix:**
```bash
# Verify certificate is valid PEM format
openssl x509 -in metadata.xml -text -noout

# Check certificate expiry
date -d "$(openssl x509 -in metadata.xml -noout -enddate | cut -d= -f2)" +%Y-%m-%d

# Update with new certificate
ve iam UpdateSAMLProvider --SAMLProviderName okta --SAMLMetadataDocument file:///new/metadata.xml --Region cn-beijing
```

### Scenario 10: Policy Attachment Limit

**Cause:** Identity already has 10 policies attached (maximum).

**Fix:**
```bash
# List attached policies
ve iam ListAttachedUserPolicies --UserName $USER --Region cn-beijing

# Combine policies or remove unused ones
ve iam DetachUserPolicy --UserName $USER --PolicyName unused-policy --Region cn-beijing

# Attach new policy
ve iam AttachUserPolicy --UserName $USER --PolicyName new-policy --Region cn-beijing
```

## Security Incident Response

### Suspicious Access Key Activity

```bash
# 1. Immediately disable key
ve iam UpdateAccessKey --UserName $USER --AccessKeyId AK... --Status Inactive --Region cn-beijing

# 2. Check credential report for key usage
ve iam GenerateCredentialReport --Region cn-beijing
ve iam GetCredentialReport --Region cn-beijing

# 3. Rotate key if confirmed compromise
ve iam DeleteAccessKey --UserName $USER --AccessKeyId AK... --Region cn-beijing
ve iam CreateAccessKey --UserName $USER --Region cn-beijing
```

### Unauthorized Role Assumption

```bash
# 1. Check trust policy for unauthorized principals
ve iam GetRole --RoleName compromised-role --Region cn-beijing | jq '.Result.Role.AssumeRolePolicyDocument'

# 2. Update trust policy to remove unauthorized access
ve iam UpdateAssumeRolePolicy --RoleName compromised-role --PolicyDocument '{...}' --Region cn-beijing

# 3. Review CloudTrail for assumption events
```

## Recovery Patterns

| Issue | Immediate Action | Follow-up |
|-------|------------------|-----------|
| Locked out of account | Contact account owner with MFA | Review root account security |
| Deleted critical user | Create new user with same policies | Update application credentials |
| Policy misconfiguration | Attach permissive policy temporarily | Fix policy, then remove temporary |
| Role assumption abuse | Update trust policy immediately | Audit all role assumptions |
| Exposed access key | Deactivate key immediately | Rotate and update applications |
