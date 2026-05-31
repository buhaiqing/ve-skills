# Troubleshooting VPN

## Common API Error Codes

| Code / HTTP | Meaning | Agent Action |
|-------------|---------|--------------|
| `InvalidParameter` / 400 | Request failed validation | Align body with OpenAPI; check parameter types and formats |
| `InvalidVpcId.NotFound` / 400 | VPC does not exist | Verify VPC ID via `DescribeVpcs`; create VPC first |
| `InvalidSubnetId.NotFound` / 400 | Subnet does not exist | Verify Subnet ID via `DescribeSubnets`; create subnet first |
| `InvalidVpnGatewayId.NotFound` / 400 | VPN Gateway not found | Verify gateway ID; list via `DescribeVpnGateways` |
| `InvalidCustomerGatewayId.NotFound` / 400 | Customer Gateway not found | Verify gateway ID; list via `DescribeCustomerGateways` |
| `InvalidVpnConnectionId.NotFound` / 400 | IPSec connection not found | Verify connection ID |
| `InvalidSslVpnServerId.NotFound` / 400 | SSL VPN Server not found | Verify server ID |
| `InvalidSslVpnClientCertId.NotFound` / 400 | SSL VPN client cert not found | Verify certificate ID |
| `VpnGateway.NotAvailable` / 400 | VPN Gateway not in Available state | Wait for gateway to become Available |
| `CustomerGateway.NotAvailable` / 400 | Customer Gateway not in Available state | Wait for gateway to become Available |
| `QuotaExceeded.VpnGateway` / 400 | VPN Gateway quota reached | Request quota increase or delete unused gateways |
| `QuotaExceeded.VpnConnection` / 400 | IPSec connection quota reached | Request quota increase or delete unused connections |
| `QuotaExceeded.SslVpnServer` / 400 | SSL VPN Server quota reached | Delete unused servers |
| `QuotaExceeded.SslVpnClientCert` / 400 | SSL VPN client cert quota reached | Delete unused certificates |
| `InvalidCidrBlock.Malformed` / 400 | CIDR format invalid | Use valid CIDR notation (e.g., 10.0.0.0/16) |
| `CidrBlock.Conflict` / 400 | CIDR overlaps with existing resource | Choose non-overlapping CIDR |
| `InvalidBandwidth.ValueNotSupported` / 400 | Bandwidth value not supported | Use value between 1-1000 Mbps |
| `InvalidPreSharedKey.Format` / 400 | PSK format invalid | Use 1-128 characters |
| `SubnetConflict` / 400 | Local and remote subnets overlap | Ensure subnets do not overlap |
| `InvalidIpAddress.Format` / 400 | IP address format invalid | Use valid IPv4 address |
| `InvalidIpAddress.NotPublic` / 400 | IP is not public | Customer Gateway requires public IP |
| `InvalidVpnGatewayName.Duplicate` / 400 | Gateway name exists | Use unique name |
| `ResourceInUse.VpnGateway` / 400 | Gateway has active connections | Delete connections first |
| `ResourceInUse.CustomerGateway` / 400 | Gateway has active connections | Delete connections first |
| `ResourceInUse.SslVpnServer` / 400 | Server has client certificates | Delete certificates first |
| `InsufficientBalance` / 400 | Account balance insufficient | Recharge account |
| `Forbidden.RAM` / 403 | Insufficient IAM permissions | Add VPN permissions to IAM policy |
| `Unauthorized` / 403 | Request not authorized | Check credentials and permissions |
| `InternalError` / 500 | Server-side error | Retry with backoff; then HALT |
| `ServiceUnavailable` / 503 | Service temporarily unavailable | Retry with backoff |
| Throttling / 429 | Rate limit exceeded | Back off and retry |

## Diagnostic Order

### 1. VPN Gateway Issues

```bash
# Check gateway status
ve vpn DescribeVpnGateways --Region cn-beijing --VpnGatewayIds '["vgw-xxxxx"]'

# Expected response path: $.Result.VpnGateways[0].Status
# Should be: "Available"
```

**Common Issues:**
- **Status stuck in "Pending"**: Wait 5 minutes; if still pending, contact support
- **Status "Deleting" for long time**: Check for active connections; deletion requires connections to be removed first

### 2. IPSec Connection Issues

```bash
# Check connection status
ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]'

# Check gateway status
ve vpn DescribeVpnGateways --Region cn-beijing --VpnGatewayIds '["vgw-xxxxx"]'

# Check customer gateway
ve vpn DescribeCustomerGateways --Region cn-beijing --CustomerGatewayIds '["cgw-xxxxx"]'
```

**Connection Status Troubleshooting:**

