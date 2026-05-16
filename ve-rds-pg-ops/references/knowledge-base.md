# RDS PostgreSQL Knowledge Base — Fault Patterns

## Pattern 1: WAL Disk Pressure — Write-Ahead Log Accumulation

### Symptoms
- pg_wal directory growing continuously
- Disk usage > 85% on data volume
- Replication lag on RO nodes increasing
- Possible instance becomes read-only if disk reaches 100%

### Root Cause
- Replication slot inactive (RO node disconnected or dead) → WAL retained
- wal_keep_size set too large, accumulating WAL segments
- Heavy write rate generating WAL faster than replication consumes
- Archive command failing → WAL not cleared

### Resolution
1. Check WAL directory: `SELECT pg_walfile_name(pg_current_wal_lsn()), pg_size_pretty(sum(size)) FROM pg_ls_waldir();`
2. Check replication slots: `SELECT slot_name, active, restart_lsn FROM pg_replication_slots;`
3. If slot inactive and safe to remove: `SELECT pg_drop_replication_slot('slot_name');`
4. If archive failing: check archive_status directory, fix archive command
5. Reduce wal_keep_size if set excessively
6. Monitor disk decreasing after cleanup

### Prevention
- Monitor replication slot activity
- WAL disk usage > 70% alert
- Archive command health check

---

## Pattern 2: Dead Tuple Accumulation — Table Bloat

### Symptoms
- Table size growing but row count stable
- Query performance degrading (more pages to scan)
- autovacuum running but not keeping up
- seq_scan time increasing

### Root Cause
- High UPDATE/DELETE rate without sufficient autovacuum cycles
- autovacuum_vacuum_scale factor too low for large tables
- Long-running transactions preventing vacuum from cleaning
- Anti-wraparound vacuum not running on time

### Resolution
1. Check dead tuples: `SELECT relname, n_dead_tup, n_live_tup FROM pg_stat_user_tables WHERE n_dead_tup > 100000;`
2. Check bloat: `SELECT schemaname, relname, round(pg_total_relation_size(oid)::numeric/1024/1024, 1) AS total_mb, round(pg_relation_size(oid)::numeric/1024/1024, 1) AS used_mb FROM pg_class;`
3. Manual vacuum: `VACUUM (VERBOSE, ANALYZE) table_name;`
4. For severe bloat: `VACUUM FULL` (requires exclusive lock, schedule off-peak) or use pg_repack
5. Tune autovacuum: increase autovacuum_vacuum_scale_factor, autovacuum_max_workers
6. Check for long-running transactions blocking vacuum: `SELECT pid, now() - xact_start AS duration, query FROM pg_stat_activity WHERE state = 'idle in transaction' ORDER BY duration DESC;`

### Prevention
- Per-table autovacuum tuning for high-churn tables
- Max transaction age alert (> 1hr)
- Weekly bloat report generated

---

## Pattern 3: Autovacuum Storm — Excessive Vacuum Impact

### Symptoms
- CPU spike caused by autovacuum > 50%
- Multiple autovacuum processes running simultaneously
- Temporary files created (vacuum spilling to disk)
- Query latency increases during vacuum activity

### Root Cause
- Too many tables need vacuum simultaneously
- maintenance_work_mem too low (vacuum spills to disk)
- autovacuum_vacuum_cost_delay too low (aggressive vacuum throttling)
- Cost limit insufficient for current I/O capacity

### Resolution
1. Check running vacuums: `SELECT pid, relid::regclass, phase, total, progress FROM pg_stat_progress_vacuum;`
2. Increase maintenance_work_mem (up to 2GB): `ALTER SYSTEM SET maintenance_work_mem = '1GB'; SELECT pg_reload_conf();`
3. Tune cost delay: `ALTER SYSTEM SET autovacuum_vacuum_cost_delay = 2;`
4. If single table problematic: schedule manual VACUUM during off-peak
5. Monitor CPU returning to normal after tuning

### Prevention
- maintenance_work_mem ≥ 1GB for production
- Per-table autovacuum settings for high-churn tables
- Monitor vacuum impact on query performance

---

## Cascade Pattern: Long Transaction → Dead Tuples Grow → Table Bloat → Scan Slows → CPU Spikes → All Queries Slow → Timeout Cascade

### Trigger Event
- Application opens transaction (e.g., ORM session) but never commits or rolls back

### Propagation Path
- **A → Long-running transaction** holds xmin horizon → vacuum cannot remove dead tuples
- **B → Dead tuples accumulate** → table bloat grows (more dead pages than live)
- **C → Sequential scans take longer** → more pages to skip, CPU bound on I/O
- **D → Query latency increases** → connections held longer
- **E → Connection pool fills** → new queries wait in queue
- **F → All queries experience high latency** → application timeouts trigger
- **G → Cascading service failure**

### Breaking the Chain
1. **Break at A**: Find and terminate long transaction: `SELECT pid, now() - xact_start, query FROM pg_stat_activity WHERE state = 'idle in transaction' AND now() - xact_start > interval '30 min';` → `SELECT pg_terminate_backend(pid);`
2. **Break at B**: After killing, run VACUUM to reclaim dead tuples
3. **Break at F**: Application-level timeout and circuit breaker
4. **Post-recovery**: Analyze application transaction management; set statement_timeout and idle_in_transaction_session_timeout
