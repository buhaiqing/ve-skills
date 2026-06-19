# Troubleshooting — IAM

## Common Error Codes (≥ 10 Codes)

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `EntityAlreadyExists` | **HALT** — user/role/policy/group already exists | Use different name or use existing resource |
| `NoSuchEntity` | **HALT** — user/role/policy/group does not exist | Verify name or create first |
| `DeleteConflict` | **HALT** — cannot delete due to dependencies | Remove all dependencies first (policies, groups, keys, profile) |
| `LimitExceeded` | **HALT** — resource limit reached | Delete unused resources or request limit increase |
| `MalformedPolicyDocument` | **HALT** — policy JSON has syntax errors | Fix JSON syntax and validate with `jq` |
| `InvalidUserName` | **HALT** — name violates naming rules | Use 1-64 chars with alphanumeric + `+=,.@_-` |
| `InvalidParameter` | **HALT** — request parameter invalid | Check parameter format and constraints |
| `EntityNotFound` | **HALT** — referenced entity not found | Verify entity exists before referencing |
| `DuplicatePolicyAttachment` | **HALT** — policy already attached | Policy already attached to identity |
| `AccessDenied` | **HALT** — access denied by policy | Check policy allows the action |
| `Unauthorized` | **HALT** — insufficient IAM permissions | User needs IAM permissions to perform action |
| `PasswordPolicyViolation` | **HALT** — password not complex enough | Use stronger password per policy |
| `TooManyAccessKeys` | **HALT** — user already has 2 access keys | Delete existing key first |
| `Throttling` | Exponential backoff (1s, 2s, 4s, 8s) | Max 3 retries |
| `InternalError` | Retry with backoff; **HALT** after 3 retries | Capture RequestId for escalation |
| `ServiceUnavailable` | Retry with exponential backoff | Retry until service recovers |

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
