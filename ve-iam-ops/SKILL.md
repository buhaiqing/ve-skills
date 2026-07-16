---
name: ve-iam-ops
description: >-
  Use when the user needs to manage Volcengine (火山引擎) IAM (Identity and Access Management) —
  users, policies, roles, groups, identity providers, and audit logs. User mentions IAM, 身份与访问管理,
  access control, permissions, users, policies, roles, groups, SSO, or audit trails even without
  naming the product directly. Not for product-specific resource management (delegate to product skills).
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.21+ runtime
  (JIT SDK fallback; scripts compatible with Go 1.14+ syntax), valid API credentials,
  network access to Volcengine IAM endpoints.
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  cli_applicability: dual-path
  cli_support_evidence: >-
    IAM API is accessible via `ve iam --help`. Full coverage for user, policy, role,
    group, and identity provider operations.
    See: https://www.volcengine.com/docs/6257
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine IAM Operations Skill

## Overview

IAM (Identity and Access Management, 身份与访问管理) on Volcengine (火山引擎) provides centralized access control for all Volcengine resources. This skill enables agents to manage users, policies, roles, groups, identity providers, and audit logs using the `ve` CLI (primary) or JIT Go SDK (fallback).

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports IAM operations.
  - **`ve iam`**: User, policy, role, group, identity provider management
  - **`ve sts`**: Security Token Service for temporary credentials

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env vars), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 IAM-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | IAM only; cross-product resource permissions reference product skills |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine IAM", "火山引擎 IAM", "身份与访问管理", "access control", "permissions"
- Task involves user management: CreateUser, ListUsers, GetUser, UpdateUser, DeleteUser
- Task involves policy management: CreatePolicy, ListPolicies, AttachPolicy, DetachPolicy, DeletePolicy
- Task involves role management: CreateRole, ListRoles, AssumeRole, DeleteRole
- Task involves group management: CreateGroup, ListGroups, AddUserToGroup, RemoveUserFromGroup
- Task involves identity providers: CreateSAMLProvider, CreateOIDCProvider, ListIdentityProviders
- Task involves audit logs: ListAccessKeys, GetCredentialReport, ListLoginProfiles
- Task involves temporary credentials: AssumeRole, GetSessionToken
- Task involves service-linked roles or permission boundaries

### SHOULD NOT Use This Skill When

- Task is about specific product resource operations (ECS, RDS, etc.) → delegate to respective product skill
- Task is purely billing → delegate to billing ops
- Task is about KMS key management → delegate to `ve-kms-ops` (when present)

### Delegation Rules

- IAM policies grant access to product resources → reference product skills for resource-specific actions
- Cross-account access requires IAM roles → complete IAM role setup before cross-account operations
- Service-linked roles are created automatically by product services → reference service documentation

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Access key from environment | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Secret key from environment | NEVER ask user; fail if unset; mask as `<masked>` |
| `{{env.VOLCENGINE_REGION}}` | Region (e.g., cn-beijing) | Use default if skill allows |
| `{{user.user_name}}` | IAM user name | Ask once; follow naming conventions |
| `{{user.policy_name}}` | IAM policy name | Ask once |
| `{{user.role_name}}` | IAM role name | Ask once |
| `{{user.group_name}}` | IAM group name | Ask once |
| `{{user.provider_name}}` | Identity provider name | Ask once |
| `{{user.access_key_id}}` | Access key ID to delete | Ask once; confirm before deletion |
| `{{user.account_id}}` | Volcengine account ID | Ask once; used in AssumeRole and cross-account trust |
| `{{user.org_unit}}` | Organization unit path | Ask once; optional, used in CreateUser path |
| `{{user.password}}` | Console password for user | Ask once; never log or echo |
| `{{user.policy_description}}` | Policy description | Ask once; optional |
| `{{user.role_description}}` | Role description | Ask once; optional |
| `{{user.session_name}}` | Role session name | Ask once; used in AssumeRole |
| `{{user.trusted_account_id}}` | Trusted account ID for cross-account trust | Ask once; used in CreateRole trust policy |
| `{{output.user_id}}` | User ID from API response | Parse from response |
| `{{output.policy_arn}}` | Policy ARN from response | Parse from response |
| `{{output.role_arn}}` | Role ARN from response | Parse from response |
| `{{output.group_id}}` | Group ID from CreateGroup response | Parse from response |
| `{{output.access_key_id}}` | Access key ID from CreateAccessKey response | Parse from response |

