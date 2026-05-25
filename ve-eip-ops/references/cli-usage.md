# CLI Usage — Volcengine EIP

> **Purpose:** Comprehensive CLI usage reference for Volcengine EIP operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [Installation and Setup](#1-installation-and-setup)
2. [EIP Discovery Commands](#2-eip-discovery-commands)
3. [Lifecycle Commands](#3-lifecycle-commands)
4. [Binding Commands](#4-binding-commands)
5. [Bandwidth Commands](#5-bandwidth-commands)
6. [Attribute Commands](#6-attribute-commands)
7. [Output Formatting](#7-output-formatting)
8. [Common Patterns](#8-common-patterns)

---

## 1. Installation and Setup

```bash
# Verify EIP service support
ve eip --help
```

Expected: Lists EIP actions including AllocateEipAddress, DescribeEipAddresses, etc.

---

## 2. EIP Discovery Commands

### List All EIPs

```bash
ve eip DescribeEipAddresses --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by Status

```bash
# Unbound EIPs only
ve eip DescribeEipAddresses --Region "{{env.VOLCENGINE_REGION}}" --Status "Available"

# Bound EIPs only
ve eip DescribeEipAddresses --Region "{{env.VOLCENGINE_REGION}}" --Status "InUse"
```

### Filter by IP Address

```bash
ve eip DescribeEipAddresses --Region "{{env.VOLCENGINE_REGION}}" --EipAddress "120.78.x.x"
```

### Get EIP Details

```bash
ve eip DescribeEipAddressAttributes \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}"
```

---

## 3. Lifecycle Commands

### Allocate EIP

```bash
ve eip AllocateEipAddress \
  --Region "{{user.region}}" \
  --LineType "BGP" \
  --Bandwidth "{{user.bandwidth}}" \
  --Name "{{user.eip_name}}"
```

### Release EIP

```bash
# ⚠️ IRREVERSIBLE: EIP must be unbound
ve eip ReleaseEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}"
```

---

## 4. Binding Commands

### Bind EIP to ECS

```bash
ve eip AssociateEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --InstanceId "{{user.instance_id}}" \
  --InstanceType "EcsInstance"
```

### Bind EIP to CLB

```bash
ve eip AssociateEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --InstanceId "{{user.clb_id}}" \
  --InstanceType "ClbInstance"
```

### Bind EIP to NAT Gateway

```bash
ve eip AssociateEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --InstanceId "{{user.nat_id}}" \
  --InstanceType "Nat"
```

### Unbind EIP

```bash
ve eip DisassociateEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}"
```

### Bind EIP to ENI

```bash
ve eip AssociateEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --InstanceId "{{user.eni_id}}" \
  --InstanceType "NetworkInterface"
```

---

## 5. Bandwidth Commands

### Modify Bandwidth

```bash
ve eip ModifyEipBandwidth \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --Bandwidth "{{user.bandwidth}}"
```

### Query Current Bandwidth

```bash
ve eip DescribeEipBandwidth \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}"
```

---

## 6. Attribute Commands

### Modify Name/Description

```bash
ve eip ModifyEipAddressAttributes \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --Name "{{user.new_name}}" \
  --Description "{{user.new_description}}"
```

### Add Tags

```bash
ve eip TagEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --Tags '[{"Key": "env", "Value": "production"}, {"Key": "team", "Value": "ops"}]'
```

### Renew Prepaid EIP

```bash
ve eip RenewEipAddress \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --AllocationId "{{user.eip_id}}" \
  --RenewalPeriod "Month" \
  --Quantity 1
```

---

## 7. Output Formatting

### Extract EIP ID and Address

```bash
RESULT=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" --Bandwidth 5)
EIP_ID=$(echo "$RESULT" | jq -r '.Result.AllocationId')
EIP_ADDR=$(echo "$RESULT" | jq -r '.Result.EipAddress')
echo "EIP: $EIP_ADDR ($EIP_ID)"
```

### Table of All EIPs

```bash
ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.EipAddresses[] |
  [.AllocationId, .EipAddress, .Status, (.Bandwidth | tostring) + "Mbps", .InstanceType // "-", .InstanceId // "-"] |
  @tsv
' | column -t -s $'\t'
```

---

## 8. Common Patterns

### Pattern: Allocate and Bind EIP

```bash
# Step 1: Allocate
EIP_ID=$(ve eip AllocateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --Bandwidth 5 \
  --Name "web-eip" \
  | jq -r '.Result.AllocationId')

# Step 2: Poll until Available
for i in {1..30}; do
  STATUS=$(ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" \
    --AllocationIds "[\"$EIP_ID\"]" | jq -r '.Result.EipAddresses[0].Status')
  [ "$STATUS" = "Available" ] && break
  sleep 2
done

# Step 3: Bind to ECS
ve eip AssociateEipAddress \
  --Region "$VOLCENGINE_REGION" \
  --AllocationId "$EIP_ID" \
  --InstanceId "$ECS_ID" \
  --InstanceType "EcsInstance"
```

### Pattern: Scale Bandwidth

```bash
# Check current bandwidth
CURRENT=$(ve eip DescribeEipBandwidth --Region "$VOLCENGINE_REGION" --AllocationId "$EIP_ID" \
  | jq -r '.Result.Bandwidth')

if [ "$CURRENT" -lt 100 ]; then
  echo "Scaling from ${CURRENT}Mbps to 100Mbps"
  ve eip ModifyEipBandwidth --Region "$VOLCENGINE_REGION" \
    --AllocationId "$EIP_ID" --Bandwidth 100
else
  echo "Bandwidth adequate: ${CURRENT}Mbps"
fi
```

---

*This reference document is part of the ve-eip-ops skill.*
