---
name: ve-security-group-ops-prompt-templates
description: GCL prompt templates for ve-security-group-ops. G/C/O + safety prompts for DeleteSecurityGroup, RevokeSecurityGroup*, AuthorizeSecurityGroup* (0.0.0.0/0 on sensitive ports).
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-security-group-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-security-group-ops

## 1. Generator Prompt

```text
Generator for ve-security-group-ops. Execute via `ve vpc <Action>` or Go SDK.
- NEVER ask {{env.*}}. Dual-path. Trace to gcl-trace-*.json.
- Surface safety prompts from §4 for destructive/state-changing ops.
- For AuthorizeSecurityGroup* with 0.0.0.0/0 on sensitive ports (22,3389,3306,6379): warn.
- For RevokeSecurityGroup*: warn about breaking access. NEVER print VOLCENGINE_SECRET_KEY.
Output JSON: { "status", "operation", "sg_id", "rule_summary", "command", "request_id", "validation", "trace_path" }
```

## 2. Critic Prompt (MUST NOT see raw user request)

```text
Independent SG auditor. Score against rubric.
- Verify VOLCENGINE_SECRET_KEY not in trace.
- Verify DeleteSecurityGroup: instance-usage check in trace.
- Verify RevokeSecurityGroup* on 0.0.0.0/0 sensitive port: double confirmation.
- Verify AuthorizeSecurityGroup* 0.0.0.0/0 sensitive port: internet-exposure warning.
Output: { "scores": {...}, "suggestions": [...], "blocking": bool }
```

## 3. Orchestrator Prompt

```text
Orchestrator. Safety=0 AND destructive/state-changing → ABORT. All pass → return. iter<max_iter → loop.
```

## 4. Safety Prompts

### 4.1 DeleteSecurityGroup
```text
[DESTRUCTIVE] Delete SG {{user.sg_id}} ({{user.sg_name}}).
Attached instances: {{output.attached_instance_count}}. IRREVERSIBLE.
Type "yes" to delete, or "cancel".
```

### 4.2 RevokeSecurityGroupIngress (sensitive port variant)
```text
[DESTRUCTIVE] Revoke inbound rule on {{user.sg_id}}:
  CIDR: {{user.cidr}}, Port: {{user.port}}, Protocol: {{user.protocol}}
WARNING: If this is the only rule allowing traffic on port {{user.port}},
instances may lose connectivity. Type "yes" to revoke, or "cancel".
```

### 4.3 RevokeSecurityGroupEgress
```text
[DESTRUCTIVE] Revoke outbound rule on {{user.sg_id}}:
  CIDR: {{user.cidr}}, Port: {{user.port}}, Protocol: {{user.protocol}}
WARNING: Instances may lose outbound connectivity. Type "yes", or "cancel".
```

### 4.4 AuthorizeSecurityGroupIngress (0.0.0.0/0 sensitive port)
```text
[STATE-CHANGING] Add inbound rule on {{user.sg_id}}:
  CIDR: 0.0.0.0/0, Port: {{user.port}}, Protocol: {{user.protocol}}
WARNING: This opens port {{user.port}} to THE INTERNET. If this is 22/3389/3306/6379,
instances are exposed to brute-force attacks. Type "yes" to confirm, or "cancel".
```

### 4.5 AuthorizeSecurityGroupEgress (0.0.0.0/0)
```text
[STATE-CHANGING] Add outbound rule on {{user.sg_id}}:
  CIDR: 0.0.0.0/0, Port: all
WARNING: Unrestricted outbound to internet (data exfiltration risk).
Type "yes" to confirm, or "cancel".
```

## 5. Changelog
| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-security-group-ops |