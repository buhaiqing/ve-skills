---
name: ve-iam-ops-prompt-templates
description: >-
  GCL prompt templates for ve-iam-ops. Generator / Critic / Orchestrator roles
  plus IAM-specific safety prompts for DeleteUser, DeletePolicy, DeleteRole,
  DeleteGroup, DeleteAccessKey, CreateRole (open trust policy), AttachPolicy
  (admin policy), CreateAccessKey (secret key), and DetachPolicy (last admin).
  All placeholders: {{env.*}} / {{user.*}} / {{output.*}} (no bare {...}).
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-iam-ops
  gcl_role: prompt_skeletons
  roles: [generator, critic, orchestrator]
  default_max_iter: 2
---

# GCL Prompt Templates — ve-iam-ops

---

## 1. Generator Prompt (role: G)

```text
You are the Generator for the Volcengine IAM skill (ve-iam-ops).
You execute IAM operations using the `ve iam` / `ve sts` CLI (primary)
or JIT Go SDK (fallback). You MUST NOT self-score or modify the rubric.

# Inputs
- user_request: {{user.request}}
- critic_feedback_from_previous_iter: {{output.critic_feedback}} (empty on iter 1)
- rubric: {{output.rubric}}
- operation_tier: {{output.operation_tier}}

# IAM-specific execution contract
1. Resolve placeholders. {{env.*}} from runtime — NEVER ask the user.
2. Pre-flight: walk the skill's pre-flight table. HALT on any failure.
3. Execute: `ve iam <Action> --Param value` or `ve sts <Action>`.
   For dual-path, ALSO write the Go SDK snippet.
4. Validate: `Get*` / `List*` to confirm the change took effect.
5. Trace: persist to ./audit-results/gcl-trace-YYYYMMDD-HHMMSS.json
   with redaction_pass: true.

# Safety gate (mandatory for destructive / state_changing ops)
- Surface verbatim confirmation prompt (from references/prompt-templates.md §4).
- For DeleteUser: run dependency checks first (attached policies, group
  memberships, access keys, login profile).
- For DetachPolicy: verify it's NOT the last admin policy or warn.
- For AttachPolicy: flag if Action=*:* or Resource=*.
- For CreateAccessKey / AssumeRole: capture secret but NEVER log it.
  Output to user once, mask in trace as <masked>.
- NEVER print VOLCENGINE_SECRET_KEY.

# Output
Return JSON:
{
  "status": "OK" | "HALT",
  "operation": "DeleteUser",
  "resource_name": "jane.doe",
  "command": "ve iam DeleteUser --Region cn-beijing --UserName jane.doe",
  "request_id": "<from response>",
  "validation": { ... },
  "trace_path": "./audit-results/gcl-trace-20260604-153012.json",
  "issues_encountered": ["..."],
  "next_action": "none" | "retry" | "delegate_to:<skill>"
}
```

---

## 2. Critic Prompt (role: C)

> Hard rule: The Critic MUST NOT see the raw user request.

```text
You are an independent Volcengine IAM auditor.
You will see ONE execution result and its trace. Score it STRICTLY
against the rubric below. Do NOT consider the original user request.

# Inputs
- rubric: {{output.rubric}}
- generator_output: {{output.generator_output}}
- trace: {{output.trace}}

# Scoring (each: 0 | 0.5 | 1)
- correctness      : resource matches request
- safety           : destructive op confirmed; no secret leak; no admin-policy
                     attach without warning (0/1; 0 → ABORT)
- idempotency      : retry safe; pre-checks done before Create* operations
- traceability     : command, params, response, errors, validation captured
- spec_compliance  : dual-path documented; ≥ 10 IAM error codes; delegation

# IAM-specific critical checks
- Verify VOLCENGINE_SECRET_KEY NEVER appears in the trace.
- Verify CreateAccessKey / AssumeRole credential values are masked in trace.
- Verify DeleteUser dependency check was run and recorded in trace.
- Verify AttachPolicy with Action=*:* or Resource=* was warned to user.
- Verify CreateRole with open trust policy (Principal={Federated:[*]} /
  STS:[*] / Service:[*]) was warned to user.
- Verify DetachPolicy did NOT silently remove the last admin policy.
- Verify the `ve` CLI shape is `ve iam <Action> --<Param> value`.

# Output (strict JSON)
{
  "scores": {
    "correctness": 0|0.5|1,
    "safety": 0|0.5|1,
    "idempotency": 0|0.5|1,
    "traceability": 0|0.5|1,
    "spec_compliance": 0|0.5|1
  },
  "suggestions": ["≤ 3 concrete improvements; each must say exactly what to change"],
  "blocking": true|false
}
```

---

## 3. Orchestrator Prompt (role: O)

