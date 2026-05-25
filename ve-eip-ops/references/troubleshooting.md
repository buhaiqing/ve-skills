# Troubleshooting Guide — Volcengine EIP

> **Purpose:** Systematic troubleshooting guide for common EIP operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [Allocation Errors](#2-allocation-errors)
3. [Binding Errors](#3-binding-errors)
4. [Bandwidth Errors](#4-bandwidth-errors)
5. [Lifecycle Errors](#5-lifecycle-errors)
6. [Rate Limiting](#6-rate-limiting)
7. [Debugging Strategies](#7-debugging-strategies)

---

## 1. Error Taxonomy

| Category | Code Pattern | HALT or Retry | Example |
|----------|-------------|---------------|---------|
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.EipAddress` |
| **Resource Error** | `*.NotFound` | HALT | `InvalidAllocationId.NotFound` |
| **Status Error** | `IncorrectStatus.*` | HALT | `IncorrectStatus.EipAddress` |
| **Conflict Error** | `AlreadyAssociated` | HALT or rebind | `AlreadyAssociated.Instance` |
| **Parameter Error** | `Invalid*.*` | HALT | `InvalidBandwidth.Malformed` |
| **IAM Error** | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| **Billing Error** | `InsufficientBalance` | HALT | `InsufficientBalance` |
| **Rate Limit** | `Throttling` | Retry with backoff | `Throttling` |
| **Server Error** | `InternalError` | Retry with backoff | `InternalError` |
| **Service Down** | `ServiceUnavailable` | Retry, then HALT | `ServiceUnavailable` |

---

## 2. Allocation Errors

### QuotaExceeded.EipAddress

```
Error Code: QuotaExceeded.EipAddress
Message: The maximum number of EIPs per region has been reached.
```

**Resolution:**
```bash
# Check current usage
ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" | jq '.Result.TotalCount'
# Request quota increase via Volcengine console
```

### InvalidLineType.NotSupported

```
Error Code: InvalidLineType.NotSupported
Message: The specified LineType is not supported.
```

**Resolution:** Use `BGP` — the only supported line type in most regions.

### InsufficientBalance

```
Error Code: InsufficientBalance
Message: Account balance is insufficient to allocate EIP.
```

**Resolution:** Recharge account via Volcengine console → Billing Management.

---

## 3. Binding Errors

### AlreadyAssociated.Instance

```
Error Code: AlreadyAssociated.Instance
Message: The specified instance already has an EIP associated.
```

**Root Cause:** Target resource (ECS/CLB/NAT) already has an EIP bound.

**Resolution:**
```bash
# Check current EIP bindings for instance
ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.EipAddresses[] | select(.InstanceId == "'"$INSTANCE_ID"'") | .AllocationId'

# Option 1: Disassociate existing EIP first
ve eip DisassociateEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EXISTING_EIP_ID"

# Option 2: Bind to a different instance
```

### IncorrectStatus.EipAddress

```
Error Code: IncorrectStatus.EipAddress
Message: The EIP is in status InUse and cannot be associated.
```

**Root Cause:** EIP is already bound. Must disassociate first.

**Resolution:**
```bash
ve eip DisassociateEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID"
```

### InvalidInstanceType.NotSupported

```
Error Code: InvalidInstanceType.NotSupported
Message: The specified InstanceType is not supported for EIP binding.
```

**Resolution:** Use valid types: `EcsInstance`, `ClbInstance`, `Nat`, `HaVip`, `NetworkInterface`.

### InvalidAllocationId.NotFound

```
Error Code: InvalidAllocationId.NotFound
Message: The specified AllocationId does not exist.
```

**Resolution:**
```bash
ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" --AllocationIds "[\"$EIP_ID\"]"
```

---

## 4. Bandwidth Errors

### InvalidBandwidth.ValueNotSupported

```
Error Code: InvalidBandwidth.ValueNotSupported
Message: The specified bandwidth value is not supported.
```

**Resolution:**
- PayByTraffic: 1–200 Mbps
- PayByBandwidth: 1–500 Mbps

---

## 5. Lifecycle Errors

### IncorrectStatus.EipAddress (Release while InUse)

```
Error Code: IncorrectStatus.EipAddress
Message: The EIP is in status InUse and cannot be released.
```

**Resolution:**
```bash
# Disassociate first
ve eip DisassociateEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID"
# Then release
ve eip ReleaseEipAddress --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID"
```

---

## 6. Rate Limiting

### Throttling

```
Error Code: Throttling
Message: Request was denied due to request throttling.
```

**Resolution:** Retry with exponential backoff (2s, 4s, 8s). Max 3 retries.

---

## 7. Debugging Strategies

### Capture RequestId

```bash
ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" 2>&1 \
  | jq -r '.ResponseMetadata.RequestId'
```

### Verify EIP Binding Chain

```bash
# Verify: EIP → Instance connectivity
EIP_ID="eipalloc-xxx"
ve eip DescribeEipAddressAttributes --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" \
  | jq '{
    eip_address: .Result.EipAddress,
    status: .Result.Status,
    bound_to: .Result.InstanceType,
    target_id: .Result.InstanceId,
    bandwidth: .Result.Bandwidth
  }'
```

### Verify Network Connectivity After Binding

```bash
# After binding EIP to ECS, test from external
ping -c 3 {{output.eip_address}}
curl -s --connect-timeout 5 "http://{{output.eip_address}}" || echo "Connection refused (expected if no service running)"
```

---

*This reference document is part of the ve-eip-ops skill.*
