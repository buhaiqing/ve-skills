---
name: ve-kms-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) KMS (密钥管理服务 / Key Management Service) — customer master key (CMK)
  lifecycle, encryption/decryption operations, data key generation, key rotation,
  key policies, and grant management. User mentions KMS, 密钥管理服务, encryption,
  decryption, data key, CMK, or describes key management scenarios even without
  naming the product directly. Not for IAM access keys, certificate management,
  or secret storage services.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`) for KMS API calls, Go SDK
  `github.com/volcengine/volc-sdk-golang/service/kms`, valid API credentials,
  network access to KMS endpoints (kms.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.1.0"
  last_updated: "2026-06-04"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  cli_applicability: dual-path
  cli_support_evidence: >-
    KMS API is accessible via `ve kms --help`. CLI supports DescribeKeys,
    CreateKey, EnableKey, DisableKey, ScheduleKeyDeletion, CancelKeyDeletion,
    DescribeKeyRotation, UpdateKeyRotation, Encrypt, Decrypt, GenerateDataKey
    operations.
    API docs: https://www.volcengine.com/docs/6291
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine KMS Operations Skill

## Overview

KMS (Key Management Service, 密钥管理服务) on Volcengine (火山引擎) provides secure key management for encryption operations, supporting customer master keys (CMK), symmetric (AES) and asymmetric (RSA) keys, encryption/decryption, data key generation, automatic key rotation, key policies, and grants. This skill is an **operational runbook** for agents with **dual-path execution**: `ve` CLI for API calls and JIT Go SDK fallback.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports KMS operations including key lifecycle, encryption/decryption, and key rotation management.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API/CLI response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 KMS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (KMS), one primary resource (CMK); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine KMS", "火山引擎 KMS", "密钥管理服务", "Key Management Service", or "CMK"
- Task involves key operations: CreateKey, DescribeKey, EnableKey, DisableKey, ScheduleKeyDeletion, CancelKeyDeletion
- Task involves encryption/decryption: Encrypt, Decrypt
- Task involves data key generation: GenerateDataKey, GenerateDataKeyWithoutPlaintext
- Task involves key rotation: DescribeKeyRotation, UpdateKeyRotation, manual rotation
- Task involves key policies: PutKeyPolicy, GetKeyPolicy
- Task involves grants: CreateGrant, ListGrants, RevokeGrant, RetireGrant
- Task involves key material import: GetParametersForImport, ImportKeyMaterial, DeleteKeyMaterial
- Task describes key management scenarios: "encrypt data", "generate encryption key", "rotate keys"

### SHOULD NOT Use This Skill When

- Task is about IAM access keys (AK/SK) → delegate to: `ve-iam-ops`
- Task is about certificate management (SSL/TLS) → delegate to certificate ops skill
- Task is about Secrets Manager (secret storage) → delegate to secrets manager ops skill
- Task is purely billing → delegate to billing ops

### Delegation Rules

- KMS key policies depend on IAM principals → reference `ve-iam-ops` for IAM user/role/policy operations
- Encrypted resources (ECS disks, RDS databases) depend on KMS keys → reference respective product skills

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | Access key ID | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | Secret access key | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Region (e.g., cn-beijing) | Use documented default if skill allows |
| `{{user.key_id}}` | KMS key ID | Ask once; format: `key-xxxxxxxx` |
| `{{user.key_alias}}` | KMS key alias | Ask once; format: `alias/my-key` |
| `{{user.key_spec}}` | Key specification | Ask once; default: `AES_256` or `RSA_2048` |
| `{{user.key_usage}}` | Key usage type | Ask once; default: `ENCRYPT_DECRYPT` |
| `{{user.plaintext}}` | Data to encrypt (base64) | Ask once; mask in output |
| `{{user.ciphertext}}` | Encrypted data (base64) | Ask once |
| `{{user.data_key_spec}}` | Data key specification | Ask once; `AES_256` or `AES_128` |
| `{{user.rotation_period}}` | Key rotation period in days | Ask once; default: 365 |
| `{{user.pending_window}}` | Deletion waiting period (7-30 days) | Ask once; default: 7 |
| `{{user.grantee_principal}}` | IAM principal for grant | Ask once |
| `{{user.retiring_principal}}` | IAM principal that can retire grant | Ask once |
| `{{user.operations}}` | Grant operations list | Ask once; e.g., `Encrypt,Decrypt` |
| `{{output.key_id}}` | Key ID from API response | Parse from `$.Result.KeyId` |
| `{{output.key_arn}}` | Key ARN from API response | Parse from `$.Result.KeyArn` |
| `{{output.ciphertext_blob}}` | Encrypted data from response | Parse from `$.Result.CiphertextBlob` |
| `{{output.plaintext}}` | Decrypted data from response | Parse from `$.Result.Plaintext` |
| `{{output.data_key_ciphertext}}` | Encrypted data key | Parse from `$.Result.CiphertextBlob` |
| `{{output.data_key_plaintext}}` | Plaintext data key (sensitive) | Parse and mask immediately |

> **Security Warning (Credential Masking):** NEVER echo or log `VOLCENGINE_SECRET_KEY`, plaintext data keys, or any credential values. Verify existence only with `test -n "$VOLCENGINE_SECRET_KEY"`. Always mask sensitive data in output.

## API and Response Conventions (Agent-Readable)

- **KMS uses JSON REST API** with standard Volcengine API patterns
- **Endpoint:** `https://kms.volcengineapi.com`
- **Go SDK:** `github.com/volcengine/volc-sdk-golang/service/kms`
- **Key States:** `Enabled`, `Disabled`, `PendingDeletion`, `PendingImport`
- **Key Specs:** `AES_256`, `AES_128`, `RSA_2048`, `RSA_3072`, `RSA_4096`, `SM2`
- **Key Usage:** `ENCRYPT_DECRYPT`, `SIGN_VERIFY`

