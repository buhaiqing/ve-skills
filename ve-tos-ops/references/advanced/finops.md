# FinOps — TOS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7.

## Billing Model

| Tier | Standard | Infrequent Access (IA) | Archive | Cold Archive |
|------|----------|------------------------|---------|--------------|
| Storage | Per GB/month | ~60% of Standard | ~20% of Standard | ~10% of Standard |
| Request | Per 10K requests | Per 10K + retrieval fee | Per 10K + retrieval fee | Per 10K + retrieval fee |
| Min bill | — | 30d min storage | 90d min storage | 180d min storage |

## Cost Optimization

| Situation | Action | 💰 Savings |
|-----------|--------|---------|
| Data accessed <1x/month | Lifecycle rule: Standard → IA | ~60% storage cost |
| Data not accessed >90d | Lifecycle rule: IA → Archive | ~80% on stored objects |
| Orphaned multipart uploads | Lifecycle rule: abort incomplete >7d | Recovers orphaned storage |
| Cross-region replication | Disable if not needed | 100% of replication cost |
| Public-read objects | Use pre-signed URL instead | Avoids bandwidth surcharge |

## Cost Anomaly Detection

| Sign | Cause | Investigation |
|------|-------|-------------|
| ⚠️ Storage spike >20% in 1d | Unplanned upload, data sync | `ve tos ListObjects --bucket {{output.BucketName}}` |
| ⚠️ Request count surge | Misconfigured app, DDoS | Check access logs, enable CDN |
| ⚠️ Cross-region replication cost | Replication rules enabled | `ve tos GetBucketReplication --bucket {{output.BucketName}}` |
| ⚠️ Unexpected storage tier cost | Lifecycle rule not applied | Verify lifecycle policy exists |

## Query Pricing

```bash
# TOS pricing is per-GB tiered by region — use Billing API for cost analysis
ve billing DescribeBillSummaryByMonth --body '{"BillingPeriod":"{{env.BILLING_MONTH}}"}'
# Check bucket location (pricing varies by region)
ve tos GetBucketLocation --bucket {{output.BucketName}}
```

## Operations

### Operation: DescribeStorageAnalysis — Analyze Storage Class Distribution

Analyzes bucket storage distribution across storage classes to identify optimization opportunities.

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Bucket exists | `tosutil ls tos://{{user.bucket}}` | Bucket found | HALT |
| Credentials | `test -n "$TOS_ACCESS_KEY" && test -n "$TOS_SECRET_KEY"` | Both set | HALT |

#### Execution

```bash
# Get bucket storage class distribution
tosutil ls tos://{{user.bucket}} -s -ab | awk '{print $2, $3}' | sort | uniq -c | sort -rn

# Get bucket size summary
tosutil du tos://{{user.bucket}}
```

#### Analysis Logic

| Storage Class | Recommended Use | Cost Relative to Standard |
|--------------|-----------------|--------------------------|
| Standard | Frequent access (daily) | 100% (baseline) |
| IA (Infrequent Access) | Occasional access (monthly) | ~60% |
| Archive | Rare access (quarterly, restore needed) | ~40% |
| ColdArchive | Compliance retention (yearly) | ~20% |

---

### Operation: DetectStaleObjects — Find Objects Not Accessed Recently

Identifies objects not accessed for a specified period.

#### Stale Classification

| Last Access | Classification | Recommended Action |
|-------------|---------------|-------------------|
| > 30 days | Warm | Consider IA storage class |
| > 90 days | Cold | Consider Archive storage class |
| > 365 days | Frozen | Consider deletion or ColdArchive |

#### Execution

```bash
# List objects with last modified date
tosutil ls tos://{{user.bucket}} -s -ab
```

---

### Operation: CleanupMultipartUploads — Abort Incomplete Uploads

Finds and aborts multipart uploads incomplete beyond a threshold.

#### Execution

```bash
# List incomplete multipart uploads
ve tos ListMultipartUploads --bucket "{{user.bucket}}" --Region "{{env.VOLCENGINE_REGION}}"

# Abort a specific upload
ve tos AbortMultipartUpload --bucket "{{user.bucket}}" --key "{{user.object_key}}" --upload-id "{{user.upload_id}}" --Region "{{env.VOLCENGINE_REGION}}"
```

---

### Operation: OptimizeStorageClass — Apply Storage Class Transitions

Transitions objects to a more cost-effective storage class.

#### Pre-flight (Safety Gate)

- **MUST** list all objects to be transitioned with current and target class
- **MUST** warn about retrieval costs and restore times for Archive/ColdArchive
- **MUST** confirm with user before proceeding

#### Execution

```bash
# Set lifecycle rule for automatic transition
ve tos PutBucketLifecycle \
  --bucket "{{user.bucket}}" \
  --body '{"Rules": [{"ID": "auto-transition-to-ia", "Status": "Enabled", "Prefix": "logs/", "Transitions": [{"Days": 30, "StorageClass": "IA"}, {"Days": 90, "StorageClass": "Archive"}]}]}'
```

---

### Operation: DescribeCostSummary — Generate TOS Cost Report

Generates a cost summary for all TOS buckets.

#### Execution

```bash
# List all buckets with sizes
tosutil ls -s

# Query billing data for TOS
ve billing DescribeBillDetail --BillingCycle "{{user.billing_cycle}}" --ProductType tos
```

#### Output Format

```markdown
## TOS Cost Summary — {{user.billing_cycle}}

| Bucket | Storage | Requests | Bandwidth | Monthly Cost |
|--------|---------|----------|-----------|-------------|
| prod-assets | 500 GB (Standard) | 2M | 100 GB | ¥280 |
| prod-logs | 200 GB (IA) | 500K | 10 GB | ¥65 |
| **Total** | **700 GB** | **2.5M** | **110 GB** | **¥345** |
```

---

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md)
