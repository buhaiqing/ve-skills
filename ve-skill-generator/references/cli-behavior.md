# ve CLI Behavioral Reference

> **Purpose:** Verified behavioral notes and invocation patterns for the Volcengine `ve` CLI, derived from source code analysis and official documentation. Every generated skill MUST follow these conventions.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15

---

## Table of Contents

1. [Command Prefix Change](#1-command-prefix-change)
2. [Default Output is JSON](#2-default-output-is-json)
3. [Native Environment Variable Support](#3-native-environment-variable-support)
4. [JSON Config File Format](#4-json-config-file-format)
5. [Correct CLI Invocation Patterns](#5-correct-cli-invocation-patterns)
6. [Profile Management](#6-profile-management)
7. [SSO Support](#7-sso-support)
8. [Console Login](#8-console-login)
9. [Common Mistakes to Avoid](#9-common-mistakes-to-avoid)

---

## 1. Command Prefix Change

Since **v1.0.20**, the Volcengine CLI command prefix changed from `volcengine-cli` to `ve`:

```bash
# NEW (v1.0.20+):
ve ecs DescribeInstances

# OLD (pre-v1.0.20, still works but deprecated):
volcengine-cli ecs DescribeInstances

# Recommended: set alias for backward compatibility
alias volcengine-cli=ve
```

All generated skills MUST use `ve` as the command prefix.

---

## 2. Default Output is JSON

The `ve` CLI outputs **JSON by default**:

```bash
# Works fine — output is JSON by default
ve ecs DescribeInstances --Region cn-beijing
```

---

## 3. Native Environment Variable Support

The `ve` CLI reads credentials from environment variables natively:

```bash
export VOLCENGINE_ACCESS_KEY="your_ak"
export VOLCENGINE_SECRET_KEY="your_sk"
export VOLCENGINE_REGION="cn-beijing"

# Also supports:
export VOLCENGINE_ENDPOINT="open.volcengineapi.com"       # optional, default endpoint
export VOLCENGINE_ENDPOINT_RESOLVER="standard"            # optional, uses standard resolver
export VOLCENGINE_SESSION_TOKEN="your_token"              # for STS credentials
export VOLCENGINE_DISABLE_SSL="false"                     # optional
export VOLCENGINE_USE_DUALSTACK="false"                   # optional
```

**Supported env vars:**

| Purpose | Variable Name |
|---------|---------------|
| Access Key | `VOLCENGINE_ACCESS_KEY` |
| Secret Key | `VOLCENGINE_SECRET_KEY` |
| Region | `VOLCENGINE_REGION` |
| Endpoint | `VOLCENGINE_ENDPOINT` |
| Endpoint Resolver | `VOLCENGINE_ENDPOINT_RESOLVER` |
| Session Token | `VOLCENGINE_SESSION_TOKEN` |
| Disable SSL | `VOLCENGINE_DISABLE_SSL` |
| Dual Stack | `VOLCENGINE_USE_DUALSTACK` |

---

## 4. JSON Config File Format

The `ve` CLI stores config in `~/.volcengine/config.json` as JSON:

```json
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "AKID",
      "secret_key": "SECRET",
      "region": "cn-beijing",
      "endpoint": "open.volcengineapi.com"
    }
  ]
}
```

---

## 5. Correct CLI Invocation Patterns

### Standard API Calls

```bash
ve <service> <action> --<parameter> value
```

Examples:
```bash
ve ecs DescribeInstances --Region cn-beijing
ve rds_mysql ListDBInstanceIPLists --InstanceId "xxxxxx"
```

### JSON Parameter Passing

```bash
# For array parameters
ve rds_mysql ModifyDBInstanceIPList --InstanceId "xxxxxx" --GroupName "xxxxxx" --IPList '["10.20.30.40", "50.60.70.80"]'

# For full JSON body (Content-Type: application/json APIs)
ve rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"xxxxxx", "GroupName": "xxxxxx", "IPList": ["10.20.30.40", "50.60.70.80"]}'
```

### Help / Parameter Discovery

```bash
# List supported services
ve --help

# List actions for a service
ve ecs --help

# Show parameters for a specific action
ve ecs DescribeInstances --help
```

### Version Check

```bash
ve version
# or
ve -v
```

### Multi-Cloud Credential Namespace Convention

To avoid credential conflicts when mixing cloud providers:

```ini
# Volcengine — use VOLCENGINE_* prefix
VOLCENGINE_ACCESS_KEY=...
VOLCENGINE_SECRET_KEY=...
VOLCENGINE_REGION=cn-beijing

# JD Cloud — use JDCLOUD_* prefix
JDCLOUD_ACCESS_KEY=...
JDCLOUD_SECRET_KEY=...
JDCLOUD_REGION=cn-north-1

# AWS — use AWS_* prefix (standard)
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_DEFAULT_REGION=us-east-1
```

---

## 6. Profile Management

```bash
# Create/modify profile
ve configure set --profile myprofile --region cn-beijing --access-key ak --secret-key sk

# List all profiles
ve configure list

# Get specific profile
ve configure get --profile myprofile

# Switch current profile
ve configure profile --profile myprofile

# Delete profile
ve configure delete --profile myprofile
```

---

## 7. SSO Support

```bash
# 1. Create SSO session
ve configure sso-session --name my-sso --start-url https://{custom}.volccloudidentity.com/userportal --region cn-beijing

# 2. Create SSO profile (triggers device code authorization)
ve configure sso --profile my-dev --sso-session my-sso

# 3. Switch to SSO profile
ve configure profile --profile my-dev

# 4. Re-login (when token expires)
ve sso login --profile my-dev

# 5. Logout
ve sso logout --sso-session my-sso
```

**Note:** `ve configure sso` writes the SSO profile but does **NOT** switch the current default profile. You must run `ve configure profile --profile [name]` afterwards.

---

## 8. Console Login

```bash
# Login via console (OAuth 2.0 + PKCE)
ve login --profile dev --region cn-beijing

# Remote/cross-device login
ve login --profile dev --region cn-beijing --remote

# Logout
ve logout --profile dev

# Logout all console-login profiles
ve logout --all
```

---

## 9. Common Mistakes to Avoid

### Mistake 1: Using Old Command Prefix

```bash
# WRONG (pre-v1.0.20):
volcengine-cli ecs DescribeInstances

# CORRECT (v1.0.20+):
ve ecs DescribeInstances
```

### Mistake 2: Hardcoding Regions

```bash
# WRONG: Hardcoded region
ve ecs DescribeInstances --Region cn-beijing

# CORRECT: Use placeholder (or env var)
ve ecs DescribeInstances --Region "{{env.VOLCENGINE_REGION}}"
```

### Mistake 3: Wrong Env Var Names

```bash
# WRONG: Using wrong variable names that aren't recognized by ve CLI
export ACCESS_KEY=...
export SECRET_KEY=...

# CORRECT: Use Volcengine-prefixed env var names
export VOLCENGINE_ACCESS_KEY=...
export VOLCENGINE_SECRET_KEY=...
export VOLCENGINE_REGION=cn-beijing
```

### Mistake 4: Wrong Config File Path

```bash
# WRONG: Using a non-Volcengine config location
~/.cloud-credentials/config.json

# CORRECT: Volcengine config path
~/.volcengine/config.json
```

---

## See Also

- [Volcengine CLI Source Code](https://github.com/volcengine/volcengine-cli)
- [Execution Environment Setup](execution-environment.md)
- [Enhanced Self-Healing Framework](enhanced-self-healing-framework.md)