### Key Response Fields

| Operation | Response Field | Type | Description |
|-----------|---------------|------|-------------|
| CreateKey | `$.Result.KeyId` | string | New CMK ID |
| CreateKey | `$.Result.KeyArn` | string | Full ARN of the key |
| CreateKey | `$.Result.KeyState` | string | Initial state: `Enabled` |
| DescribeKey | `$.Result.KeyMetadata.KeyId` | string | Key ID |
| DescribeKey | `$.Result.KeyMetadata.KeyState` | string | Current state |
| DescribeKey | `$.Result.KeyMetadata.KeySpec` | string | Key specification |
| Encrypt | `$.Result.CiphertextBlob` | string | Base64-encoded encrypted data |
| Decrypt | `$.Result.Plaintext` | string | Base64-encoded plaintext |
| GenerateDataKey | `$.Result.CiphertextBlob` | string | Encrypted data key |
| GenerateDataKey | `$.Result.Plaintext` | string | Plaintext data key (sensitive) |

### Key State Transitions

| Operation | Initial State | Target State | Waiting Period |
|-----------|---------------|--------------|----------------|
| CreateKey | — | `Enabled` | Immediate |
| EnableKey | `Disabled` | `Enabled` | Immediate |
| DisableKey | `Enabled` | `Disabled` | Immediate |
| ScheduleKeyDeletion | `Enabled`/`Disabled` | `PendingDeletion` | 7-30 days |
| CancelKeyDeletion | `PendingDeletion` | `Disabled` | Immediate |
| DeleteKeyMaterial | `Enabled` | `PendingImport` | Immediate |

## Quick Start

### What This Skill Does
This skill enables you to manage Volcengine KMS keys — create and manage CMKs, encrypt/decrypt data, generate data keys, configure key rotation, manage key policies and grants — using `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed
- [ ] Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`
- [ ] Region: `VOLCENGINE_REGION` (e.g., `cn-beijing`)

