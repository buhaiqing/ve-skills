# API & SDK Usage — Volcengine NAT Gateway

> **Purpose:** Detailed API reference for NAT Gateway, SNAT, and DNAT operations.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [API Overview](#1-api-overview)
2. [NAT Gateway Operations](#2-nat-gateway-operations)
3. [SNAT Rule Operations](#3-snat-rule-operations)
4. [DNAT Rule Operations](#4-dnat-rule-operations)
5. [Response Parsing Patterns](#5-response-parsing-patterns)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `nat` |
| API Version | `2020-04-01` |
| Endpoint | `nat.volcengineapi.com` (default: `open.volcengineapi.com`) |
| Protocol | HTTPS |

### Common Response Structure

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "...", "Service": "nat" },
  "Result": { ... }
}
```

---

## 2. NAT Gateway Operations

### 2.1 CreateNatGateway

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `VpcId` | String | Yes | Parent VPC ID | `vpc-xxx` |
| `SubnetId` | String | Yes | Deployment subnet | `subnet-xxx` |
| `NatGatewaySpec` | String | Yes | Spec | `Small`, `Medium`, `Large` |
| `NatGatewayName` | String | No | NAT name | `prod-nat` |
| `Description` | String | No | Description | — |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `NatGatewayId` | String | NAT Gateway ID (`ngw-xxx`) |
| `OrderNo` | String | Order number (for PrePaid) |

### 2.2 DescribeNatGateways

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `NatGatewayIds.N` | String Array | No | Filter by NAT IDs |
| `VpcId` | String | No | Filter by VPC ID |
| `NatGatewayName` | String | No | Filter by name (fuzzy) |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `NatGateways` | Array | NAT Gateway list |
| `NatGateways[].NatGatewayId` | String | NAT ID |
| `NatGateways[].NatGatewayName` | String | NAT name |
| `NatGateways[].VpcId` | String | Parent VPC |
| `NatGateways[].SubnetId` | String | Deployment subnet |
| `NatGateways[].NatGatewaySpec` | String | Spec |
| `NatGateways[].Status` | String | Status (`Available`) |
| `NatGateways[].EipAddresses` | Array | Bound EIP addresses |
| `TotalCount` | Integer | Total matching NATs |

### 2.3 ModifyNatGatewayAttribute

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `NatGatewayId` | String | Yes | NAT ID |
| `NatGatewayName` | String | No | New name |
| `NatGatewaySpec` | String | No | New spec |
| `Description` | String | No | New description |

### 2.4 DeleteNatGateway

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `NatGatewayId` | String | Yes | NAT ID |

**Prerequisites:** All SNAT rules, DNAT rules deleted; EIP unbound.

---

## 3. SNAT Rule Operations

### 3.1 CreateSnatRule

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `NatGatewayId` | String | Yes | Parent NAT ID |
| `SourceCidr` | String | Yes | Source CIDR for SNAT |
| `EipAddresses` | String Array | Yes | EIP addresses |
| `SnatRuleName` | String | No | Rule name |
| `Description` | String | No | Description |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `SnatRuleId` | String | SNAT rule ID |

### 3.2 DescribeSnatRules

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `SnatRules` | Array | SNAT rule list |
| `SnatRules[].SnatRuleId` | String | Rule ID |
| `SnatRules[].NatGatewayId` | String | Parent NAT ID |
| `SnatRules[].SourceCidr` | String | Source CIDR |
| `SnatRules[].EipAddresses` | Array | EIP addresses |
| `SnatRules[].Status` | String | Status |

### 3.3 DeleteSnatRule

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `SnatRuleId` | String | Yes | SNAT rule ID |

---

## 4. DNAT Rule Operations

### 4.1 CreateDnatRule

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `NatGatewayId` | String | Yes | Parent NAT ID | `ngw-xxx` |
| `EipAddress` | String | Yes | EIP address | `120.78.x.x` |
| `IpProtocol` | String | Yes | Protocol | `TCP`, `UDP` |
| `ExternalPort` | Integer | Yes | External port | `8080` |
| `InternalIp` | String | Yes | Target private IP | `10.0.2.10` |
| `InternalPort` | Integer | Yes | Target port | `80` |
| `DnatRuleName` | String | No | Rule name | `web-dnat` |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `DnatRuleId` | String | DNAT rule ID |

### 4.2 DescribeDnatRules

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `DnatRules` | Array | DNAT rule list |
| `DnatRules[].DnatRuleId` | String | Rule ID |
| `DnatRules[].NatGatewayId` | String | Parent NAT ID |
| `DnatRules[].EipAddress` | String | EIP address |
| `DnatRules[].IpProtocol` | String | Protocol |
| `DnatRules[].ExternalPort` | Integer | External port |
| `DnatRules[].InternalIp` | String | Target IP |
| `DnatRules[].InternalPort` | Integer | Target port |

### 4.3 DeleteDnatRule

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `DnatRuleId` | String | Yes | DNAT rule ID |

---

## 5. Response Parsing Patterns

### Extract NAT Gateway ID

```bash
NAT_ID=$(ve nat Gateway CreateNatGateway --Region "$VOLCENGINE_REGION" \
  --VpcId "$VPC_ID" --SubnetId "$SUBNET_ID" --NatGatewaySpec "Small" \
  | jq -r '.Result.NatGatewayId')
```

### List SNAT Rules with Details

```bash
ve nat Gateway DescribeSnatRules --Region "$VOLCENGINE_REGION" --NatGatewayId "$NAT_ID" \
  | jq -r '.Result.SnatRules[] | "\(.SnatRuleId)\t\(.SourceCidr)\t\(.EipAddresses | join(","))"'
```

### Poll Until NAT Available

```bash
for i in {1..30}; do
  STATUS=$(ve nat Gateway DescribeNatGateways --Region "$VOLCENGINE_REGION" \
    --NatGatewayIds "[\"$NAT_ID\"]" | jq -r '.Result.NatGateways[0].Status')
  [ "$STATUS" = "Available" ] && echo "NAT Gateway ready" && break
  sleep 2
done
```

---

## 7. Go SDK Examples

### CreateNatGateway

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/natgateway"
)

func main() {
    instance := natgateway.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateNatGateway(&natgateway.CreateNatGatewayInput{
        VpcId:          "vpc-xxx",
        SubnetId:       "subnet-xxx",
        NatGatewayName: "prod-nat",
        NatGatewaySpec: "Small",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("NAT Gateway ID: %s\n", resp.Result.NatGatewayId)
}
```

### DescribeNatGateways

```go
func listNATs() {
    instance := natgateway.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.DescribeNatGateways(&natgateway.DescribeNatGatewaysInput{
        VpcId: "vpc-xxx",
    })
    if err != nil {
        panic(err)
    }
    for _, ngw := range resp.Result.NatGateways {
        fmt.Printf("%s: %s (%s) spec=%s\n", ngw.NatGatewayId, ngw.NatGatewayName, ngw.Status, ngw.NatGatewaySpec)
    }
}
```

### CreateSnatRule

```go
func createSNATRule(natID, eipAddr, cidr string) {
    instance := natgateway.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateSnatRule(&natgateway.CreateSnatRuleInput{
        NatGatewayId: natID,
        SourceCidr:   cidr,
        SnatRuleName: "private-subnet-snat",
        EipAddresses: []string{eipAddr},
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("SNAT Rule ID: %s\n", resp.Result.SnatRuleId)
}
```

### CreateDnatRule

```go
func createDNATRule(natID, eipAddr, internalIP string, extPort, intPort int) {
    instance := natgateway.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.CreateDnatRule(&natgateway.CreateDnatRuleInput{
        NatGatewayId: natID,
        EipAddress:   eipAddr,
        IpProtocol:   "TCP",
        ExternalPort: extPort,
        InternalIp:   internalIP,
        InternalPort: intPort,
        DnatRuleName: "https-to-web",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("DNAT Rule ID: %s\n", resp.Result.DnatRuleId)
}
```

---

*This reference document is part of the ve-nat-ops skill.*
