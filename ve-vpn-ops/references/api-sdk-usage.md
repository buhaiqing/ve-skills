# API & SDK — VPN

## OpenAPI Reference

- **Service:** VPN
- **API Version:** 2020-04-01
- **Endpoint:** `vpc.volcengineapi.com`
- **Protocol:** HTTPS
- **Documentation:** https://www.volcengine.com/docs/6491/130519

## SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/vpn"
```

## API Operations Map

### VPN Gateway Operations

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create VPN Gateway | CreateVpnGateway | CreateVpnGateway | `ve vpn CreateVpnGateway` |
| Describe VPN Gateways | DescribeVpnGateways | DescribeVpnGateways | `ve vpn DescribeVpnGateways` |
| Delete VPN Gateway | DeleteVpnGateway | DeleteVpnGateway | `ve vpn DeleteVpnGateway` |
| Modify VPN Gateway | ModifyVpnGatewayAttribute | ModifyVpnGatewayAttribute | `ve vpn ModifyVpnGatewayAttribute` |

### Customer Gateway Operations

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create Customer Gateway | CreateCustomerGateway | CreateCustomerGateway | `ve vpn CreateCustomerGateway` |
| Describe Customer Gateways | DescribeCustomerGateways | DescribeCustomerGateways | `ve vpn DescribeCustomerGateways` |
| Delete Customer Gateway | DeleteCustomerGateway | DeleteCustomerGateway | `ve vpn DeleteCustomerGateway` |
| Modify Customer Gateway | ModifyCustomerGatewayAttribute | ModifyCustomerGatewayAttribute | `ve vpn ModifyCustomerGatewayAttribute` |

### IPSec Connection Operations

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create IPSec Connection | CreateVpnConnection | CreateVpnConnection | `ve vpn CreateVpnConnection` |
| Describe IPSec Connections | DescribeVpnConnections | DescribeVpnConnections | `ve vpn DescribeVpnConnections` |
| Delete IPSec Connection | DeleteVpnConnection | DeleteVpnConnection | `ve vpn DeleteVpnConnection` |
| Modify IPSec Connection | ModifyVpnConnectionAttribute | ModifyVpnConnectionAttribute | `ve vpn ModifyVpnConnectionAttribute` |

### SSL VPN Server Operations

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create SSL VPN Server | CreateSslVpnServer | CreateSslVpnServer | `ve vpn CreateSslVpnServer` |
| Describe SSL VPN Servers | DescribeSslVpnServers | DescribeSslVpnServers | `ve vpn DescribeSslVpnServers` |
| Delete SSL VPN Server | DeleteSslVpnServer | DeleteSslVpnServer | `ve vpn DeleteSslVpnServer` |
| Modify SSL VPN Server | ModifySslVpnServer | ModifySslVpnServer | `ve vpn ModifySslVpnServer` |

### SSL VPN Client Operations

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create Client Cert | CreateSslVpnClientCert | CreateSslVpnClientCert | `ve vpn CreateSslVpnClientCert` |
| Describe Client Certs | DescribeSslVpnClientCerts | DescribeSslVpnClientCerts | `ve vpn DescribeSslVpnClientCerts` |
| Delete Client Cert | DeleteSslVpnClientCert | DeleteSslVpnClientCert | `ve vpn DeleteSslVpnClientCert` |
| Download Client Config | DownloadSslVpnClientConfig | DownloadSslVpnClientConfig | `ve vpn DownloadSslVpnClientConfig` |

## Request/Response Examples

### CreateVpnGateway

**Request:**
```json
{
  "Region": "cn-beijing",
  "VpcId": "vpc-2feppibx2c6kobss3p442****",
  "SubnetId": "subnet-2fel1rkvk9q3w2i7e1p3z****",
  "Bandwidth": 10,
  "VpnGatewayName": "prod-vpn-gateway",
  "Description": "Production VPN Gateway",
  "ChargeType": "PostPaid"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567890",
    "Action": "CreateVpnGateway",
    "Version": "2020-04-01",
    "Service": "vpn",
    "Region": "cn-beijing"
  },
  "Result": {
    "VpnGatewayId": "vgw-2fdcrq7jvi6co5tfc2n8h****"
  }
}
```

### DescribeVpnGateways

**Request:**
```json
{
  "Region": "cn-beijing",
  "VpnGatewayIds": ["vgw-2fdcrq7jvi6co5tfc2n8h****"],
  "MaxResults": 10
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567891",
    "Action": "DescribeVpnGateways",
    "Version": "2020-04-01"
  },
  "Result": {
    "TotalCount": 1,
    "VpnGateways": [
      {
        "VpnGatewayId": "vgw-2fdcrq7jvi6co5tfc2n8h****",
        "VpnGatewayName": "prod-vpn-gateway",
        "VpcId": "vpc-2feppibx2c6kobss3p442****",
        "SubnetId": "subnet-2fel1rkvk9q3w2i7e1p3z****",
        "Bandwidth": 10,
        "Status": "Available",
        "ChargeType": "PostPaid",
        "CreationTime": "2024-05-27T12:00:00Z",
        "ExpiredTime": "2099-12-31T23:59:59Z",
        "Description": "Production VPN Gateway"
      }
    ]
  }
}
```

### CreateVpnConnection

**Request:**
```json
{
  "Region": "cn-beijing",
  "VpnGatewayId": "vgw-2fdcrq7jvi6co5tfc2n8h****",
  "CustomerGatewayId": "cgw-2fdcrq7jvi6co5tfc2n8i****",
  "VpnConnectionName": "hq-office-connection",
  "PreSharedKey": "<masked>",
  "LocalSubnet": ["10.0.0.0/16"],
  "RemoteSubnet": ["192.168.0.0/16"],
  "IKEConfig": {
    "Version": "ikev2",
    "Mode": "main",
    "EncAlg": "aes256",
    "AuthAlg": "sha256",
    "DhGroup": "group14",
    "Lifetime": 86400
  },
  "IPsecConfig": {
    "EncAlg": "aes256",
    "AuthAlg": "sha256",
    "Pfs": "group14",
    "Lifetime": 3600
  }
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567892"
  },
  "Result": {
    "VpnConnectionId": "vpn-2fdcrq7jvi6co5tfc2n8j****"
  }
}
```

### DescribeVpnConnections

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567893"
  },
  "Result": {
    "TotalCount": 1,
    "VpnConnections": [
      {
        "VpnConnectionId": "vpn-2fdcrq7jvi6co5tfc2n8j****",
        "VpnConnectionName": "hq-office-connection",
        "VpnGatewayId": "vgw-2fdcrq7jvi6co5tfc2n8h****",
        "CustomerGatewayId": "cgw-2fdcrq7jvi6co5tfc2n8i****",
        "LocalSubnet": ["10.0.0.0/16"],
        "RemoteSubnet": ["192.168.0.0/16"],
        "Status": "Available",
        "IKEConfig": {
          "Version": "ikev2",
          "Mode": "main",
          "EncAlg": "aes256",
          "AuthAlg": "sha256",
          "DhGroup": "group14",
          "Lifetime": 86400
        },
        "IPsecConfig": {
          "EncAlg": "aes256",
          "AuthAlg": "sha256",
          "Pfs": "group14",
          "Lifetime": 3600
        },
        "CreationTime": "2024-05-27T12:00:00Z"
      }
    ]
  }
}
```

