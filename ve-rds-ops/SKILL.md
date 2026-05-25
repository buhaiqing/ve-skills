---
name: ve-rds-ops
description: >-
  Use when the user needs to create, manage, or troubleshoot Volcengine (火山引擎)
  RDS MySQL — Instance lifecycle (CreateDBInstance, DescribeDBInstances, DeleteDBInstance),
  databases, accounts, parameters, backups, white lists, and monitoring.
  User mentions RDS, MySQL, 云数据库, or describes database creation,
  account management, backup, parameter tuning scenarios.
  Not for billing, IAM, or compute/network resources.
license: MIT
compatibility: >-
  Official Volcengine CLI (`ve`, Go binary, no runtime), Go 1.14+ runtime
  (for JIT SDK fallback), valid API credentials, network access to Volcengine
  endpoints (rds-mysql.volcengineapi.com).
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-05-25"
  runtime: Harness AI Agent, Claude Code, Cursor, or compatible Agent runtimes
  go_version_minimum: "1.14"
  go_jit_runtime_version: "1.21+"
  api_profile: "RDS MySQL API V2 2022-01-01"
  cli_applicability: dual-path
  cli_support_evidence: >-
    Confirmed via `ve rds_mysql --help` — RDS MySQL is supported by the ve CLI.
  environment:
    - VOLCENGINE_ACCESS_KEY
    - VOLCENGINE_SECRET_KEY
    - VOLCENGINE_REGION
---

