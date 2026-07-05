# FinOps — TOS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Use IA storage tier | Objects accessed <1x/month → IA tier instead of Standard | ~60% on storage cost |
| Archive cold data | Data not accessed for 90d+ → transition to Archive via lifecycle rule | ~80% on storage cost |
| Clean up incomplete uploads | Abort multipart uploads older than 7 days via lifecycle policy | Recovers orphaned storage |

## Query Current Prices

```bash
# Query current TOS pricing per storage tier
ve tos DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve tos DescribePrice` for current quotes.