### CreateSslVpnServer

**Request:**
```json
{
  "Region": "cn-beijing",
  "VpnGatewayId": "vgw-2fdcrq7jvi6co5tfc2n8h****",
  "SslVpnServerName": "remote-access-server",
  "LocalSubnets": ["10.0.0.0/16"],
  "ClientIpPool": "10.0.100.0/24",
  "SslVpnServerProtocol": "UDP",
  "SslVpnServerPort": 1194,
  "Cipher": "AES-256-CBC",
  "Auth": "SHA256",
  "Compress": false
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567894"
  },
  "Result": {
    "SslVpnServerId": "ssl-2fdcrq7jvi6co5tfc2n8k****"
  }
}
```

### CreateSslVpnClientCert

**Request:**
```json
{
  "Region": "cn-beijing",
  "SslVpnServerId": "ssl-2fdcrq7jvi6co5tfc2n8k****",
  "SslVpnClientCertName": "user-laptop-cert"
}
```

**Response:**
```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567895"
  },
  "Result": {
    "SslVpnClientCertId": "sc-2fdcrq7jvi6co5tfc2n8l****",
    "Certificate": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIQC2Z...\n-----END CERTIFICATE-----",
    "PrivateKey": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0B...\n-----END PRIVATE KEY-----",
    "CaCert": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIQ...\n-----END CERTIFICATE-----"
  }
}
```