### Verify Setup
```bash
# List KMS keys
ve kms DescribeKeys --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# Create a symmetric encryption key
ve kms CreateKey --KeySpec AES_256 --KeyUsage ENCRYPT_DECRYPT --Region {{env.VOLCENGINE_REGION}}
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand KMS architecture
- [Common Operations](#execution-flows) — Create, manage, and manage encryption keys
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level | | safety_class | blast_radius |
|-----------|-------------|------------|------------||---|---|
| CreateKey | Create a new CMK | Medium | Low | | state-changing | single |
| DescribeKey | View key details and metadata | Low | ✅ None | | read-only | single |
| EnableKey | Enable a disabled key | Low | Low | | state-changing | single |
| DisableKey | Disable a key (reversible) | Low | Medium | | state-changing | single |
| ScheduleKeyDeletion | Schedule key deletion (irreversible after waiting period) | Medium | 🔴 **High** | | state-changing | single |
| CancelKeyDeletion | Cancel pending deletion | Low | Low | | destructive | single |
| Encrypt | Encrypt plaintext data | Low | Low | | state-changing | single |
| Decrypt | Decrypt ciphertext | Low | ✅ None | | state-changing | single |
| GenerateDataKey | Generate a data key (returns plaintext + encrypted) | Medium | 🔴 **High** — handle plaintext securely | | state-changing | single |
| GenerateDataKeyWithoutPlaintext | Generate data key (encrypted only) | Low | Low | | state-changing | single |
| DescribeKeyRotation | View rotation status | Low | ✅ None | | read-only | single |
| UpdateKeyRotation | Enable/disable automatic rotation | Low | Medium | | state-changing | single |
| PutKeyPolicy | Set key access policy | Medium | 🔴 **High** — affects access control | | state-changing | single |
| GetKeyPolicy | Get key access policy | Low | ✅ None | | read-only | single |
| CreateGrant | Create a grant for key access | Medium | Medium | | state-changing | single |
| ListGrants | List grants for a key | Low | ✅ None | | read-only | single |
| RevokeGrant | Revoke a grant | Low | Medium | | destructive | single |

## Changelog
| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-07-13 | T04: annotate operation table with safety_class + blast_radius leaf-op metadata columns (L3 policy inputs); see ve-skill-generator/references/leaf-op-metadata-spec.md |
| 1.0.0 | 2026-05-27 | Initial release with key lifecycle, encryption/decryption, data keys, rotation, policies, grants |
| 1.1.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md`; `max_iter=2` for destructive / state_changing ops, `max_iter=3` for read-only ops |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-kms-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for the scoring dimensions and
> [`references/prompt-templates.md`](references/prompt-templates.md) for the G/C/O
> prompt skeletons and verbatim safety prompts. The Critic and Generator MUST
> live in **isolated prompt contexts**.

### Operation Tiers

| Tier | Operations | `max_iter` | Safety floor |
|---|---|---|---|
| **Destructive** | `ScheduleKeyDeletion` | 2 | 1.0 (mandatory) — `PendingWindowInDays` ≥ 7 required |
| **State-changing** | `DisableKey`, `UpdateKeyRotation`, `PutKeyPolicy`, `RevokeGrant`, `DeleteKeyMaterial` | 2 | 1.0 (mandatory) |
| **Mutating** | `CreateKey`, `EnableKey`, `CreateGrant`, `CancelKeyDeletion`, `Encrypt`, `Decrypt`, `GenerateDataKey`, `GenerateDataKeyWithoutPlaintext` | 2 | ≥ 0.5 |
| **Read-only** | `DescribeKey`, `DescribeKeys`, `DescribeKeyRotation`, `ListGrants`, `GetKeyPolicy` | 3 | ≥ 0 |

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

### KMS-specific safety rules

- **ScheduleKeyDeletion**: `PendingWindowInDays` MUST be ≥ 7 (default 7;
   production keys should use ≥ 30). User must confirm the exact window in days.
- **Encrypt / Decrypt / GenerateDataKey**: plaintext values are NEVER in the
   trace — only `<masked>[length=N]`. The user sees plaintext once.
- **PutKeyPolicy** with `Principal: "*"` or `Action: "kms:*"`: user MUST be
   warned about broad permissions.
- **DisableKey** on a key actively used by production resources: user must
   confirm the impact.
- **GenerateDataKey**: user warned that the plaintext key is returned once
   and must be saved securely.

### Trace

Every GCL run persists a JSON trace to `./audit-results/gcl-trace-*.json`.
Trace MUST NOT contain `VOLCENGINE_SECRET_KEY`, `Encrypt.Plaintext`,
`Decrypt.Plaintext`, or `GenerateDataKey.Plaintext` — only `<masked>`.
`CiphertextBlob` is safe to record (encrypted). See rubric §4 for mandatory
fields.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| IAM principal/user/policy needed for key policy or grant | `ve-iam-ops` |
| Encrypted resource (disk, database) reporting key access issues | respective product skill |
| Billing quota exceeded | `ve-billing-ops` |

The Critic MUST NOT call any skill — it only emits suggestions.

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateKey — Create a Customer Master Key

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | `test -n "$VOLCENGINE_REGION"` | Set | HALT; set VOLCENGINE_REGION |
| Key spec | `{{user.key_spec}}` in `AES_256`, `AES_128`, `RSA_2048`, `RSA_3072`, `RSA_4096`, `SM2` | Valid spec | Use default `AES_256` |
| Key usage | `{{user.key_usage}}` in `ENCRYPT_DECRYPT`, `SIGN_VERIFY` | Valid usage | Use default `ENCRYPT_DECRYPT` |
| Quota | DescribeKeys count < quota | Within quota | HALT; request quota increase |

