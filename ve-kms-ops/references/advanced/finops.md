# FinOps — KMS Cost Optimization (Advanced)

> Deep FinOps analysis per TE-7. `core-concepts.md` links here.

## Cost Optimization Tips

| Tip | Action | 💰 Savings |
|-----|--------|---------|
| Delete unused keys | Orphaned KMS keys (not referenced by any resource) — schedule deletion | 100% of key storage + API cost |
| Consolidate keys | Use one key with multiple aliases instead of multiple keys where possible | Reduces key count fees |
| Reduce API call volume | Cache decrypted data and reuse instead of re-decrypting on every access | Variable (key usage fee) |

## Query Current Prices

```bash
# Query current KMS pricing
ve kms DescribePrice
```

> ⚠️ Pricing data sourced from ve DescribePrice command — use `ve kms DescribePrice` for current quotes.