| Status | Meaning | Action |
|--------|---------|--------|
| `Available` | Tunnel established | Normal operation |
| `Pending` | Creating/updating | Wait for completion |
| `Modifying` | Configuration updating | Wait for completion |
| `Deleting` | Being deleted | Wait for completion |
| `Deleted` | Deleted | Create new connection if needed |
| `Error` | Error state | Check configuration; retry creation |
| `Unavailable` | Tunnel down | Check IKE/IPsec config; verify peer is reachable |

### 3. IKE/IPsec Negotiation Failures

**Symptoms:**
- Connection status shows `Unavailable`
- No traffic flowing through tunnel
- Logs show negotiation failures

**Diagnostic Steps:**

1. **Verify IKE Configuration Match:**
   ```bash
   # Compare local and remote IKE config
   ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]'
   ```
   Must match exactly:
   - IKE version (v1 or v2)
   - Encryption algorithm
   - Authentication algorithm
   - DH Group
   - Lifetime

2. **Verify Pre-Shared Key:**
   - PSK must be identical on both sides
   - Check for whitespace or special characters
   - Never log the actual PSK value

3. **Verify Network Reachability:**
   ```bash
   # From on-premises gateway, test connectivity
   ping <VPN_Gateway_IP>
   telnet <VPN_Gateway_IP> 500  # IKE port
   telnet <VPN_Gateway_IP> 4500  # NAT-T port
   ```

4. **Check IPsec Configuration Match:**
   - Encryption algorithm
   - Authentication algorithm
   - PFS group
   - Lifetime

**Common IKE/IPsec Mismatches:**

| Issue | Local Config | Remote Config | Fix |
|-------|--------------|---------------|-----|
| IKE version mismatch | IKEv2 | IKEv1 | Align versions |
| Encryption mismatch | AES-256 | AES-128 | Use same algorithm |
| DH group mismatch | Group14 | Group2 | Use same group |
| PSK mismatch | (value A) | (value B) | Set identical PSK |
| Lifetime mismatch | 86400s | 3600s | Set same lifetime |

### 4. SSL VPN Server Issues

```bash
# Check SSL VPN server status
ve vpn DescribeSslVpnServers --Region cn-beijing --SslVpnServerIds '["ssl-xxxxx"]'
```

**Common Issues:**

| Issue | Diagnosis | Solution |
|-------|-----------|----------|
| Server not responding | Check gateway type | Ensure VPN Gateway supports SSL |
| Port unreachable | Check security groups | Allow UDP 1194 or TCP 443 |
| Client cannot connect | Verify client config | Download fresh client config |
| Certificate expired | Check cert dates | Create new client certificate |

### 5. SSL VPN Client Issues

```bash
# Check client certificate status
ve vpn DescribeSslVpnClientCerts --Region cn-beijing --SslVpnClientCertIds '["sc-xxxxx"]'
```

**Client Connection Issues:**

| Symptom | Cause | Solution |
|---------|-------|----------|
| "TLS handshake failed" | Certificate mismatch | Verify CA cert, client cert, key match |
| "Authentication failed" | Certificate revoked | Create new client certificate |
| "Connection timeout" | Server unreachable | Check server status, firewall rules |
| "Cannot assign IP" | Client pool exhausted | Expand client IP pool |

### 6. Routing Issues

**Symptoms:**
- Tunnel is up (`Available`)
- Cannot reach resources across VPN

**Diagnostic Steps:**

1. **Verify Local Subnets:**
   ```bash
   # Check configured local subnets
   ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]' | jq '.Result.VpnConnections[0].LocalSubnet'
   ```

2. **Verify Remote Subnets:**
   ```bash
   # Check configured remote subnets
   ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]' | jq '.Result.VpnConnections[0].RemoteSubnet'
   ```

3. **Check VPC Route Tables:**
   ```bash
   # Routes must exist for remote subnets pointing to VPN Gateway
   ve vpc DescribeRouteTables --Region cn-beijing --VpcId vpc-xxxxx
   ```

4. **Check On-Premises Routes:**
   - Ensure routes to cloud subnets exist
   - Point to VPN device/tunnel interface

**Common Routing Issues:**

| Issue | Check | Fix |
|-------|-------|-----|
| No route to cloud | Local subnet missing from config | Add correct local subnet |
| No route from cloud | VPC route table missing entry | Add route to VPN Gateway |
| Asymmetric routing | Return path differs | Ensure symmetric routing |
| Subnet overlap | Local/remote CIDRs overlap | Use non-overlapping CIDRs |

## Specific Error Scenarios

### Error: "IKE negotiation failed"

**Causes:**
1. Mismatched IKE parameters
2. Incorrect PSK
3. Network/firewall blocking IKE ports
4. NAT traversal not enabled when needed

