# Redis Knowledge Base — Fault Patterns

## Pattern 1: Big Key — Memory Spike + Latency Surge

### Symptoms
- Memory spike w/o proportionate key growth
- Single cmd latency > 100ms
- Slow log → DEL/EXPIRE on large keys (> 10MB)
- Network bandwidth spike on single shard

### Root Cause
- App stores large data structure as single key (hash/map w/ millions of fields)
- Key expiration not configured → growing indefinitely
- Bulk write (e.g., HSET w/ thousands of fields)

### Resolution
1. Identify: `ve redis BigKeys <instance>` or `redis-cli --bigkeys`
2. Check size: `MEMORY USAGE <key>`
3. Break into smaller keys (hash shards by user prefix)
4. Set TTL on keys
5. Use async delete: `UNLINK <key>` instead of `DEL`
6. Monitor memory stabilization

### Prevention
- App code review for hash/set size limits
- Big key monitor threshold (e.g., > 10MB)
- Use Redis Streams/Lists instead of massive hashes

---

## Pattern 2: Connection Leak — Connection Exhaustion

### Symptoms
- Active connections steadily climbing toward max
- App errors: "Can't handle max connections"
- Connections not decreasing after app restart
- Memory stable, CPU normal

### Root Cause
- App code creates connections w/o closing/reusing
- Connection pool cfg mismatch (poolMax > Redis max)
- Pool not releasing idle connections

### Resolution
1. Check connections: `ve redis DescribeDBInstanceDetail` → connection count
2. Identify source IPs via `CLIENT LIST`
3. Fix pool: set maxIdle, maxActive, idleTimeout
4. Restart affected app pods
5. Monitor connections → return to normal

### Prevention
- Pool best practice: maxPoolSize = 2 × threads
- App-level monitoring: pool active vs Redis connections
- Alert on connection growth rate > 10/min

---

## Pattern 3: Cache Stampede — Hit Rate Collapse

### Symptoms
- Hit rate drops from > 95% to < 50% suddenly
- Backend DB CPU spikes (DB fallback traffic)
- Massive new keys created in short window
- Eviction count spikes

### Root Cause
- Popular keys expire simultaneously (same TTL)
- App doesn't protect cache miss w/ mutex/distributed lock
- Redis restart/flush → full cache miss

### Resolution
1. Protect hot keys: check cache w/ mutex, regenerate if miss
2. Add jitter to TTL (e.g., 3600 ± 300s)
3. Implement cache warming for critical keys
4. Temporarily rate-limit backend while cache rebuilds

### Prevention
- TTL jitter for all cached keys
- Cache warming after deploy/restart
- L1 (local) + L2 (Redis) cache for ultra-hot data

---

## Cascade Pattern: Memory Full → Evictions → Hit Rate Drop → DB Overload → Timeout

### Trigger Event
- Redis memory reaches 100% w/ allkeys-lru eviction

### Propagation Path
- **A → Memory full** ⇒ LRU evicts aggressively
- **B → Hit rate drops** ⇒ App gets more cache misses
- **C → DB fallback traffic spikes** → DB CPU 100%
- **D → DB response time increases** → Redis pool waits longer
- **E → Connections stack up** → New connections rejected
- **F → App errors cascade** → Service degradation/outage

### Breaking the Chain
1. **Break at A**: Increase memory or scale to larger instance
2. **Break at B**: Serve stale data from local cache fallback
3. **Break at C**: Rate-limit non-critical read paths to DB
4. **Post-recovery**: Warm cache w/ top 100 keys before full traffic resume