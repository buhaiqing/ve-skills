# CLI Usage — TOS

## Install and Config

### tosutil CLI (TOS-specific tool)

```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/tosutil/releases/latest/download/tosutil-linux-amd64 -o /usr/local/bin/tosutil
chmod +x /usr/local/bin/tosutil

# Initialize configuration
tosutil config
# Enter AK, SK, endpoint (https://tos-cn-beijing.volces.com), region (cn-beijing)
```

### TOS Credential Configuration

TOS credentials are configured via:
1. **Environment variables:** `TOS_ACCESS_KEY`, `TOS_SECRET_KEY`
2. **tosutil config file:** `~/.tosconfig` (created via `tosutil config`)
3. **ve CLI profile:** Shared credentials from `~/.volcengine/config.json`

## TOS CLI Command Map

### tosutil (Bulk Data Operations)

| Command | Example | Description |
|---------|---------|-------------|
| `mb` (make bucket) | `tosutil mb tos://my-bucket` | Create a bucket |
| `rb` (remove bucket) | `tosutil rb tos://my-bucket` | Delete a bucket (must be empty) |
| `ls` (list) | `tosutil ls tos://my-bucket` | List bucket contents |
| `ls` (recursive) | `tosutil ls tos://my-bucket/ -s` | Recursive list |
| `cp` (copy up) | `tosutil cp local.txt tos://bucket/key` | Upload file |
| `cp` (copy down) | `tosutil cp tos://bucket/key local.txt` | Download file |
| `cp` (recurse) | `tosutil cp ./dir/ tos://bucket/prefix/ -r` | Upload directory |
| `rm` (remove) | `tosutil rm tos://bucket/key` | Delete object |
| `rm` (recurse) | `tosutil rm tos://bucket/prefix/ -r` | Delete recursively |
| `stat` | `tosutil stat tos://bucket/key` | Object metadata |
| `presign` | `tosutil presign tos://bucket/key --expires 3600` | Pre-signed URL |
| `du` | `tosutil du tos://bucket/` | Bucket size/count |
| `probe` | `tosutil probe tos://bucket/` | Speed test |
| `version` | `tosutil version` | Tool version |

### ve tos (API Operations)

| API Action | Example `ve` Invocation |
|-----------|------------------------|
| ListBuckets | `ve tos ListBuckets --Region cn-beijing` |
| CreateBucket | `ve tos CreateBucket --bucket my-bucket --Region cn-beijing` |
| DeleteBucket | `ve tos DeleteBucket --bucket my-bucket --Region cn-beijing` |
| HeadBucket | `ve tos HeadBucket --bucket my-bucket --Region cn-beijing` |
| ListObjects | `ve tos ListObjects --bucket my-bucket --Region cn-beijing` |
| GetObject | `ve tos GetObject --bucket my-bucket --key mykey --Region cn-beijing` |
| PutObject | `ve tos PutObject --bucket my-bucket --key mykey --Region cn-beijing` |
| DeleteObject | `ve tos DeleteObject --bucket my-bucket --key mykey --Region cn-beijing` |
| CopyObject | `ve tos CopyObject --bucket my-bucket --key target --CopySource source-bucket/source-key --Region cn-beijing` |
| CreateMultipartUpload | `ve tos CreateMultipartUpload --bucket my-bucket --key bigfile --Region cn-beijing` |
| UploadPart | `ve tos UploadPart --bucket my-bucket --key bigfile --upload-id xxx --part-number 1 --Region cn-beijing` |
| CompleteMultipartUpload | `ve tos CompleteMultipartUpload --bucket my-bucket --key bigfile --upload-id xxx --Region cn-beijing` |
| AbortMultipartUpload | `ve tos AbortMultipartUpload --bucket my-bucket --key bigfile --upload-id xxx --Region cn-beijing` |
| ListMultipartUploads | `ve tos ListMultipartUploads --bucket my-bucket --Region cn-beijing` |
| PutBucketLifecycle | `ve tos PutBucketLifecycle --bucket my-bucket --body '...' --Region cn-beijing` |
| GetBucketLifecycle | `ve tos GetBucketLifecycle --bucket my-bucket --Region cn-beijing` |
| PutBucketVersioning | `ve tos PutBucketVersioning --bucket my-bucket --Status Enabled --Region cn-beijing` |
| GetBucketVersioning | `ve tos GetBucketVersioning --bucket my-bucket --Region cn-beijing` |
| PutBucketACL | `ve tos PutBucketACL --bucket my-bucket --ACL private --Region cn-beijing` |
