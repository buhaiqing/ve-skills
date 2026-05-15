# Optimization Analysis Framework

> **Purpose:** Provides a three-dimensional optimization framework for analyzing and improving generated skills. Use this framework during skill generation and review to ensure comprehensive coverage.
> **Version:** 1.0.0
> **Last Updated:** 2026-05-15

---

## Table of Contents

1. [Three-Dimensional Optimization Model](#1-three-dimensional-optimization-model)
2. [Dimension 1: Fault Diagnosis (故障诊断)](#2-dimension-1-fault-diagnosis)
3. [Dimension 2: Root Cause Localization (根因定位)](#3-dimension-2-root-cause-localization)
4. [Dimension 3: Rapid Resolution (快速恢复)](#4-dimension-3-rapid-resolution)
5. [Cross-Dimensional Analysis](#5-cross-dimensional-analysis)
6. [Optimization Checklist](#6-optimization-checklist)

---

## 1. Three-Dimensional Optimization Model

Every generated skill is evaluated across three independent dimensions:

```
                    Fault Diagnosis
                         /\
                        /  \
                       /    \
          Root Cause  ------  Rapid
         Localization        Resolution
```

| Dimension | Focus Question | Success Metric |
|-----------|---------------|----------------|
| **Fault Diagnosis** | "Can we detect anomalies quickly and accurately?" | Detection time < 30s, false positive rate < 5% |
| **Root Cause Localization** | "Can we pinpoint the exact cause?" | Mean time to isolate < 5min, accuracy > 85% |
| **Rapid Resolution** | "Can we recover with minimal impact?" | Mean time to recover < 10min, automation rate > 70% |

---

## 2. Dimension 1: Fault Diagnosis (故障诊断)

### 2.1 Pre-flight State Validation

Every operation MUST validate the current state before execution:

- **Credential validation:** Check env vars exist before API calls
- **Region availability:** Verify region is valid and service is available
- **Resource state:** Confirm resource exists and is in expected state
- **Quota check:** Verify sufficient quota before create operations

### 2.2 Health Check Patterns

All installation flows MUST include health checks with scoring:

```bash
# Example health check scoring
HEALTH_SCORE=0
# Binary exists: +2, In PATH: +2, Version works: +2, API call works: +2
# Total: 8/10 minimum for pass
```

### 2.3 Error Taxonomy

Each skill MUST define ≥ 10 product-specific error codes with:
- Error code name and HTTP status
- Human-readable description
- Detection pattern (regex, JSON path, status code)
- Severity level (critical, warning, info)

---

## 3. Dimension 2: Root Cause Localization (根因定位)

### 3.1 Dependency Mapping

Document resource dependencies in `core-concepts.md`:

```
VPC → VSwitch → SecurityGroup → Instance
                      ↓
                  LoadBalancer
```

### 3.2 Causal Chain Analysis

For each error pattern, trace the causal chain:

| Symptom | Possible Cause | Verification | Confidence |
|---------|---------------|--------------|------------|
| 403 Forbidden | Insufficient IAM permissions | Call IAM policy API | High |
| 403 Forbidden | Wrong region endpoint | Check Region parameter | Medium |
| 403 Forbidden | Expired credentials | Check token expiry | High |

### 3.3 Log Correlation

When available, document how to correlate:
- API request IDs with service logs
- Resource IDs with monitoring metrics
- Timestamps across multiple systems

---

## 4. Dimension 3: Rapid Resolution (快速恢复)

### 4.1 Auto-Remediation Patterns

Define which errors are safe to auto-remediate:

| Error Type | Auto-Remediation | Max Retries | Fallback |
|-----------|-----------------|-------------|----------|
| Throttling | Exponential backoff | 3 | HALT with escalation |
| InternalError | Retry with same params | 3 | HALT with RequestId |
| NetworkTimeout | Retry with increased timeout | 2 | Switch endpoint |
| InvalidParameter | Fix and retry once | 1 | HALT with parameter guide |

### 4.2 Recovery Macros

Provide reusable recovery macros for common patterns:

```bash
# Macro: Retry with exponential backoff
retry_with_backoff() {
    local max_retries=$1
    local base_delay=$2
    local attempt=1
    
    while [ $attempt -le $max_retries ]; do
        if "$@"; then
            return 0
        fi
        local delay=$((base_delay * 2 ** (attempt - 1)))
        echo "Retrying in ${delay}s... (Attempt ${attempt}/${max_retries})"
        sleep $delay
        attempt=$((attempt + 1))
    done
    return 1
}
```

### 4.3 Escalation Templates

Standardized escalation template for HALT scenarios:
- Request ID
- Operation name and parameters (masked)
- Resource ID
- Error code and message
- Timestamp
- Support channel URL (Volcengine: https://console.volcengine.com/ticket)

---

## 5. Cross-Dimensional Analysis

### 5.1 Dimension Coverage Matrix

| Operation | Fault Diagnosis | Root Cause | Rapid Resolution | Coverage |
|-----------|----------------|------------|------------------|----------|
| Create | Pre-flight checks | Dependency map | Retry macros | 3/3 ✅ |
| Describe | State validation | N/A | N/A | 1/3 ⚠️ |
| Modify | Pre-flight checks | Causal chain | Retry macros | 3/3 ✅ |
| Delete | Safety gate | N/A | HALT pattern | 2/3 ⚠️ |

### 5.2 Gap Analysis

For any operation with < 3/3 coverage:
1. Identify missing dimension
2. Add appropriate checks/patterns
3. Re-evaluate coverage

---

## 6. Optimization Checklist

- [ ] **Fault Diagnosis:** Pre-flight state validation for every operation
- [ ] **Fault Diagnosis:** Health check scoring ≥ 8/10 for installations
- [ ] **Fault Diagnosis:** Error taxonomy ≥ 10 codes per product
- [ ] **Root Cause:** Dependency mapping documented in core-concepts.md
- [ ] **Root Cause:** Causal chain analysis for ambiguous errors (403, 500)
- [ ] **Root Cause:** Log correlation patterns documented
- [ ] **Rapid Resolution:** Auto-remediation defined for safe errors
- [ ] **Rapid Resolution:** Recovery macros available for retry patterns
- [ ] **Rapid Resolution:** Escalation template includes all required fields
- [ ] **Cross-Dimensional:** Coverage matrix shows ≥ 2/3 for all operations

---

*Use this framework during skill generation review. Update quarterly based on incident learnings.*
