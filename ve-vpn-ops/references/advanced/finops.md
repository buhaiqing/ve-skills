# FinOps — VPN Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Billing Model

| Component | Pricing Model | Typical Cost |
|-----------|---------------|--------------|
| VPN gateway | Per connection / month | Low–moderate |
| Data transfer | Per GB outbound | Variable (dominant) |
| VPN tunnel | Per tunnel / month | Low |

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete unused tunnels | No traffic > 30d → remove tunnel + customer gateway | 100% of idle tunnel fee |
| Consolidate connections | Multiple tunnels to same site → merge with BGP ECMP | ~50% on connection fees |
| Right-size bandwidth | Monitor peak usage → downgrade excess bandwidth spec | 20-40% |

## Cost Anomaly Detection

⚠️ **VPN cost anomalies** — investigate when:

| Anomaly | Investigation | Action |
|---------|---------------|--------|
| Data transfer spike > 200% | `ve vpn DescribeVpnConnections` → check new tunnels, traffic pattern | Rate-limit or audit |
| New tunnel count spikes | `ve vpn DescribeVpnGateways` → verify creation events | Confirm authorized |
| Bandwidth consistently 0% | Flag idle tunnels for deletion | Review → `ve vpn DeleteVpnConnection` |

> 💡 VPN cost is dominated by **data transfer**, not gateway fees — monitor outbound traffic monthly.

## Query Current Resources

```bash
# List active VPN gateways and connections
ve vpn DescribeVpnGateways --body '{}'
ve vpn DescribeVpnConnections --body '{}'
```

> VPN costs come from running resources + data transfer. Use `ve billing DescribeBillSummaryByMonth` for aggregated cost data.

## Related Resources

- [ve-billing-ops FinOps](../../ve-billing-ops/references/advanced/finops.md) — billing models, budget alerts, tag allocation