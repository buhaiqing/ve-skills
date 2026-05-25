# Design: Fix HIGH Issues in P0 Cloud Resource Skills

**Date:** 2026-05-26
**Author:** Claude Code
**Status:** Approved

---

## Problem Statement

Code review identified 3 HIGH severity issues in P0 cloud resource skills that must be fixed before merge:

1. **ve-nat-ops CLI syntax inconsistency** — Commands use `ve nat Gateway DescribeNatGateways` but should use `ve natgateway DescribeNatGateways`
2. **ve-clb-ops HTTPS listener missing CertificateId** — HTTPS listener example lacks required certificate parameter
3. **ve-eip-ops/ve-vpc-ops pre-flight shell syntax** — Linux-specific `test -n "$VAR"` in documentation

---

## Design Overview

**Approach:** Minimal targeted fixes — address only the 3 HIGH issues with minimal, verifiable changes.

**Scope:** 3 fixes affecting 8 files total.

---

## Fix 1: NAT Gateway CLI Syntax

### Problem

CLI commands in ve-nat-ops use `ve nat Gateway DescribeNatGateways` format. This is inconsistent with the standard `ve <service> <action>` pattern used by other services:
- VPC: `ve vpc DescribeVpcs`
- EIP: `ve eip DescribeEipAddresses`
- CLB: `ve clb DescribeLoadBalancers`

### Solution

Replace all `ve nat Gateway` with `ve natgateway` throughout ve-nat-ops skill.

### Files Affected

| File | Changes |
|------|---------|
| `ve-nat-ops/SKILL.md` | ~15 command references + metadata update |
| `ve-nat-ops/references/cli-usage.md` | ~20 command references |
| `ve-nat-ops/references/api-sdk-usage.md` | Service name description |

### Specific Changes

1. **SKILL.md metadata (line 24):**
   ```yaml
   cli_support_evidence: >-
     Confirmed via `ve natgateway --help` — NAT Gateway is supported by the ve CLI.
   ```

2. **SKILL.md examples:** Replace all occurrences:
   - `ve nat Gateway DescribeNatGateways` → `ve natgateway DescribeNatGateways`
   - `ve nat Gateway CreateNatGateway` → `ve natgateway CreateNatGateway`
   - `ve nat Gateway CreateSnatRule` → `ve natgateway CreateSnatRule`
   - `ve nat Gateway CreateDnatRule` → `ve natgateway CreateDnatRule`
   - `ve nat Gateway DeleteNatGateway` → `ve natgateway DeleteNatGateway`
   - `ve nat Gateway DescribeSnatRules` → `ve natgateway DescribeSnatRules`
   - `ve nat Gateway DescribeDnatRules` → `ve natgateway DescribeDnatRules`

3. **cli-usage.md:** Same replacements throughout

4. **api-sdk-usage.md:** Update service name description (line 23)

### Verification

- Count replacements: should be ~35 occurrences
- No changes to API endpoint (service remains `nat` internally)

---

## Fix 2: CLB HTTPS Listener Certificate

### Problem

HTTPS listener creation example in ve-clb-ops lacks `CertificateId` parameter, which is required for HTTPS listeners. Users following the example will receive InvalidParameter errors.

### Solution

Add `CertificateId` parameter to HTTPS listener examples and document the prerequisite.

### Files Affected

| File | Changes |
|------|---------|
| `ve-clb-ops/SKILL.md` | HTTPS example (line ~220) + Variable Convention table |
| `ve-clb-ops/references/cli-usage.md` | HTTPS example (lines 103-110) |

### Specific Changes

1. **SKILL.md Variable Convention table:** Add new placeholder:
   ```
   | `{{user.certificate_id}}` | SSL certificate ID | Format `cert-xxxxxxxxx` |
   ```

2. **SKILL.md HTTPS listener example:**
   ```bash
   ve clb CreateListener \
     --Region "{{user.region}}" \
     --LoadBalancerId "{{user.clb_id}}" \
     --Protocol "HTTPS" \
     --Port 443 \
     --ListenerName "https-listener" \
     --CertificateId "{{user.certificate_id}}"
   ```

3. **cli-usage.md HTTPS example:** Same addition

4. **Add prerequisite note in both files:**
   ```
   > **Prerequisite:** HTTPS listeners require an SSL certificate. Obtain certificate ID from Volcengine SSL Certificate Management before creating HTTPS listeners.
   ```

### Verification

- Check CertificateId appears in both HTTPS examples
- Check new placeholder in Variable Convention table
- Check prerequisite note exists

---

## Fix 3: Pre-flight Shell Syntax

### Problem

Pre-flight check tables use Linux-specific shell syntax:
```
`test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"`
```

This is platform-specific and may confuse agents or users on different systems.

### Solution

Replace with generic, platform-agnostic description.

### Files Affected

| File | Changes |
|------|---------|
| `ve-eip-ops/SKILL.md` | Pre-flight table (line 175) |
| `ve-vpc-ops/SKILL.md` | Pre-flight table (line 176) |
| `ve-nat-ops/SKILL.md` | Pre-flight table (line 177) |
| `ve-clb-ops/SKILL.md` | No shell syntax found — skip |

### Specific Changes

Replace shell syntax with:
```
| Credentials | Check VOLCENGINE_ACCESS_KEY and VOLCENGINE_SECRET_KEY are set | Both non-empty | HALT |
```

### Verification

- Search for `test -n` pattern across all 4 skills
- Replace all occurrences
- Ensure HALT action preserved

---

## Implementation Order

1. **Fix 1 (NAT CLI)** — Most impactful, affects command correctness
2. **Fix 2 (CLB HTTPS)** — Prevents user errors
3. **Fix 3 (Pre-flight)** — Documentation clarity

---

## Success Criteria

- All `ve nat Gateway` replaced with `ve natgateway`
- HTTPS listener examples include CertificateId
- No shell-specific syntax in pre-flight tables
- Git diff shows only targeted changes
- No new lines added beyond specified fixes

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Missed occurrences | Use `grep -r` to find all patterns |
| Breaking existing patterns | Verify changes match existing CLI pattern |
| Incomplete CertificateId documentation | Add both example and prerequisite note |

---

## Post-Fix Verification

```bash
# Verify NAT CLI fix
grep -r "ve nat Gateway" ve-nat-ops/  # Should return nothing
grep -r "ve natgateway" ve-nat-ops/   # Should show ~35 matches

# Verify CLB HTTPS fix
grep -n "CertificateId" ve-clb-ops/SKILL.md ve-clb-ops/references/cli-usage.md

# Verify pre-flight fix
grep -r "test -n" ve-*/  # Should return nothing
```

---

## References

- [Code Review Report](../code-review-report.md)
- [Volcengine CLI Pattern](https://github.com/volcengine/volcengine-cli) — `ve <service> <action>`