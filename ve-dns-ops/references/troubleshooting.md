# Troubleshooting — Volcengine DNS

## Diagnostic Order

> Follow this sequence for DNS troubleshooting:

1. **Verify credentials** — check `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` are set
2. **Check region** — verify `VOLCENGINE_REGION` is correct
3. **Describe the domain** — `ve dns DescribeDomain --Region ... --DomainName ...`
4. **List records** — `ve dns ListRecords --Region ... --DomainName ...`
5. **Test resolution** — `dig @8.8.8.8 example.com +short` or `nslookup example.com 8.8.8.8`
6. **Check statistics** — `ve dns DescribeDomainStatistics --Region ... --DomainName ...`
7. **Verify CLI** — `ve dns --help` to list available operations

## Common API Error Codes

| Error Code | HTTP | Meaning | Agent Action |
|------------|------|---------|--------------|
| `InvalidDomainName` | 400 | Domain name format invalid | Use FQDN format (e.g., `example.com`) |
| `DomainAlreadyExists` | 400 | Domain already in account | Use a different domain or check existing |
| `DomainNotFound` | 404 | Domain does not exist | Verify the domain name or create it first |
| `InvalidRecordType` | 400 | Record type not supported | Use A, AAAA, CNAME, MX, TXT, NS, SRV, or CAA |
| `InvalidRecordValue` | 400 | Record value format invalid | Check value format per record type |
| `DuplicateRecord` | 400 | Duplicate DNS record | Use UpdateRecord to modify existing |
| `RecordNotFound` | 404 | Record does not exist | Verify RecordId |
| `RecordLimitExceeded` | 400 | Record count limit reached | Delete unused records first |
| `QuotaExceeded` | 400 | Domain quota exceeded | Request quota increase from support |
| `InsufficientBalance` | 400 | Account balance insufficient | Recharge the account |
| `Unauthorized` | 403 | IAM permission denied | Check IAM policies for DNS actions |
| `Throttling` | 429 | Rate limit exceeded | Implement exponential backoff |
| `InternalError` | 500 | Server-side error | Retry with backoff; escalate if persists |

## Common DNS Issues

### Issue: Domain Not Resolving

```bash
# Step 1: Check if domain exists in your account
ve dns ListDomains --Region "cn-beijing" | jq '.Domains[] | select(.DomainName=="example.com")'

# Step 2: Verify records exist
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com"

# Step 3: Check resolution status
ve dns DescribeDNSResolution --Region "cn-beijing" --DomainName "example.com"

# Step 4: Test from public DNS
dig @8.8.8.8 example.com +short
```

| Possible Cause | Diagnostic | Fix |
|---------------|------------|-----|
| Domain not added to account | `ListDomains` shows nothing | `CreateDomain` first |
| No A/AAAA record | `ListRecords` empty | Add an A or AAAA record |
| Wrong nameservers | Domain registered with different NS | Update NS at registrar to Volcengine NS |
| DNS propagation delay | Changes < 24h old | Wait for TTL to expire |

### Issue: Record Update Not Taking Effect

```bash
# Step 1: Verify the record was updated
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RecordId=="r-xxxxx")'

# Step 2: Check TTL for expected propagation time
# Step 3: Clear local DNS cache
sudo dscacheutil -flushcache  # macOS
sudo systemd-resolve --flush-caches  # Linux
ipconfig /flushdns  # Windows
```

### Issue: CNAME Conflict

**⚠️ Problem:** CNAME records cannot coexist with other record types at the same name.

```bash
# Check for conflicting records
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RR=="www")'
```

**✅ Fix:** Delete conflicting records before adding the CNAME, or use a different hostname for the CNAME.

### Issue: Duplicate Record Error

```bash
# Check existing records
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RR=="www" and .Type=="A")'

# If exists, use UpdateRecord instead of AddRecord
ve dns UpdateRecord --Region "cn-beijing" --DomainName "example.com" --RecordId "r-xxxxx" --Value "10.0.0.1"
```

### Issue: Quota Exceeded

```bash
# Check current domain count
ve dns ListDomains --Region "cn-beijing" | jq '.Domains | length'

# Check record count per domain
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records | length'

# Solution: Delete unused domains/records or request quota increase
```

## CLI Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `command not found: ve` | CLI not installed | Download from GitHub releases |
| `AccessKey is empty` | Env vars not set | Set `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` |
| `Region is empty` | Env var not set | Set `VOLCENGINE_REGION` |
| `InvalidParameter` | Wrong parameter name | Check `ve dns <Action> --help` |
| `[ERROR]` with no message | Network issue | Check internet connectivity to `open.volcengineapi.com` |

## Escalation

If issues persist:

1. Capture `RequestId` from error response
2. Contact [Volcengine Support](https://console.volcengine.com/support)
3. Provide: RequestId, operation, parameters, and error code
