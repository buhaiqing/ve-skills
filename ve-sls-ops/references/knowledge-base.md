# SLS Knowledge Base

## Fault Pattern Library

### Pattern: Missing Logs

**Symptoms:** Expected logs not appearing in SLS topic

**Root Causes:**
1. Logtail agent not installed or not running
2. Machine group not configured for the topic
3. File path mismatch in collector config
4. Network connectivity issues

**Resolution Steps:**
1. Check Logtail status on ECS
2. Verify machine group assignment
3. Validate file path and log format
4. Test network to SLS endpoint

### Pattern: Log Volume Spike

**Symptoms:** Unexpected increase in log ingestion, rising costs

**Root Causes:**
1. Debug logging enabled in production
2. Log loop in application code
3. DDoS or malicious traffic
4. New feature generating excessive logs

**Resolution Steps:**
1. Identify offending topic: `ve tls DescribeTopics`
2. Review recent application deployments
3. Adjust log level or sampling
4. Set up volume alerts