> **Security Warning (Credential Masking):** NEVER log, print, or expose `VOLCENGINE_SECRET_KEY` or any credential value. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`.

## API and Response Conventions (Agent-Readable)

- **IAM uses JSON REST API** with standard Volcengine response format
- **Endpoint:** `https://iam.volcengineapi.com`
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/iam`
- **Error responses:** JSON with `ResponseMetadata.Error` structure

### Key Response Fields (Centralized JSON Paths)

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateUser | `$.Result.User.UserId` | string | Unique user ID |
| CreateUser | `$.Result.User.Arn` | string | User ARN |
| CreatePolicy | `$.Result.Policy.PolicyArn` | string | Policy ARN |
| CreateRole | `$.Result.Role.RoleArn` | string | Role ARN |
| ListUsers | `$.Result.Users[].UserName` | array | User names |
| ListPolicies | `$.Result.Policies[].PolicyName` | array | Policy names |
| CreateAccessKey | `$.Result.AccessKey.AccessKeyId` | string | Access key ID |
| CreateAccessKey | `$.Result.AccessKey.SecretKey` | string | **Show once** — mask in logs |
| CreateAccessKey | `$.Result.AccessKey.Status` | string | Active/Inactive |
| AssumeRole | `$.Result.Credentials.AccessKeyId` | string | Temporary access key |
| AssumeRole | `$.Result.Credentials.SecretKey` | string | Temporary secret key |
| AssumeRole | `$.Result.Credentials.SessionToken` | string | Session token |
| AssumeRole | `$.Result.Credentials.Expiration` | string | Credential expiration time |
| CreateRole | `$.Result.Role.RoleId` | string | Created role ID — parse to `{{output.role_id}}` |
| CreateGroup | `$.Result.Group.GroupId` | string | Created group ID — parse to `{{output.group_id}}` |
| DeleteUser | `$.Result.UserId` | string | Deleted user ID — empty when gone (poll GetUser → 404) |
| DeletePolicy | `$.Result.PolicyName` | string | Deleted policy name — empty when gone (poll GetPolicy → 404) |
| DeleteRole | `$.Result.RoleName` | string | Deleted role name — empty when gone (poll GetRole → 404) |
| DeleteGroup | `$.Result.GroupId` | string | Deleted group ID — empty when gone (poll GetGroup → 404) |
| DeleteAccessKey | `$.Result.AccessKeyId` | string | Deleted access key ID — empty when gone (poll ListAccessKeys → 404) |

## Quick Start

### What This Skill Does
This skill enables you to manage IAM identities and permissions on Volcengine — create users, manage policies, configure roles, set up groups, and review audit logs using `ve iam` CLI or JIT Go SDK.

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region set: `VOLCENGINE_REGION`

### Verify Setup
```bash
# Check CLI and list users
ve iam ListUsers --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all IAM users
ve iam ListUsers --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand IAM architecture
- [Common Operations](#execution-flows) — Create, manage, and manage access control and permissions
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level | | safety_class | blast_radius |
|-----------|-------------|------------|------------||---|---|
| CreateUser | Create a new IAM user | Low | Low | | state-changing | single |
| ListUsers | List all IAM users | Low | ✅ None | | read-only | single |
| GetUser | Get user details | Low | ✅ None | | read-only | single |
| UpdateUser | Update user attributes | Low | Low | | state-changing | single |
| DeleteUser | Delete an IAM user | Low | 🔴 **High** — check dependencies | | destructive | single |
| CreatePolicy | Create a custom policy | Medium | Medium | | state-changing | single |
| AttachPolicy | Attach policy to user/role/group | Low | Medium | | state-changing | single |
| DetachPolicy | Detach policy from identity | Low | Medium | | state-changing | single |
| DeletePolicy | Delete a custom policy | Low | 🔴 **High** — check attachments | | destructive | single |
| CreateRole | Create an IAM role | Medium | Low | | state-changing | single |
| AssumeRole | Get temporary credentials | Medium | Medium | | state-changing | single |
| DeleteRole | Delete an IAM role | Low | 🔴 **High** — check assumptions | | destructive | single |
| CreateGroup | Create an IAM group | Low | Low | | state-changing | single |
| AddUserToGroup | Add user to group | Low | Low | | state-changing | single |
| RemoveUserFromGroup | Remove user from group | Low | Low | | state-changing | single |
| DeleteGroup | Delete an IAM group | Low | 🔴 **High** — check members | | destructive | single |
| CreateSAMLProvider | Create SAML identity provider | High | Medium | | state-changing | single |
| CreateOIDCProvider | Create OIDC identity provider | High | Medium | | state-changing | single |
| ListAccessKeys | List user's access keys | Low | ✅ None | | read-only | single |
| CreateAccessKey | Create access key for user | Low | 🔴 **High** — secret key shown once | | state-changing | single |
| DeleteAccessKey | Delete user's access key | Low | 🔴 **High** — irreversible | | destructive | single |
| GetCredentialReport | Generate credential report | Medium | ✅ None | | read-only | single |
| UpdateLoginProfile | Set user console password | Low | Medium | | state-changing | single |

## Changelog
| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-07-13 | T04: annotate operation table with safety_class + blast_radius leaf-op metadata columns (L3 policy inputs); see ve-skill-generator/references/leaf-op-metadata-spec.md |
| 1.0.0 | 2026-05-27 | Initial release with user, policy, role, group, identity provider, and audit operations |
| 1.1.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md`; `max_iter=2` for destructive / state_changing ops, `max_iter=3` for read-only ops |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-iam-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for the scoring dimensions and
> [`references/prompt-templates.md`](references/prompt-templates.md) for the G/C/O
> prompt skeletons and verbatim safety prompts. The Critic and Generator MUST
> live in **isolated prompt contexts**.

### Operation Tiers

> See [`references/rubric.md` §0](references/rubric.md#0-operation-tier) for the full operation tier table.

### Loop

1. **Pre-flight (Orchestrator)** — resolve `{{env.*}}` and `{{user.*}}`; classify
   the operation into one of the four tiers; load `references/rubric.md`.
2. **Generate** — execute per the `## Execution Flows` chapter. Capture full
   command, parameters, raw response excerpt, `RequestId`, validation output,
   retries, and final state into `./audit-results/gcl-trace-*.json` with
   `redaction_pass: true`.
3. **Critique** — isolated prompt; score correctness / safety / idempotency /
   traceability / spec_compliance per the rubric. The Critic MUST NOT see the
   raw user request.
4. **Decide** — Safety=0 on Destructive/State-changing → **ABORT**; all pass
   → return; `iter < max_iter` → inject suggestions; else → return best +
   unresolved rubric items.

### IAM-specific safety rules

- **DeleteUser**: dependency pre-check required before any execution.
- **AttachPolicy with `Action=*:*` or `Resource=*`**: user MUST be warned.
- **CreateRole with open trust policy** (`Principal={Federated:["*"]}` / `STS:["*"]`):
   user MUST be warned.
- **DetachPolicy** when it's the last admin policy: user MUST confirm the risk.
- **CreateAccessKey / AssumeRole**: secret credentials output once to user;
   NEVER in trace (only `<masked>` or `sha256:<prefix>`).

### Trace

Every GCL run persists a JSON trace to `./audit-results/gcl-trace-*.json`.
Trace MUST NOT contain `VOLCENGINE_SECRET_KEY`, `CreateAccessKey.SecretKey`, or
`AssumeRole.Credentials.SecretKey` — only `<masked>`. See rubric §4 for
mandatory trace fields.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| KMS key / secret needed for the operation | `ve-kms-ops` |
| Product-specific resource permissions (ECS, RDS, etc.) | respective product skill |
| Billing quota exceeded | `ve-billing-ops` |

The Critic MUST NOT call any skill — it only emits suggestions.

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateUser — Create an IAM User

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| User name format | Alphanumeric plus `+=,.@_-`; 1-64 chars | Valid format | Fix name format |
| User name unique | Query `ListUsers` | No conflict | Use different name |

#### Execution — CLI (`ve`)

```bash
# Create a user
ve iam CreateUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"

# Create user with path (for organization)
ve iam CreateUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  --Path "/{{user.org_unit}}/"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/iam"
)

func main() {
    instance := iam.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    params["UserName"] = os.Getenv("USER_NAME")

    resp, err := instance.Client.Request("iam", "CreateUser", params)
    if err != nil {
        log.Fatalf("Failed to create user: %v", err)
    }
    fmt.Println(string(resp))
}
```

#### Validation

```bash
# Verify user was created
ve iam GetUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `EntityAlreadyExists` | HALT; user name already exists — use different name |
| `InvalidUserName` | HALT; name must be 1-64 chars with allowed characters |
| `LimitExceeded` | HALT; user limit reached (default 1000 per account) |
| `Unauthorized` | HALT; check IAM permissions for CreateUser |

---

### Operation: DeleteUser — Delete an IAM User

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** IAM user `{{user.user_name}}`.
> This is **IRREVERSIBLE** — attached policies, group memberships, access keys, and login profile are also removed.
> Type the user name `{{user.user_name}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** verify user has no dependencies:
  - No attached policies (detach first)
  - No group memberships (remove first)
  - No access keys (delete first)
  - No login profile (delete first)

```bash
# Check dependencies before deletion

# 1. Check attached policies
ve iam ListAttachedUserPolicies \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"

# 2. Check group memberships
ve iam ListGroupsForUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"

# 3. Check access keys
ve iam ListAccessKeys \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"

# 4. Check login profile
ve iam GetLoginProfile \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"
```

#### Execution

```bash
# Delete user (only after all dependencies removed)
ve iam DeleteUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"
```

#### Validation

```bash
# Verify user no longer exists
ve iam GetUser \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  && echo "USER STILL EXISTS" || echo "USER DELETED"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `DeleteConflict` | HALT; user has attached policies, groups, or other dependencies — remove all first |
| `NoSuchEntity` | User already deleted; skip |
| `Unauthorized` | HALT; check IAM permissions for DeleteUser |

---

### Operation: CreatePolicy — Create a Custom Policy

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Policy name format | Alphanumeric plus `_+=,.@-`; 1-128 chars | Valid format | Fix name format |
| Policy document | Valid JSON with correct structure | Valid policy | Fix policy syntax |

#### Execution

```bash
# Create policy from file
ve iam CreatePolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PolicyName "{{user.policy_name}}" \
  --PolicyDocument "$(cat policy.json)" \
  --Description "{{user.policy_description}}"
```

Example policy.json structure:
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
      "Resource": "*"
    }
  ]
}
```

#### Validation

```bash
# Get policy details
ve iam GetPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PolicyName "{{user.policy_name}}"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `EntityAlreadyExists` | HALT; policy name already exists |
| `MalformedPolicyDocument` | HALT; policy JSON is invalid or has syntax errors |
| `InvalidPolicyName` | HALT; name must follow naming conventions |

---

### Operation: AttachPolicy — Attach Policy to User

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| User exists | `GetUser` | User found | HALT; create user first |
| Policy exists | `GetPolicy` | Policy found | HALT; create policy first |

#### Execution

```bash
# Attach policy to user
ve iam AttachUserPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  --PolicyName "{{user.policy_name}}"

# Attach policy to role
ve iam AttachRolePolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}" \
  --PolicyName "{{user.policy_name}}"

# Attach policy to group
ve iam AttachGroupPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}" \
  --PolicyName "{{user.policy_name}}"
