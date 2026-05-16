# RDS MySQL Knowledge Base — Fault Patterns

## Pattern 1: Slow Query Cascade — Single Query Blocks Application

### Symptoms
- QPS drops significantly while connections increase
- Connections approach max_connections limit
- CPU spikes to > 80%
- Slow log shows same query pattern repeatedly

### Root Cause
- Missing index on frequently filtered column
- Full table scan on growing table (millions of rows)
- Statistics not updated → optimizer picks suboptimal plan
- ORM generates inefficient N+1 queries

### Resolution
1. Query slow log: `ve rds_mysql DescribeSlowLog` (or via SQL)
2. EXPLAIN the problematic query
3. Add appropriate index: `CREATE INDEX idx_col ON t(col)`
4. If table is large and online ALTER needed, use `ALGORITHM=INPLACE`
5. Monitor `Handler_read_next` and `Slow_queries` for improvement
6. If ORM N+1, coordinate with application team to add EAGER loading

### Prevention
- Slow query alert at 10/min (before it becomes catastrophic)
- Monthly index review for new tables
- Database review in PR review for new queries

---

## Pattern 2: Connection Exhaustion — Max Reached, New Connections Rejected

### Symptoms
- Active connections = max_connections
- New connection attempts fail
- Application returns 500 errors
- No slow queries, CPU normal

### Root Cause
- Connection leak in application (not closing connections)
- Connection pool misconfigured (too many pools × large pool size)
- Long-running queries holding connections (unoptimized reports)

### Resolution
1. Check connection sources: `SHOW FULL PROCESSLIST`
2. Identify sleeping connections: `COUNT(*) FROM information_schema.processlist WHERE Command='Sleep'`
3. Fix connection pool: reduce poolSize, enable idle timeout
4. Kill idle connections: `KILL <connection_id>` (sleeping > 30min)
5. If long queries, optimize query or move to read replica
6. As emergency: temporarily increase max_connections (requires resource headroom)

### Prevention
- Connection pool standard: poolSize ≤ 10 per service instance
- Application monitoring: track active vs idle connection ratio
- Alert at 70% of max_connections (leave 30% buffer)

---

## Pattern 3: Disk Full — Binary Logs and Storage Exhaustion

### Symptoms
- Disk usage > 95%
- Binlog directory growing rapidly
- Instance becomes read-only (InnoDB protection)
- New writes fail with "disk full" or "log file full"

### Root Cause
- Binlog retention too long (default or misconfigured)
- Large data import or batch job generating excessive binlog
- Binlog format ROW (logs full row images for every change)
- No automatic cleanup or retention policy

### Resolution
1. Check binlog size: `SHOW BINARY LOGS`
2. Purge old binlogs: `PURGE BINARY LOGS BEFORE '2026-05-09 00:00:00'`
3. If read-only, emergency: delete temporary files in data directory
4. Reduce binlog retention: 7 days (balance between PITR and disk)
5. For batch imports: `SET sql_log_bin=OFF` (for non-replicated data)
6. Monitor disk stabilizing after cleanup

### Prevention
- Binlog retention ≤ 7 days with alert at 80% disk
- Monitor binlog growth rate daily
- Pre-allocate extra disk buffer for batch operations
- Auto-purge binlog policy configured

---

## Cascade Pattern: Lock Contention → Query Queue → Connection Pool Full → Application Timeout → Service Outage

### Trigger Event
- Long-running transaction (uncommitted) holding row lock (e.g., bulk UPDATE without WHERE, or uncommitted transaction in application)

### Propagation Path
- **A → Long transaction holds lock** → Other transactions queue waiting for lock
- **B → Lock wait timeout increases** → Each new transaction adds to queue
- **C → Connection pool fills** → All connections are in Lock Wait state
- **D → New connections rejected** → Application starts getting connection errors
- **E → Application threads block** → Thread pool exhausted
- **F → Health checks fail** → Load balancer marks instance down
- **G → Service outage**

### Breaking the Chain
1. **Break at A**: Identify and kill blocking transaction: `SHOW ENGINE INNODB STATUS\G` → find blocking thread → `KILL <id>`
2. **Break at C**: Temporarily increase connection pool to flush queue while root cause is fixed
3. **Break at E**: Application circuit breaker — fail fast instead of blocking
4. **Post-recovery**: Analyze transaction code; add query timeout; implement automatic transaction rollback

### Prevention
- `innodb_lock_wait_timeout` set to reasonable value (e.g., 30s, not default 50s)
- Application-level transaction timeout
- Monitor `InnoDB_row_lock_waits` and `InnoDB_row_lock_time_avg`
- Code review: avoid long transactions
