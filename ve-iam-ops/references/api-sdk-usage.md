# API & SDK — IAM

## Go SDK

**Package:** `github.com/volcengine/volc-sdk-golang/service/iam`
**Minimum Go version:** 1.14
**Latest version:** Latest from GitHub

## Client Initialization

```go
package main

import (
    "github.com/volcengine/volc-sdk-golang/service/iam"
    "os"
)

func main() {
    instance := iam.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.Client.SetRegion(os.Getenv("VOLCENGINE_REGION"))
}
```

## SDK Operations Map

### User Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create user | `CreateUser` | `CreateUserInput` |
| List users | `ListUsers` | `ListUsersInput` |
| Get user | `GetUser` | `GetUserInput` |
| Update user | `UpdateUser` | `UpdateUserInput` |
| Delete user | `DeleteUser` | `DeleteUserInput` |
| List groups for user | `ListGroupsForUser` | `ListGroupsForUserInput` |
| List attached policies | `ListAttachedUserPolicies` | `ListAttachedUserPoliciesInput` |

### Policy Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create policy | `CreatePolicy` | `CreatePolicyInput` |
| List policies | `ListPolicies` | `ListPoliciesInput` |
| Get policy | `GetPolicy` | `GetPolicyInput` |
| Delete policy | `DeletePolicy` | `DeletePolicyInput` |
| Attach to user | `AttachUserPolicy` | `AttachUserPolicyInput` |
| Detach from user | `DetachUserPolicy` | `DetachUserPolicyInput` |
| Attach to role | `AttachRolePolicy` | `AttachRolePolicyInput` |
| Attach to group | `AttachGroupPolicy` | `AttachGroupPolicyInput` |
| List entities for policy | `ListEntitiesForPolicy` | `ListEntitiesForPolicyInput` |

### Role Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create role | `CreateRole` | `CreateRoleInput` |
| List roles | `ListRoles` | `ListRolesInput` |
| Get role | `GetRole` | `GetRoleInput` |
| Update role | `UpdateRole` | `UpdateRoleInput` |
| Delete role | `DeleteRole` | `DeleteRoleInput` |
| Update trust policy | `UpdateAssumeRolePolicy` | `UpdateAssumeRolePolicyInput` |

### Group Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create group | `CreateGroup` | `CreateGroupInput` |
| List groups | `ListGroups` | `ListGroupsInput` |
| Get group | `GetGroup` | `GetGroupInput` |
| Delete group | `DeleteGroup` | `DeleteGroupInput` |
| Add user to group | `AddUserToGroup` | `AddUserToGroupInput` |
| Remove user from group | `RemoveUserFromGroup` | `RemoveUserFromGroupInput` |
| List group users | `GetGroup` (with user list) | `GetGroupInput` |
| List attached policies | `ListAttachedGroupPolicies` | `ListAttachedGroupPoliciesInput` |

### Access Key Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create access key | `CreateAccessKey` | `CreateAccessKeyInput` |
| List access keys | `ListAccessKeys` | `ListAccessKeysInput` |
| Update access key | `UpdateAccessKey` | `UpdateAccessKeyInput` |
| Delete access key | `DeleteAccessKey` | `DeleteAccessKeyInput` |

### Login Profile Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create login profile | `CreateLoginProfile` | `CreateLoginProfileInput` |
| Get login profile | `GetLoginProfile` | `GetLoginProfileInput` |
| Update login profile | `UpdateLoginProfile` | `UpdateLoginProfileInput` |
| Delete login profile | `DeleteLoginProfile` | `DeleteLoginProfileInput` |

### STS Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Assume role | `AssumeRole` | `AssumeRoleInput` |
| Get session token | `GetSessionToken` | `GetSessionTokenInput` |

### Identity Provider Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create SAML provider | `CreateSAMLProvider` | `CreateSAMLProviderInput` |
| Create OIDC provider | `CreateOIDCProvider` | `CreateOIDCProviderInput` |
| List providers | `ListIdentityProviders` | `ListIdentityProvidersInput` |
| Get provider | `GetSAMLProvider` / `GetOIDCProvider` | Provider-specific input |
| Delete provider | `DeleteSAMLProvider` / `DeleteOIDCProvider` | Provider-specific input |

### Credential Report Operations

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Generate report | `GenerateCredentialReport` | `GenerateCredentialReportInput` |
| Get report | `GetCredentialReport` | `GetCredentialReportInput` |

## Pagination Pattern

```go
// ListUsers pagination
input := &iam.ListUsersInput{
    Limit: 100,
}

for {
    output, err := client.ListUsers(input)
    if err != nil {
        log.Fatal(err)
    }

    for _, user := range output.Result.Users {
        fmt.Println(user.UserName)
    }

    if output.Result.IsTruncated {
        input.Marker = output.Result.Marker
    } else {
        break
    }
}
```

## Policy Document Examples

### Administrator Access
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
```

### Read-Only Access to ECS
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:Describe*",
        "ecs:List*"
      ],
      "Resource": "*"
    }
  ]
}
```

### Access to Specific Resources
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeInstances",
        "ecs:StartInstance",
        "ecs:StopInstance"
      ],
      "Resource": [
        "trn:ecs::123456:instance/i-1234567890abcdef0"
      ]
    }
  ]
}
```

### Conditional Access (Time-based)
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ecs:*",
      "Resource": "*",
      "Condition": {
        "DateGreaterThan": {
          "iam:CurrentTime": "2026-01-01T00:00:00Z"
        },
        "DateLessThan": {
          "iam:CurrentTime": "2026-12-31T23:59:59Z"
        }
      }
    }
  ]
}
```

## Trust Policy Examples

### Cross-Account Access
```json
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
```

### Service Role
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": ["ecs", "rds"]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

### SAML Federation
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": ["trn:iam::123456:saml-provider/okta"]
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "SAML:aud": "https://signin.volcengine.com/saml"
        }
      }
    }
  ]
}
```

### OIDC Federation
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": ["trn:iam::123456:oidc-provider/auth0.com"]
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "auth0.com:aud": "your-client-id"
        }
      }
    }
  ]
}
```