```

#### Validation

```bash
# Verify attachment
ve iam ListAttachedUserPolicies \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"
```

---

### Operation: DetachPolicy — Detach Policy from User

#### Execution

```bash
# Detach policy from user
ve iam DetachUserPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  --PolicyName "{{user.policy_name}}"

# Detach from role
ve iam DetachRolePolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}" \
  --PolicyName "{{user.policy_name}}"

# Detach from group
ve iam DetachGroupPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}" \
  --PolicyName "{{user.policy_name}}"
```

---

### Operation: DeletePolicy — Delete a Custom Policy

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** IAM policy `{{user.policy_name}}`.
> This is **IRREVERSIBLE** — the policy is removed from all attached users, roles, and groups.
> Type the policy name `{{user.policy_name}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** verify policy is not attached to any user, role, or group

```bash
# Check policy attachments
ve iam ListEntitiesForPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PolicyName "{{user.policy_name}}"
```

#### Execution

```bash
# Delete policy
ve iam DeletePolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PolicyName "{{user.policy_name}}"
```

---

### Operation: CreateRole — Create an IAM Role

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Role name format | Alphanumeric plus `_+=,.@-`; 1-64 chars | Valid format | Fix name format |
| Trust policy | Valid assume role policy document | Valid JSON | Fix trust policy |

#### Execution

```bash
# Create role with trust policy
ve iam CreateRole \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}" \
  --AssumeRolePolicyDocument "$(cat trust-policy.json)" \
  --Description "{{user.role_description}}"
```

Example trust-policy.json for cross-account access:
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "STS": ["trn:sts::{{user.trusted_account_id}}:root"]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Example trust-policy.json for service role:
```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": ["ecs"]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

#### Validation

```bash
# Get role details
ve iam GetRole \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}"
```

---

### Operation: DeleteRole — Delete an IAM Role

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** IAM role `{{user.role_name}}`.
> This is **IRREVERSIBLE** — any user or service assuming this role will lose access.
> Type the role name `{{user.role_name}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** verify no resources are granted this role via AssumeRole:
  - No active AssumeRole sessions (revoke sessions first)
  - No trust policy conditions that would be broken

```bash
# Check which entities assume this role
ve iam ListEntitiesForPolicy \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --PolicyName "{{user.role_name}}"
```

#### Execution

```bash
# Delete role
ve iam DeleteRole \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}"
```

#### Validation

```bash
# Verify role no longer exists
ve iam GetRole \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleName "{{user.role_name}}" \
  && echo "ROLE STILL EXISTS" || echo "ROLE DELETED"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `DeleteConflict` | HALT; role has active sessions or attached policies — detach and revoke sessions first |
| `NoSuchEntity` | Role already deleted; skip |
| `Unauthorized` | HALT; check IAM permissions for DeleteRole |

---

### Operation: AssumeRole — Get Temporary Credentials

#### Execution

```bash
# Assume role
ve sts AssumeRole \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RoleTrn "trn:iam::{{user.account_id}}:role/{{user.role_name}}" \
  --RoleSessionName "{{user.session_name}}"
```

#### Response Fields

See [Key Response Fields](#key-response-fields-centralized-json-paths) table above for `$.Result.Credentials.*` paths.

---

### Operation: CreateGroup — Create an IAM Group

#### Execution

```bash
# Create group
ve iam CreateGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}"
```

---

### Operation: AddUserToGroup — Add User to Group

#### Execution

```bash
# Add user to group
ve iam AddUserToGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}" \
  --UserName "{{user.user_name}}"
