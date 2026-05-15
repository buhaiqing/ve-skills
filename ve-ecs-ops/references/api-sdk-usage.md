# API & SDK — ECS

## OpenAPI

- **API Version:** 2020-04-01
- **Endpoint:** `ecs.volcengineapi.com` (or default `open.volcengineapi.com`)
- **Protocol:** HTTPS POST with JSON or form-encoded body
- **Region:** Required for most operations via `Region` parameter
- **Pagination:** Uses `MaxResults` (max 100) and `NextToken` for all List operations
- **Documentation:** https://www.volcengine.com/docs/6396/69513

## SDK Package

```
github.com/volcengine/volc-sdk-golang/service/ecs
```

## SDK Operations Map

| Goal | API Action | SDK Method |
|------|-----------|------------|
| Create instances | `RunInstances` | `instance.Client.Request("RunInstances", nil, params)` |
| Query instances | `DescribeInstances` | `instance.Client.Request("DescribeInstances", nil, params)` |
| Start instance | `StartInstance` | `instance.Client.Request("StartInstance", nil, params)` |
| Start instances (batch) | `StartInstances` | `instance.Client.Request("StartInstances", nil, params)` |
| Stop instance | `StopInstance` | `instance.Client.Request("StopInstance", nil, params)` |
| Stop instances (batch) | `StopInstances` | `instance.Client.Request("StopInstances", nil, params)` |
| Reboot instance | `RebootInstance` | `instance.Client.Request("RebootInstance", nil, params)` |
| Reboot instances (batch) | `RebootInstances` | `instance.Client.Request("RebootInstances", nil, params)` |
| Delete instance | `DeleteInstance` | `instance.Client.Request("DeleteInstance", nil, params)` |
| Delete instances (batch) | `DeleteInstances` | `instance.Client.Request("DeleteInstances", nil, params)` |
| Resize instance | `ModifyInstanceSpec` | `instance.Client.Request("ModifyInstanceSpec", nil, params)` |
| Modify attributes | `ModifyInstanceAttribute` | `instance.Client.Request("ModifyInstanceAttribute", nil, params)` |
| Create snapshot | `CreateSnapshot` | `instance.Client.Request("CreateSnapshot", nil, params)` |
| Describe snapshots | `DescribeSnapshots` | `instance.Client.Request("DescribeSnapshots", nil, params)` |
| Create image | `CreateImage` | `instance.Client.Request("CreateImage", nil, params)` |
| Describe images | `DescribeImages` | `instance.Client.Request("DescribeImages", nil, params)` |
| Delete image | `DeleteImage` | `instance.Client.Request("DeleteImage", nil, params)` |
| Create key pair | `CreateKeyPair` | `instance.Client.Request("CreateKeyPair", nil, params)` |
| Describe key pairs | `DescribeKeyPairs` | `instance.Client.Request("DescribeKeyPairs", nil, params)` |
| Import key pair | `ImportKeyPair` | `instance.Client.Request("ImportKeyPair", nil, params)` |
| Delete key pair | `DeleteKeyPair` | `instance.Client.Request("DeleteKeyPair", nil, params)` |
| Describe regions | `DescribeRegions` | `instance.Client.Request("DescribeRegions", nil, params)` |
| Describe zones | `DescribeZones` | `instance.Client.Request("DescribeZones", nil, params)` |
| Describe instance types | `DescribeInstanceTypes` | `instance.Client.Request("DescribeInstanceTypes", nil, params)` |
| Describe resource quota | `DescribeResourceQuota` | `instance.Client.Request("DescribeResourceQuota", nil, params)` |
| Attach volume | `AttachVolume` | `instance.Client.Request("AttachVolume", nil, params)` |
| Detach volume | `DetachVolume` | `instance.Client.Request("DetachVolume", nil, params)` |
| Create volume | `CreateVolume` | `instance.Client.Request("CreateVolume", nil, params)` |
| Describe volumes | `DescribeVolumes` | `instance.Client.Request("DescribeVolumes", nil, params)` |
| Delete volume | `DeleteVolume` | `instance.Client.Request("DeleteVolume", nil, params)` |

## Request / Response Notes

### Common Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `Action` | string | Yes | API action name |
| `Version` | string | Yes | API version: `2020-04-01` |
| `Region` | string | Yes | Region ID (e.g., `cn-beijing`) |

### RunInstances Key Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `InstanceType` | string | Yes | — | Instance spec (e.g., `ecs.g3i.large`) |
| `ImageId` | string | Yes | — | Image ID (e.g., `image-xxxxx`) |
| `VpcId` | string | Yes | — | VPC ID |
| `SubnetId` | string | Yes | — | Subnet ID |
| `InstanceName` | string | No | — | Instance name (1-64 chars) |
| `Password` | string | No | — | Admin password |
| `KeyPairName` | string | No | — | SSH key pair name |
| `ChargeType` | string | No | `PostPaid` | `PostPaid`, `PrePaid`, `Spot` |
| `SecurityGroupIds` | array | Yes | — | Security group ID list |
| `Amount` | integer | No | 1 | Number of instances |
| `ClientToken` | string | No | — | Idempotency token (1-64 chars) |

### RunInstances Response

```json
{
  "ResponseMetadata": {
    "RequestId": "request-id-string",
    "Action": "RunInstances",
    "Version": "2020-04-01",
    "Service": "ecs",
    "Region": "cn-beijing"
  },
  "Result": {
    "InstanceIds": ["i-xxxxxxxx"],
    "OrderId": ""
  }
}
```

### DescribeInstances Response

```json
{
  "ResponseMetadata": { "RequestId": "..." },
  "Result": {
    "Instances": {
      "Instance": [{
        "InstanceId": "i-xxxxxxxx",
        "InstanceName": "my-server",
        "InstanceType": "ecs.g3i.large",
        "Status": "RUNNING",
        "PrivateIpAddress": "192.168.1.10",
        "PublicIpAddress": "120.0.0.1",
        "ImageId": "image-xxxxxxxx",
        "VpcId": "vpc-xxxxxxxx",
        "SubnetId": "subnet-xxxxxxxx",
        "ZoneId": "cn-beijing-a",
        "CreatedAt": "2026-01-15T10:00:00Z",
        "ExpiredAt": "2027-01-15T10:00:00Z"
      }]
    },
    "TotalCount": 1,
    "NextToken": ""
  }
}
```

### Pagination

All list operations support:
- `MaxResults`: Maximum items per response (1-100)
- `NextToken`: Opaque token for next page

**Pagination pattern:**
```bash
# First page
ve ecs DescribeInstances --Region cn-beijing --MaxResults 50

# Subsequent pages (use NextToken from previous response)
ve ecs DescribeInstances --Region cn-beijing --MaxResults 50 --NextToken "previous-token"
```