#### Execution — ve CLI (Primary)

```bash
# Create a symmetric encryption key (AES-256)
ve kms CreateKey \
  --KeySpec AES_256 \
  --KeyUsage ENCRYPT_DECRYPT \
  --Description "{{user.key_description}}" \
  --Region {{env.VOLCENGINE_REGION}}

# Create an asymmetric key for signing
ve kms CreateKey \
  --KeySpec RSA_2048 \
  --KeyUsage SIGN_VERIFY \
  --Description "Signing key for application" \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/base"
    "github.com/volcengine/volc-sdk-golang/service/kms"
)

func main() {
    instance := kms.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    params := map[string]interface{}{
        "Region":      os.Getenv("VOLCENGINE_REGION"),
        "KeySpec":     "AES_256",
        "KeyUsage":    "ENCRYPT_DECRYPT",
        "Description": "CMK for data encryption",
    }
    
    resp, err := instance.Client.Request("kms", "CreateKey", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create key: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

#### Validation

```bash
# Describe the newly created key
ve kms DescribeKey --KeyId "{{output.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Verify key state is Enabled
KEY_STATE=$(ve kms DescribeKey --KeyId "{{output.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyState')
[ "$KEY_STATE" = "Enabled" ] && echo "Key created and enabled successfully"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidParameter` | HALT; check KeySpec and KeyUsage combination |
| `QuotaExceeded` | HALT; request quota increase from Volcengine console |
| `AccessDenied` | HALT; verify IAM permissions for `kms:CreateKey` |
| `InternalError` | Retry 3x with exponential backoff; then HALT |

---

### Operation: ScheduleKeyDeletion — Schedule Key Deletion

#### Pre-flight (Safety Gate)

**CRITICAL SAFETY REQUIREMENTS:**

- **MUST** obtain explicit confirmation: irreversible deletion of key `{{user.key_id}}`
- **MUST** verify key is not the default key for any service
- **MUST** list all encrypted resources using this key and warn about impact
- **MUST** require `{{user.pending_window}}` between 7-30 days (default: 7)
- **MUST NOT** proceed without clear user assent

```bash
# Check key metadata
ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# List grants (indicates other principals using the key)
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Warning: Cannot recover key after deletion completes
# Warning: All data encrypted with this key will be unrecoverable
```

#### Execution — ve CLI

```bash
# Schedule deletion with 7-day waiting period
ve kms ScheduleKeyDeletion \
  --KeyId "{{user.key_id}}" \
  --PendingWindowInDays 7 \
  --Region {{env.VOLCENGINE_REGION}}

# Schedule deletion with 30-day waiting period (maximum)
ve kms ScheduleKeyDeletion \
  --KeyId "{{user.key_id}}" \
  --PendingWindowInDays 30 \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Execution — JIT Go SDK

```go
params := map[string]interface{}{
    "Region":             os.Getenv("VOLCENGINE_REGION"),
    "KeyId":              os.Getenv("KEY_ID"),
    "PendingWindowInDays": 7, // or 30 for maximum delay
}

resp, err := instance.Client.Request("kms", "ScheduleKeyDeletion", params)
```

#### Validation

```bash
# Verify key state changed to PendingDeletion
KEY_STATE=$(ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyState')
[ "$KEY_STATE" = "PendingDeletion" ] && echo "Key deletion scheduled"

# Check deletion date
DELETION_DATE=$(ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.DeletionDate')
echo "Key will be deleted on: $DELETION_DATE"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidArn` | HALT; verify KeyId format is correct |
| `NotFound` | Key already deleted or never existed |
| `KMSInvalidState` | HALT; key is already pending deletion |
| `DependencyViolation` | HALT; key is in use by other resources |
| `AccessDenied` | HALT; verify IAM permissions |

---

### Operation: CancelKeyDeletion — Cancel Pending Deletion

#### Execution — ve CLI

```bash
# Cancel scheduled deletion (restores key to Disabled state)
ve kms CancelKeyDeletion --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}

# Enable the key after canceling deletion
ve kms EnableKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

#### Validation

```bash
# Verify key state changed to Disabled
KEY_STATE=$(ve kms DescribeKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.KeyMetadata.KeyState')
[ "$KEY_STATE" = "Disabled" ] && echo "Key deletion canceled, key is now Disabled"

# Enable if needed
ve kms EnableKey --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

---

### Operation: Encrypt — Encrypt Data

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Key exists | `ve kms DescribeKey --KeyId` | Key found | HALT; create key first |
| Key state | `$.Result.KeyMetadata.KeyState` | `Enabled` | HALT; enable key first |
| Key usage | `$.Result.KeyMetadata.KeyUsage` | `ENCRYPT_DECRYPT` | HALT; use compatible key |
| Plaintext format | Base64 encoded | Valid base64 | Encode if needed |

#### Execution — ve CLI

```bash
# Encrypt plaintext (base64 encoded)
ve kms Encrypt \
  --KeyId "{{user.key_id}}" \
  --Plaintext "{{user.plaintext}}" \
  --Region {{env.VOLCENGINE_REGION}}

# Encrypt with encryption context (additional authenticated data)
ve kms Encrypt \
  --KeyId "{{user.key_id}}" \
  --Plaintext "{{user.plaintext}}" \
  --EncryptionContext '{"Purpose":"ApplicationData","Environment":"Production"}' \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Execution — JIT Go SDK

```go
params := map[string]interface{}{
    "Region":    os.Getenv("VOLCENGINE_REGION"),
    "KeyId":     os.Getenv("KEY_ID"),
    "Plaintext": os.Getenv("PLAINTEXT"), // Base64 encoded
    "EncryptionContext": map[string]string{
        "Purpose":     "ApplicationData",
        "Environment": "Production",
    },
}

resp, err := instance.Client.Request("kms", "Encrypt", params)
```

#### Validation

```bash
# Capture ciphertext from response
CIPHERTEXT=$(ve kms Encrypt --KeyId "{{user.key_id}}" --Plaintext "{{user.plaintext}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.CiphertextBlob')
echo "Encryption successful"
echo "Ciphertext: ${CIPHERTEXT:0:50}..."

# Verify ciphertext is not empty
[ -n "$CIPHERTEXT" ] && echo "Ciphertext received"
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `Disabled` | HALT; key is disabled, enable first |
| `InvalidKeyUsage` | HALT; key does not support encryption |
| `KeyUnavailable` | Retry 3x; then HALT |
| `AccessDenied` | HALT; verify IAM permissions for `kms:Encrypt` |

---

### Operation: Decrypt — Decrypt Data

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Ciphertext provided | `{{user.ciphertext}}` | Non-empty | HALT; provide ciphertext |
| Same encryption context | If used during encrypt, must match | Match | HALT; provide matching context |

#### Execution — ve CLI

```bash
# Decrypt ciphertext
ve kms Decrypt \
  --CiphertextBlob "{{user.ciphertext}}" \
  --Region {{env.VOLCENGINE_REGION}}

# Decrypt with encryption context (must match encryption context used)
ve kms Decrypt \
  --CiphertextBlob "{{user.ciphertext}}" \
  --EncryptionContext '{"Purpose":"ApplicationData","Environment":"Production"}' \
  --Region {{env.VOLCENGINE_REGION}}
```

> Note: Decrypt does not require `--KeyId` because the ciphertext contains the key ID.

#### Validation

```bash
# Decrypt and verify output
PLAINTEXT=$(ve kms Decrypt --CiphertextBlob "{{user.ciphertext}}" --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.Plaintext')
echo "Decryption successful"
echo "Plaintext (masked): ${PLAINTEXT:0:10}..."
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `InvalidCiphertext` | HALT; ciphertext is malformed or corrupted |
| `IncorrectEncryptionContext` | HALT; encryption context does not match |
| `AccessDenied` | HALT; verify IAM permissions for `kms:Decrypt` |
| `KeyUnavailable` | Key may be disabled or pending deletion |

---

### Operation: GenerateDataKey — Generate a Data Encryption Key

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Key exists and enabled | `ve kms DescribeKey` | `Enabled` | HALT; check key state |
| Key supports encryption | `$.Result.KeyMetadata.KeyUsage` | `ENCRYPT_DECRYPT` | HALT; use compatible key |
| Data key spec | `{{user.data_key_spec}}` | `AES_256` or `AES_128` | Use default `AES_256` |

#### Execution — ve CLI

```bash
# Generate a 256-bit data key
ve kms GenerateDataKey \
  --KeyId "{{user.key_id}}" \
  --KeySpec AES_256 \
  --Region {{env.VOLCENGINE_REGION}}

# Generate with encryption context
ve kms GenerateDataKey \
  --KeyId "{{user.key_id}}" \
  --KeySpec AES_256 \
  --EncryptionContext '{"Purpose":"DataEncryption","Application":"MyApp"}' \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Execution — JIT Go SDK

```go
params := map[string]interface{}{
    "Region":  os.Getenv("VOLCENGINE_REGION"),
    "KeyId":   os.Getenv("KEY_ID"),
    "KeySpec": "AES_256",
}

resp, err := instance.Client.Request("kms", "GenerateDataKey", params)
```

#### Validation

```bash
# Generate and capture both keys
RESPONSE=$(ve kms GenerateDataKey --KeyId "{{user.key_id}}" --KeySpec AES_256 --Region {{env.VOLCENGINE_REGION}})
PLAINTEXT_KEY=$(echo "$RESPONSE" | jq -r '.Result.Plaintext')
CIPHERTEXT_KEY=$(echo "$RESPONSE" | jq -r '.Result.CiphertextBlob')

echo "Data key generated"
echo "Plaintext key (first 16 chars): ${PLAINTEXT_KEY:0:16}... (HANDLE WITH CARE)"
echo "Encrypted key (first 50 chars): ${CIPHERTEXT_KEY:0:50}..."

# Verify both keys present
[ -n "$PLAINTEXT_KEY" ] && [ -n "$CIPHERTEXT_KEY" ] && echo "Both keys received successfully"
```

#### Security Handling

**CRITICAL: The plaintext data key is sensitive**

- Use the plaintext key to encrypt your data locally
- Store only the encrypted data key (`CiphertextBlob`)
- Discard the plaintext key from memory after use
- To decrypt data later, call `Decrypt` on the encrypted data key

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `Disabled` | HALT; key is disabled |
| `InvalidKeyUsage` | HALT; key does not support data key generation |
| `AccessDenied` | HALT; verify IAM permissions for `kms:GenerateDataKey` |

---

### Operation: GenerateDataKeyWithoutPlaintext — Generate Encrypted Data Key Only

#### Execution — ve CLI

```bash
# Generate encrypted data key (no plaintext returned)
ve kms GenerateDataKeyWithoutPlaintext \
  --KeyId "{{user.key_id}}" \
  --KeySpec AES_256 \
  --Region {{env.VOLCENGINE_REGION}}
```

> Use this when you want to generate a data key for someone else to use, or when you don't need the plaintext immediately.

---

### Operation: UpdateKeyRotation — Enable/Disable Automatic Key Rotation

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Key exists | `ve kms DescribeKey` | Key found | HALT |
| Key is symmetric | `$.Result.KeyMetadata.KeySpec` | `AES_256` or `AES_128` | HALT; asymmetric keys don't support rotation |
| Key origin | `$.Result.KeyMetadata.Origin` | `VOLCENGINE_KMS` | HALT; external keys don't support rotation |

#### Execution — ve CLI

```bash
# Enable automatic key rotation
ve kms UpdateKeyRotation \
  --KeyId "{{user.key_id}}" \
  --EnableAutomaticKeyRotation true \
  --Region {{env.VOLCENGINE_REGION}}

# Disable automatic key rotation
ve kms UpdateKeyRotation \
  --KeyId "{{user.key_id}}" \
  --EnableAutomaticKeyRotation false \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Validation

```bash
# Verify rotation status
ve kms DescribeKeyRotation --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

#### Failure Recovery

| Error Pattern | Agent Action |
|--------------|-------------|
| `UnsupportedOperation` | HALT; key type does not support rotation |
| `AccessDenied` | HALT; verify IAM permissions |

---

### Operation: PutKeyPolicy — Set Key Access Policy

#### Pre-flight (Safety Gate)

- **MUST** validate policy JSON syntax before applying
- **MUST** warn about broad permissions (`"Principal": "*"`, `"Action": "kms:*"`)
- **MUST** require explicit confirmation for destructive policy changes

#### Execution — ve CLI

```bash
# Apply key policy from JSON file
ve kms PutKeyPolicy \
  --KeyId "{{user.key_id}}" \
  --PolicyName default \
  --Policy file://key-policy.json \
  --Region {{env.VOLCENGINE_REGION}}

# Apply inline policy
ve kms PutKeyPolicy \
  --KeyId "{{user.key_id}}" \
  --PolicyName default \
  --Policy '{"Version":"2012-10-17","Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789:root"},"Action":"kms:*","Resource":"*"}]}' \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Validation

```bash
# Verify policy applied
ve kms GetKeyPolicy --KeyId "{{user.key_id}}" --PolicyName default --Region {{env.VOLCENGINE_REGION}}
```

---

### Operation: CreateGrant — Create a Grant for Key Access

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Grantee principal | `{{user.grantee_principal}}` | Valid IAM ARN | HALT |
| Operations | `{{user.operations}}` | Valid KMS operations | HALT |
| Key state | `Enabled` | Enabled | HALT |

#### Execution — ve CLI

```bash
# Create grant for encrypt/decrypt operations
ve kms CreateGrant \
  --KeyId "{{user.key_id}}" \
  --GranteePrincipal "{{user.grantee_principal}}" \
  --Operations '["Encrypt","Decrypt","GenerateDataKey"]' \
  --Region {{env.VOLCENGINE_REGION}}

# Create grant with retiring principal
ve kms CreateGrant \
  --KeyId "{{user.key_id}}" \
  --GranteePrincipal "{{user.grantee_principal}}" \
  --RetiringPrincipal "{{user.retiring_principal}}" \
  --Operations '["Encrypt","Decrypt"]' \
  --Constraints '{"EncryptionContextSubset":{"Purpose":"ApplicationData"}}' \
  --Region {{env.VOLCENGINE_REGION}}
```

#### Validation

```bash
# List grants to verify creation
ve kms ListGrants --KeyId "{{user.key_id}}" --Region {{env.VOLCENGINE_REGION}}
```

---

## Error Taxonomy (≥ 10 Product-Specific Codes)

| Error Code | Meaning | Resolution |
|------------|---------|-----------|
| `InvalidParameter` | Request parameter invalid | 0 retries; **RETRY** — Fix per OpenAPI docs |
| `InvalidArn` | ARN format incorrect | 0 retries; **HALT** — Fix ARN format |
| `NotFound` | Key or resource not found | 0 retries; **HALT** — Verify key ID exists |
| `Disabled` | Key is disabled | 0 retries; **RETRY** — Enable key first |
| `KMSInvalidState` | Key in wrong state for operation | 0 retries; **HALT** — Check key state and retry |
| `InvalidKeyUsage` | Key usage doesn't support operation | 0 retries; **HALT** — Use compatible key type |
| `DependencyViolation` | Key in use by other resources | 0 retries; **HALT** — Remove dependencies first |
| `UnsupportedOperation` | Operation not supported for key type | 0 retries; **HALT** — Use different key or operation |
| `IncorrectEncryptionContext` | Encryption context mismatch | 0 retries; **HALT** — Provide matching context |
| `InvalidCiphertext` | Ciphertext malformed | 0 retries; **HALT** — Verify ciphertext integrity |
| `KeyUnavailable` | Key temporarily unavailable | 3 retries/2s/4s/8s; **RETRY** — Retry with backoff |
| `AccessDenied` | Insufficient IAM permissions | 0 retries; **HALT** — Add IAM policy |
| `QuotaExceeded` | Resource quota limit reached | 0 retries; **HALT** — Request quota increase |
| `Throttling` | Rate limit exceeded | 3 retries/exponential; **RETRY** — Back off and retry |
| `InternalError` | Server-side error | 3 retries/2s/4s/8s; **RETRY** — Retry, escalate with RequestId |

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring & Alerts](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../ve-skill-generator/references/cli-behavior.md)
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + KMS-specific safety prompts
- [SecurityOps (Advanced)](references/advanced/securityops.md) — Key security baseline, encryption governance, key lifecycle incident response

## Operational Best Practices

- **Key rotation:** Enable automatic rotation for symmetric keys; rotate annually minimum
- **Encryption context:** Always use encryption context for additional authentication
- **Data keys:** Use data keys for application-level encryption; never encrypt large data directly with CMK
- **Key deletion:** Use 30-day waiting period for production keys; document all encrypted resources
- **Access control:** Use key policies + IAM policies for defense in depth
- **Grants:** Use grants for temporary access; prefer grants over key policy changes
- **Monitoring:** Monitor key usage via CloudTrail; set alarms on key deletion schedules
