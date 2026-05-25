# Troubleshooting Guide — Volcengine NAT Gateway

> **Purpose:** Systematic troubleshooting guide for common NAT Gateway operational issues.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [NAT Gateway Creation Errors](#2-nat-gateway-creation-errors)
3. [SNAT Rule Errors](#3-snat-rule-errors)
4. [DNAT Rule Errors](#4-dnat-rule-errors)
5. [Network Connectivity Issues](#5-network-connectivity-issues)
6. [Rate Limiting](#6-rate-limiting)

---

## 1. Error Taxonomy

| Category | Code Pattern | HALT or Retry | Example |
|----------|-------------|---------------|---------|
| **Parameter Error** | `Invalid*.*` | HALT | `InvalidNatGatewayId.NotFound` |
| **Resource Error** | `*.NotFound` | HALT | `InvalidVpcId.NotFound` |
| **Conflict Error** | `*.Conflict` | HALT or fix | `SnatRule.Conflict` |
| **Status Error** | `IncorrectStatus.*` | HALT | `IncorrectStatus.NatGateway` |
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.NatGateway` |
| **Dependency Error** | `DependencyViolation` | HALT | `DependencyViolation.SnatRule` |
| **Billing Error** | `InsufficientBalance` | HALT | `InsufficientBalance` |
| **Rate Limit** | `Throttling` | Retry with backoff | `Throttling` |
| **Server Error** | `InternalError` | Retry with backoff | `InternalError` |

---

## 2. NAT Gateway Creation Errors

### QuotaExceeded.NatGateway

```
Error Code: QuotaExceeded.NatGateway
Message: The maximum number of NAT Gateways per VPC has been reached.
```

**Default:** 5 NAT Gateways per VPC.

**Resolution:**
```bash
ve nat Gateway DescribeNatGateways --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" | jq '.Result.TotalCount'
# Request quota increase if needed
```

### InvalidSubnetId.NotFound

```
Error Code: InvalidSubnetId.NotFound
Message: The specified SubnetId does not exist in the specified VPC.
```

**Resolution:**
```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" --SubnetIds "[\"$SUBNET_ID\"]"
```

---

## 3. SNAT Rule Errors

### SnatRule.Conflict

```
Error Code: SnatRule.Conflict
Message: An SNAT rule with the same SourceCidr already exists.
```

**Root Cause:** Cannot create two SNAT rules with the same source CIDR on the same NAT Gateway.

**Resolution:** Delete existing conflicting rule, or use a more specific CIDR.

### InvalidSourceCidr.Malformed

```
Error Code: InvalidSourceCidr.Malformed
Message: The specified SourceCidr format is invalid.
```

**Resolution:** Use valid CIDR notation within the VPC CIDR range (e.g., `10.0.2.0/24`).

---

## 4. DNAT Rule Errors

### DnatRule.PortConflict

```
Error Code: DnatRule.PortConflict
Message: A DNAT rule with the same EIP and ExternalPort already exists.
```

**Resolution:** Use a different external port, or delete the conflicting rule.

### InvalidInternalIp.Malformed

```
Error Code: InvalidInternalIp.Malformed
Message: The specified InternalIp is not valid.
```

**Resolution:** Ensure the internal IP is within the VPC subnet CIDR range.

---

## 5. Network Connectivity Issues

### Private Subnet Cannot Access Internet via SNAT

**Checklist:**

1. SNAT rule exists and is `Available`:
   ```bash
   ve nat Gateway DescribeSnatRules --Region "$VOLCENGINE_REGION" --NatGatewayId "$NAT_ID"
   ```

2. EIP is bound to NAT Gateway:
   ```bash
   ve nat Gateway DescribeNatGateways --Region "$VOLCENGINE_REGION" --NatGatewayIds "[\"$NAT_ID\"]" | jq '.Result.NatGateways[0].EipAddresses'
   ```

3. Route table has default route to NAT Gateway:
   ```bash
   ve vpc DescribeRouteTableList --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" | jq '.Result.RouteTables[].RouteEntries[] | select(.DestinationCidrBlock == "0.0.0.0/0")'
   ```

4. Security group allows outbound traffic (egress rules).

### DNAT Port Not Accessible from Internet

**Checklist:**

1. DNAT rule exists and is `Available`
2. EIP is bound to NAT Gateway and accessible from internet
3. Internal target server is running the service on InternalPort
4. Security group allows inbound traffic from NAT Gateway subnet
5. The internal IP is correct and the server is in the same VPC

---

## 6. Rate Limiting

### Throttling

Standard backoff retry: 2s, 4s, 8s. Max 3 retries.

---

*This reference document is part of the ve-nat-ops skill.*
