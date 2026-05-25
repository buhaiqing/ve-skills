# CLI Usage — Volcengine NAT Gateway

> **Purpose:** CLI usage reference for Volcengine NAT Gateway operations using the `ve` CLI.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [NAT Gateway Commands](#1-nat-gateway-commands)
2. [SNAT Commands](#2-snat-commands)
3. [DNAT Commands](#3-dnat-commands)
4. [Output Formatting](#4-output-formatting)
5. [Common Patterns](#5-common-patterns)

---

## 1. NAT Gateway Commands

### List All NAT Gateways

```bash
ve nat Gateway DescribeNatGateways --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by VPC

```bash
ve nat Gateway DescribeNatGateways --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Create NAT Gateway

```bash
ve nat Gateway CreateNatGateway \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --NatGatewaySpec "Small" \
  --NatGatewayName "{{user.nat_name}}"
```

### Modify NAT Gateway

```bash
ve nat Gateway ModifyNatGatewayAttribute \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --NatGatewayId "{{user.nat_id}}" \
  --NatGatewayName "{{user.new_name}}" \
  --NatGatewaySpec "Medium"
```

### Delete NAT Gateway

```bash
# ⚠️ IRREVERSIBLE: All rules must be deleted first
ve nat Gateway DeleteNatGateway --Region "{{env.VOLCENGINE_REGION}}" --NatGatewayId "{{user.nat_id}}"
```

---

## 2. SNAT Commands

### List SNAT Rules

```bash
ve nat Gateway DescribeSnatRules --Region "{{env.VOLCENGINE_REGION}}" --NatGatewayId "{{user.nat_id}}"
```

### Create SNAT Rule

```bash
ve nat Gateway CreateSnatRule \
  --Region "{{user.region}}" \
  --NatGatewayId "{{user.nat_id}}" \
  --SourceCidr "{{user.snat_cidr}}" \
  --SnatRuleName "{{user.snat_rule_name}}" \
  --EipAddresses '["{{user.eip_address}}"]'
```

### Delete SNAT Rule

```bash
ve nat Gateway DeleteSnatRule \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --SnatRuleId "{{user.snat_rule_id}}"
```

---

## 3. DNAT Commands

### List DNAT Rules

```bash
ve nat Gateway DescribeDnatRules --Region "{{env.VOLCENGINE_REGION}}" --NatGatewayId "{{user.nat_id}}"
```

### Create DNAT Rule

```bash
ve nat Gateway CreateDnatRule \
  --Region "{{user.region}}" \
  --NatGatewayId "{{user.nat_id}}" \
  --EipAddress "{{user.eip_address}}" \
  --IpProtocol "TCP" \
  --ExternalPort "{{user.dnat_external_port}}" \
  --InternalIp "{{user.dnat_internal_ip}}" \
  --InternalPort "{{user.dnat_internal_port}}" \
  --DnatRuleName "{{user.dnat_rule_name}}"
```

### Delete DNAT Rule

```bash
ve nat Gateway DeleteDnatRule \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --DnatRuleId "{{user.dnat_rule_id}}"
```

---

## 4. Output Formatting

### NAT Gateway Table

```bash
ve nat Gateway DescribeNatGateways --Region "$VOLCENGINE_REGION" | jq -r '
  .Result.NatGateways[] |
  [.NatGatewayId, .NatGatewayName, .VpcId, .NatGatewaySpec, .Status] |
  @tsv
' | column -t -s $'\t'
```

### Complete NAT + Rules View

```bash
for NAT_ID in $(ve nat Gateway DescribeNatGateways --Region "$VOLCENGINE_REGION" | jq -r '.Result.NatGateways[].NatGatewayId'); do
  echo "=== NAT: $NAT_ID ==="
  echo "--- SNAT Rules ---"
  ve nat Gateway DescribeSnatRules --Region "$VOLCENGINE_REGION" --NatGatewayId "$NAT_ID" | jq -r '.Result.SnatRules[] | "\(.SnatRuleId): \(.SourceCidr) → \(.EipAddresses | join(","))"'
  echo "--- DNAT Rules ---"
  ve nat Gateway DescribeDnatRules --Region "$VOLCENGINE_REGION" --NatGatewayId "$NAT_ID" | jq -r '.Result.DnatRules[] | "\(.DnatRuleId): \(.EipAddress):\(.ExternalPort) → \(.InternalIp):\(.InternalPort) (\(.IpProtocol))"'
done
```

---

## 5. Common Patterns

### Pattern: Full NAT Gateway Setup

```bash
# Prerequisites: VPC, Subnet (ve-vpc-ops), EIP (ve-eip-ops)
VPC_ID="vpc-xxx"
SUBNET_ID="subnet-xxx"  # Public subnet for NAT
EIP_ADDR="120.78.x.x"

# Step 1: Create NAT Gateway
NAT_ID=$(ve nat Gateway CreateNatGateway \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --SubnetId "$SUBNET_ID" \
  --NatGatewaySpec "Small" \
  --NatGatewayName "prod-nat" \
  | jq -r '.Result.NatGatewayId')

# Step 2: Bind EIP to NAT Gateway
NAT_EIP_ID=$(ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" \
  --EipAddress "$EIP_ADDR" | jq -r '.Result.EipAddresses[0].AllocationId')
ve eip AssociateEipAddress --Region "$VOLCENGINE_REGION" \
  --AllocationId "$NAT_EIP_ID" --InstanceId "$NAT_ID" --InstanceType "Nat"

# Step 3: Create SNAT rule for private subnet
ve nat Gateway CreateSnatRule \
  --Region "$VOLCENGINE_REGION" \
  --NatGatewayId "$NAT_ID" \
  --SourceCidr "10.0.2.0/24" \
  --SnatRuleName "private-subnet-snat" \
  --EipAddresses "[\"$EIP_ADDR\"]"

echo "NAT Gateway setup complete: $NAT_ID"
```

### Pattern: Add DNAT Port Mapping

```bash
ve nat Gateway CreateDnatRule \
  --Region "$VOLCENGINE_REGION" \
  --NatGatewayId "$NAT_ID" \
  --EipAddress "$EIP_ADDR" \
  --IpProtocol "TCP" \
  --ExternalPort 443 \
  --InternalIp "10.0.2.10" \
  --InternalPort 443 \
  --DnatRuleName "https-to-web"

echo "DNAT rule: $EIP_ADDR:443 → 10.0.2.10:443"
```

---

*This reference document is part of the ve-nat-ops skill.*
