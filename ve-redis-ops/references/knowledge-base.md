# Redis Knowledge Base — Fault Patterns

## Pattern 1: Big Key — Memory Spike + Latency Surge

### Symptoms
- Memory usage spikes suddenly without proportionate key growth
- Single command latency > 100ms on random samples
- Slow log shows DEL/EXPIRE on large keys (> 10MB)
- Network bandwidth spike on single shard

### Root Cause
- Application stores large data structure as single key (hash/map with millions of fields)
- Key expiration not configured, growing indefinitely
- Bulk write operation (e.g., HSET with thousands of fields)

### Resolution
1. Identify big key: `ve redis BigKeys <instance>` or `redis-cli --bigkeys`
2. Check key size: `MEMORY USAGE <key>`
3. Break into smaller keys (hash shards by user prefix)
4. Set TTL on keys if applicable
5. Use async delete: `UNLINK <key>` instead of `DEL`
6. Monitor memory stabilization

### Prevention
- Application code review for hash/set size limits
- Big key monitoring threshold (e.g., > 10MB)
- Use Redis Streams or Lists instead of massive hashes

---

## Pattern 2: Connection Leak — Connection Exhaustion

### Symptoms
- Active connections steadily climbing toward max
- Application errors: "Can't handle max connections"
- Connections not decreasing after application restart
- Memory stable, CPU normal

### Root Cause
- Application code creates connection without closing/reusing
- Connection pool configuration mismatch (poolMax > Redis max)
- Connection pool not releasing idle connections

### Resolution
1. Check current connections: `ve redis DescribeDBInstanceDetail` → connection count
2. Identify source IPs via `CLIENT LIST`
3. Fix connection pool: set maxIdle, maxActive, idleTimeout
4. Restart affected application pods
5. Monitor connections returning to normal

### Prevention
- Connection pool best practice: maxPoolSize = 2 * threads
- Application-level monitoring: pool active connections vs Redis connections
- Alert on connection growth rate > 10/min

---

## Pattern 3: Cache Stampede — Hit Rate Collapse

### Symptoms
- Hit rate drops from > 95% to < 50% suddenly
- Backend database CPU spikes (DB fallback traffic)
- Massive new keys created in short window
- Eviction count spikes

### Root Cause
- Popular keys expire simultaneously (same TTL)
- Application does not protect cache miss with mutex/distributed lock
- Redis restart or flush causes full cache miss

### Resolution
1. Protect hot keys: check cache with mutex, regenerate if miss
2. Add jitter to TTL (e.g., 3600 ± 300 seconds)
3. Implement cache warming for critical keys
4. Temporarily rate-limit backend while cache rebuilds

### Prevention
- TTL jitter for all cached keys
- Cache warming after deployment or restart
- L1 (local) + L2 (Redis) cache for ultra-hot data

---

## Cascade Pattern: Memory Full → Evictions → Hit Rate Drop → DB Overload → Instance Timeout

### Trigger Event
- Redis memory reaches 100% with allkeys-lru eviction

### Propagation Path
- **A → Memory full** → LRU starts evicting keys aggressively
- **B → Hit rate drops** → Application gets more cache misses
- **C → DB fallback traffic increases** → Database CPU spikes to 100%
- **D → Database response time increases** → Redis connection pool waits longer
- **E → Connections stack up** → New connections rejected
- **F → Application errors cascade** → Service degradation or outage

### Breaking the Chain
1. **Break at A**: Immediately increase memory or scale to larger instance
2. **Break at B**: Enable application-side fallback (serve stale data from local cache)
3. **Break at C**: Rate-limit non-critical read paths to database
4. **Post-recovery**: Warm cache with top 100 keys before full traffic resume
