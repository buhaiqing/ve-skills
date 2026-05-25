# API & SDK Usage — Volcengine VPC

> **Purpose:** Detailed API and SDK usage reference for Volcengine VPC operations. Covers OpenAPI specifications, Go SDK patterns, and response parsing.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [API Overview](#1-api-overview)
2. [VPC Operations](#2-vpc-operations)
3. [Subnet Operations](#3-subnet-operations)
4. [Route Table Operations](#4-route-table-operations)
5. [Route Entry Operations](#5-route-entry-operations)
6. [Go SDK Usage](#6-go-sdk-usage)
7. [Response Parsing Patterns](#7-response-parsing-patterns)
8. [Pagination](#8-pagination)
9. [Idempotency](#9-idempotency)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `vpc` |
| API Version | `2020-04-01` |
| Endpoint | `vpc.volcengineapi.com` (default: `open.volcengineapi.com`) |
| Protocol | HTTPS |
| Request Format | JSON |
| Signature | HMAC-SHA256 |

### Common Request Parameters

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `Action` | Yes | API operation name | `CreateVpc` |
| `Version` | Yes | API version | `2020-04-01` |
| `Region` | Yes | Region ID | `cn-beijing` |

### Common Response Structure

```json
{
  "ResponseMetadata": {
    "RequestId": "20260525120000010206084039004D0054",
    "Action": "CreateVpc",
    "Version": "2020-04-01",
    "Service": "vpc",
    "Region": "cn-beijing"
  },
  "Result": { ... }
}
```

---

## 2. VPC Operations

### 2.1 CreateVpc

**Description:** Create a VPC with a specified CIDR block.

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `CidrBlock` | String | Yes | IPv4 CIDR block | `10.0.0.0/16` |
| `VpcName` | String | No | VPC name | `my-vpc` |
| `Description` | String | No | VPC description | `Production VPC` |
| `EnableIpv6` | Boolean | No | Enable IPv6 | `false` |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `VpcId` | String | Created VPC ID |

**Example Response:**

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "CreateVpc" },
  "Result": {
    "VpcId": "vpc-2d5zg1qk7rj0w0x5tq3w8xxxx"
  }
}
```

### 2.2 DescribeVpcs

**Description:** Query VPC list with optional filters.

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `VpcIds.N` | String Array | No | Filter by VPC IDs | `["vpc-xxx"]` |
| `VpcName` | String | No | Filter by name (fuzzy) | `my-vpc` |
| `ProjectName` | String | No | Filter by project | `default` |
| `PageNumber` | Integer | No | Page number (1-based) | `1` |
| `PageSize` | Integer | No | Items per page (default 10, max 100) | `10` |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Vpcs` | Array | VPC list |
| `Vpcs[].VpcId` | String | VPC ID |
| `Vpcs[].VpcName` | String | VPC name |
| `Vpcs[].CidrBlock` | String | Primary CIDR block |
| `Vpcs[].SecondaryCidrBlocks` | Array | Additional CIDR blocks |
| `Vpcs[].Status` | String | VPC status (`Available`) |
| `Vpcs[].CreationTime` | String | ISO 8601 creation time |
| `Vpcs[].UpdateTime` | String | ISO 8601 update time |
| `Vpcs[].ZoneId` | String | Primary zone |
| `Vpcs[].Tags` | Array | Tags |
| `TotalCount` | Integer | Total matching VPCs |

### 2.3 DescribeVpcAttributes

**Description:** Query detailed attributes of a specific VPC.

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `VpcId` | String | Yes | VPC ID |

**Response Fields (additions over DescribeVpcs):**

| Field | Type | Description |
|-------|------|-------------|
| `NatGatewayIds` | Array | NAT gateway IDs in this VPC |
| `RouteTableIds` | Array | Route table IDs in this VPC |
| `SubnetIds` | Array | Subnet IDs in this VPC |
| `SecurityGroupIds` | Array | Security group IDs in this VPC |
| `ProjectName` | String | Project name |

### 2.4 DeleteVpc

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `VpcId` | String | Yes | VPC ID to delete |

**Response:** Empty Result on success.

**Prerequisites for Deletion:**
- All subnets must be deleted
- All route table associations must be removed
- No active resources (ECS, CLB, etc.) in the VPC

### 2.5 ModifyVpcAttribute

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `VpcId` | String | Yes | VPC ID |
| `VpcName` | String | No | New VPC name |
| `Description` | String | No | New description (max 255 chars) |

---

## 3. Subnet Operations

### 3.1 CreateSubnet

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `VpcId` | String | Yes | Parent VPC ID | `vpc-xxx` |
| `CidrBlock` | String | Yes | Subnet CIDR (mask /16–/29) | `10.0.1.0/24` |
| `ZoneId` | String | Yes | Availability zone | `cn-beijing-a` |
| `SubnetName` | String | No | Subnet name | `web-subnet-a` |
| `Description` | String | No | Description | — |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `SubnetId` | String | Created subnet ID |

### 3.2 DescribeSubnets

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `SubnetIds.N` | String Array | No | Filter by subnet IDs |
| `VpcId` | String | No | Filter by VPC ID |
| `ZoneId` | String | No | Filter by zone |
| `SubnetName` | String | No | Filter by name (fuzzy) |
| `PageNumber` | Integer | No | Page number |
| `PageSize` | Integer | No | Page size |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Subnets` | Array | Subnet list |
| `Subnets[].SubnetId` | String | Subnet ID |
| `Subnets[].SubnetName` | String | Subnet name |
| `Subnets[].CidrBlock` | String | IPv4 CIDR |
| `Subnets[].VpcId` | String | Parent VPC ID |
| `Subnets[].ZoneId` | String | Availability zone |
| `Subnets[].Status` | String | Status (`Available`) |
| `Subnets[].AvailableIpAddressCount` | Integer | Available IPs |
| `Subnets[].TotalIpAddressCount` | Integer | Total IPs |
| `Subnets[].IsDefault` | Boolean | Is default subnet |
| `Subnets[].RouteTable` | Object | Associated route table |
| `Subnets[].CreationTime` | String | Creation time |
| `TotalCount` | Integer | Total matching subnets |

### 3.3 DescribeSubnetAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `SubnetId` | String | Yes | Subnet ID |

### 3.4 DeleteSubnet

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `SubnetId` | String | Yes | Subnet ID |

**Prerequisites:**
- No instances, ENIs, or other resources in the subnet

---

## 4. Route Table Operations

### 4.1 CreateRouteTable

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `VpcId` | String | Yes | Parent VPC ID | `vpc-xxx` |
| `RouteTableName` | String | No | Route table name | `custom-rt` |
| `Description` | String | No | Description | — |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `RouteTableId` | String | Created route table ID |

### 4.2 DescribeRouteTableList

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableIds.N` | String Array | No | Filter by IDs |
| `VpcId` | String | No | Filter by VPC ID |
| `RouteTableName` | String | No | Filter by name |
| `PageNumber` | Integer | No | Page number |
| `PageSize` | Integer | No | Page size |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `RouteTables` | Array | Route table list |
| `RouteTables[].RouteTableId` | String | Route table ID |
| `RouteTables[].RouteTableName` | String | Route table name |
| `RouteTables[].VpcId` | String | Parent VPC ID |
| `RouteTables[].RouteTableType` | String | `System` or `Custom` |
| `RouteTables[].RouteEntryCount` | Integer | Number of entries |
| `RouteTables[].SubnetIds` | Array | Associated subnet IDs |
| `RouteTables[].RouteEntries` | Array | Route entries |
| `TotalCount` | Integer | Total matching route tables |

### 4.3 ModifyRouteTableAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableId` | String | Yes | Route table ID |
| `RouteTableName` | String | No | New name |
| `Description` | String | No | New description |

### 4.4 AssociateRouteTable

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableId` | String | Yes | Route table ID |
| `SubnetId` | String | Yes | Subnet ID to associate |

### 4.5 DisassociateRouteTable

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableId` | String | Yes | Route table ID |
| `SubnetId` | String | Yes | Subnet ID to disassociate |

### 4.6 DeleteRouteTable

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableId` | String | Yes | Route table ID |

**Prerequisites:**
- No custom route entries
- No subnet associations

---

## 5. Route Entry Operations

### 5.1 CreateRouteEntry

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `RouteTableId` | String | Yes | Route table ID | `rtb-xxx` |
| `DestinationCidrBlock` | String | Yes | Destination CIDR | `0.0.0.0/0` |
| `NextHopType` | String | Yes | Next hop type | `NatGateway`, `Instance`, `Local` |
| `NextHopId` | String | Yes | Next hop ID | `ngw-xxx` |
| `Description` | String | No | Description | — |

**Valid NextHopType Values:**

| Value | Description |
|-------|-------------|
| `Local` | Local VPC route (cannot delete) |
| `NatGateway` | NAT Gateway |
| `Instance` | ECS instance |
| `HaVip` | High-availability VIP |
| `NetworkInterface` | Elastic network interface |
| `VpnConnection` | VPN connection |

### 5.2 DeleteRouteEntry

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `RouteTableId` | String | Yes | Route table ID |
| `DestinationCidrBlock` | String | Yes | Destination CIDR |
| `NextHopId` | String | Yes | Next hop ID |

> **Note:** System routes (e.g., `Local` to VPC CIDR, default `0.0.0.0/0`) cannot be deleted through this API.

---

## 6. Go SDK Usage

### 6.1 SDK Initialization

```go
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/base"
)

func newVpcClient() *base.Client {
    client := base.NewClient(
        &base.Credentials{
            AK: os.Getenv("VOLCENGINE_ACCESS_KEY"),
            SK: os.Getenv("VOLCENGINE_SECRET_KEY"),
        },
        base.Region(os.Getenv("VOLCENGINE_REGION")),
    )
    client.ServiceInfo = &base.ServiceInfo{
        Timeout: 5 * time.Second,
        Schema:  "https",
        Host:    "open.volcengineapi.com",
        Header:  base.NewHeader(),
        Credentials: &base.Credentials{
            AK: os.Getenv("VOLCENGINE_ACCESS_KEY"),
            SK: os.Getenv("VOLCENGINE_SECRET_KEY"),
        },
    }
    client.ServiceInfo.ServiceName = "vpc"
    return &client
}
```

### 6.2 JIT Go SDK Pattern

For the JIT Go SDK path used in this skill, the dynamic script pattern is:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/base"
)

func main() {
    instance := newVpcClient()

    body := make(map[string]interface{})
    body["CidrBlock"] = "10.0.0.0/16"
    body["VpcName"] = os.Getenv("VPC_NAME")

    resp, err := instance.Request("CreateVpc", nil, body)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}
```

---

## 7. Response Parsing Patterns

### Extract VPC ID from CreateVpc

```bash
# Using jq (assuming CLI output or SDK stdout)
VPC_ID=$(ve vpc CreateVpc --Region "$VOLCENGINE_REGION" --CidrBlock "10.0.0.0/16" | jq -r '.Result.VpcId')
```

### Extract Subnet List from DescribeSubnets

```bash
ve vpc DescribeSubnets --Region "$VOLCENGINE_REGION" --VpcId "$VPC_ID" \
  | jq -r '.Result.Subnets[] | "\(.SubnetId)\t\(.SubnetName)\t\(.CidrBlock)\t\(.Status)"'
```

### Poll Until Available

```bash
# Poll VPC status until Available
for i in {1..30}; do
    STATUS=$(ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION" --VpcIds "[\"$VPC_ID\"]" \
      | jq -r '.Result.Vpcs[0].Status')
    if [ "$STATUS" = "Available" ]; then
        echo "VPC $VPC_ID is Available"
        break
    fi
    echo "Waiting for VPC to be available... (attempt $i)"
    sleep 2
done
```

---

## 8. Pagination

All `Describe*` APIs support pagination:

```bash
# First page
ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION" --PageSize 50 --PageNumber 1

# Check if more pages exist
TOTAL=$(ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION" | jq -r '.Result.TotalCount')
PAGE_SIZE=50
TOTAL_PAGES=$(( (TOTAL + PAGE_SIZE - 1) / PAGE_SIZE ))

# Iterate all pages
for PAGE in $(seq 1 $TOTAL_PAGES); do
    ve vpc DescribeVpcs --Region "$VOLCENGINE_REGION" --PageSize $PAGE_SIZE --PageNumber $PAGE
done
```

---

## 9. Idempotency

### CreateVpc

| Parameter | Idempotency Behavior |
|-----------|---------------------|
| `CidrBlock` + `VpcName` | Not inherently idempotent; creating twice with same name may fail with `InvalidVpcName.Duplicate` |

### CreateSubnet

| Parameter | Idempotency Behavior |
|-----------|---------------------|
| `VpcId` + `CidrBlock` + `ZoneId` | May fail with `InvalidCidrBlock.Conflict` if CIDR already exists in VPC |

**Recommendation:** Use `Describe*` before `Create*` to check for existing resources and achieve idempotent behavior:

1. Call `DescribeVpcs` with `VpcName` filter
2. If VPC exists with matching CIDR → reuse `{{output.vpc_id}}`
3. If not → call `CreateVpc`

---

*This reference document is part of the ve-vpc-ops skill. For operational procedures, see the main SKILL.md.*
