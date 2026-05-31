# Monitoring VPN

## Key Metrics

### VPN Gateway Metrics

| Metric | Namespace | Description | Unit | Dimensions |
|--------|-----------|-------------|------|------------|
| `VpnGatewayNetworkIn` | Volcengine/VPN | Network traffic into VPN Gateway | bytes | VpnGatewayId |
| `VpnGatewayNetworkOut` | Volcengine/VPN | Network traffic out of VPN Gateway | bytes | VpnGatewayId |
| `VpnGatewayBandwidthUsage` | Volcengine/VPN | Bandwidth utilization percentage | percent | VpnGatewayId |
| `VpnGatewayActiveConnections` | Volcengine/VPN | Number of active IPSec connections | count | VpnGatewayId |
| `VpnGatewaySslConnections` | Volcengine/VPN | Number of active SSL VPN connections | count | VpnGatewayId |

### IPSec Connection Metrics

| Metric | Namespace | Description | Unit | Dimensions |
|--------|-----------|-------------|------|------------|
| `VpnConnectionStatus` | Volcengine/VPN | Tunnel status (1=up, 0=down) | count | VpnConnectionId |
| `VpnConnectionNetworkIn` | Volcengine/VPN | Traffic into tunnel | bytes | VpnConnectionId |
| `VpnConnectionNetworkOut` | Volcengine/VPN | Traffic out of tunnel | bytes | VpnConnectionId |
| `VpnConnectionPacketDrop` | Volcengine/VPN | Dropped packets | count | VpnConnectionId |
| `VpnConnectionLatency` | Volcengine/VPN | Tunnel latency | milliseconds | VpnConnectionId |

### SSL VPN Metrics

| Metric | Namespace | Description | Unit | Dimensions |
|--------|-----------|-------------|------|------------|
| `SslVpnServerConnections` | Volcengine/VPN | Active SSL VPN connections | count | SslVpnServerId |
| `SslVpnServerMaxConnections` | Volcengine/VPN | Maximum allowed connections | count | SslVpnServerId |
| `SslVpnServerAuthFailures` | Volcengine/VPN | Authentication failures | count | SslVpnServerId |
| `SslVpnClientBytesIn` | Volcengine/VPN | Bytes received from client | bytes | SslVpnClientCertId |
| `SslVpnClientBytesOut` | Volcengine/VPN | Bytes sent to client | bytes | SslVpnClientCertId |

## CloudMonitor (CMS) Integration

### Query Metrics via CLI

```bash
# Query VPN Gateway network traffic
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine/VPN \
  --MetricName VpnGatewayNetworkIn \
  --Dimensions '[{"Name":"VpnGatewayId","Value":"vgw-xxxxx"}]' \
  --StartTime "2024-05-01T00:00:00Z" \
  --EndTime "2024-05-27T00:00:00Z" \
  --Period 300

# Query IPSec connection status
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine/VPN \
  --MetricName VpnConnectionStatus \
  --Dimensions '[{"Name":"VpnConnectionId","Value":"vpn-xxxxx"}]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 60

# Query SSL VPN active connections
ve cms GetMetricData \
  --Region cn-beijing \
  --Namespace Volcengine/VPN \
  --MetricName SslVpnServerConnections \
  --Dimensions '[{"Name":"SslVpnServerId","Value":"ssl-xxxxx"}]' \
  --StartTime "2024-05-27T00:00:00Z" \
  --EndTime "2024-05-27T23:59:59Z" \
  --Period 300
```

## Alert Rules

### VPN Gateway Bandwidth Alert

```json
{
  "RuleName": "vpn-gateway-high-bandwidth",
  "Namespace": "Volcengine/VPN",
  "MetricName": "VpnGatewayBandwidthUsage",
  "Dimensions": [
    {
      "Name": "VpnGatewayId",
      "Value": "vgw-xxxxx"
    }
  ],
  "ComparisonOperator": "GreaterThanThreshold",
  "Threshold": 80,
  "Period": 300,
  "EvaluationCount": 3,
  "Level": "warning",
  "ContactGroup": ["vpn-admins"],
  "SilenceTime": 3600
}
```

### IPSec Connection Down Alert

```json
{
  "RuleName": "vpn-connection-down",
  "Namespace": "Volcengine/VPN",
  "MetricName": "VpnConnectionStatus",
  "Dimensions": [
    {
      "Name": "VpnConnectionId",
      "Value": "vpn-xxxxx"
    }
  ],
  "ComparisonOperator": "LessThanThreshold",
  "Threshold": 1,
  "Period": 60,
  "EvaluationCount": 2,
  "Level": "critical",
  "ContactGroup": ["vpn-admins", "oncall"],
  "SilenceTime": 300
}
```

### SSL VPN Connection Limit Alert

```json
{
  "RuleName": "ssl-vpn-connections-high",
  "Namespace": "Volcengine/VPN",
  "MetricName": "SslVpnServerConnections",
  "Dimensions": [
    {
      "Name": "SslVpnServerId",
      "Value": "ssl-xxxxx"
    }
  ],
  "ComparisonOperator": "GreaterThanThreshold",
  "Threshold": 80,
  "Period": 300,
  "EvaluationCount": 2,
  "Level": "warning",
  "ContactGroup": ["vpn-admins"],
  "SilenceTime": 3600
}
```

## Log Collection

### Enable VPC Flow Logs for VPN Subnet

