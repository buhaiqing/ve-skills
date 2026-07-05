# CLI Usage — Volcengine VPC

> **Purpose:** Comprehensive CLI usage reference for Volcengine VPC operations using the `ve` CLI. Covers installation, configuration, all supported commands, output formats, and best practices.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [ve CLI Overview](#1-ve-cli-overview)
2. [Installation](#2-installation)
3. [Credential Configuration](#3-credential-configuration)
4. [Command Discovery](#4-command-discovery)
5. [VPC Commands](#5-vpc-commands)
6. [Subnet Commands](#6-subnet-commands)
7. [Route Table Commands](#7-route-table-commands)
8. [Route Entry Commands](#8-route-entry-commands)
9. [Output Formatting](#9-output-formatting)
10. [Common Patterns](#10-common-patterns)

---

## 1. ve CLI Overview

The Volcengine `ve` CLI is the primary execution path for all VPC operations in this skill.

| Property | Value |
|----------|-------|
| Command Prefix | `ve` (since v1.0.20) |
| Service Prefix | `ve vpc` |
| Default Output | JSON |
| Credential Source | Environment variables or `~/.volcengine/config.json` |
| GitHub Releases | https://github.com/volcengine/volcengine-cli/releases |

> **Note:** Prior to v1.0.20, the command prefix was `volcengine-cli`. All skills use `ve`.

---

## 2. Installation

### Download and Install

```bash
# Auto-detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

VE_VERSION="{{env.ve_version}}"
curl -fsSL "https://github.com/volcengine/volcengine-cli/releases/download/${VE_VERSION}/ve-${OS}-${ARCH}" -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify
ve version
```

### Verify VPC Service Support

```bash
ve vpc --help
```

Expected output should list available VPC actions:

```
Available Actions:
  CreateVpc
  DescribeVpcs
  DeleteVpc
  ModifyVpcAttribute
  CreateSubnet
  DescribeSubnets
  DeleteSubnet
  CreateRouteTable
  DescribeRouteTableList
  CreateRouteEntry
  DeleteRouteEntry
  ...
```

---

## 3. Credential Configuration

### Environment Variables (Recommended)

```bash
export VOLCENGINE_ACCESS_KEY="YOUR_ACCESS_KEY"
export VOLCENGINE_SECRET_KEY="YOUR_SECRET_KEY"
export VOLCENGINE_REGION="cn-beijing"
```

### Profile-Based Configuration

```bash
# Configure a profile
ve configure set --profile default --region cn-beijing --access-key "YOUR_AK" --secret-key "YOUR_SK"

# List all profiles
ve configure list

# Switch profile
ve configure profile --profile production
```

### Credential Verification

```bash
# Verify with a read-only API call
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}"
```

Expected: A JSON response with an empty or populated `Vpcs` array, **not** an authentication error.

---

## 4. Command Discovery

### Explore Available Actions

```bash
# All VPC actions
ve vpc --help

# Help for a specific action
ve vpc DescribeVpcs --help

# All services
ve --help
```

### Understanding CLI Syntax

```
ve <service> <action> --<ParameterName> <value>
```

Examples:
```bash
ve vpc DescribeVpcs --Region cn-beijing
ve vpc CreateVpc --Region cn-beijing --CidrBlock "10.0.0.0/16" --VpcName "my-vpc"
```

### Array and JSON Parameters

```bash
# Filter by multiple VPC IDs (JSON array)
ve vpc DescribeVpcs --Region cn-beijing --VpcIds '["vpc-xxx", "vpc-yyy"]'

# Using --body for full JSON
ve vpc CreateVpc --body '{"Region":"cn-beijing","CidrBlock":"10.0.0.0/16","VpcName":"my-vpc"}'
```

---

## 5. VPC Commands

### List All VPCs

```bash
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}"
```

### Filter by Name

```bash
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}" --VpcName "{{user.vpc_name}}"
```

### Filter by ID

```bash
ve vpc DescribeVpcs --Region "{{env.VOLCENGINE_REGION}}" --VpcIds "[\"{{user.vpc_id}}\"]"
```

### Get VPC Details

```bash
ve vpc DescribeVpcAttributes --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Create VPC

```bash
ve vpc CreateVpc \
  --Region "{{user.region}}" \
  --CidrBlock "{{user.cidr_block}}" \
  --VpcName "{{user.vpc_name}}"
```

### Delete VPC

```bash
# ⚠️ IRREVERSIBLE: VPC must be empty
ve vpc DeleteVpc --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"
```

### Modify VPC Name/Description

```bash
ve vpc ModifyVpcAttribute \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --VpcId "{{user.vpc_id}}" \
  --VpcName "{{user.new_vpc_name}}"
```

---

## 6. Subnet Commands

### List All Subnets

```bash
ve vpc DescribeSubnets --Region "{{env.VOLCENGINE_REGION}}"
```

### List Subnets in a VPC

```bash
ve vpc DescribeSubnets \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --VpcId "{{user.vpc_id}}"
```

### Get Subnet Details

```bash
ve vpc DescribeSubnetAttributes \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --SubnetId "{{user.subnet_id}}"
```

### Create Subnet

```bash
ve vpc CreateSubnet \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --CidrBlock "{{user.subnet_cidr}}" \
  --ZoneId "{{user.zone_id}}" \
  --SubnetName "{{user.subnet_name}}"
```

### Delete Subnet

```bash
# ⚠️ IRREVERSIBLE: Subnet must be empty
ve vpc DeleteSubnet \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --SubnetId "{{user.subnet_id}}"
```

---

## 7. Route Table Commands

### List Route Tables

```bash
ve vpc DescribeRouteTableList --Region "{{env.VOLCENGINE_REGION}}"
```

### List Route Tables in a VPC

```bash
ve vpc DescribeRouteTableList \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --VpcId "{{user.vpc_id}}"
```

### Create Route Table

```bash
ve vpc CreateRouteTable \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --VpcId "{{user.vpc_id}}" \
  --RouteTableName "{{user.route_table_name}}"
```

### Modify Route Table

```bash
ve vpc ModifyRouteTableAttributes \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --RouteTableName "{{user.new_name}}"
```

### Associate Route Table with Subnet

```bash
ve vpc AssociateRouteTable \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --SubnetId "{{user.subnet_id}}"
```

### Disassociate Route Table

```bash
ve vpc DisassociateRouteTable \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --SubnetId "{{user.subnet_id}}"
```

### Delete Route Table

```bash
# ⚠️ Must have no custom route entries and no subnet associations
ve vpc DeleteRouteTable \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}"
```

---

## 8. Route Entry Commands

### Create Route Entry

```bash
ve vpc CreateRouteEntry \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --DestinationCidrBlock "{{user.destination_cidr}}" \
  --NextHopType "{{user.next_hop_type}}" \
  --NextHopId "{{user.next_hop_id}}"
```

### Delete Route Entry

```bash
ve vpc DeleteRouteEntry \
  --Region "{{env.VOLCENGINE_REGION}}" \
  --RouteTableId "{{user.route_table_id}}" \
  --DestinationCidrBlock "{{user.destination_cidr}}" \
  --NextHopId "{{user.next_hop_id}}"
```

> **Note:** `NextHopType` is not required for deletion — `RouteTableId`, `DestinationCidrBlock`, and `NextHopId` uniquely identify a route entry.

---

## 9. Output Formatting

### Default JSON Output

```bash
ve vpc DescribeVpcs --Region cn-beijing
```

### Extract Specific Fields with jq

```bash
# Get VPC IDs only
ve vpc DescribeVpcs --Region cn-beijing | jq -r '.Result.Vpcs[].VpcId'

# Get VPC names
ve vpc DescribeVpcs --Region cn-beijing | jq -r '.Result.Vpcs[].VpcName'

# Get VPC name and CIDR pairs
ve vpc DescribeVpcs --Region cn-beijing | jq -r '.Result.Vpcs[] | "\(.VpcName): \(.CidrBlock)"'

# Count total VPCs
ve vpc DescribeVpcs --Region cn-beijing | jq -r '.Result.TotalCount'
```

### Table Output

```bash
# Pretty-print VPC table
ve vpc DescribeVpcs --Region cn-beijing | jq -r '
  .Result.Vpcs[] |
  [.VpcId, .VpcName, .CidrBlock, .Status] |
  @tsv
' | column -t -s $'\t'
```

---

## 10. Common Patterns

### Pattern: Create VPC, then Create Subnet

```bash
# Step 1: Create VPC, capture VpcId
VPC_ID=$(ve vpc CreateVpc \
  --Region "$VOLCENGINE_REGION" \
  --CidrBlock "10.0.0.0/16" \
  --VpcName "prod-vpc" \
  | jq -r '.Result.VpcId')

echo "Created VPC: $VPC_ID"

# Step 2: Poll until VPC is Available
for i in {1..30}; do
  STATUS=$(ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION" --VpcIds "[\"$VPC_ID\"]" \
    | jq -r '.Result.Vpcs[0].Status')
  [ "$STATUS" = "Available" ] && break
  sleep 2
done

# Step 3: Create subnet
SUBNET_ID=$(ve vpc CreateSubnet \
  --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" \
  --CidrBlock "10.0.1.0/24" \
  --ZoneId "cn-beijing-a" \
  --SubnetName "web-subnet" \
  | jq -r '.Result.SubnetId')

echo "Created Subnet: $SUBNET_ID"
```

### Pattern: Idempotent Resource Creation

```bash
# Check if VPC exists before creating
EXISTING=$(ve vpc DescribeVpcs \
  --Region "$VOLCENGINE_REGION" \
  --VpcName "my-vpc" \
  | jq -r '.Result.Vpcs[0].VpcId')

if [ "$EXISTING" != "null" ] && [ -n "$EXISTING" ]; then
  echo "VPC already exists: $EXISTING"
else
  VPC_ID=$(ve vpc CreateVpc \
    --Region "$VOLCENGINE_REGION" \
    --CidrBlock "10.0.0.0/16" \
    --VpcName "my-vpc" \
    | jq -r '.Result.VpcId')
  echo "Created VPC: $VPC_ID"
fi
```

### Pattern: Multi-AZ Subnet Creation

```bash
VPC_ID="vpc-xxx"
ZONES=("cn-beijing-a" "cn-beijing-b" "cn-beijing-c")
CIDRS=("10.0.1.0/24" "10.0.2.0/24" "10.0.3.0/24")
NAMES=("web-a" "web-b" "web-c")

for i in "${!ZONES[@]}"; do
  SUBNET_ID=$(ve vpc CreateSubnet \
    --Region "$VOLCENGINE_REGION" \
    --VpcId "$VPC_ID" \
    --CidrBlock "${CIDRS[$i]}" \
    --ZoneId "${ZONES[$i]}" \
    --SubnetName "${NAMES[$i]}" \
    | jq -r '.Result.SubnetId')
  echo "Created subnet in ${ZONES[$i]}: $SUBNET_ID"
done
```

---

*This reference document is part of the ve-vpc-ops skill. For JIT Go SDK fallback patterns, see [api-sdk-usage.md](api-sdk-usage.md).*
