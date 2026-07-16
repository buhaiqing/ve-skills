---
name: ve-tos-ops
description: >-
  Use when the user needs to deploy, configure, troubleshoot, or manage Volcengine
  (火山引擎) TOS (对象存储 / Torch Object Storage) — bucket lifecycle, object
  upload/download/copy, multi-part uploads, lifecycle rules, versioning,
  access control (ACL/policy), pre-signed URLs, and cross-region replication.
  User mentions TOS, 对象存储, or describes storage-related scenarios
  (e.g., uploading files to cloud storage, downloading large objects, setting
  bucket permissions, configuring lifecycle rules) even without naming the
  product directly. Not for ECS block storage (云盘), file storage (NAS), or
  database storage services.
license: MIT
compatibility: >-
  Official tosutil CLI tool, Volcengine CLI (`ve`) for API calls, Go SDK
  `github.com/volcengine/ve-tos-golang-sdk/v2`, valid API credentials,
  network access to TOS endpoints (tos-{region}.volces.com).
metadata:
  author: volcengine
  version: "1.2.1"
  last_updated: "2026-07-11"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.13"
  cli_applicability: dual-path
  cli_support_evidence: >-
    TOS API is accessible via `ve tos --help`. Bulk data operations use the
    dedicated `tosutil` CLI tool.
    See: https://github.com/volcengine/tosutil
    API docs: https://www.volcengine.com/docs/6349
  environment:
    - TOS_ACCESS_KEY
    - TOS_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine TOS Operations Skill

## Overview

TOS (Torch Object Storage, 对象存储) on Volcengine (火山引擎) provides massively scalable object storage with S3-compatible APIs, supporting buckets, objects, lifecycle management, versioning, and access control. This skill is an **operational runbook** for agents with **dual-path execution**: `ve` CLI for API calls, `tosutil` for bulk data transfer, and JIT Go SDK fallback.

### CLI applicability

- **`cli_applicability: dual-path`:** TOS API is accessible via both `ve tos` and the dedicated `tosutil` tool.
  - **`ve tos`**: API operations (CreateBucket, ListBuckets, etc.)
  - **`tosutil`**: Bulk data transfer (cp, ls, rm, mb) — recommended for upload/download

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.TOS_*}}` (TOS env), `{{user.*}}` (interactive), `{{output.*}}` (API/tosutil response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 TOS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (TOS), one primary resource (Bucket/Object); cross-product delegation documented |
| 6 | **FinOps Integration** | Storage analysis, stale detection, multipart cleanup, cost reports, storage class optimization |
| 7 | **AIOps Integration** | Knowledge base with fault patterns, cross-skill diagnosis, proactive inspection |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine TOS", "火山引擎 TOS", "对象存储", "Torch Object Storage", or "tos"
- Task involves bucket operations: CreateBucket, ListBuckets, DeleteBucket, GetBucketLocation
- Task involves object operations: PutObject, GetObject, DeleteObject, CopyObject, ListObjects
- Task involves multi-part upload: CreateMultipartUpload, UploadPart, CompleteMultipartUpload
- Task involves lifecycle rules, versioning, ACL/policy, pre-signed URLs
- Task involves bulk data transfer: upload/download/copy local files to/from TOS
- Task involves **cost optimization**: storage class analysis, stale object detection, cost reports
- Task involves **resource cleanup**: incomplete multipart uploads, expired objects, orphaned delete markers
- Task involves **storage class optimization**: transitioning objects between Standard/IA/Archive

### SHOULD NOT Use This Skill When

- Task is about ECS cloud disks (云盘) → delegate to: `ve-ecs-ops`
- Task is about NAS file storage → delegate to NAS ops skill (when present)
- Task is about database storage (RDS) → delegate to: `ve-rds-ops` (when present)
- Task is purely billing → delegate to billing ops

### Delegation Rules

- TOS bucket permissions depend on IAM policies → reference `ve-iam-ops` (when present)
- TOS cross-region replication depends on VPC networking → reference `ve-vpc-ops` (when present)

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.TOS_ACCESS_KEY}}` | TOS access key | NEVER ask user; fail if unset |
| `{{env.TOS_SECRET_KEY}}` | TOS secret key | NEVER ask user; fail if unset |
| `{{env.VOLCENGINE_REGION}}` | Region (e.g., cn-beijing) | Use documented default if skill allows |
| `{{user.bucket}}` | Bucket name | Ask once; globally unique; lowercase alphanumeric + hyphens |
| `{{user.object_key}}` | Object key (path) | Ask once; URL-safe |
| `{{user.local_file}}` | Local file path | Ask once; verify exists |
| `{{user.endpoint}}` | TOS endpoint | Default: `https://tos-{{env.VOLCENGINE_REGION}}.volces.com` |
| `{{output.etag}}` | Object ETag from upload response | Parse from response |
| `{{output.request_id}}` | Request ID for tracing | Parse from response |