```

---

### Operation: RemoveUserFromGroup — Remove User from Group

#### Execution

```bash
# Remove user from group
ve iam RemoveUserFromGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}" \
  --UserName "{{user.user_name}}"
```

---

### Operation: DeleteGroup — Delete an IAM Group

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** IAM group `{{user.group_name}}`.
> This is **IRREVERSIBLE** — all members lose group permissions immediately.
> Type the group name `{{user.group_name}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** verify group has no members and no attached policies

```bash
# Check group members
ve iam GetGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}"

# Check attached policies
ve iam ListAttachedGroupPolicies \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}"
```

#### Execution

```bash
# Delete group
ve iam DeleteGroup \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --GroupName "{{user.group_name}}"
```

---

### Operation: CreateAccessKey — Create Access Key for User

#### Pre-flight (Safety Gate)

- **MUST** warn user: secret key is shown only once at creation
- **MUST** recommend: download/save credentials immediately

#### Execution

```bash
# Create access key
ve iam CreateAccessKey \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}"
```

#### Response Fields

See [Key Response Fields](#key-response-fields-centralized-json-paths) table above for `$.Result.AccessKey.*` paths.

---

### Operation: DeleteAccessKey — Delete Access Key

> ⚠️ **Destructive Action Confirmation**
> You are about to **permanently delete** access key `{{user.access_key_id}}` for user `{{user.user_name}}`.
> This is **IRREVERSIBLE** — any application using this key will lose access immediately.
> Type the access key ID `{{user.access_key_id}}` to confirm, or reply `abort` to cancel.

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation (type-to-confirm above) — **MUST NOT** proceed without clear user assent
- **MUST** warn: this will invalidate any applications using this key

#### Execution

```bash
# Delete access key
ve iam DeleteAccessKey \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  --AccessKeyId "{{user.access_key_id}}"
```

---

### Operation: UpdateLoginProfile — Set Console Password

#### Execution

```bash
# Create/update login profile
ve iam UpdateLoginProfile \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --UserName "{{user.user_name}}" \
  --Password "{{user.password}}" \
  --PasswordResetRequired false
```

---

### Operation: GetCredentialReport — Generate Credential Report

#### Execution

```bash
# Generate and get credential report
ve iam GenerateCredentialReport \
  --Region "{{env.VOLCENGINE_REGION}}"

# Get the report
ve iam GetCredentialReport \
  --Region "{{env.VOLCENGINE_REGION}}"
```

---

## Failure Feedback Format

When an operation fails, present the result to the user using this standardized block so failures are actionable and consistent:

```
❌ **Operation Failed: <OperationName>**
- **Error code**: `<code>` (from the table below)
- **What happened**: <one-line plain-language explanation>
- **Why it matters**: <impact on the user's resources / data / delivery>
- **Action required**: <concrete next step — e.g. fix input, wait for state, or HALT and escalate>
- **Retry policy**: <0 retries; HALT> or <N retries with backoff> — state explicitly, never silent-retry
```

Rules:
- **MUST** surface the raw error code from the API — do not paraphrase into a generic "something went wrong".
- **MUST** state the retry policy (0 retries → HALT, or bounded retries) so the user knows whether the action auto-repeats.
- **MUST NOT** log or echo `{{env.VOLCENGINE_SECRET_KEY}}` in any failure output.
- On `**HALT**` conditions, stop the runbook and wait for explicit user direction — do not fall through to the next operation.

## Error Taxonomy

| Code | Description | Resolution |
|------|-------------|------------|
| `NoSuchEntity` | 指定的 IAM 实体不存在 | 0 retries; **HALT** |
| `EntityAlreadyExists` | 同名的 IAM 实体已存在 | 0 retries; **HALT** |
| `MalformedPolicyDocument` | 策略文档 JSON 格式不合法 | 0 retries; **HALT** |
| `DeleteConflict` | 实体存在依赖资源，无法删除 | 0 retries; **HALT** |
| `LimitExceeded.User` | 用户数量超出配额限制 | 0 retries; **HALT** |
| `LimitExceeded.Policy` | 策略数量超出配额限制 | 0 retries; **HALT** |
| `LimitExceeded.Role` | 角色数量超出配额限制 | 0 retries; **HALT** |
| `AccessKeyLimitExceeded` | AccessKey 数量超出上限 | 0 retries; **HALT** |
| `PasswordNotComplex` | 密码不满足复杂度策略要求 | 0 retries; **HALT** |
| `InvalidUserName.Malformed` | 用户名格式不符合规范 | 0 retries; **HALT** |
| `InvalidPolicyName.Malformed` | 策略名格式不符合规范 | 0 retries; **HALT** |
| `AttachedEntityLimitExceeded` | 策略附加的实体数量超出限制 | 0 retries; **HALT** |
| `PolicyVersionLimitExceeded` | 策略版本数量超出上限 | 0 retries; **HALT** |
| `Throttling` | IAM API 请求频率超限 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | IAM 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** |
 
## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + IAM-specific safety prompts
- [SecurityOps (Advanced)](references/advanced/securityops.md) — Identity security baseline, credential management, policy audit, incident response

## Operational Best Practices

- **Least privilege:** Grant only necessary permissions; use specific actions and resources
- **Policy versioning:** Use policy versions for safe updates; test before applying
- **Regular audit:** Review credential reports quarterly; rotate access keys every 90 days
- **MFA enforcement:** Require MFA for privileged operations
- **Service roles:** Use service-linked roles instead of long-term credentials
- **Cross-account:** Use role assumption instead of sharing access keys
- **Permission boundaries:** Use permission boundaries to limit maximum permissions
- **Naming conventions:** Use consistent naming for users, roles, and policies
