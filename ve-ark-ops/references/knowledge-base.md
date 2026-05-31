# Ark Knowledge Base

## Pattern: Endpoint Creation Fails

**Root Causes:**
1. Insufficient endpoint quota
2. Model not available in region
3. Invalid endpoint configuration

**Resolution Steps:**
1. Check available quota vs requested
2. Verify model availability in region
3. Check endpoint configuration parameters

## Pattern: Training Job Fails

**Root Causes:**
1. Dataset format incorrect
2. Insufficient compute resources
3. Training parameter misconfiguration

**Resolution Steps:**
1. Validate dataset format (JSONL, etc.)
2. Check training job logs for errors
3. Verify hyperparameter ranges
