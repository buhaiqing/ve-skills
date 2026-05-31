# Knowledge Base — DNS Fault Patterns and Diagnostics

## Fault Pattern 001: DNS Resolution Failure After Domain Creation

**Symptom:** Newly created domain does not resolve.

**Root Cause Analysis:**
1. Domain was created in Volcengine DNS but nameservers at registrar still point to old provider
2. No DNS records exist yet (empty zone)
3. Domain registration may have expired

**Diagnostic Steps:**
```bash
# 1. Verify domain exists in account
ve dns ListDomains --Region "cn-beijing" | jq '.Domains[] | select(.DomainName=="example.com")'

# 2. Check if records exist
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com"

# 3. Check DNS resolution status
ve dns DescribeDNSResolution --Region "cn-beijing" --DomainName "example.com"

# 4. Check current nameservers (from registrar)
dig NS example.com +short

# 5. Test resolution from a public resolver
dig @8.8.8.8 example.com +short
```

**Resolution Path:**
1. If no records exist: `AddRecord` for A/AAAA/CNAME as needed
2. If nameservers are wrong: Update nameservers at domain registrar to Volcengine NS
3. Wait for DNS propagation (up to 48 hours for NS changes)

---

## Fault Pattern 002: DNS Propagation Delay

**Symptom:** Updated DNS records not resolving globally after update.

**Root Cause Analysis:**
- TTL was high before the change
- Intermediate resolvers cached the old record
- Some regions update faster than others

**Diagnostic Steps:**
```bash
# Check current TTL on the record
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RecordId=="r-xxxxx") | {RR, Type, TTL}'

# Check what different resolvers see
dig @8.8.8.8 www.example.com +short    # Google DNS
dig @1.1.1.1 www.example.com +short    # Cloudflare DNS
dig @208.67.222.222 www.example.com +short  # OpenDNS
```

**Resolution Path:**
1. For future changes: lower TTL to 60-300s at least 24 hours before planned changes
2. Current fix: Wait out the remaining TTL or contact ISP support to clear their cache

**Prevention:**
- Always lower TTL before planned changes
- Use a TTL of 60s for change windows
- Restore to normal TTL (600-3600s) after changes stabilize

---

## Fault Pattern 003: Email Delivery Failure

**Symptom:** Emails to the domain are not being delivered or bounce.

**Root Cause Analysis:**
1. Missing or incorrect MX records
2. Missing SPF (TXT) records causing receivers to reject
3. Missing DKIM (TXT) records for signing
4. Missing or incorrect DMARC (TXT) policy

**Diagnostic Steps:**
```bash
# 1. Check MX records
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.Type=="MX")'

# 2. Check SPF, DKIM, DMARC
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.Type=="TXT")'

# 3. Validate DNS configuration from outside
dig MX example.com +short
dig TXT example.com +short
dig TXT _dmarc.example.com +short
```

**Resolution Path:**
```bash
# Add MX record (if missing)
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "@" --Type "MX" --Value "mail.example.com" --Priority 10

# Add SPF TXT record (allow only your mail server to send)
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "@" --Type "TXT" --Value "v=spf1 mx include:_spf.example.com ~all"

# Add DMARC policy
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "_dmarc" --Type "TXT" --Value "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"
```

---

## Fault Pattern 004: SSL/TLS Certificate Validation Failure

**Symptom:** Certificate Authority cannot validate domain ownership, or certificate issuance fails.

**Root Cause Analysis:**
1. CAA records block the intended CA
2. DNS validation TXT record is missing or incorrect
3. DNS changes haven't propagated before CA retry timeout

**Diagnostic Steps:**
```bash
# 1. Check CAA records
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.Type=="CAA")'

# 2. Check for validation TXT records
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RR=="_acme-challenge")'

# 3. Verify from outside
dig CAA example.com +short
dig TXT _acme-challenge.example.com +short
```

**Resolution Path:**
```bash
# Add CAA record to allow Let's Encrypt
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "@" --Type "CAA" --Value "0 issue "letsencrypt.org""

# Or check if CAA blocks everything — remove restrictive CAA
# (Use DeleteRecord if needed)
```

---

## Fault Pattern 005: Subdomain Delegation Not Working

**Symptom:** Subdomain (e.g., `sub.example.com`) does not resolve, but parent domain works.

**Root Cause Analysis:**
1. NS records for subdomain are missing or incorrect
2. Target nameservers are not authoritative for the subdomain
3. Glue records are missing (in-bailiwick delegation)

**Diagnostic Steps:**
```bash
# Check NS records for subdomain
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RR=="sub" and .Type=="NS")'

# Test delegation
dig NS sub.example.com +short
dig @ns1.external-dns.com sub.example.com +short
```

**Resolution Path:**
```bash
# Add NS record for subdomain delegation
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "sub" --Type "NS" --Value "ns1.external-dns.com" --TTL 86400
```

---

## Fault Pattern 006: API Throttling

**Symptom:** API calls returning HTTP 429 / `Throttling` errors during batch operations.

**Diagnostic Steps:**
- Check frequency of API calls
- Check if `Retry-After` header is present in 429 response

**Resolution Path:**
1. Implement exponential backoff in calling code
2. Reduce batch sizes for operations like `BatchImportRecords`
3. Space out requests with 100-500ms intervals
4. If throttling persists, request rate limit increase from support

---

## Fault Pattern 007: Unauthorized API Access

**Symptom:** API returns HTTP 403 / `Unauthorized` error.

**Root Cause Analysis:**
1. IAM policy does not grant DNS permissions
2. Access/Secret keys are incorrect or rotated
3. The IAM user does not have `dns:*` or specific action permissions

**Resolution Path:**
1. Verify credentials: `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"`
2. Check IAM policy includes `dns:*` actions
3. Regenerate keys if they may be compromised

---

## Quick Reference Card

| Symptom | Most Likely Cause | First Diagnostic Command |
|---------|------------------|------------------------|
| Domain not resolving | No records or wrong NS | `dig NS example.com +short` |
| Record update not appearing | High TTL / DNS cache | `ve dns ListRecords` |
| Email bouncing | Missing MX or SPF | `dig MX example.com +short` |
| SSL cert won't issue | CAA block or missing TXT | `dig CAA example.com +short` |
| API returns 403 | Wrong IAM permissions | Check access/secret keys |
| API returns 429 | Rate limited | Check Retry-After header |
| CLI "command not found" | `ve` not installed | `ve version` |
| CLI "AccessKey is empty" | Env vars not set | `echo $VOLCENGINE_ACCESS_KEY` |
