# API & SDK Usage — Volcengine EIP

> **Purpose:** Detailed API and SDK usage reference for Volcengine EIP operations. Covers OpenAPI specifications, Go SDK patterns, and response parsing.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-25

---

## Table of Contents

1. [API Overview](#1-api-overview)
2. [EIP Lifecycle Operations](#2-eip-lifecycle-operations)
3. [EIP Binding Operations](#3-eip-binding-operations)
4. [Bandwidth Management](#4-bandwidth-management)
5. [EIP Attribute Operations](#5-eip-attribute-operations)
6. [Go SDK Usage](#6-go-sdk-usage)
7. [Response Parsing Patterns](#7-response-parsing-patterns)

---

## 1. API Overview

| Property | Value |
|----------|-------|
| Service | `eip` |
| API Version | `2020-04-01` |
| Endpoint | `eip.volcengineapi.com` (default: `open.volcengineapi.com`) |
| Protocol | HTTPS |
| Request Format | JSON |
| Signature | HMAC-SHA256 |

### Common Response Structure

```json
{
  "ResponseMetadata": {
    "RequestId": "20260525120000010206084039004D0054",
    "Action": "AllocateEipAddress",
    "Version": "2020-04-01",
    "Service": "eip",
    "Region": "cn-beijing"
  },
  "Result": { ... }
}
```

---

## 2. EIP Lifecycle Operations

### 2.1 AllocateEipAddress

**Description:** Allocate a new EIP address.

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `LineType` | String | No | Line type | `BGP` (default) |
| `ChargeType` | String | No | Billing type | `PostPaid` (default) |
| `Bandwidth` | Integer | No | Bandwidth in Mbps | `5` |
| `BillingType` | String | No | Billing method | `PayByTraffic`, `PayByBandwidth` |
| `Name` | String | No | EIP name | `prod-eip-1` |
| `Description` | String | No | Description | — |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `AllocationId` | String | EIP Allocation ID |
| `EipAddress` | String | Allocated IP address |

**Example Response:**

```json
{
  "ResponseMetadata": { "RequestId": "...", "Action": "AllocateEipAddress" },
  "Result": {
    "AllocationId": "eipalloc-3d5zg1qk7rj0w0x5tq3w8xxxx",
    "EipAddress": "120.78.100.50"
  }
}
```

### 2.2 DescribeEipAddresses

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationIds.N` | String Array | No | Filter by EIP IDs |
| `EipAddress` | String | No | Filter by IP address |
| `Status` | String | No | Filter by status (`Available`, `InUse`) |
| `Name` | String | No | Filter by name (fuzzy) |
| `IsDefault` | Boolean | No | Filter by default EIP |
| `PageNumber` | Integer | No | Page number |
| `PageSize` | Integer | No | Page size |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `EipAddresses` | Array | EIP list |
| `EipAddresses[].AllocationId` | String | EIP ID |
| `EipAddresses[].EipAddress` | String | IP address |
| `EipAddresses[].Status` | String | Status |
| `EipAddresses[].LineType` | String | Line type |
| `EipAddresses[].Bandwidth` | Integer | Bandwidth (Mbps) |
| `EipAddresses[].ChargeType` | String | Billing type |
| `EipAddresses[].BillingType` | String | Billing method |
| `EipAddresses[].InstanceType` | String | Bound instance type |
| `EipAddresses[].InstanceId` | String | Bound instance ID |
| `EipAddresses[].Name` | String | EIP name |
| `EipAddresses[].CreatedAt` | String | Creation time |
| `TotalCount` | Integer | Total matching EIPs |

### 2.3 DescribeEipAddressAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP Allocation ID |

### 2.4 ReleaseEipAddress

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID to release |

**Prerequisites:** EIP must be `Available` (not bound).

---

## 3. EIP Binding Operations

### 3.1 AssociateEipAddress

**Request Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `AllocationId` | String | Yes | EIP ID | `eipalloc-xxx` |
| `InstanceId` | String | Yes | Target instance ID | `i-xxx` |
| `InstanceType` | String | Yes | Instance type | `EcsInstance`, `ClbInstance`, `Nat` |

**Valid InstanceType Values:**

| Value | Description |
|-------|-------------|
| `EcsInstance` | ECS instance |
| `ClbInstance` | CLB load balancer |
| `Nat` | NAT Gateway |
| `HaVip` | High-availability VIP |
| `NetworkInterface` | Elastic network interface |

### 3.2 DisassociateEipAddress

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID to unbind |

---

## 4. Bandwidth Management

### 4.1 ModifyEipBandwidth

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID |
| `Bandwidth` | Integer | Yes | New bandwidth (Mbps) |

**Bandwidth Range:**

| Billing Type | Range |
|--------------|-------|
| PayByTraffic | 1–200 Mbps |
| PayByBandwidth | 1–500 Mbps |

### 4.2 DescribeEipBandwidth

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID |

**Response (Result):**

| Field | Type | Description |
|-------|------|-------------|
| `Bandwidth` | Integer | Current bandwidth (Mbps) |
| `BandwidthUnit` | String | Unit (`Mbps`) |

---

## 5. EIP Attribute Operations

### 5.1 ModifyEipAddressAttributes

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID |
| `Name` | String | No | New name |
| `Description` | String | No | New description |

### 5.2 TagEipAddress

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID |
| `Tags.N.Key` | String | Yes | Tag key |
| `Tags.N.Value` | String | Yes | Tag value |

### 5.3 RenewEipAddress

**Request Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `AllocationId` | String | Yes | EIP ID |
| `RenewalPeriod` | String | Yes | Renewal period | `Month`, `Year` |
| `Quantity` | Integer | Yes | Number of periods | — |

---

## 6. Go SDK Usage

### JIT Go SDK Pattern

```go
// main.go — AllocateEipAddress example
package main

import (
    "fmt"
    "os"

    "github.com/volcengine/volc-sdk-golang/base"
)

func main() {
    client := newEipClient()

    body := make(map[string]interface{})
    body["LineType"] = "BGP"
    body["Bandwidth"] = 5
    body["Name"] = os.Getenv("EIP_NAME")

    resp, err := client.Request("AllocateEipAddress", nil, body)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }

    fmt.Println(string(resp))
}

func newEipClient() *base.Client {
    // same pattern as VPC, ServiceName = "eip"
    ...
}
```

---

## 7. Response Parsing Patterns

### Extract EIP ID and Address

```bash
EIP_ID=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.AllocationId')
EIP_ADDR=$(ve eip AllocateEipAddress --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.EipAddress')
```

### List EIPs with Binding Status

```bash
ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" \
  | jq -r '.Result.EipAddresses[] |
    "\(.AllocationId)\t\(.EipAddress)\t\(.Status)\t\(.Bandwidth)Mbps\t\(.InstanceType // "none")\t\(.InstanceId // "-")"'
```

### Poll Until Available

```bash
for i in {1..30}; do
    STATUS=$(ve eip DescribeEipAddresses --Region "$VOLCENGINE_REGION" \
      --AllocationIds "[\"$EIP_ID\"]" | jq -r '.Result.EipAddresses[0].Status')
    [ "$STATUS" = "Available" ] && break
    sleep 2
done
```

---

## 7. Go SDK Examples

### AllocateEipAddress

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/eip"
)

func main() {
    instance := eip.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.AllocateEipAddress(&eip.AllocateEipAddressInput{
        LineType:  "BGP",
        Bandwidth: 5,
        Name:      "web-eip",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("EIP: %s (%s)\n", resp.Result.EipAddress, resp.Result.AllocationId)
}
```

### DescribeEipAddresses

```go
func listEIPs() {
    instance := eip.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    resp, err := instance.DescribeEipAddresses(&eip.DescribeEipAddressesInput{})
    if err != nil {
        panic(err)
    }
    for _, e := range resp.Result.EipAddresses {
        bound := "-"
        if e.InstanceId != "" {
            bound = fmt.Sprintf("%s:%s", e.InstanceType, e.InstanceId)
        }
        fmt.Printf("%s: %s status=%s bw=%dMbps bound=%s\n",
            e.AllocationId, e.EipAddress, e.Status, e.Bandwidth, bound)
    }
}
```

### AssociateEipAddress

```go
func associateEIP(eipID, instanceID, instanceType string) {
    instance := eip.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    _, err := instance.AssociateEipAddress(&eip.AssociateEipAddressInput{
        AllocationId: eipID,
        InstanceId:   instanceID,
        InstanceType: instanceType,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("EIP %s bound to %s (%s)\n", eipID, instanceID, instanceType)
}
```

### ModifyEipBandwidth

```go
func modifyBandwidth(eipID string, bandwidth int) {
    instance := eip.NewInstance()
    instance.SetCredential(os.Getenv("VOLCENGINE_ACCESS_KEY"), os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.SetRegion(os.Getenv("VOLCENGINE_REGION"))

    _, err := instance.ModifyEipBandwidth(&eip.ModifyEipBandwidthInput{
        AllocationId: eipID,
        Bandwidth:    bandwidth,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Bandwidth updated to %d Mbps for EIP %s\n", bandwidth, eipID)
}
```

---

*This reference document is part of the ve-eip-ops skill. For operational procedures, see the main SKILL.md.*