## Field Reference

### VpnGateway

| Field | Type | Description |
|-------|------|-------------|
| VpnGatewayId | string | Gateway ID (vgw-xxx) |
| VpnGatewayName | string | Gateway name |
| VpcId | string | Associated VPC ID |
| SubnetId | string | Associated Subnet ID |
| Bandwidth | int | Bandwidth in Mbps |
| Status | string | Available, Pending, Deleting, etc. |
| ChargeType | string | PostPaid or PrePaid |
| CreationTime | string | ISO 8601 timestamp |
| ExpiredTime | string | ISO 8601 timestamp |
| Description | string | Description |

### VpnConnection

| Field | Type | Description |
|-------|------|-------------|
| VpnConnectionId | string | Connection ID (vpn-xxx) |
| VpnConnectionName | string | Connection name |
| VpnGatewayId | string | VPN Gateway ID |
| CustomerGatewayId | string | Customer Gateway ID |
| LocalSubnet | []string | Local CIDR blocks |
| RemoteSubnet | []string | Remote CIDR blocks |
| Status | string | Connection status |
| IKEConfig | object | IKE configuration |
| IPsecConfig | object | IPsec configuration |
| CreationTime | string | ISO 8601 timestamp |

### IKEConfig

| Field | Type | Options |
|-------|------|---------|
| Version | string | ikev1, ikev2 |
| Mode | string | main, aggressive (IKEv1 only) |
| EncAlg | string | aes128, aes192, aes256 |
| AuthAlg | string | md5, sha1, sha256, sha384, sha512 |
| DhGroup | string | group1, group2, group5, group14, group24 |
| Lifetime | int | 900-86400 seconds |

### IPsecConfig

| Field | Type | Options |
|-------|------|---------|
| EncAlg | string | aes128, aes192, aes256 |
| AuthAlg | string | md5, sha1, sha256, sha384, sha512 |
| Pfs | string | disabled, group1, group2, group5, group14 |
| Lifetime | int | 900-86400 seconds |

## Pagination

List operations support pagination:

| Parameter | Type | Description |
|-----------|------|-------------|
| MaxResults | int | Maximum results per page (1-100) |
| NextToken | string | Token for next page |

**Response pagination fields:**
| Field | Type | Description |
|-------|------|-------------|
| TotalCount | int | Total matching results |
| NextToken | string | Token for next page (null if last) |

## Error Response Format

```json
{
  "ResponseMetadata": {
    "RequestId": "2024052712345678901234567890"
  },
  "Error": {
    "Code": "InvalidParameter",
    "Message": "The request parameter is invalid"
  }
}
```

## IAM Permissions Required

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "vpn:CreateVpnGateway",
        "vpn:DescribeVpnGateways",
        "vpn:DeleteVpnGateway",
        "vpn:ModifyVpnGatewayAttribute",
        "vpn:CreateCustomerGateway",
        "vpn:DescribeCustomerGateways",
        "vpn:DeleteCustomerGateway",
        "vpn:CreateVpnConnection",
        "vpn:DescribeVpnConnections",
        "vpn:DeleteVpnConnection",
        "vpn:ModifyVpnConnectionAttribute",
        "vpn:CreateSslVpnServer",
        "vpn:DescribeSslVpnServers",
        "vpn:DeleteSslVpnServer",
        "vpn:CreateSslVpnClientCert",
        "vpn:DescribeSslVpnClientCerts",
        "vpn:DeleteSslVpnClientCert"
      ],
      "Resource": "*"
    }
  ]
}
```