**Resolution:**
```bash
# 1. Verify connection config
ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]'

# 2. Check IKE config matches peer:
#    - Version (IKEv1 vs IKEv2)
#    - Mode (Main vs Aggressive)
#    - Encryption (AES-128/192/256)
#    - Authentication (MD5/SHA1/SHA256)
#    - DH Group (Group1/2/5/14/24)
#    - Lifetime (seconds)

# 3. Verify PSK is correct (do not log actual value)
# 4. Check firewall rules allow UDP 500 and 4500
# 5. Ensure NAT-T is enabled if behind NAT
```

### Error: "IPsec SA establishment failed"

**Causes:**
1. Mismatched IPsec parameters
2. PFS group mismatch
3. Proxy ID mismatch

**Resolution:**
```bash
# Verify IPsec config matches peer:
#    - Protocol (ESP)
#    - Encryption (AES-128/192/256)
#    - Authentication (MD5/SHA1/SHA256)
#    - PFS (disabled/group1/2/5/14)
#    - Lifetime (seconds)
```

### Error: "SSL VPN client cannot connect"

**Diagnostic Steps:**

1. **Verify Server Status:**
   ```bash
   ve vpn DescribeSslVpnServers --Region cn-beijing --SslVpnServerIds '["ssl-xxxxx"]'
   # Should show Status: Available
   ```

2. **Verify Client Certificate:**
   ```bash
   ve vpn DescribeSslVpnClientCerts --Region cn-beijing --SslVpnClientCertIds '["sc-xxxxx"]'
   # Should show Status: Available
   ```

3. **Check Security Groups:**
   - VPN Gateway subnet must allow inbound UDP 1194 (or configured port)
   - VPC security groups must allow traffic from client IP pool

4. **Verify Client Configuration:**
   - Download fresh config: `ve vpn DownloadSslVpnClientConfig`
   - Check OpenVPN client logs for specific errors

## Recovery Procedures

### Procedure: Reset IPSec Connection

If tunnel is stuck in error state:

1. **Document current configuration:**
   ```bash
   ve vpn DescribeVpnConnections --Region cn-beijing --VpnConnectionIds '["vpn-xxxxx"]' > connection_backup.json
   ```

2. **Delete and recreate:**
   ```bash
   # Delete connection
   ve vpn DeleteVpnConnection --Region cn-beijing --VpnConnectionId vpn-xxxxx
   
   # Wait for deletion
   sleep 30
   
   # Recreate with same config
   ve vpn CreateVpnConnection --Region cn-beijing ...
   ```

### Procedure: Rotate SSL VPN Certificates

When certificates are expiring or compromised:

1. **Create new certificate:**
   ```bash
   ve vpn CreateSslVpnClientCert --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnClientCertName "new-cert"
   # Save Certificate and PrivateKey immediately
   ```

2. **Distribute new certificate to client**

3. **Verify new connection works**

4. **Delete old certificate:**
   ```bash
   ve vpn DeleteSslVpnClientCert --Region cn-beijing --SslVpnServerId ssl-xx --SslVpnClientCertId sc-old-xx
   ```

### Procedure: Recover from Deleted Gateway

If VPN Gateway was accidentally deleted:

1. **Check if backup exists** (connection details, certificates)
2. **Create new VPN Gateway:**
   ```bash
   ve vpn CreateVpnGateway --Region cn-beijing --VpcId vpc-xx --SubnetId subnet-xx --Bandwidth 10 --VpnGatewayName recovered-gateway
   ```
3. **Recreate connections** using saved configuration
4. **Update customer gateway** configuration with new gateway IP
5. **Recreate SSL VPN Server** and certificates as needed

## Health Check Commands

```bash
# Quick health check for all VPN resources
REGION="cn-beijing"

echo "=== VPN Gateways ==="
ve vpn DescribeVpnGateways --Region $REGION | jq -r '.Result.VpnGateways[] | "\(.VpnGatewayId): \(.Status)"'

echo "=== Customer Gateways ==="
ve vpn DescribeCustomerGateways --Region $REGION | jq -r '.Result.CustomerGateways[] | "\(.CustomerGatewayId): \(.IpAddress)"'

echo "=== IPSec Connections ==="
ve vpn DescribeVpnConnections --Region $REGION | jq -r '.Result.VpnConnections[] | "\(.VpnConnectionId): \(.Status)"'

echo "=== SSL VPN Servers ==="
ve vpn DescribeSslVpnServers --Region $REGION | jq -r '.Result.SslVpnServers[] | "\(.SslVpnServerId): \(.Status)"'
```

## Support Escalation

When to escalate to Volcengine support:

| Issue | Information to Provide |
|-------|----------------------|
| Gateway stuck in Pending > 30 min | RequestId, Gateway ID, timestamp |
| Connection repeatedly fails | RequestId, Connection ID, peer device logs |
| SSL VPN server unreachable | RequestId, Server ID, client logs |
| Quota increase needed | Current usage, required quota, business justification |

**Support Contact:**
- Console: https://console.volcengine.com/ticket
- API: Include `RequestId` from error response
