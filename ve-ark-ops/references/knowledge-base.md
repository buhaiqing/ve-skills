# Ark (方舟大模型平台) Knowledge Base

## FAQ

### Endpoints

**Q: How long does it take to create an endpoint?**
A: Typically 2-10 minutes depending on model size and resource availability. Large models (70B+ parameters) may take longer. Poll with `DescribeEndpoint` every 10s for up to 600s.

**Q: Can I change the model version of a running endpoint?**
A: No. You must create a new endpoint with the new model version, then delete the old one. Use `ModifyEndpoint` to change scaling or description fields only.

**Q: Can I scale an endpoint to zero replicas?**
A: Yes, set `MinReplicas=0` to allow the endpoint to scale down when idle (reduces cost). However, cold start latency applies on first request after idle period.

**Q: What is the maximum endpoint quota per account?**
A: Default quotas vary. Check your account limit via the console or contact support. Request increases through the quota management system.

### Training

**Q: What training methods does Ark support?**
A: SFT (Supervised Fine-Tuning), DPO (Direct Preference Optimization), and RL (Reinforcement Learning). Method availability depends on the base model.

**Q: What dataset formats are supported for SFT?**
A: JSONL format with conversation structure. Each line contains a JSON object with `messages` array containing `role` and `content` fields. Example:
```json
{"messages": [{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "What is AI?"}, {"role": "assistant", "content": "AI stands for Artificial Intelligence."}]}
```

**Q: How long does training take?**
A: Depends on model size, dataset size, and GPU configuration. Small models (1B-7B) with small datasets may finish in 1-2 hours. Large models (70B+) may take days.

**Q: Can I resume failed training jobs?**
A: No. Failed training jobs cannot be resumed. You must create a new training job with the same configuration. Check training logs to identify and fix the failure cause.

### Datasets

**Q: What is the maximum dataset size?**
A: Dataset size limits depend on the model and training method. Typically up to several GB. Very large datasets (>10GB) should be split into multiple datasets.

**Q: Can I update an existing dataset?**
A: No. Datasets are immutable after creation. To modify data, create a new dataset with updated content.

**Q: What regions can I use TOS buckets from?**
A: The TOS bucket must be in the same region as the Ark dataset. Cross-region TOS access is not supported for dataset creation.

### Evaluation

**Q: What metrics does evaluation provide?**
A: Metrics depend on the evaluation job configuration. Common metrics include accuracy, BLEU, ROUGE, F1 score, and custom metrics defined in the evaluation template.

**Q: Can I use my own evaluation dataset?**
A: Yes. Create a dataset with ground truth data and reference it in the evaluation job.

## Best Practices

### Endpoint Management
1. **Naming convention:** Use `{environment}-{purpose}-{model-abbrev}` (e.g., `prod-chat-doubao-pro`, `staging-embedding-bge`)
2. **Auto-scaling:** Configure `MinReplicas` and `MaxReplicas` based on traffic patterns. Start with `MinReplicas=1, MaxReplicas=3` for new endpoints
3. **Testing:** Create a staging endpoint before production deployment to validate model behavior
4. **Version tracking:** Document which model version each endpoint uses. Use endpoint descriptions to track deployment dates

### Training Best Practices
1. **Start small:** Run 2-3 epochs on a small dataset subset to validate format and training pipeline
2. **Monitor loss:** Track training/validation loss — diverging curves indicate overfitting or data issues
3. **Learning rate:** Start with 1e-5 for SFT, 5e-6 for DPO. Adjust based on loss convergence
4. **Data quality:** Use clean, deduplicated data. Check for formatting errors before training
5. **Checkpointing:** Training jobs automatically save checkpoints. Use the output model for downstream evaluation

### Cost Optimization
1. **Pause idle endpoints:** Use `ModifyEndpoint --MinReplicas 0` for non-production endpoints during off-hours
2. **Right-size model:** Use smaller models for simple tasks (e.g., Doubao-lite vs Doubao-pro)
3. **Batch inference:** Use the Batch Inference API for large-scale non-real-time workloads instead of sustained endpoint usage
4. **Monitor token usage:** Track token consumption via CMS metrics to forecast costs

## Related Documentation

- [Volcengine Ark Product Page](https://www.volcengine.com/product/ark)
- [Ark Console](https://console.volcengine.com/ark)
- [Ark OpenAPI Documentation](https://www.volcengine.com/docs/82379)
- [Volcengine SDK for Go](https://github.com/volcengine/volc-sdk-golang)
- [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- [Model List](https://www.volcengine.com/docs/82379/1330310)
- [Model Pricing](https://www.volcengine.com/docs/82379/1544106)
- [IAM Access Control](https://www.volcengine.com/docs/82379/1263493)

## Glossary

| Term | Description |
|------|-------------|
| Inference Endpoint | Deployment of a model version for serving inference requests |
| Model Version | A specific snapshot/version of a model |
| SFT | Supervised Fine-Tuning — training with labeled data |
| DPO | Direct Preference Optimization — alignment training |
| RL | Reinforcement Learning — training with reward signals |
| JSONL | JSON Lines format — each line is a valid JSON object |
| TOS | Volcengine Object Storage — used for dataset storage |
| TPM | Tokens Per Minute — throughput guarantee for endpoints |
| Cold Start | Initial latency when scaling from zero replicas |