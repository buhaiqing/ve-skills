# Knowledge Base — TOS Fault Patterns

## Pattern 1: AccessDenied Flood

### Symptoms
- Sudden spike in 403 AccessDenied errors
- Error rate > 5% of total requests
- Multiple source IPs affected
- Bucket ACL or policy recently changed

### Root Cause
- Bucket ACL changed from `public-read` to `private`
- IAM policy revoked for application role
- Pre-signed URLs expired
- IP-based bucket policy blocking legitimate sources

### Resolution
1. Check recent ACL changes: `ve tos GetBucketACL --bucket {{user.bucket}}`
2. Check bucket policy: `ve tos GetBucketPolicy --bucket {{user.bucket}}`
3. If ACL change was accidental: restore previous ACL
4. If IAM policy issue: re-attach `TOSReadOnlyAccess` or custom policy
5. If pre-signed URLs expired: regenerate with longer expiration

### Prevention
- Monitor 403 error rate with alert at > 1%
- Require approval for ACL/policy changes
- Use lifecycle rules instead of manual ACL changes

---

## Pattern 2: Storage Growth Anomaly

### Symptoms
- StorageUsed growth > 50% in 1 hour
- StorageUsed growth > 100 GB in 24 hours
- ObjectCount growth rate significantly above baseline

### Root Cause
- Application writing logs to TOS instead of log service
- Backup job misconfigured (full backup instead of incremental)
- Infinite loop in upload process
- Compromessed credentials being abused

### Resolution
1. Identify the source: Check `RequestCount` by prefix/IP
2. If logs: redirect to proper log service, delete TOS logs
3. If backup: verify backup configuration, switch to incremental
4. If abuse: rotate credentials immediately, review bucket policy
5. Set up lifecycle rule to auto-expire unexpected data

### Prevention
- Set storage growth rate alert (> 20% per day)
- Use bucket quotas to cap maximum storage
- Separate buckets by purpose (logs, backups, assets)
- Enable versioning with lifecycle to auto-expire

---

## Pattern 3: High Latency on GET Requests

### Symptoms
- FirstByteLatency > 5000ms for GET requests
- Intermittent timeouts on object downloads
- Latency spikes correlated with bandwidth usage

### Root Cause
- Bandwidth saturation (bucket approaching egress limit)
- Hot object pattern (single object requested by many clients)
- Cross-region access without acceleration
- Client-side network issues

### Resolution
1. Check bandwidth usage: `ve cms DescribeMetricData --Namespace Volcengine_TOS --MetricName BandwidthOut`
2. If bandwidth saturated: enable CDN acceleration for hot objects
3. If cross-region: use VPC endpoint or transfer acceleration
4. If client-side: test from different network location

### Prevention
- Use CDN for frequently accessed objects
- Enable transfer acceleration for cross-region access
- Monitor bandwidth with alert at 80% of limit
- Implement client-side caching

---

## Pattern 4: Multipart Upload Stuck

### Symptoms
- Large upload hangs at 99% or specific part
- `NetworkError` or `RequestTimeout` during upload
- Upload speed drops to near zero

### Root Cause
- Network instability causing part upload failures
- Part size too large for network conditions
- Concurrent upload limit reached
- Temporary server-side issue

### Resolution
1. Reduce part size: `tosutil cp file tos://bucket/key -ps=5mb`
2. Resume failed upload: `tosutil cp file tos://bucket/key --task-id <id>`
3. Reduce concurrency: configure `-p 2 -j 2` in tosutil
4. If persistent: try from different network location

### Prevention
- Use adaptive part sizing based on network speed
- Enable automatic retry in tosutil config
- Monitor upload completion rate

---

## Cascade Pattern: Storage Growth → Cost Spike → Budget Alert Storm

### Trigger Event
- Application bug writing excessive data to TOS
- Backup job duplication creating N copies
- Missing lifecycle rules allowing data accumulation

### Propagation Path
- A (Data surge) → B (Storage cost increases) → C (Budget alert fires) → D (Multiple bucket alerts) → E (Finance team alerted) → F (Panic investigation)

### Breaking the Chain
- **Primary break point:** Set bucket-level storage quotas
- **Secondary break point:** Configure lifecycle rules to auto-expire
- **Tertiary break point:** Set storage growth rate alerts

### Resolution
1. Stop the bleeding: Identify and stop the source of data surge
2. Clean up: Delete unintended data (after verification)
3. Add controls: Set bucket quota, lifecycle rules, growth alerts
4. Review: Ensure all buckets have appropriate lifecycle policies
