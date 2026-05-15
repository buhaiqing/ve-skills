# Troubleshooting — TOS

## Common Error Codes

| Error Code | HTTP Status | Meaning | Agent Action |
|-----------|-------------|---------|-------------|
| `BucketAlreadyExists` | 409 | Bucket name taken globally | HALT; use a different bucket name |
| `NoSuchBucket` | 404 | Bucket does not exist | HALT; create bucket first |
| `NoSuchKey` | 404 | Object does not exist | HALT; verify object key |
| `AccessDenied` | 403 | Insufficient permissions | Check IAM policy and bucket ACL |
| `InvalidBucketName` | 400 | Bucket name format invalid | Fix name: 3-63 chars, lowercase, alphanumeric+hyphens |
| `InvalidObjectName` | 400 | Object key is invalid | Fix key format |
| `TooManyBuckets` | 400 | Bucket limit reached | HALT; delete unused or request increase |
| `InvalidPart` | 400 | Invalid part for multipart upload | Retry upload part |
| `EntityTooLarge` | 400 | Object exceeds size limit | Use multipart upload |
| `IncompleteBody` | 400 | Request body incomplete | Retry with full body |
| `InternalError` | 500 | Server-side error | Retry with backoff; HALT after 3 |
| `Throttling` | 429 / 503 | Rate limit exceeded | Exponential backoff |
| `SignatureDoesNotMatch` | 403 | Credential/signature mismatch | Verify AK/SK and timestamp |
| `InvalidAccessKeyId` | 403 | Access key not found | Verify AK is correct |
| `RequestTimeout` | 400 | Request timed out | Increase timeout; check network |
| `MalformedXML` | 400 | XML body is malformed | Fix request body |
| `BucketNotEmpty` | 409 | Bucket has objects | Delete objects first |
| `ObjectNotAppendable` | 400 | Cannot append to non-appendable object | Use PUT or multipart |

## Diagnostic Order

1. **Check credentials:**
   ```bash
   test -n "$TOS_ACCESS_KEY" && test -n "$TOS_SECRET_KEY" && echo "OK"
   ```

2. **Verify endpoint:**
   ```bash
   # Correct format: https://tos-{region}.volces.com
   echo "https://tos-{{env.VOLCENGINE_REGION}}.volces.com"
   ```

3. **List buckets:**
   ```bash
   tosutil ls -s
   ```

4. **Check object exists:**
   ```bash
   tosutil stat tos://{{user.bucket}}/{{user.object_key}}
   ```

5. **Verify ACL:**
   ```bash
   ve tos GetBucketACL --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"
   ```

## Common Scenarios

### Scenario 1: Domain Error — `server returned an invalid body`

**Cause:** Using S3 protocol domain (`tos-s3-{region}.volces.com`) instead of TOS protocol domain.

**Fix:** Use `https://tos-{region}.volces.com` format.

### Scenario 2: Slow Upload / Progress bar rollback

**Cause:** Slow or unstable network triggering retries.

**Fix:** Reduce part size: `tosutil cp local.file tos://bucket/key -ps=5mb`

### Scenario 3: `use of closed network connection`

**Cause:** Network instability or client bandwidth saturated.

**Fix:** Reduce concurrency (`-p` and `-j` parameters in tosutil config).