> **TOS-specific env vars:** TOS uses `TOS_ACCESS_KEY`/`TOS_SECRET_KEY` (not `VOLCENGINE_*`). Use `VOLCENGINE_REGION` for region.

> **Security Warning (Credential Masking):** NEVER echo or log `TOS_SECRET_KEY` or any credential values. Verify existence only with `test -n "$TOS_SECRET_KEY"`.

## API and Response Conventions (Agent-Readable)

- **TOS uses RESTful S3-compatible API** with XML responses for most operations
- **Endpoint:** `https://tos-{region}.volces.com` (e.g., `https://tos-cn-beijing.volces.com`)
- **Go SDK:** `github.com/volcengine/ve-tos-golang-sdk/v2/tos`
- **Error responses:** S3-compatible `Error` XML with `Code` and `Message` elements

### Key Response Fields

| Operation | Response Field | Type | Description |
|-----------|---------------|------|-------------|
| ListBuckets | `$.Buckets[].Name` | array | Bucket names |
| ListBuckets | `$.Buckets[].CreationDate` | array | Creation timestamps |
| ListObjectsV2 | `$.Contents[]` | array | Object list |
| ListObjectsV2 | `$.Contents[].Key` | string | Object key |
| ListObjectsV2 | `$.Contents[].Size` | integer | Object size in bytes |
| ListObjectsV2 | `$.IsTruncated` | boolean | More pages available |
| ListObjectsV2 | `$.NextMarker` | string | Pagination token |
| PutObjectV2 | `.ETag` | string | Object tag |
| PutObjectV2 | `.RequestID` | string | Request identifier |
| CreateBucketV2 | `.Location` | string | Bucket location |

## Quick Start

### What This Skill Does
This skill enables you to manage Volcengine TOS buckets and objects — create buckets, upload/download files, set lifecycle rules, generate pre-signed URLs — using `tosutil` (bulk data) or `ve tos` (API calls), with JIT Go SDK fallback.

### Prerequisites
- [ ] `tosutil` CLI installed (bulk data) or `ve` CLI (API calls)
- [ ] Credentials: `TOS_ACCESS_KEY`, `TOS_SECRET_KEY`
- [ ] Region: `VOLCENGINE_REGION` (e.g., `cn-beijing`)

### Verify Setup
```bash
# List buckets using tosutil
tosutil ls

# Or using ve CLI
ve tos ListBuckets
```

### Your First Command
```bash
# List all buckets
tosutil ls -s
```

