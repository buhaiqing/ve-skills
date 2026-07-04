# CLI — VPN (`ve`)

> **jq Paths (`.Result.*`):** `.VpnGateways[0]` (first gateway), `.VpnGateways[]` (gateway ID list), `.VpnConnections[]` (connection list)

## Install and Config

- Install: see [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json` (JSON format).
- Environment variables are preferred for Agent execution.

### Credential Verification

```bash
# Verify credentials exist (never echo the actual values)
test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ Credentials configured"
test -n "$VOLCENGINE_REGION" && echo "✅ Region configured: $VOLCENGINE_REGION"
```

### Masking Example

```bash
# Safe: Only show that secret key exists, not its value
echo "VOLCENGINE_SECRET_KEY=<masked>"

# Unsafe: Never do this
echo "VOLCENGINE_SECRET_KEY=$VOLCENGINE_SECRET_KEY"  # ❌ NEVER
```

## Conventions (Agent Execution)

- Output is **JSON by default**
- Document **exact** JSON paths after verifying with real invocation
- CLI invocation: `ve vpn <action> --parameter value`
- Arrays passed as JSON strings: `'["item1","item2"]'`

## CLI vs API Coverage Gap

| Operation (API / SDK) | Available via `ve`? | Notes |
|------------------------|---------------------|-------|
| CreateVpnGateway | yes | Full support |
| DescribeVpnGateways | yes | Full support |
| DeleteVpnGateway | yes | Full support |
| ModifyVpnGatewayAttribute | yes | Full support |
| CreateCustomerGateway | yes | Full support |
| DescribeCustomerGateways | yes | Full support |
| DeleteCustomerGateway | yes | Full support |
| ModifyCustomerGatewayAttribute | yes | Full support |
| CreateVpnConnection | yes | Full support |
| DescribeVpnConnections | yes | Full support |
| DeleteVpnConnection | yes | Full support |
| ModifyVpnConnectionAttribute | yes | Full support |
| CreateSslVpnServer | yes | Full support |
| DescribeSslVpnServers | yes | Full support |
| DeleteSslVpnServer | yes | Full support |
| ModifySslVpnServer | yes | Full support |
| CreateSslVpnClientCert | yes | Full support |
| DescribeSslVpnClientCerts | yes | Full support |
| DeleteSslVpnClientCert | yes | Full support |
| DownloadSslVpnClientConfig | yes | Full support |

**Coverage Status:** Complete — all VPN operations are supported via `ve` CLI.

## Command Map

### VPN Gateway Commands

| Goal | Example `ve` Invocation | Notes |
|------|--------------------------|-------|
| List gateways | `ve vpn DescribeVpnGateways --Region cn-beijing` | JSON output by default |
| Get gateway by ID | `ve vpn DescribeVpnGateways --Region cn-beijing --VpnGatewayIds '["vgw-xxxxx"]'` | Filter by ID |
| Filter by VPC | `ve vpn DescribeVpnGateways --Region cn-beijing --VpcId vpc-xxxxx` | Filter by VPC |
| Create gateway | `ve vpn CreateVpnGateway --Region cn-beijing --VpcId vpc-xx --SubnetId subnet-xx --Bandwidth 10 --VpnGatewayName my-vpn` | Returns gateway ID |
| Delete gateway | `ve vpn DeleteVpnGateway --Region cn-beijing --VpnGatewayId vgw-xxxxx` | Irreversible |
| Modify gateway | `ve vpn ModifyVpnGatewayAttribute --Region cn-beijing --VpnGatewayId vgw-xx --VpnGatewayName new-name` | Update attributes |

### Customer Gateway Commands

| Goal | Example `ve` Invocation | Notes |
|------|--------------------------|-------|
| List gateways | `ve vpn DescribeCustomerGateways --Region cn-beijing` | All customer gateways |
| Get by ID | `ve vpn DescribeCustomerGateways --Region cn-beijing --CustomerGatewayIds '["cgw-xxxxx"]'` | Filter by ID |
| Create gateway | `ve vpn CreateCustomerGateway --Region cn-beijing --IpAddress 203.0.113.1 --CustomerGatewayName hq-gateway` | IP must be public |
| Delete gateway | `ve vpn DeleteCustomerGateway --Region cn-beijing --CustomerGatewayId cgw-xxxxx` | Cannot have connections |
| Modify gateway | `ve vpn ModifyCustomerGatewayAttribute --Region cn-beijing --CustomerGatewayId cgw-xx --CustomerGatewayName new-name` | Update name/desc |

### IPSec Connection Commands

| Goal | Example `ve` Invocation | Notes |
|------|--------------------------|-------|
| List connections | `ve vpn DescribeVpnConnections --Region cn-beijing` | All connections |
| Get by ID | `ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]'` | Filter by ID |
| Filter by gateway | `ve vpn DescribeVpnConnections --Region cn-beijing --VpnGatewayId vgw-xxxxx` | Filter by VPN GW |
| Create connection | `ve vpn CreateVpnConnection --Region cn-beijing --VpnGatewayId vgw-xx --CustomerGatewayId cgw-xx --VpnConnectionName conn --PreSharedKey "<masked>" --LocalSubnet '["10.0.0.0/16"]' --RemoteSubnet '["192.168.0.0/16"]'` | Complex params |
| Delete connection | `ve vpn DeleteVpnConnection --Region cn-beijing --VpnConnectionId vpn-xxxxx` | Terminates tunnel |
| Modify connection | `ve vpn ModifyVpnConnectionAttribute --Region cn-beijing --VpnConnectionId vpn-xx --VpnConnectionName new-name` | Update attributes |

