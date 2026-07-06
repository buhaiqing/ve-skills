# FinOps — Skill Generator FinOps Meta-Operation

> FinOps deep content per TE-7. This is a meta-skill; FinOps here means:
> (1) enforcing finops.md coverage across all generated skills,
> (2) cost-aware skill generation practices.

## FinOps Coverage Enforcement

Every generated skill MUST have `references/advanced/finops.md`. The generator SHOULD:

### Per-Skill Requirements

| Skill Tier | FinOps File Required | Min Content |
|------------|:-------------------:|-------------|
| required (13) | ✅ **yes** | ≥ 20 lines, product-specific pricing + ≥ 2 optimizations |
| recommended (10) | ✅ **yes** | ≥ 15 lines, product-specific pricing + ≥ 1 optimization |
| optional (6) | ✅ **yes** | ≥ 10 lines, pricing sources + ≥ 1 optimization |

### Validation Checklist
```
□ references/advanced/finops.md exists
□ Contains product-specific pricing (not generic "check DescribePrice")
□ Has ≥ 2 actionable cost optimization items
□ Includes real ve CLI query example (not placeholder)
□ Cross-references ve-billing-ops for billing queries
```

## Cost Optimization Patterns by Product Category

### Compute (ECS, VKE, FG)
- Right-size instances: `ve ecs DescribeInstanceTypes` + utilization analysis
- Spot/preemptible instances for stateless workloads
- Auto-scaling policies to match demand

### Database (RDS, Redis, MongoDB, ES, Kafka)
- Storage tiering: SSD ↔高效云盘 cost comparison
- Backup retention window optimization
- Read replica for analytics workload isolation

### Network (VPC, EIP, NAT, CLB, ALB, VPN)
- EIP: release unattached IPs, `ve eip DescribeEipAddresses --Status "Available"`
- NAT: SNAT/DNAT usage review
- Cross-region traffic cost awareness

### Storage (TOS)
- Lifecycle rules for tiered storage
- Delete incomplete multipart uploads
- CDN origin-pull cost optimization

## Proactive Inspection Checklist

```markdown
## Generator FinOps Proactive Inspection — [Date]

### Coverage
- [ ] All 29 skills have advanced/finops.md
- [ ] Required tier (13): each has ≥ 20 lines product-specific content
- [ ] Recommended tier (10): each has ≥ 15 lines product-specific content
- [ ] Optional tier (6): each has ≥ 10 lines product-specific content

### Quality
- [ ] No "DescribePrice" placeholder commands (verify with `ve <svc> --help`)
- [ ] Each skill has ≥ 2 real cost optimization actions
- [ ] ve-billing-ops cross-reference present for billing queries
- [ ] All CLI query examples are runnable (verified command syntax)

### Consistency
- [ ] No copy-paste boilerplate across different product categories
- [ ] Cost optimization items are product-specific, not generic
```
