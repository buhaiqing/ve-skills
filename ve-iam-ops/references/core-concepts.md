# Core Concepts — Volcengine IAM

## Architecture

IAM (Identity and Access Management) provides centralized access control for all Volcengine resources.

```
┌─────────────────────────────────────────────────────────────────┐
│                        IAM Service                               │
│  ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌──────────────────┐  │
│  │ Users   │  │ Policies │  │ Roles   │  │ Identity         │  │
│  │         │  │          │  │         │  │ Providers        │  │
│  │ - Alice │  │ - Admin  │  │ - XAcct │  │                  │  │
│  │ - Bob   │  │ - ReadOnly│ │ - ECS   │  │ - SAML (SSO)     │  │
│  │ -_svc   │  │ - Custom │  │ - Lambda│  │ - OIDC (OIDC)    │  │
│  └────┬────┘  └────┬─────┘  └────┬────┘  └──────────────────┘  │
│       │            │             │                              │
│       └────────────┴─────────────┘                              │
│                    │                                            │
│            ┌───────┴───────┐                                    │
│            │   Groups      │                                    │
│            │  - Developers │                                    │
│            │  - Admins     │                                    │
│            └───────────────┘                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
       ┌──────────────────────┼──────────────────────┐
       │                      │                      │
  ┌────┴─────┐          ┌────┴─────┐          ┌────┴─────┐
  │   ECS    │          │   RDS    │          │   TOS    │
  │          │          │          │          │          │
  │ Resources│          │ Resources│          │ Resources│
  └──────────┘          └──────────┘          └──────────┘
```

## Core Entities

### Users

| Attribute | Description | Constraints |
|-----------|-------------|-------------|
| `UserName` | Human-readable name | 1-64 chars, alphanumeric + `+=,.@_-` |
| `UserId` | Unique identifier | Auto-generated, immutable |
| `Arn` | Resource name | Format: `trn:iam::{account}:user/{path}{name}` |
| `Path` | Organizational prefix | Optional, e.g., `/engineering/` |
| `CreateDate` | Creation timestamp | ISO 8601 format |

### Policies

| Type | Description | Use Case |
|------|-------------|----------|
| **System policies** | Managed by Volcengine | Common access patterns (e.g., `IAMFullAccess`) |
| **Custom policies** | User-defined | Organization-specific permissions |
| **Inline policies** | Embedded in user/role | Tightly coupled, no reuse |

Policy Document Structure:
```json
{
  "Version": "2021-04-01",
  "Statement": [
    {
      "Effect": "Allow|Deny",
      "Action": ["service:action"],
      "Resource": ["trn:service::account:resource"],
      "Condition": {
        "StringEquals": {
          "iam:ResourceTag/Key": "value"
        }
      }
    }
  ]
}
```

### Roles

| Use Case | Trust Policy Principal | Example |
|----------|----------------------|---------|
| Cross-account access | `STS` account | `"STS": ["trn:sts::123456:root"]` |
| Service role | `Service` | `"Service": ["ecs", "rds"]` |
| Federated access | `Federated` | `"Federated": ["trn:iam::account:saml-provider/okta"]` |

### Groups

- Collection of users
- Cannot be nested (no groups within groups)
- Maximum 1000 users per group
- Maximum 10 policies attached per group

## Access Control Flow

```
User/Role Request
       │
       ▼
┌─────────────────┐
│ Authentication  │ ──► Verify identity (AK/SK, password, or session token)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Authorization   │ ──► Evaluate all applicable policies
│   - User policy │     (user inline + attached + group + permission boundary)
│   - Group policy│
│   - Role policy │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Decision        │ ──► Allow if any policy grants AND no explicit deny
│   - Allow/Deny  │
└────────┬────────┘
         │
         ▼
   Resource Access
```

## Policy Evaluation Logic

1. **Default deny**: All requests are denied by default
2. **Explicit allow**: Matching `Allow` statement grants access
3. **Explicit deny**: Matching `Deny` statement overrides all allows
4. **Permission boundary**: Maximum permissions a user/role can have

## ARN Format

| Resource | ARN Format |
|----------|------------|
| User | `trn:iam::{account}:user/{path}{user-name}` |
| Policy | `trn:iam::{account}:policy/{policy-name}` |
| Role | `trn:iam::{account}:role/{role-name}` |
| Group | `trn:iam::{account}:group/{group-name}` |
| SAML Provider | `trn:iam::{account}:saml-provider/{name}` |
| OIDC Provider | `trn:iam::{account}:oidc-provider/{url}` |

## Resource Limits

| Resource | Default Limit | Adjustable |
|----------|---------------|------------|
| Users per account | 1000 | Yes |
| Groups per account | 300 | Yes |
| Roles per account | 1000 | Yes |
| Custom policies | 1500 | Yes |
| Access keys per user | 2 | No |
| MFA devices per user | 1 | No |
| Groups per user | 10 | No |
| Users per group | 1000 | Yes |
| Policies per identity | 10 | No |
| Role session duration | 12 hours max | No |

## Permission Boundaries

Permission boundaries limit the maximum permissions a user or role can have, regardless of what policies are attached.

```
┌─────────────────────────────────────┐
│      Permission Boundary            │
│  (Maximum possible permissions)     │
│                                     │
│  ┌─────────────────────────────┐   │
│  │   Attached Policies         │   │
│  │   (Actual permissions)      │   │
│  │                             │   │
│  │   Effective = Intersection  │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

## Identity Federation

### SAML 2.0

- Enterprise SSO integration
- Supports identity providers: Okta, Azure AD, ADFS, etc.
- User attributes mapped to IAM roles

### OIDC (OpenID Connect)

- Web and mobile application authentication
- Supports identity providers: Auth0, Google, GitHub, etc.
- JWT tokens exchanged for temporary credentials

## Audit and Compliance

### Credential Report

CSV format containing:
- User ARN and creation date
- Access key status and last used
- Console password enabled and last login
- MFA status

### Access Records

- Who performed the action
- What action was performed
- Which resource was accessed
- When the action occurred
- Source IP and user agent

## Dependency Map

```
IAM Operations are foundational:
  ├── All product skills reference IAM for permissions
  ├── Cross-account operations require IAM roles
  ├── Service-linked roles created by product services
  └── Audit logs stored in security services
```
