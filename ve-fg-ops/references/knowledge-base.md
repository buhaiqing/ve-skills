# FunctionGraph Knowledge Base

## Fault Pattern Library

### Pattern: Function Invocation Timeout

**Symptoms:** Invocation returns timeout error after configured timeout period

**Root Causes:**
1. Function code has infinite loop or long-running operation
2. Network calls within function are slow (downstream API timeouts)
3. Cold start delay for functions with heavy dependencies

**Resolution Steps:**
1. Check function logs for slow operations
2. Increase `Timeout` configuration if operation legitimately needs more time
3. Optimize code to reduce execution time
4. Use reserved concurrency to reduce cold starts

**Prevention:**
- Set realistic timeout values
- Implement proper error handling and retries in function code
- Use async processing for long-running tasks

### Pattern: Out of Memory

**Symptoms:** Function fails with memory exceeded error

**Root Causes:**
1. Inefficient memory usage in function code
2. Loading large datasets into memory
3. Memory leak in function runtime

**Resolution Steps:**
1. Check function logs for memory usage patterns
2. Increase `MemorySize` configuration (also increases CPU allocation)
3. Optimize code to reduce memory footprint
4. Stream large data instead of loading entirely into memory

### Pattern: Function Not Triggered

**Symptoms:** Trigger configured but function not executing

**Root Causes:**
1. Trigger configuration incorrect (wrong cron, wrong event source)
2. Function status not `Active` (still creating or failed)
3. IAM permissions missing

**Resolution Steps:**
1. Verify trigger configuration with `ListTriggers`
2. Check function status with `GetFunction`
3. Verify IAM policies have required permissions
4. Test invoke function directly with `InvokeFunction --InvocationType RequestResponse`

### Pattern: Concurrent Execution Limit Hit

**Symptoms:** Function invocations being throttled

**Root Causes:**
1. Sudden spike in traffic
2. Function has low concurrency quota

**Resolution Steps:**
1. Check `ConcurrentExecutions` metric
2. Request concurrency quota increase
3. Implement request queuing or rate limiting
4. Consider using async invocation pattern
