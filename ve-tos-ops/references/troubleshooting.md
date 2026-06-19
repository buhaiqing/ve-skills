# Troubleshooting — TOS

## Common Error Codes

| Error Code | Agent Action | Recovery |
|-----------|-------------|----------|
| `BucketAlreadyExists` | **HALT** — bucket name taken globally | Use a different bucket name |
| `NoSuchBucket` | **HALT** — bucket does not exist | Create bucket first |
| `NoSuchKey` | **HALT** — object does not exist | Verify object key |
| `AccessDenied` | Check IAM policy and bucket ACL | Verify permissions allow the action |
| `InvalidBucketName` | **HALT** — name format invalid | Use 3-63 chars, lowercase, alphanumeric+hyphens |
| `InvalidObjectName` | **HALT** — object key invalid | Fix key format |
| `TooManyBuckets` | **HALT** — bucket limit reached | Delete unused buckets or request increase |
| `InvalidPart` | Retry upload part | Retry the failed multipart part |
| `EntityTooLarge` | Use multipart upload | Object exceeds size limit for single PUT |
| `IncompleteBody` | Retry with full body | Request body incomplete |
| `InternalError` | Retry with backoff; **HALT** after 3 | Capture RequestId for escalation |
| `Throttling` | Exponential backoff | Max 3 retries; respect `Retry-After` |
| `SignatureDoesNotMatch` | **HALT** — credential/signature mismatch | Verify AK/SK and timestamp |
| `InvalidAccessKeyId` | **HALT** — access key not found | Verify AK is correct |
| `RequestTimeout` | Increase timeout; check network | Request timed out |
| `BucketNotEmpty` | **HALT** — bucket has objects | Delete objects first |
| `ObjectNotAppendable` | Use PUT or multipart | Cannot append to non-appendable object |

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
