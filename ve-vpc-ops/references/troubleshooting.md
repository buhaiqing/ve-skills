# Troubleshooting Guide — Volcengine VPC

> **Purpose:** Systematic troubleshooting guide for common VPC operational issues. Organized by error category with root cause analysis and recovery procedures.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Error Taxonomy](#1-error-taxonomy)
2. [Credential Issues](#2-credential-issues)
3. [Parameter Validation Errors](#3-parameter-validation-errors)
4. [Resource Lifecycle Errors](#4-resource-lifecycle-errors)
5. [Quota and Limit Errors](#5-quota-and-limit-errors)
6. [Network and Availability Errors](#6-network-and-availability-errors)
7. [Concurrency and Rate Limiting](#7-concurrency-and-rate-limiting)
8. [Debugging Strategies](#8-debugging-strategies)

---

## 1. Error Taxonomy

### Error Response Structure

All VPC API errors follow this structure:

```json
{
  "ResponseMetadata": {
    "RequestId": "...",
    "Action": "CreateVpc",
    "Version": "2020-04-01",
    "Service": "vpc",
    "Region": "cn-beijing",
    "Error": {
      "Code": "InvalidCidrBlock.Malformed",
      "Message": "The specified CidrBlock format is invalid."
    }
  },
  "Result": null
}
```

### Error Categories

| Category | Code Pattern | HALT or Retry | Example |
|----------|-------------|---------------|---------|
| **Input Error** | `Invalid*.*` | HALT | `InvalidCidrBlock.Malformed` |
| **Resource Error** | `*.NotFound`, `*.DoesNotExist` | HALT | `InvalidVpcId.NotFound` |
| **Conflict Error** | `*.Conflict`, `ResourceAlreadyExists` | HALT or fix | `CidrBlock.Conflict` |
| **Status Error** | `IncorrectStatus.*` | HALT, then retry after status change | `IncorrectStatus.Subnet` |
| **Quota Error** | `QuotaExceeded.*` | HALT | `QuotaExceeded.Vpc` |
| **IAM Error** | `Forbidden.RAM` | HALT | `Forbidden.RAM` |
| **Rate Limit** | `Throttling` | Retry with backoff | `Throttling` |
| **Server Error** | `InternalError` | Retry with backoff | `InternalError` |
| **Service Down** | `ServiceUnavailable` | Retry, then HALT | `ServiceUnavailable` |

---

## 2. Credential Issues

### Symptom: Authentication Failure

```
Error Code: Unauthorized
Message: The Access Key ID or Secret Access Key you provided is invalid.
```

**Root Causes:**
- Incorrect `VOLCENGINE_ACCESS_KEY` or `VOLCENGINE_SECRET_KEY`
- Credentials expired or revoked
- IAM user lacks VPC permissions

**Resolution:**

```bash
# Verify credential existence (values are never printed)
test -n "$VOLCENGINE_ACCESS_KEY" && echo "✅ AK is set" || echo "❌ AK is NOT set"
test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ SK is set" || echo "❌ SK is NOT set"

# Test credentials with a read-only API call
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}"
```

**Resolution Steps:**
1. Regenerate AK/SK from Volcengine console if revoked
2. Verify IAM policy includes `vpc:DescribeVpcs` and related permissions
3. Re-configure: `ve configure set --profile default ...`

---

## 3. Parameter Validation Errors

### InvalidCidrBlock.Malformed

```
Error Code: InvalidCidrBlock.Malformed
Message: The specified CidrBlock format is invalid.
```

**Root Cause:** CIDR format does not match valid VPC or subnet CIDR patterns.

**Valid Ranges:**
- VPC: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` (and subnets)
- Subnet: `/16` to `/29` mask length

**Resolution:**
```bash
# Correct examples:
ve vpc CreateVpc --Region cn-beijing --CidrBlock "10.0.0.0/16" --VpcName "test"
ve vpc CreateSubnet --Region cn-beijing --VpcId "vpc-xxx" --CidrBlock "10.0.1.0/24"
```

### CidrBlock.Conflict

```
Error Code: CidrBlock.Conflict
Message: The specified CidrBlock overlaps with an existing CIDR.
```

**Root Cause:** The requested CIDR overlaps with an existing VPC CIDR, or a subnet CIDR overlaps with another subnet in the same VPC.

**Resolution:**
```bash
# Check existing VPC CIDRs
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}" | jq '.Result.Vpcs[] | {VpcName, CidrBlock}'

# Choose a non-overlapping CIDR
```

### InvalidVpcId.NotFound

```
Error Code: InvalidVpcId.NotFound
Message: The specified VpcId does not exist.
```

**Root Cause:** The VPC ID is incorrect or the VPC has been deleted.

**Resolution:**
```bash
# Verify VPC exists
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}" --VpcIds "[\"{{user.vpc_id}}\"]"
```

### InvalidRegion.NotFound

```
Error Code: InvalidRegion.NotFound
Message: The specified Region does not exist.
```

**Resolution:**
```bash
# List valid regions
ve vpc DescribeRegions

# Common valid regions: cn-beijing, cn-shanghai, cn-guangzhou, cn-hongkong
```

---

## 4. Resource Lifecycle Errors

### ResourceInUse

```
Error Code: ResourceInUse
Message: The {{resource}} is in use and cannot be deleted.
```

**Root Cause:** The resource has dependent resources attached.

**Common Scenarios:**

| Resource | Must Be Removed First |
|----------|----------------------|
| VPC | All subnets, route tables, instances, NAT gateways, CLBs |
| Subnet | All ECS instances, ENIs, load balancers |
| Route Table | All custom route entries, subnet associations |

**Resolution:**
```bash
# Before deleting a VPC, clean up dependencies:

# 1. List subnets
ve vpc DescribeSubnets --Region cn-beijing --VpcId "vpc-xxx" | jq -r '.Result.Subnets[].SubnetId'

# 2. Delete each subnet (must be empty)
for SUBNET_ID in $(...); do
  ve vpc DeleteSubnet --Region cn-beijing --SubnetId "$SUBNET_ID"
done

# 3. List and delete custom route tables
for RTB_ID in $(...); do
  ve vpc DeleteRouteTable --Region cn-beijing --RouteTableId "$RTB_ID"
done

# 4. Delete the VPC
ve vpc DeleteVpc --Region cn-beijing --VpcId "vpc-xxx"
```

---

## 5. Quota and Limit Errors

### QuotaExceeded.Vpc

```
Error Code: QuotaExceeded.Vpc
Message: The maximum number of VPCs per region has been reached.
```

**Default Quotas:**

| Resource | Default Limit | Per |
|----------|--------------|-----|
| VPCs | 10 | Region |
| Subnets per VPC | 200 | VPC |
| Route Tables per VPC | 50 | VPC |
| Route Entries per Route Table | 100 | Route Table |

**Resolution:**
```bash
# Check current usage
ve vpc DescribeVpcs --Region cn-beijing | jq '.Result.TotalCount'

# Request quota increase through Volcengine console:
# https://console.volcengine.com/ticket
```

---

## 6. Network and Availability Errors

### IncorrectStatus.Subnet

```
Error Code: IncorrectStatus.Subnet
Message: The subnet is in status Pending and cannot be modified.
```

**Root Cause:** Subnet is still being provisioned.

**Resolution:**
```bash
# Poll until status is Available
for i in {1..30}; do
  STATUS=$(ve vpc DescribeSubnetAttributes --Region "$VOLCENGINE_REGION" --SubnetId "$SUBNET_ID" \
    | jq -r '.Result.Status')
  [ "$STATUS" = "Available" ] && echo "Subnet is ready" && break
  echo "Waiting... (attempt $i, status: $STATUS)"
  sleep 2
done
```

### DependencyViolation

```
Error Code: DependencyViolation
Message: The {{resource}} has dependent resources.
```

**Root Cause:** Similar to ResourceInUse — check `DescribeVpcAttributes` or `DescribeSubnetAttributes` for associated resource IDs.

---

## 7. Concurrency and Rate Limiting

### Throttling

```
Error Code: Throttling
Message: Request was denied due to request throttling.
```

**Root Cause:** API call rate exceeded. VPC APIs typically allow 50–100 requests/second per account.

**Resolution:**
```bash
# Retry with exponential backoff
MAX_RETRIES=3
for i in $(seq 1 $MAX_RETRIES); do
  if ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION"; then
    break
  fi
  echo "Throttled, retrying in $((2 ** i))s..."
  sleep $((2 ** i))
done
```

### InternalError

```
Error Code: InternalError
Message: An internal error occurred. Please retry.
```

**Root Cause:** Volcengine server-side issue.

**Resolution:**
- Retry up to 3 times with exponential backoff (2s, 4s, 8s)
- If persisting > 5 minutes, check [Volcengine Status Page](https://status.volcengine.com)
- Create a support ticket with the `RequestId`

---

## 8. Debugging Strategies

### Enable Verbose Output

```bash
# Check if ve CLI supports verbose/debug mode
ve vpc DescribeVpcs --Region cn-beijing 2>&1 | head -50
```

### Capture RequestId for Support

```bash
# Extract RequestId from any response
REQUEST_ID=$(ve vpc CreateVpc --Region cn-beijing --CidrBlock "10.0.0.0/16" 2>&1 \
  | jq -r '.ResponseMetadata.RequestId')
echo "RequestId: $REQUEST_ID"
```

### Diagnose Network Connectivity

```bash
# Verify endpoint reachability
curl -s --connect-timeout 5 https://vpc.volcengineapi.com/ && echo "✅ Endpoint reachable" || echo "❌ Cannot reach VPC endpoint"
```

### Check CLI Version Compatibility

```bash
# Ensure ve CLI is compatible with current API
VPC_ACTIONS=$(ve vpc --help 2>&1 | grep -c "CreateVpc\|DescribeVpcs\|CreateSubnet")
echo "VPC actions found: $VPC_ACTIONS (expected ≥ 11)"
```

---

*This reference document is part of the ve-vpc-ops skill. For monitoring configurations, see [monitoring.md](monitoring.md).*