> This skill follows the [Agent Skill OpenSpec](https://agentskills.io/specification).

# Volcengine RDS MySQL Operations Skill

## Overview

RDS MySQL (云数据库 MySQL 版) on Volcengine (火山引擎) provides managed relational databases. This skill covers instance lifecycle, databases, accounts, parameters, backups, white lists (IP allowlists), and monitoring. It is an **operational runbook** for agents: explicit scope, credential rules, pre-flight checks, **dual-path execution** (official **SDK/API** and **`ve` CLI**), response validation, and failure recovery.

### CLI applicability

- **`cli_applicability: dual-path`:** Official `ve` CLI supports RDS MySQL. You **MUST** document **both** the SDK step **and** the `ve` CLI step for every operation.

## Five Core Standards (Quality Gates)

| # | Standard | How This Skill Fulfills It |
|---|----------|---------------------------|
| 1 | **Clear Boundaries** | SHOULD/SHOULD NOT Use conditions with precise triggers and delegation rules |
| 2 | **Structured I/O** | `{{env.VOLCENGINE_*}}` (env), `{{user.*}}` (interactive), `{{output.*}}` (API response) |
| 3 | **Explicit Actionable Steps** | Every operation: Pre-flight → Execute → Validate → Recover |
| 4 | **Complete Failure Strategies** | Error taxonomy with ≥ 10 RDS-specific codes; HALT vs retry per type |
| 5 | **Absolute Single Responsibility** | One product (RDS MySQL); cross-product delegation documented |

## Trigger & Scope (Agent-Readable)

### SHOULD Use This Skill When

- User mentions "Volcengine RDS", "火山引擎 RDS", "云数据库 MySQL", "RDS MySQL"
- Task involves **instance lifecycle**: CreateDBInstance, DescribeDBInstances, DeleteDBInstance, RestartDBInstance, ModifyDBInstanceSpec
- Task involves **databases**: CreateDB, DescribeDatabases, DeleteDB, ModifyDBDescription
- Task involves **accounts**: CreateAccount, DescribeAccounts, DeleteAccount, ResetAccountPassword, GrantAccountPrivileges, RevokeAccountPrivileges
- Task involves **parameters**: DescribeDBInstanceParameters, ModifyDBInstanceParameters
- Task involves **backups**: CreateBackup, DescribeBackups, DeleteBackup, RestoreDBInstance
- Task involves **white lists**: ModifyAllowList, DescribeAllowLists, ModifyAllowListAttributes
- Task keywords: 数据库创建, 账号管理, 备份恢复, 参数调优, IP白名单

### SHOULD NOT Use This Skill When

- Task is about **PostgreSQL** → delegate to `ve-rds-pg-ops` (when present)
- Task is about **Redis** → delegate to `ve-redis-ops`
- Task is about **ECS compute** → delegate to `ve-ecs-ops`
- Task is about **VPC networking** → delegate to `ve-vpc-ops`
- Task is purely billing → delegate to billing ops
- User insists on **console-only** flows → state limitation

### Delegation Rules

- RDS requires VPC + Subnet → verify via `ve-vpc-ops`
- RDS requires Security Group → verify via `ve-ecs-ops` or this skill
- RDS access from outside VPC may need EIP → delegate to `ve-eip-ops`
- RDS may use CLB for read/write splitting → delegate to `ve-clb-ops`

## Variable Convention (Agent-Readable)

| Placeholder | Meaning | Agent Action |
|-------------|---------|--------------|
| `{{env.VOLCENGINE_ACCESS_KEY}}` | From runtime environment | NEVER ask; fail if unset |
| `{{env.VOLCENGINE_SECRET_KEY}}` | From runtime environment | NEVER ask; fail if unset |
| `{{user.region}}` | User-supplied region | Ask once; reuse; default from env |
| `{{user.instance_id}}` | RDS instance ID | Format `mysql-xxxxxxxxx` |
| `{{user.instance_name}}` | RDS instance name | Ask once; reuse |
| `{{user.db_engine_version}}` | MySQL version | `MySQL_5_7` or `MySQL_8_0` |
| `{{user.instance_type}}` | Instance type | Single/HA/MultiNode |
| `{{user.node_spec}}` | Node spec | e.g., `rds.mysql.2c4g` |
| `{{user.storage_space}}` | Storage in GB | e.g., `100` |
| `{{user.storage_type}}` | Storage type | `LocalSSD` or `ESSD` |
| `{{user.vpc_id}}` | VPC ID | Format `vpc-xxxxxxxxx` |
| `{{user.subnet_id}}` | Subnet ID | Format `subnet-xxxxxxxxx` |
| `{{user.database_name}}` | Database name | Ask once; reuse |
| `{{user.account_name}}` | Account name | Ask once; reuse |
| `{{user.account_password}}` | Account password | Ask interactively; NEVER echo |
| `{{user.parameter_name}}` | Parameter name | e.g., `max_connections` |
| `{{output.instance_id}}` | From CreateDBInstance response | Parse from `$.Result.InstanceId` |
| `{{output.backup_id}}` | From CreateBackup response | Parse from `$.Result.BackupId` |

> **Security Warning (Credential Masking — MANDATORY):** **NEVER** log, print, or expose `VOLCENGINE_SECRET_KEY` or account passwords.

## API and Response Conventions (Agent-Readable)

- **Volcengine RDS MySQL OpenAPI V2 (2022-01-01)** is canonical.
- **Service:** `rds_mysql`
- **Endpoint:** `rds-mysql.volcengineapi.com` or `rds-mysql.{region}.volcengineapi.com`
- **Protocol:** HTTPS, JSON body
- **Errors:** Responses with `Error` object containing `Code` and `Message` fields.
- **Timestamps:** ISO 8601 format.

### Key Response Fields

| Operation | JSON Path | Type | Description |
|-----------|-----------|------|-------------|
| CreateDBInstance | `$.Result.InstanceId` | string | Instance ID |
| CreateDBInstance | `$.Result.OrderNo` | string | Order number |
| DescribeDBInstances | `$.Result.Instances` | array | Instance list |
| DescribeDBInstances | `$.Result.Instances[].InstanceId` | string | Instance ID |
| DescribeDBInstances | `$.Result.Instances[].InstanceName` | string | Instance name |
| DescribeDBInstances | `$.Result.Instances[].InstanceStatus` | string | Status |
| DescribeDBInstances | `$.Result.Instances[].DBEngineVersion` | string | MySQL version |
| DescribeDBInstances | `$.Result.Instances[].InstanceType` | string | Single/HA/MultiNode |
| DescribeDBInstances | `$.Result.Instances[].PrimaryIp` | string | Primary IP |
| DescribeDBInstances | `$.Result.Instances[].Port` | integer | Port (default 3306) |
| DescribeDBInstances | `$.Result.Instances[].VpcId` | string | VPC ID |
| DescribeDBInstances | `$.Result.Instances[].StorageSpace` | integer | Storage GB |
| DescribeDBInstances | `$.Result.TotalCount` | integer | Total matching instances |
| DescribeDBInstances | `$.Result.NextToken` | string | Pagination token |
| CreateDB | `$.Result.DBName` | string | Database name |
| DescribeDatabases | `$.Result.Databases` | array | Database list |
| CreateAccount | `$.Result.AccountName` | string | Account name |
| DescribeAccounts | `$.Result.Accounts` | array | Account list |
| DescribeDBInstanceParameters | `$.Result.Parameters` | array | Parameter list |
| CreateBackup | `$.Result.BackupId` | string | Backup ID |
| DescribeBackups | `$.Result.Backups` | array | Backup list |
| DescribeAllowLists | `$.Result.AllowLists` | array | Allow list list |

## Quick Start

### What This Skill Does
This skill enables you to create, manage, and troubleshoot Volcengine (火山引擎) RDS MySQL instances, databases, accounts, parameters, and backups using the `ve` CLI (primary) or JIT Go SDK (fallback).

### Prerequisites
- [ ] `ve` CLI installed (or Go runtime for JIT fallback)
- [ ] Credentials configured
- [ ] Region set
- [ ] VPC + Subnet ready — see `ve-vpc-ops`

### Verify Setup
```bash
ve rds_mysql DescribeDBInstances --Region {{env.VOLCENGINE_REGION}}
```

### Your First Command
```bash
# List all RDS MySQL instances in the configured region
ve rds_mysql DescribeDBInstances --Region {{env.VOLCENGINE_REGION}}
```

## Capabilities at a Glance

| Operation | Description | Complexity | Risk Level |
|-----------|-------------|------------|------------|
| CreateDBInstance | Create RDS MySQL instance | High | Low |
| DescribeDBInstances | Query instance list | Low | None |
| DeleteDBInstance | Delete instance | Low | **High** — irreversible |
| RestartDBInstance | Restart instance | Low | Medium — service interruption |
| ModifyDBInstanceSpec | Change instance spec | Medium | Medium — restart required |
| CreateDB | Create database | Low | Low |
| DescribeDatabases | List databases | Low | None |
| DeleteDB | Delete database | Low | **High** |
| CreateAccount | Create database account | Low | Low |
| DescribeAccounts | List accounts | Low | None |
| DeleteAccount | Delete account | Low | Medium |
| ResetAccountPassword | Reset account password | Low | Medium |
| GrantAccountPrivileges | Grant DB privileges to account | Low | Medium |
| DescribeDBInstanceParameters | Query parameters | Low | None |
| ModifyDBInstanceParameters | Modify parameters | Medium | Medium — some require restart |
| CreateBackup | Create backup | Low | Low |
| DescribeBackups | List backups | Low | None |
| DeleteBackup | Delete backup | Low | Medium |
| RestoreDBInstance | Restore from backup | High | Medium |
| DescribeAllowLists | Query IP allow lists | Low | None |
| ModifyAllowList | Modify IP allow list | Low | Medium |

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-25 | Initial release with instance lifecycle, databases, accounts, parameters, backups |

## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**.

### Operation: DescribeDBInstances — Query Instance List

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY"` | Set | HALT |
| CLI | `ve version` | Exit code 0 | Install ve CLI |

#### Execution

```bash
# List all RDS MySQL instances
ve rds_mysql DescribeDBInstances --Region "{{user.region}}"

# Filter by VPC
ve rds_mysql DescribeDBInstances --Region "{{user.region}}" --VpcId "{{user.vpc_id}}"

# Filter by name
ve rds_mysql DescribeDBInstances --Region "{{user.region}}" --InstanceName "{{user.instance_name}}"

# Pagination
ve rds_mysql DescribeDBInstances --Region "{{user.region}}" --PageNumber 1 --PageSize 50
```

#### Validation

1. Parse `$.Result.Instances[]` for instance details
2. Report instance IDs, names, statuses, IPs, versions

---

### Operation: CreateDBInstance — Create RDS MySQL Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | DescribeVpcs (ve-vpc-ops) | VPC found | HALT |
| Subnet exists | DescribeSubnets (ve-vpc-ops) | Subnet found | HALT |
| Quota | Check RDS quota | Sufficient | HALT; request increase |

#### Execution

```bash
ve rds_mysql CreateDBInstance \
  --Region "{{user.region}}" \
  --InstanceName "{{user.instance_name}}" \
  --DBEngineVersion "MySQL_8_0" \
  --InstanceType "HA" \
  --NodeSpec "rds.mysql.2c4g" \
  --StorageSpace 100 \
  --StorageType "ESSD" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --ChargeType "PostPaid" \
  --NumberOfInstances 1 \
  --NodeInfo '[{"ZoneId":"cn-beijing-a"},{"ZoneId":"cn-beijing-b"}]'
```

#### Post-execution Validation

1. Parse `{{output.instance_id}}` from response
2. Poll DescribeDBInstances until InstanceStatus is `Running`
3. Report instance ID, connection info (IP:Port), and creation status

#### Poll Pattern

```bash
for i in {1..60}; do
  STATUS=$(ve rds_mysql DescribeDBInstances --Region "$VOLCENGINE_REGION" --InstanceId "$INSTANCE_ID" \
    | jq -r '.Result.Instances[0].InstanceStatus')
  [ "$STATUS" = "Running" ] && echo "Instance ready" && break
  echo "Waiting... (attempt $i, status: $STATUS)"
  sleep 10
done
```

---

### Operation: DeleteDBInstance

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: irreversible deletion
- **MUST** verify instance has recent backup or data is not needed
- Note: Only PostPaid instances support API deletion; PrePaid requires unsubscription via console

```bash
ve rds_mysql DeleteDBInstance \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --DataKeepPolicy "Last" \
  --DataKeepDays 7
```

---

### Operation: CreateDB — Create Database

#### Execution

```bash
ve rds_mysql CreateDB \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --DBName "{{user.database_name}}" \
  --CharacterSet "utf8mb4" \
  --Description "{{user.db_description}}"
```

---

### Operation: CreateAccount — Create Database Account

#### Execution

```bash
ve rds_mysql CreateAccount \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.account_name}}" \
  --AccountPassword "{{user.account_password}}" \
  --AccountType "Normal"
```

#### Grant Privileges

```bash
ve rds_mysql GrantAccountPrivileges \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --AccountName "{{user.account_name}}" \
  --Privileges '[{"DBName":"{{user.database_name}}","AccountPrivilege":"ReadWrite"}]'
```

---

### Operation: DescribeDBInstanceParameters — Query Parameters

```bash
ve rds_mysql DescribeDBInstanceParameters \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --ParameterName "max_connections"
```

### Operation: ModifyDBInstanceParameters

```bash
ve rds_mysql ModifyDBInstanceParameters \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --Parameters '[{"ParameterName":"max_connections","ParameterValue":"1000"}]'
```

> **Warning:** Some parameters require instance restart (`ForceRestart: true`). The skill MUST check this field before modifying and warn the user.

---

### Operation: CreateBackup

```bash
ve rds_mysql CreateBackup \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --BackupName "{{user.backup_name}}" \
  --BackupStrategy "Snapshot"
```

### Operation: DescribeAllowLists — Query IP White List

```bash
ve rds_mysql DescribeAllowLists --Region "{{user.region}}" --InstanceId "{{user.instance_id}}"
```

### Operation: ModifyAllowList — Set IP White List

```bash
ve rds_mysql ModifyAllowList \
  --Region "{{user.region}}" \
  --InstanceId "{{user.instance_id}}" \
  --AllowListName "default" \
  --AllowListType "IPv4" \
  --AllowList '["10.0.0.0/8", "172.16.0.0/12"]'
```

---

### Operation: RestoreDBInstance — Restore from Backup

```bash
ve rds_mysql RestoreDBInstance \
  --Region "{{user.region}}" \
  --InstanceId "{{source_instance_id}}" \
  --BackupId "{{user.backup_id}}" \
  --RestoreType "Backup" \
  --NewInstanceName "{{user.new_instance_name}}"
```

---

## Reference Directory

- [Core Concepts](references/core-concepts.md)
- [API & SDK Usage](references/api-sdk-usage.md)
- [CLI Usage](references/cli-usage.md)
- [Troubleshooting Guide](references/troubleshooting.md)
- [Monitoring](references/monitoring.md)
- [Integration](references/integration.md)
- [User Experience Specification](../../ve-skill-generator/references/user-experience-spec.md)
- [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md)
- [CLI Behavioral Reference](../../ve-skill-generator/references/cli-behavior.md)

## Operational Best Practices

- **Instance Type:** Use `HA` for production (primary + read replica), `Single` for dev/test
- **Storage:** ESSD for production I/O, LocalSSD for low latency; size ≥ 100GB for production
- **Node Spec:** Match workload (rds.mysql.2c4g for light, rds.mysql.4c16g for medium, rds.mysql.8c32g for heavy)
- **Multi-AZ:** Use `NodeInfo` with different zones for HA instances
- **Account Security:** Use strong passwords, grant minimum privileges (read-only for analytics accounts)
- **Backup:** Create before parameter changes or spec modifications; enable automated backup
- **Parameter Tuning:** Always check `ForceRestart` field; plan maintenance window for restart-required parameters
- **White List:** Never use `0.0.0.0/0` in production; restrict to specific CIDRs
- **Least privilege:** Use IAM policies scoped to RDS APIs only