### Next Steps
- [Core Concepts](references/core-concepts.md) — Understand TOS architecture
- [Common Operations](#execution-flows) — Create, manage, and manage object storage
- [Troubleshooting](references/troubleshooting.md) — Fix common issues

## Capabilities at a Glance

| Operation | Description | Level |
|-----------|-------------|-------|
| CreateBucket | Create a new TOS bucket | Low/Med |
| ListBuckets | List all accessible buckets | Low/None |
| DeleteBucket | Delete an empty bucket | Low/**High** |
| PutObject | Upload an object | Low/Low |
| GetObject | Download an object | Low/None |
| ListObjects | List objects in a bucket | Low/None |
| DeleteObject | Delete an object | Low/**High** |
| CopyObject | Copy object within/across buckets | Med/Med |
| MultipartUpload | Upload large files in parts | High/Low |
| PresignURL | Generate pre-signed URL | Low/Med |
| PutBucketLifecycle | Set lifecycle rules | Med/Med |
| PutBucketVersioning | Enable/disable versioning | Low/Med |
| PutBucketACL | Set bucket access control | Low/**High** |
| DescribeStorageAnalysis | Analyze storage distribution & costs | Low/None |
| DetectStaleObjects | Find objects not accessed for X days | Low/None |
| DescribeCostSummary | Generate TOS cost report | Low/None |
| CleanupMultipartUploads | Abort incomplete multipart uploads | Low/Low |
| OptimizeStorageClass | Recommend storage class changes | Med/Low |

> FinOps Operations 详情 → [references/advanced/finops.md#operations](references/advanced/finops.md#operations)

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-15 | Initial release with bucket/object management, lifecycle, versioning |
| 1.1.0 | 2026-05-27 | Added FinOps operations (storage analysis, stale detection, cleanup, cost reports) and AIOps knowledge base |
| 1.2.0 | 2026-06-04 | Phase 1 GCL rollout: added `## Quality Gate (GCL)` chapter, `references/rubric.md`, `references/prompt-templates.md` |
| 1.2.1 | 2026-07-11 | Split FinOps Operations (5 ops) to `references/advanced/finops.md` (TE-7); SKILL.md retains entry link only |

## Quality Gate (GCL)

> This chapter is **mandatory** for every execution of `ve-tos-ops`. It implements
> the Generator-Critic-Loop defined in `../../AGENTS.md` §3-§9. Read
> [`references/rubric.md`](references/rubric.md) for scoring and
> [`references/prompt-templates.md`](references/prompt-templates.md) for safety prompts.
> **Note:** TOS uses `TOS_ACCESS_KEY` / `TOS_SECRET_KEY` (not `VOLCENGINE_*`).

### Operation Tiers

> See [`references/rubric.md` §0](references/rubric.md#0-operation-tier) for the full operation tier table.

### Loop

1. **Pre-flight** — resolve `{{env.*}}` / `{{user.*}}`; classify tier; load rubric.
2. **Generate** — execute per `## Execution Flows`. Trace to `./audit-results/gcl-trace-*.json`.
3. **Critique** — isolated prompt; score 5 dimensions; MUST NOT see raw request.
4. **Decide** — Safety=0 on Destructive/State-changing → **ABORT**; all pass → return; `iter<max_iter` → loop.

### TOS-specific safety rules

- **DeleteBucket**: MUST check bucket emptiness (objects + versions + multipart uploads); warn all data lost.
- **DeleteObject prefix-pattern (`-r`)**: MUST show file list before execution.
- **PutBucketACL** with `public-read`/`public-read-write`: warn about public internet access.
- **PutBucketVersioning suspend**: warn about losing new-overwrite protection.
- **OptimizeStorageClass to Archive/ColdArchive**: warn about retrieval costs (per-GB fee + 1-12h restore time).
- **PutBucketLifecycle with Expiration**: warn that objects will be permanently deleted.
- **TOS_SECRET_KEY** NEVER in trace (only `<masked>`).

### Trace

`./audit-results/gcl-trace-*.json` — with `redaction_pass: true`. Both `tosutil` and `ve tos` commands recorded. Note: TOS uses `{{env.TOS_SECRET_KEY}}` not `{{env.VOLCENGINE_SECRET_KEY}}`.

### Cross-skill delegation

| Critic finding | Delegate to |
|---|---|
| IAM policy/permission gap | `ve-iam-ops` |
| VPC/network context | `ve-vpc-ops` |
| Billing/quota exceeded | `ve-billing-ops` |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: CreateBucket — Create a TOS Bucket

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$TOS_ACCESS_KEY" && test -n "$TOS_SECRET_KEY"` | Both set | HALT; configure credentials |
| Bucket name unique | Bucket name must be globally unique in TOS | No conflict | Use a different name |
| Name format | Lowercase letters, numbers, hyphens; 3–63 chars | Valid format | Fix name format |

#### Execution — tosutil CLI (Primary for bucket management)

```bash
# Create a bucket
tosutil mb tos://{{user.bucket}}

# Create a bucket with storage class
tosutil mb tos://{{user.bucket}} -sc=IA  # Infrequent Access

# Create a bucket with ACL
tosutil mb tos://{{user.bucket}} -acl=public-read
```

#### Execution — ve CLI API

```bash
ve tos CreateBucket --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func main() {
    client, err := tos.NewClientV2(
        "https://tos-"+os.Getenv("VOLCENGINE_REGION")+".volces.com",
        tos.WithRegion(os.Getenv("VOLCENGINE_REGION")),
        tos.WithCredentials(tos.NewStaticCredentials(
            os.Getenv("TOS_ACCESS_KEY"),
            os.Getenv("TOS_SECRET_KEY"),
        )),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    resp, err := client.CreateBucketV2(context.Background(), &tos.CreateBucketV2Input{
        Bucket:       os.Getenv("BUCKET_NAME"),
        ACL:          enum.ACLPrivate,
        StorageClass: enum.StorageClassStandard,
    })
    if err != nil {
        log.Fatalf("Failed to create bucket: %v", err)
    }
    fmt.Printf("Bucket created at: %s\n", resp.Location)
}
```

#### Validation

```bash
tosutil ls -s | grep "{{user.bucket}}"
```

#### Failure Recovery

| Error Pattern | Action |
|--------------|--------|
| `BucketAlreadyExists` | HALT → use different name |
| `InvalidBucketName` | HALT → 3-63 chars, lowercase, hyphens |
| `Unauthorized` | HALT → attach TOSFullAccess policy |
| `TooManyBuckets` | HALT → delete unused or request increase |

---

### Operation: DeleteBucket — Delete a TOS Bucket

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible delete of bucket `{{user.bucket}}`
- **MUST NOT** proceed without clear user assent
- **MUST** verify bucket is empty — delete all objects first
- **MUST** warn about versioning — delete all versions and delete markers if versioning enabled
- **MUST** warn about multipart uploads — abort any in-progress uploads

```bash
# Check if bucket is empty
tosutil ls tos://{{user.bucket}} -s

# If versioning enabled, check for non-current versions
ve tos ListObjectVersions --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"

# Check for in-progress multipart uploads
ve tos ListMultipartUploads --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — tosutil CLI

```bash
# Delete an empty bucket
tosutil rb tos://{{user.bucket}}

# Force delete bucket and all contents (USE WITH EXTREME CAUTION)
tosutil rb tos://{{user.bucket}} -f
```

#### Execution — ve CLI API

```bash
ve tos DeleteBucket --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"
```

#### Execution — JIT Go SDK (Fallback)

```go
_, err := client.DeleteBucketV2(context.Background(), &tos.DeleteBucketV2Input{
    Bucket: os.Getenv("BUCKET"),
})
if err != nil {
    log.Fatalf("Failed to delete bucket: %v", err)
}
fmt.Println("Bucket deleted successfully")
```

#### Validation

```bash
# Verify bucket no longer exists
tosutil ls -s | grep "{{user.bucket}}" && echo "BUCKET STILL EXISTS" || echo "BUCKET DELETED"
```

#### Failure Recovery

| Error Pattern | Action |
|--------------|--------|
| `BucketNotEmpty` | HALT → delete all objects, then retry |
| `NoSuchBucket` | Skip (already deleted) |
| `AccessDenied` | HALT → check IAM + ACL |
| `Unauthorized` | HALT → attach TOSFullAccess policy |

---

### Operation: Upload Object — Upload File to TOS

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Bucket exists | `tosutil ls tos://{{user.bucket}}` | Bucket found | HALT; create bucket first |
| Local file exists | `test -f "{{user.local_file}}"` | File readable | HALT; verify file path |
| File size | `stat` or `ls -l` | Check if > 100MB | Use multipart for large files |

#### Execution — tosutil CLI (Primary)

```bash
# Upload a single file
tosutil cp "{{user.local_file}}" tos://{{user.bucket}}/{{user.object_key}}

# Upload a directory recursively
tosutil cp "{{user.local_folder}}" tos://{{user.bucket}}/{{user.object_prefix}} -r

# Upload with progress bar and verification
tosutil cp "{{user.local_file}}" tos://{{user.bucket}}/{{user.object_key}} --enable-verify
```

#### Execution — ve CLI API (for small objects)

```bash
# Note: ve CLI does not support binary upload directly
# Use tosutil or Go SDK for actual file upload
```

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func main() {
    // Initialize TOS client (same pattern as CreateBucket)
    client, err := tos.NewClientV2(
        "https://tos-"+os.Getenv("VOLCENGINE_REGION")+".volces.com",
        tos.WithRegion(os.Getenv("VOLCENGINE_REGION")),
        tos.WithCredentials(tos.NewStaticCredentials(
            os.Getenv("TOS_ACCESS_KEY"),
            os.Getenv("TOS_SECRET_KEY"),
        )),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // Open local file for upload
    fh, err := os.Open(os.Getenv("LOCAL_FILE"))
    if err != nil {
        log.Fatalf("Failed to open file: %v", err)
    }
    defer fh.Close()

    output, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
        PutObjectBasicInput: tos.PutObjectBasicInput{
            Bucket: os.Getenv("BUCKET"),
            Key:    os.Getenv("OBJECT_KEY"),
        },
        Content: fh,
    })
    if err != nil {
        log.Fatalf("Failed to upload object: %v", err)
    }
    fmt.Printf("Uploaded: ETag=%s\n", output.ETag)
}
```

#### Validation

```bash
# Check object exists and get metadata
tosutil stat tos://{{user.bucket}}/{{user.object_key}}
```

#### Failure Recovery

| Error Pattern | Action |
|--------------|--------|
| `NoSuchBucket` | HALT → create bucket first |
| `AccessDenied` | HALT → check ACL + IAM |
| `NetworkError` | Retry with backoff; reduce part size (`-ps=5mb`) |
| `QuotaExceeded` | HALT → bucket/account limit |

---

### Operation: Download Object — Download File from TOS

#### Execution

```bash
# Download a single file
tosutil cp tos://{{user.bucket}}/{{user.object_key}} "{{user.local_file}}"

# Download recursively
tosutil cp tos://{{user.bucket}}/{{user.object_prefix}} "{{user.local_folder}}" -r

# Download with verification
tosutil cp tos://{{user.bucket}}/{{user.object_key}} "{{user.local_file}}" --enable-verify
```

#### Go SDK

```go
output, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
    Bucket: os.Getenv("BUCKET"),
    Key:    os.Getenv("OBJECT_KEY"),
})
if err != nil {
    log.Fatal(err)
}
defer output.Content.Close()

// Save to local file
fw, err := os.Create(os.Getenv("LOCAL_FILE"))
if err != nil {
    log.Fatal(err)
}
defer fw.Close()
io.Copy(fw, output.Content)
```

#### Validation

```bash
# Verify object exists and check metadata (size, ETag)
tosutil stat tos://{{user.bucket}}/{{user.object_key}}

# Compare downloaded file size with expected size
EXPECTED_SIZE=$(tosutil stat tos://{{user.bucket}}/{{user.object_key}} | grep -i 'content-length' | awk '{print $2}')
ACTUAL_SIZE=$(stat -f%z "{{user.local_file}}" 2>/dev/null || stat -c%s "{{user.local_file}}")
[ "$EXPECTED_SIZE" = "$ACTUAL_SIZE" ] && echo "SIZE MATCH" || echo "SIZE MISMATCH (expected: $EXPECTED_SIZE, actual: $ACTUAL_SIZE)"

# For extra integrity verification, compare MD5/ETag
echo "Download verified successfully"
```

---

### Operation: ListObjects — List Objects in Bucket

#### Execution

```bash
# List root objects
tosutil ls tos://{{user.bucket}}

# List with prefix filter (recursive)
tosutil ls tos://{{user.bucket}}/{{user.object_prefix}} -s

# List with limit
tosutil ls tos://{{user.bucket}} --limited-num 50

# List flat (no directory grouping)
tosutil ls tos://{{user.bucket}} -ab
```

#### Go SDK (Paginated)

```go
output, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{
    Bucket: bucketName,
    Prefix: "{{user.object_prefix}}",
})

// Handle pagination
for output.IsTruncated {
    output, err = client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{
        Bucket:           bucketName,
        ListObjectsInput: tos.ListObjectsInput{Marker: output.NextMarker},
    })
}
```

---

### Operation: CopyObject — Copy Object Within/Across Buckets

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Source exists | `tosutil stat tos://{{user.src_bucket}}/{{user.src_key}}` | Object found | HALT; verify source |
| Destination bucket exists | `tosutil ls tos://{{user.dest_bucket}}` | Bucket found | HALT; create destination bucket |
| Permissions | Read access to source, write access to destination | Confirmed | HALT; check ACL/IAM |

#### Execution — tosutil CLI

```bash
# Copy object within same bucket
tosutil cp tos://{{user.bucket}}/{{user.src_key}} tos://{{user.bucket}}/{{user.dest_key}}

# Copy object across buckets
tosutil cp tos://{{user.src_bucket}}/{{user.src_key}} tos://{{user.dest_bucket}}/{{user.dest_key}}

# Copy with metadata preservation
tosutil cp tos://{{user.src_bucket}}/{{user.src_key}} tos://{{user.dest_bucket}}/{{user.dest_key}} --meta
```

#### Execution — Go SDK

```go
_, err := client.CopyObjectV2(context.Background(), &tos.CopyObjectV2Input{
    CopyBucket:           os.Getenv("SRC_BUCKET"),
    CopyKey:              os.Getenv("SRC_KEY"),
    Bucket:               os.Getenv("DEST_BUCKET"),
    Key:                  os.Getenv("DEST_KEY"),
})
if err != nil {
    log.Fatalf("Failed to copy object: %v", err)
}
fmt.Println("Object copied successfully")
```

#### Validation

```bash
# Verify destination object exists
tosutil stat tos://{{user.dest_bucket}}/{{user.dest_key}}

# Compare sizes
SRC_SIZE=$(tosutil stat tos://{{user.src_bucket}}/{{user.src_key}} | grep -i 'content-length' | awk '{print $2}')
DEST_SIZE=$(tosutil stat tos://{{user.dest_bucket}}/{{user.dest_key}} | grep -i 'content-length' | awk '{print $2}')
[ "$SRC_SIZE" = "$DEST_SIZE" ] && echo "COPY VERIFIED" || echo "COPY MISMATCH"
```

#### Failure Recovery

| Error Pattern | Action |
|--------------|--------|
| `NoSuchKey` | HALT → source missing |
| `NoSuchBucket` | HALT → source/dest bucket missing |
| `AccessDenied` | HALT → check read/write perms |
| `EntityTooLarge` | Use multipart upload + copy

---

### Operation: DeleteObject — Delete Object(s)

#### Pre-flight (Safety Gate)

- **MUST** confirm target: deleting `tos://{{user.bucket}}/{{user.object_key}}`
- **MUST** recommend backup if object is critical
- Pattern matching deletions: review the file list before `rm` for prefix patterns

#### Execution

```bash
# Delete a single object
tosutil rm tos://{{user.bucket}}/{{user.object_key}}

# Delete objects matching prefix (REVIEW LIST FIRST!)
tosutil ls tos://{{user.bucket}}/{{user.object_prefix}} -s
# After review:
tosutil rm tos://{{user.bucket}}/{{user.object_prefix}} -r
```

---

### Operation: PresignedURL — Generate Pre-signed URL

#### Execution

```bash
# Generate a pre-signed URL (default 1 hour)
tosutil presign tos://{{user.bucket}}/{{user.object_key}}

# Generate with custom expiration (seconds)
tosutil presign tos://{{user.bucket}}/{{user.object_key}} --expires 3600
```

#### Go SDK

```go
output, err := client.SignUrlHttpMethodV2(context.Background(), &tos.TrustedSignV2Input{
    Bucket: bucketName,
    Key:    objectKey,
    HTTPMethod: http.MethodGet,
    Expires:    3600, // seconds
})
fmt.Println("Signed URL:", output.SignedUrl)
```

---

### Operation: PutBucketLifecycle — Set Lifecycle Rules

#### Execution

```bash
# Set lifecycle rule via API (JSON payload)
ve tos PutBucketLifecycle \
  --bucket "{{user.bucket}}" \
  --body '{"Rules": [{"ID": "auto-delete", "Status": "Enabled", "Prefix": "logs/", "Expiration": {"Days": 30}}]}'
```

Go SDK:
```go
_, err := client.PutBucketLifecycle(context.Background(), &tos.PutBucketLifecycleInput{
    Bucket: bucketName,
    Rules: []tos.LifecycleRule{
        {
            ID:     "auto-delete",
            Status: enum.RuleStatusEnabled,
            Filter: tos.LifecycleFilter{Prefix: "logs/"},
            Expiration: &tos.LifecycleExpiration{
                Days: tos.Int(30),
            },
        },
    },
})
if err != nil {
    log.Fatalf("Failed to put bucket lifecycle: %v", err)
}
```

---

### Operation: PutBucketVersioning — Enable/disable Versioning

#### Execution

```bash
# Enable versioning
ve tos PutBucketVersioning --bucket "{{user.bucket}}" --Status Enabled

# Suspend versioning
ve tos PutBucketVersioning --bucket "{{user.bucket}}" --Status Suspended
```

---

### Operation: MultipartUpload — Upload Large Files

For files > 100MB, use multipart upload via `tosutil`:

```bash
# tosutil automatically uses multipart for large files
# Customize part size with -ps (e.g., 10MB per part)
tosutil cp "{{user.local_file}}" tos://{{user.bucket}}/{{user.object_key}} -ps=10mb

# Resume a failed upload
tosutil cp "{{user.local_file}}" tos://{{user.bucket}}/{{user.object_key}} --task-id <task-id>
```

## FinOps Operations (Advanced)

详细 FinOps Operation 步骤（DescribeStorageAnalysis / CleanupMultipartUploads / DetectStaleObjects / OptimizeStorageClass / DescribeCostSummary）见 [`references/advanced/finops.md`](references/advanced/finops.md)。

> SKILL.md 中仅保留 FinOps 入口；具体执行步骤在 `references/advanced/finops.md`（TE-7 专业内容分层）。

## Error Taxonomy

| 错误码 | 含义 | 分辨率 |
|--------|------|--------|
| `InvalidBucketName` | Bucket 名称不符合命名规范 | 0 retries; **HALT** — 3-63 字符，小写字母/数字/连字符 |
| `BucketAlreadyExists` | Bucket 名称已被占用 | 0 retries; **HALT** — 使用全局唯一的 bucket 名称 |
| `BucketNotFound` | 指定的 Bucket 不存在 | 0 retries; **HALT** — 确认 bucket 名称和区域 |
| `InvalidObjectName` | 对象键（Key）格式无效 | 0 retries; **HALT** — 使用 URL 安全字符，编码特殊字符 |
| `ObjectNotFound` | 指定的对象不存在 | 0 retries; **HALT** — 确认对象键和 bucket 名称 |
| `QuotaExceeded.Bucket` | Bucket 数量已达账户配额 | 0 retries; **HALT** — 删除未使用的 bucket 或提额 |
| `QuotaExceeded.Storage` | 存储容量已达配额上限 | 0 retries; **HALT** — 清理过期对象或提额 |
| `AccessDenied` | 访问被拒绝（IAM/ACL 权限不足） | 0 retries; **HALT** — 检查 IAM 策略和 bucket ACL |
| `InvalidCorsRule` | CORS 规则配置无效 | 0 retries; **HALT** — 修正 AllowedOrigin/AllowedMethod/AllowedHeader |
| `InvalidLifecycleRule` | 生命周期规则格式无效 | 0 retries; **HALT** — 检查 Days/Date/Prefix/Status 字段 |
| `InvalidPolicyDocument` | Bucket 策略文档格式错误 | 0 retries; **HALT** — 检查 JSON 语法和 IAM 策略规范 |
| `Throttling` | 请求限流 | 3 retries/exponential/1s/2s/4s; **RETRY** |
| `InternalError` | 服务端内部错误 | 3 retries/exponential/2s/4s/8s; **RETRY** — 记录 RequestId |

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
- [Enhanced Self-Healing Framework](../ve-skill-generator/references/enhanced-self-healing-framework.md)
- [FinOps Best Practices](../ve-skill-generator/references/finops-best-practices.md)
- [Knowledge Base](references/knowledge-base.md)
- [GCL Rubric](references/rubric.md) — Scoring dimensions for the Generator-Critic-Loop
- [GCL Prompt Templates](references/prompt-templates.md) — G/C/O prompt skeletons + TOS-specific safety prompts

## Operational Best Practices

- **Bucket naming:** lowercase, alphanumeric, hyphens; avoid dots; globally unique
- **Lifecycle rules:** auto-expire logs, temp files, and old versions
- **Versioning:** enable for production buckets to prevent accidental deletion
- **Access control:** use bucket policies and ACLs; prefer least privilege
- **Data integrity:** use `--enable-verify` flag on uploads/downloads
- **Cost optimization:** use IA/Archive storage classes for infrequently accessed data
- **Large files:** always use multipart upload for files > 100MB
