# FinOps — Elasticsearch Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Right-size node specs | Over-provisioned data nodes → downsize instance type | ~50% |
| Reduce replica count | Index with 2+ replicas during hot phase → reduce to 1 after indexing completes | ~33% on storage cost |
| Convert steady PostPaid to PrePaid | Running > 30d → 1yr PrePaid | ~35% |

## Query Current Prices

```bash
# Query current Elasticsearch instance pricing
ve elasticsearch DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve elasticsearch DescribePrice` for current quotes.