```bash
# Create flow log for VPN subnet
ve vpc CreateFlowLog \
  --Region cn-beijing \
  --VpcId vpc-xxxxx \
  --SubnetId subnet-xxxxx \
  --TrafficType ALL \
  --LogStoreType tls \
  --ProjectName vpn-logs \
  --TopicName vpn-traffic
```

### TLS (Log Service) Query Examples

```sql
-- Query VPN traffic by source IP
* | SELECT src_ip, dst_ip, count(*) as connection_count 
GROUP BY src_ip, dst_ip 
ORDER BY connection_count DESC 
LIMIT 100

-- Query failed IKE negotiations
* | WHERE event_type = 'ike_negotiation_failed' 
SELECT timestamp, src_ip, reason 
ORDER BY timestamp DESC 
LIMIT 50

-- Query high latency connections
* | WHERE latency > 100 
SELECT timestamp, connection_id, latency 
ORDER BY latency DESC 
LIMIT 20
```

## Health Check Automation

### Script: VPN Health Dashboard

```bash
#!/bin/bash
REGION="cn-beijing"

echo "=== VPN Health Check ==="
echo "Time: $(date)"
echo ""

# Check VPN Gateways
echo "--- VPN Gateways ---"
ve vpn DescribeVpnGateways --Region $REGION | jq -r '
  .Result.VpnGateways[] |
  "\(.VpnGatewayId): \(.Status) | Bandwidth: \(.Bandwidth) Mbps"
'

# Check IPSec Connections
echo ""
echo "--- IPSec Connections ---"
ve vpn DescribeVpnConnections --Region $REGION | jq -r '
  .Result.VpnConnections[] |
  "\(.VpnConnectionId): \(.Status) | Gateway: \(.VpnGatewayId) | Customer: \(.CustomerGatewayId)"
'

# Check SSL VPN Servers
echo ""
echo "--- SSL VPN Servers ---"
ve vpn DescribeSslVpnServers --Region $REGION | jq -r '
  .Result.SslVpnServers[] |
  "\(.SslVpnServerId): \(.Status) | Gateway: \(.VpnGatewayId) | Port: \(.SslVpnServerPort)"
'

# Check recent metrics
echo ""
echo "--- Recent Bandwidth Usage (5 min avg) ---"
for vgw in $(ve vpn DescribeVpnGateways --Region $REGION | jq -r '.Result.VpnGateways[].VpnGatewayId'); do
  echo -n "$vgw: "
  ve cms GetMetricData \
    --Region $REGION \
    --Namespace Volcengine/VPN \
    --MetricName VpnGatewayBandwidthUsage \
    --Dimensions "[{\"Name\":\"VpnGatewayId\",\"Value\":\"$vgw\"}]" \
    --StartTime "$(date -u -v-5M +%Y-%m-%dT%H:%M:%SZ)" \
    --EndTime "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --Period 300 | jq -r '.Result.Data[0].Value // "N/A"'
done
```

## Monitoring Best Practices

### Critical Metrics to Monitor

| Priority | Metric | Threshold | Action |
|----------|--------|-----------|--------|
| P0 | VpnConnectionStatus | < 1 (down) | Page on-call immediately |
| P0 | VpnGatewayBandwidthUsage | > 90% | Scale bandwidth or alert |
| P1 | SslVpnServerConnections | > 80% of max | Plan capacity increase |
| P1 | VpnConnectionLatency | > 100ms | Investigate network path |
| P2 | VpnConnectionPacketDrop | > 0.1% | Check MTU settings |
| P2 | SslVpnServerAuthFailures | > 10/hour | Check for brute force |

### Monitoring Checklist

- [ ] Bandwidth utilization monitored with 80% and 90% thresholds
- [ ] IPSec connection status monitored with 2-minute evaluation
- [ ] SSL VPN connection count monitored near capacity limits
- [ ] Authentication failures tracked for security
- [ ] Flow logs enabled for VPN subnet
- [ ] Alert notifications routed to appropriate teams
- [ ] Runbook documented for common alerts
- [ ] Health check script scheduled to run regularly

### Capacity Planning

| Metric | Warning Threshold | Critical Threshold | Planning Action |
|--------|-------------------|-------------------|-----------------|
| Bandwidth Usage | 70% sustained | 90% sustained | Upgrade bandwidth |
| SSL Connections | 70% of max | 90% of max | Add SSL VPN Server |
| IPSec Connections | 70% of quota | 90% of quota | Request quota increase |
| Gateway Count | 70% of quota | 90% of quota | Plan multi-region deployment |

## Troubleshooting with Metrics

### Symptom: High Latency

**Metrics to Check:**
- `VpnConnectionLatency` > 100ms
- `VpnGatewayBandwidthUsage` approaching limit

**Diagnostic Steps:**
1. Check if bandwidth is saturated
2. Verify routing path is optimal
3. Check for packet loss on underlying network
4. Review MTU settings (recommend 1400 for VPN)

### Symptom: Intermittent Disconnects

**Metrics to Check:**
- `VpnConnectionStatus` fluctuating
- `VpnConnectionPacketDrop` > 0

**Diagnostic Steps:**
1. Check IKE/IPsec lifetime settings
2. Verify NAT-T is enabled if behind NAT
3. Review DPD (Dead Peer Detection) configuration
4. Check for network path instability

### Symptom: SSL VPN Slow Performance

**Metrics to Check:**
- `SslVpnServerConnections` near limit
- `VpnGatewayBandwidthUsage` high
- Client connection metrics

**Diagnostic Steps:**
1. Check if connection pool is exhausted
2. Verify bandwidth is not saturated
3. Review compression settings
4. Check client-side network quality
