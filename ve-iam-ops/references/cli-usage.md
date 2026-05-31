# CLI — IAM (`ve`)

## Install and config

- Install: see [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json` (JSON format).
- For sandbox environments, set env vars directly (preferred).

## Conventions (agent execution)

- Output is **JSON by default**
- Document **exact** JSON paths after verifying with a real invocation
- CLI invocation: `ve iam <action> --parameter value`
- STS operations use: `ve sts <action> --parameter value`

## CLI vs API coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateUser | yes | Full support |
| ListUsers | yes | Full support with pagination |
| GetUser | yes | Full support |
| UpdateUser | yes | Full support |
| DeleteUser | yes | Full support |
| CreatePolicy | yes | Full support |
| ListPolicies | yes | Full support |
| GetPolicy | yes | Full support |
| DeletePolicy | yes | Full support |
| AttachUserPolicy | yes | Full support |
| DetachUserPolicy | yes | Full support |
| AttachRolePolicy | yes | Full support |
| AttachGroupPolicy | yes | Full support |
| CreateRole | yes | Full support |
| ListRoles | yes | Full support |
| GetRole | yes | Full support |
| DeleteRole | yes | Full support |
| AssumeRole | yes | Via `ve sts AssumeRole` |
| CreateGroup | yes | Full support |
| ListGroups | yes | Full support |
| GetGroup | yes | Full support |
| DeleteGroup | yes | Full support |
| AddUserToGroup | yes | Full support |
| RemoveUserFromGroup | yes | Full support |
| CreateAccessKey | yes | Full support |
| ListAccessKeys | yes | Full support |
| DeleteAccessKey | yes | Full support |
| UpdateLoginProfile | yes | Full support |
| GetLoginProfile | yes | Full support |
| DeleteLoginProfile | yes | Full support |
| CreateSAMLProvider | yes | Full support |
| CreateOIDCProvider | yes | Full support |
| ListIdentityProviders | yes | Full support |
| GenerateCredentialReport | yes | Full support |
| GetCredentialReport | yes | Full support |

## Command map

### User Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create user | `ve iam CreateUser --UserName alice --Region cn-beijing` | JSON output by default |
| List users | `ve iam ListUsers --Region cn-beijing` | Supports pagination |
| Get user | `ve iam GetUser --UserName alice --Region cn-beijing` | Returns user details |
| Update user | `ve iam UpdateUser --UserName alice --NewUserName bob --Region cn-beijing` | Rename user |
| Delete user | `ve iam DeleteUser --UserName alice --Region cn-beijing` | Requires no dependencies |
| List user groups | `ve iam ListGroupsForUser --UserName alice --Region cn-beijing` | Returns attached groups |
| List user policies | `ve iam ListAttachedUserPolicies --UserName alice --Region cn-beijing` | Returns attached policies |

### Policy Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create policy | `ve iam CreatePolicy --PolicyName mypolicy --PolicyDocument '{"Statement":[{"Effect":"Allow","Action":"ecs:*","Resource":"*"}]}' --Region cn-beijing` | JSON document required |
| List policies | `ve iam ListPolicies --Scope Local --Region cn-beijing` | Scope: All, Local, System |
| Get policy | `ve iam GetPolicy --PolicyName mypolicy --Region cn-beijing` | Returns policy document |
| Delete policy | `ve iam DeletePolicy --PolicyName mypolicy --Region cn-beijing` | No attachments allowed |
| Attach to user | `ve iam AttachUserPolicy --UserName alice --PolicyName mypolicy --Region cn-beijing` | Attaches custom policy |
| Detach from user | `ve iam DetachUserPolicy --UserName alice --PolicyName mypolicy --Region cn-beijing` | Detaches policy |
| Attach to role | `ve iam AttachRolePolicy --RoleName myrole --PolicyName mypolicy --Region cn-beijing` | Attaches to role |
| Attach to group | `ve iam AttachGroupPolicy --GroupName mygroup --PolicyName mypolicy --Region cn-beijing` | Attaches to group |
| List policy attachments | `ve iam ListEntitiesForPolicy --PolicyName mypolicy --Region cn-beijing` | Shows all attachments |

### Role Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create role | `ve iam CreateRole --RoleName myrole --AssumeRolePolicyDocument '{"Statement":[{"Effect":"Allow","Principal":{"Service":["ecs"]},"Action":"sts:AssumeRole"}]}' --Region cn-beijing` | Trust policy required |
| List roles | `ve iam ListRoles --Region cn-beijing` | Returns all roles |
| Get role | `ve iam GetRole --RoleName myrole --Region cn-beijing` | Returns role details |
| Update role | `ve iam UpdateRole --RoleName myrole --Description 'New description' --Region cn-beijing` | Updates metadata |
| Delete role | `ve iam DeleteRole --RoleName myrole --Region cn-beijing` | No attached policies allowed |
| Update trust policy | `ve iam UpdateAssumeRolePolicy --RoleName myrole --PolicyDocument '{...}' --Region cn-beijing` | Updates trust relationship |
| Assume role | `ve sts AssumeRole --RoleTrn trn:iam::123456:role/myrole --RoleSessionName mysession --Region cn-beijing` | Returns temporary credentials |

### Group Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create group | `ve iam CreateGroup --GroupName developers --Region cn-beijing` | Creates empty group |
| List groups | `ve iam ListGroups --Region cn-beijing` | Returns all groups |
| Get group | `ve iam GetGroup --GroupName developers --Region cn-beijing` | Returns group and members |
| Delete group | `ve iam DeleteGroup --GroupName developers --Region cn-beijing` | Must be empty |
| Add user | `ve iam AddUserToGroup --GroupName developers --UserName alice --Region cn-beijing` | Adds to group |
| Remove user | `ve iam RemoveUserFromGroup --GroupName developers --UserName alice --Region cn-beijing` | Removes from group |
| List group policies | `ve iam ListAttachedGroupPolicies --GroupName developers --Region cn-beijing` | Returns attached policies |

### Access Key Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create key | `ve iam CreateAccessKey --UserName alice --Region cn-beijing` | Returns AK/SK pair |
| List keys | `ve iam ListAccessKeys --UserName alice --Region cn-beijing` | Returns key metadata |
| Update key | `ve iam UpdateAccessKey --UserName alice --AccessKeyId AK... --Status Inactive --Region cn-beijing` | Enable/disable key |
| Delete key | `ve iam DeleteAccessKey --UserName alice --AccessKeyId AK... --Region cn-beijing` | Irreversible |

### Login Profile Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create profile | `ve iam CreateLoginProfile --UserName alice --Password 'SecureP@ss123' --PasswordResetRequired true --Region cn-beijing` | Console access |
| Get profile | `ve iam GetLoginProfile --UserName alice --Region cn-beijing` | Returns profile status |
| Update profile | `ve iam UpdateLoginProfile --UserName alice --Password 'NewSecureP@ss456' --Region cn-beijing` | Change password |
| Delete profile | `ve iam DeleteLoginProfile --UserName alice --Region cn-beijing` | Removes console access |

### Identity Provider Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Create SAML provider | `ve iam CreateSAMLProvider --SAMLProviderName okta --SAMLMetadataDocument file:///path/to/metadata.xml --Region cn-beijing` | SSO integration |
| Create OIDC provider | `ve iam CreateOIDCProvider --OIDCProviderName auth0 --Url https://auth0.com --ClientIDList 'client1,client2' --ThumbprintList 'thumb1,thumb2' --Region cn-beijing` | OIDC integration |
| List providers | `ve iam ListIdentityProviders --Region cn-beijing` | Returns all providers |
| Get SAML provider | `ve iam GetSAMLProvider --SAMLProviderName okta --Region cn-beijing` | Returns SAML details |
| Get OIDC provider | `ve iam GetOIDCProvider --OIDCProviderName auth0 --Region cn-beijing` | Returns OIDC details |
| Delete SAML provider | `ve iam DeleteSAMLProvider --SAMLProviderName okta --Region cn-beijing` | Removes provider |
| Delete OIDC provider | `ve iam DeleteOIDCProvider --OIDCProviderName auth0 --Region cn-beijing` | Removes provider |

### Credential Report Operations

| Goal | Example `ve` invocation | Notes |
|------|------------------------|-------|
| Generate report | `ve iam GenerateCredentialReport --Region cn-beijing` | Starts generation |
| Get report | `ve iam GetCredentialReport --Region cn-beijing` | Returns CSV content |

## Policy Document via File

For complex policies, use file input:

```bash
# Create policy from file
cat > /tmp/policy.json << 'EOF'
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeInstances",
        "ecs:StartInstance",
        "ecs:StopInstance"
      ],
      "Resource": "*"
    }
  ]
}
EOF

ve iam CreatePolicy \
  --PolicyName ec2-management \
  --PolicyDocument "$(cat /tmp/policy.json)" \
  --Region cn-beijing
```

## Trust Policy via File

```bash
# Create trust policy for cross-account access
cat > /tmp/trust.json << 'EOF'
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "STS": ["trn:sts::123456789012:root"]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

ve iam CreateRole \
  --RoleName cross-account-role \
  --AssumeRolePolicyDocument "$(cat /tmp/trust.json)" \
  --Region cn-beijing
```