### SSL VPN Server Commands

| Goal | Example `ve` Invocation | Notes |
|------|--------------------------|-------|
| List servers | `ve vpn DescribeSslVpnServers --Region cn-beijing` | All SSL VPN servers |
| Get by ID | `ve vpn DescribeSslVpnServers --Region cn-beijing --SslVpnServerIds '["ssl-xxxxx"]'` | Filter by ID |
| Create server | `ve vpn CreateSslVpnServer --Region cn-beijing --VpnGatewayId vgw-xx --SslVpnServerName server --LocalSubnets '["10.0.0.0/16"]' --ClientIpPool 10.0.100.0/24` | Requires SSL-enabled GW |
| Delete server | `ve vpn DeleteSslVpnServer --Region cn-beijing --SslVpnServerId ssl-xxxxx` | Cannot have clients |
| Modify server | `ve vpn ModifySslVpnServer --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnServerName new-name` | Update configuration |

### SSL VPN Client Commands

| Goal | Example `ve` Invocation | Notes |
|------|--------------------------|-------|
| List certs | `ve vpn DescribeSslVpnClientCerts --Region cn-beijing` | All client certificates |
| List by server | `ve vpn DescribeSslVpnClientCerts --Region cn-beijing --SslVpnServerId ssl-xxxxx` | Filter by server |
| Create cert | `ve vpn CreateSslVpnClientCert --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnClientCertName user-laptop` | Returns cert+key |
| Delete cert | `ve vpn DeleteSslVpnClientCert --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnClientCertId sc-xxxxx` | Revokes access |
| Download config | `ve vpn DownloadSslVpnClientConfig --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnClientCertId sc-xx` | Get client config file |

## Complex Parameter Examples

### Create VPN Connection with Full IKE/IPsec Config

```bash
ve vpn CreateVpnConnection \
  --Region cn-beijing \
  --VpnGatewayId "vgw-2fdcrq7jvi6co5tfc2n8h****" \
  --CustomerGatewayId "cgw-2fdcrq7jvi6co5tfc2n8i****" \
  --VpnConnectionName "hq-office-connection" \
  --PreSharedKey "<masked>" \
  --LocalSubnet '["10.0.0.0/16","10.1.0.0/16"]' \
  --RemoteSubnet '["192.168.0.0/16","192.168.1.0/24"]' \
  --IKEConfig '{"Version":"ikev2","Mode":"main","EncAlg":"aes256","AuthAlg":"sha256","DhGroup":"group14","Lifetime":86400}' \
  --IPsecConfig '{"EncAlg":"aes256","AuthAlg":"sha256","Pfs":"group14","Lifetime":3600}'
```

### Create SSL VPN Server

```bash
ve vpn CreateSslVpnServer \
  --Region cn-beijing \
  --VpnGatewayId "vgw-2fdcrq7jvi6co5tfc2n8h****" \
  --SslVpnServerName "remote-access-server" \
  --LocalSubnets '["10.0.0.0/16"]' \
  --ClientIpPool "10.0.100.0/24" \
  --SslVpnServerProtocol "UDP" \
  --SslVpnServerPort 1194 \
  --Cipher "AES-256-CBC" \
  --Auth "SHA256" \
  --Compress false
```

## Help Commands

```bash
# List all VPN actions
ve vpn --help

# Help for specific action
ve vpn CreateVpnGateway --help
ve vpn DescribeVpnConnections --help
ve vpn CreateSslVpnServer --help
```

## JSON Output Processing

```bash
# Extract VPN Gateway ID from creation response
VGW_ID=$(ve vpn CreateVpnGateway --Region cn-beijing ... | jq -r '.Result.VpnGatewayId')

# Extract status from describe
STATUS=$(ve vpn DescribeVpnGateways --Region cn-beijing --VpnGatewayIds "[\"$VGW_ID\"]" | jq -r '.Result.VpnGateways[0].Status')

# List all gateway IDs
ve vpn DescribeVpnGateways --Region cn-beijing | jq -r '.Result.VpnGateways[].VpnGatewayId'

# Filter connections by status
ve vpn DescribeVpnConnections --Region cn-beijing | jq '.Result.VpnConnections[] | select(.Status == "Available")'
```

## Pagination

```bash
# First page
ve vpn DescribeVpnGateways --Region cn-beijing --MaxResults 10

# Next page (using NextToken from previous response)
ve vpn DescribeVpnGateways --Region cn-beijing --MaxResults 10 --NextToken "token-from-previous-response"
```

## Common Filters

```bash
# Filter by VPC
ve vpn DescribeVpnGateways --Region cn-beijing --VpcId vpc-xxxxx

# Filter by subnet
ve vpn DescribeVpnGateways --Region cn-beijing --SubnetId subnet-xxxxx

# Filter by status
ve vpn DescribeVpnConnections --Region cn-beijing | jq '.Result.VpnConnections[] | select(.Status == "Available")'

# Filter SSL VPN servers by gateway
ve vpn DescribeSslVpnServers --Region cn-beijing --VpnGatewayId vgw-xxxxx
```

## Security Notes

| Item | Safe | Unsafe |
|------|------|--------|
| Pre-shared key in command | Use placeholder `<masked>` | Never use real PSK in logs |
| Secret key env var | `test -n "$VOLCENGINE_SECRET_KEY"` | `echo $VOLCENGINE_SECRET_KEY` |
| Certificate output | Save to file with 600 permissions | Display in console output |
| Private key | Save securely, never commit | Store in version control |
