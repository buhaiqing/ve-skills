# NAS Knowledge Base

## Fault Pattern Library

### Pattern: Mount Target Unreachable

**Symptoms:** ECS instance cannot mount or reach NAS mount target

**Root Causes:**
1. VPC/subnet mismatch between mount target and ECS
2. Security group rules blocking NFS/SMB traffic
3. Mount target in `creating` state

**Resolution Steps:**
1. Verify mount target status: `ve nas DescribeMountTargets`
2. Check VPC routing between ECS and mount target
3. Verify security group allows NFS (port 2049) or SMB (port 445)
4. Recreate mount target if status is stuck

### Pattern: Performance Degradation

**Symptoms:** Slow file operations, high latency

**Root Causes:**
1. Wrong storage tier for workload
2. Single directory with too many files (>100K)
3. Network congestion between ECS and NAS

**Resolution Steps:**
1. Check current IOPS/latency metrics via CMS
2. Review storage type vs workload requirements
3. Restructure file hierarchy if needed
4. Consider upgrading to Performance or Extreme tier

### Pattern: Capacity Exhaustion

**Symptoms:** Write failures, out-of-space errors

**Root Causes:**
1. Unexpected file growth
2. Log files not rotated
3. Backup retention too long

**Resolution Steps:**
1. Identify largest directories: `du -sh /*/ | sort -hr`
2. Clean up old/stale files
3. Archive cold data to TOS
4. Plan storage expansion