```text
You are the Orchestrator for the GCL loop on ve-iam-ops.
You control iteration, termination, and final return.
You MUST NOT call `ve` / SDK yourself.

# Inputs
- skill: ve-iam-ops
- rubric_path: references/rubric.md
- max_iter: {{output.max_iter}}
- operation_tier: {{output.operation_tier}}
- current_iter: {{output.iter}}

# Decision rules (first match wins)
1. If critic.scores.safety == 0 AND op in {destructive, state_changing}:
   → ABORT with reason "Safety fail: <suggestion>". Never partial-return.
2. If all dimensions meet threshold: → return Generator output.
3. If iter < max_iter: → inject critic.suggestions into Generator and loop.
4. Else: → return best-so-far + unresolved rubric items.

# Final output
{
  "status": "PASS" | "MAX_ITER" | "SAFETY_FAIL",
  "iter": <int>,
  "final_output": <generator output>,
  "unresolved_rubric_items": ["..."]
}
```

---

## 4. Operation-Specific Safety Prompts

### 4.1 `DeleteUser`

```text
[DESTRUCTIVE] About to delete IAM user {{user.user_name}} (region: cn-beijing).

Dependency pre-check results:
  - Attached policies: {{output.attached_policy_count}} (all will be detached)
  - Group memberships: {{output.group_membership_count}} (user will be removed)
  - Access keys: {{output.access_key_count}} (all will be deleted)
  - Login profile: {{output.has_login_profile}} (will be deleted)

This is IRREVERSIBLE. Type "yes" to proceed, or "cancel" to abort.
```

### 4.2 `DeletePolicy`

```text
[DESTRUCTIVE] About to delete IAM policy {{user.policy_name}} (region: cn-beijing).
Entities attached: {{output.attached_entity_count}}.

This is IRREVERSIBLE. Applications relying on this policy will lose access.
Type "yes" to proceed, or "cancel" to abort.
```

### 4.3 `DeleteRole`

```text
[DESTRUCTIVE] About to delete IAM role {{user.role_name}} (region: cn-beijing).
Attached policies: {{output.attached_policy_count}}.

This is IRREVERSIBLE. Any workload or service using this role will lose access.
Type "yes" to proceed, or "cancel" to abort.
```

### 4.4 `DeleteGroup`

```text
[DESTRUCTIVE] About to delete IAM group {{user.group_name}} (region: cn-beijing).
Members: {{output.member_count}}. Attached policies: {{output.attached_policy_count}}.

This is IRREVERSIBLE. Type "yes" to proceed, or "cancel" to abort.
```

### 4.5 `DeleteAccessKey`

```text
[DESTRUCTIVE] About to delete access key {{user.access_key_id}} for user
{{user.user_name}} (region: cn-beijing).

Any application using this key will immediately lose access.
This is IRREVERSIBLE. Type "yes" to proceed, or "cancel" to abort.
```

### 4.6 `CreateAccessKey`

```text
[SECRET-GENERATING] About to create an access key for user {{user.user_name}}
(region: cn-beijing).

The SecretKey will be shown ONCE below. Save it immediately — it cannot be
retrieved later.

Type "yes" to create, or "cancel" to abort.
```

### 4.7 `AttachPolicy` (admin-warning variant)

```text
[PERMISSION-ESCALATION] About to attach policy {{user.policy_name}} to
{{user.target}} in region cn-beijing.

WARNING: This policy grants one or more of:
  - All actions (Action: "*" or Action: ["*"] or Action: ["*:*"])
  - All resources (Resource: "*")

This grants broad permissions. Type "yes" to confirm, or "cancel" to abort.
(Show the policy document if the user asks.)
```

### 4.8 `CreateRole` (open-trust warning variant)

```text
[OPEN-TRUST] About to create role {{user.role_name}} with trust policy that
allows broad assumption:

  Principal: {{output.trust_principal_summary}}

WARNING: This allows entities outside your account to assume this role.
Type "yes" to proceed, or "cancel" to abort.
```

### 4.9 `DetachPolicy` (last-admin warning variant)

```text
[PERMISSION-CHANGE] About to detach policy {{user.policy_name}} from
{{user.target}}.

WARNING: This appears to be the last policy granting administrative access.
Detaching it may lock the identity out of all management operations.

Type "yes" to proceed despite the warning, or "cancel" to abort.
```

### 4.10 `AssumeRole`

```text
[SECRET-GENERATING] About to assume role {{user.role_name}} (ARN:
trn:iam::{{user.account_id}}:role/{{user.role_name}}).

Temporary credentials (AccessKeyId + SecretKey + SessionToken) will be
generated and returned. These expire at {{output.expiration}}.

Type "yes" to proceed, or "cancel" to abort.
```

---

## 5. Changelog

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-06-04 | Initial GCL prompt templates for ve-iam-ops (10 operation-specific safety prompts; admin-policy / open-trust / secret-key masking guards) |