# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

> **Content hierarchy**: This file is the entry point. Detailed specifications live in `docs/`:
> - `docs/gcl-spec.md` — full GCL specification (purpose, roles, loop flow, trace, prompt templates, changelog)
> - `docs/token-efficiency.md` — full TE rules with code examples (TE-1 through TE-9)
>
> Keep this file in sync with `docs/` files. When updating either, verify the other stays aligned.

---

## Repository Type — Read This First

This repo contains **only Markdown skill specifications**. There is:

- **No build system, no tests, no lint, no package manager.** Do not run `go build`, `go test`, `npm`, `pip`, `make`, etc. — they will fail or do nothing.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md` (see also **Files that DO NOT exist** below).

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

---

## MANDATORY: Two-Round Self-Review After Every Skill Update

After ANY edit to a `ve-*-ops/SKILL.md`, its `references/*.md`, `assets/*`, or to `ve-skill-generator/**`, you MUST run **two rounds of self-review** before declaring done. Fix every issue surfaced — do not just list them.

### Round 1 — Structural & Specification Compliance

Verify against `ve-skill-generator/SKILL.md` and `ve-skill-generator/references/ve-skill-template.md`:

| # | Check | Detail |
|---|-------|--------|
| C1 | Frontmatter | name, description, license, compatibility, metadata complete |
| C2 | cli_applicability | set; if `dual-path`, both CLI and SDK steps exist |
| C3 | Five Core Standards | boundaries, I/O, steps, failures, single responsibility present & accurate |
| C4 | Placeholder channels | `{{env.*}}` for secrets/region, `{{user.*}}` for interactive, `{{output.*}}` for captured |
| C5 | Credential safety | No credential value logged/printed/echoed — only `<masked>` or existence checks |
| C6 | **Token Efficiency** | **See §Token Efficiency below; verify TE-1~TE-7** |
| C7 | Safety gates | Every destructive operation (delete, stop, release, restore) has explicit confirmation |
| C8 | Error taxonomy | ≥ 10 product-specific codes with HALT vs retry classification |
| C9 | Reference files | All 6 standard files exist when applicable |
| C10 | Cross-product delegation | ECS skill delegates IAM work to IAM skill, etc. |
| C11 | Version bump | `last_updated` and `version` bumped if behavior changed |
| C12 | **Content dedup** | No duplicate operation flows across SKILL.md ↔ references/; TE-6 verified (see §Link & Dedup) |

### Round 2 — Accuracy & Anti-Pattern Sweep

Verify against reality, not memory:

| # | Check | Detail |
|---|-------|--------|
| F1 | API fidelity | Every field/flag grounded in official OpenAPI or `ve <service> --help` |
| F2 | Command prefix | `ve` (not `volcengine-cli`, renamed at v1.0.20) |
| F3 | CLI shape | `ve <service> <Action> --<Param> value` (PascalCase actions and params) |
| F4 | JSON bodies | `--body '{"Key":"Value"}'` form, not invented flags |
| F5 | Originality | No generic prose, speculative claims, or copy-pasted irrelevant boilerplate |
| F6 | Links | Sibling references use correct relative paths |
| F7 | Examples | Minimal, runnable, not pseudocode |
| F8 | Rendering | Tables aligned, code fences closed, headings hierarchical |
| **F9** | **Link integrity** | Cross-references between docs/ and AGENTS.md verified; all internal `[...](...)` resolve to existing files (see §Link & Dedup) |
| **F10** | **Dedup integrity** | No repeated content across multiple files; after every change run dedup check (see §Link & Dedup) |

**Fix every issue found in both rounds before responding "done".** If a finding requires information you don't have, mark it `[blocked: needs OpenAPI verification]`.

### Validation Command Matrix

| Change scope | Required command |
|---|---|
| Full local validation before PR / handoff | Walk P0/P1 checklist in `ve-skill-generator/SKILL.md` (no automated runner in this repo) |
| Any `SKILL.md` frontmatter or metadata change | Re-check P0/P1 checklist items C1–C12 manually |
| Any GCL rubric, prompt template, or `## Quality Gate (GCL)` section change | Verify against `docs/gcl-spec.md` §Rubric + §Prompt Templates |
| Any Markdown spec, README, or path-reference change | Run the link-integrity scan in **§Document Integrity & Link Validation** below |
| Any `docs/*.md` change | Verify cross-reference symmetry (see Layer 2 below) |
| Any Go SDK example code change | `python3 -m py_compile <file>` if Python; Go examples are illustrative — no build step |

> **Note:** This repo has automated validation scripts in `scripts/` for pre-commit checks (see **Files that DO NOT exist** below). After changes, run `python3 scripts/validate_local.py` if available, then supplement with manual checks against the P0/P1 checklist.

---

## Token Efficiency Requirements (P0 — 强制)

> 在保持 Agent 可执行性的前提下，最小化每个 Skill 的 Token 消耗。
> 完整规则详解（含代码示例）见 [docs/token-efficiency.md](docs/token-efficiency.md)。

| 规则 | 要点 | 节省 |
|------|------|------|
| **TE-1** API 查询 > 静态表格 | 用 `ve` 命令获取版本/配额，不硬编码 | ~200-500/文件 |
| **TE-2** 省略不必要 docstring | Go SDK 用 `#` 注释代替函数级 docstring | ~100-200/函数 |
| **TE-3** 紧凑错误表 | 每行 1 个错误码，≤3 列 | ~300-500/文件 |
| **TE-4** JSON paths 集中声明 | 文件顶部统一声明，不重复 | ~50-100/文件 |
| **TE-5** YAML anchors | `example-config.yaml` 用 `&anchor` 消除重复 | ~200-400/文件 |
| **TE-6** 消除跨文件重复 | SKILL.md 已有完整流程，references 不重复 | 因 Skill 而异 |
| **TE-7** 专业内容分层 | AIOps/FinOps 放 `references/advanced/`；安全敏感操作单独标注并显式确认 | ~3,000-8,000/文件 |
| **TE-8** 符号与缩写 | 使用 `→` `⇒` `✅` `❌` 等符号和标准缩写替代冗余自然语言 | ~100-300/文件 |
| **TE-9** 压缩级别选择 | 按文档用途选择 Minimal / Efficient / Compressed / Critical / Emergency | ~500-2,000/文件 |

**不可压缩的内容**：Agent 可执行命令本身（参数、JSON paths）、错误恢复逻辑、安全门、Credential 规则、跨技能编排链。

**发现任一违规 → 立即修复 → 重新检查直到全部通过。**

### Symbol System

Standard Token Efficiency symbols for compact skill documentation:

| Category | Symbols | Usage |
|----------|---------|-------|
| **Core Logic** | `→` (leads to), `⇒` (transforms to), `←` (rollback), `⇄` (bidirectional), `&` (and), `\|` (separator), `:` (define), `»` (sequence), `∴` (therefore), `∵` (because), `≡` (equivalent), `≈` (approximately), `≠` (not equal) | Flow mapping, state transitions, causality |
| **Status** | `✅` (completed), `❌` (failed), `⚠️` (warning), `ℹ️` (info), `🔄` (in progress), `⏳` (pending), `🚨` (critical), `🎯` (target), `📊` (metrics), `💡` (insight) | Checkpoint gates, validation results |
| **Domain** | `⚡` (perf), `🔍` (analysis), `🔧` (config), `🛡️` (security), `📦` (deploy), `🎨` (design), `🌐` (network), `📱` (mobile), `🏗️` (architecture), `🧩` (components) | Section markers, scope indicators |

### Abbreviations

Standard abbreviations to reduce token count while preserving meaning:

| Abbr | Meaning | Abbr | Meaning |
|------|---------|------|---------|
| `cfg` | configuration | `impl` | implementation |
| `arch` | architecture | `perf` | performance |
| `ops` | operations | `env` | environment |
| `req` | requirements | `deps` | dependencies |
| `val` | validation | `test` | testing |
| `docs` | documentation | `std` | standards |
| `qual` | quality | `sec` | security |
| `err` | error | `rec` | recovery |
| `sev` | severity | `opt` | optimization |

### Compression Levels

Adaptive compression strategy for token management:

| Level | Range | Use Case |
|-------|-------|----------|
| **Minimal** | 0-40% | Full detail, persona-optimized clarity |
| **Efficient** | 40-70% | Balanced compression with domain awareness |
| **Compressed** | 70-85% | Aggressive optimization, quality gates active |
| **Critical** | 85-95% | Maximum compression, preserve essential context |
| **Emergency** | 95%+ | Ultra-compression, information validation |

### TE 规则速查

| TE 规则 | 检查方法 | 不通过则 |
|---------|---------|---------|
| TE-1 | 检查 references/ 中是否有硬编码的版本号/配额数字 | 替换为 `ve` 查询命令 |
| TE-2 | 检查 Go SDK 代码块是否有函数级 docstring | 删除，改用 `#` 行注释 |
| TE-3 | 检查错误表是否超过 3 列 | 合并列，每行 1 个错误码 |
| TE-4 | 检查 JSON path 是否在文件顶部集中声明 | 移至文件顶部统一声明 |
| TE-5 | 检查 example-config.yaml 是否有重复字段 | 用 YAML anchors 消除 |
| TE-6 | 检查 SKILL.md 与 references/ 是否有内容重复 | 删除 references 中的重复 |
| TE-7 | 检查 AIOps/FinOps 是否在 `references/advanced/`；SQL 执行是否标注 Security-Sensitive | 分层修复 |
| TE-8 | 检查 Markdown 中是否混用了自然语言描述与 TE symbols | 统一使用 `→` `⇒` `✅` `❌` 等符号替代冗余文字 |
| TE-9 | 检查当前压缩级别是否匹配文档目的（Minimal/Efficient/Compressed） | 调整压缩策略 |

---

## Skill Authoring Guardrails

- **Never create a new `ve-*-ops/` directory by hand.** Use the `ve-skill-generator` meta-skill flow.
- **Never invent CLI flags or API parameters.** Only official OpenAPI or `ve <service> <action> --help` verified fields.
- **Single-product rule.** One `ve-*-ops` skill = one product = one primary resource model.
- **Prefer editing existing skills over creating new files.** No tutorial-style `.md` at the repo root.

## File Layout Anchors (do not relocate without reason)

```
ve-skill-generator/                  meta-skill: how to author new skills
  SKILL.md                           generator workflow + P0/P1 checklist
  references/
    ve-skill-template.md             canonical skill template
    cli-behavior.md                  verified `ve` CLI conventions
    execution-environment.md         CLI + Go SDK setup
    user-experience-spec.md          UX requirements every generated skill must follow
    governance-and-adversarial-review.md
ve-[product]-ops/                    one per Volcengine product
  SKILL.md                           main runbook
  references/                        6 standard reference files
  assets/example-config.yaml
docs/
  gcl-spec.md                        full GCL specification
  token-efficiency.md                detailed TE rules with code examples
```

`.omc/` and `.omo/` are tool-local state and are gitignored — do not commit anything inside them.

### Asset & Schema Placement Rules (mandatory)

Skill-owned artifacts MUST NOT be placed at repo root. Use this split:

| Location | Allowed contents | Forbidden |
|---|---|---|
| `ve-*-ops/assets/` | `eval_queries.json`, `example-config.yaml`, `*.schema.json`, skill-specific templates | Cross-skill executables |
| `ve-*-ops/references/` | Runbooks, output contracts in Markdown, delegation stubs | Duplicate JSON schemas that belong in `assets/` |
| `scripts/` (repo root) | Shared validation scripts (`validate_local.py`, `check_gcl_conformance.py`, etc.) — pre-commit use | JSON Schema, handoff contracts, example YAML |

**Owner skill rule:** the skill that **defines and primarily consumes** the contract owns the file. Secondary consumers link to the owner via relative path — they do not copy or re-home the schema.

**When adding a new `*.schema.json` or handoff contract:**
1. Pick the owner skill (primary consumer of the JSON contract).
2. Create under `ve-<owner>-ops/assets/<name>.schema.json`.
3. Reference from owner `SKILL.md` / `references/` and owner `example-config.yaml` if config-driven.
4. Secondary skills cite the owner path (e.g. `../ve-<owner>-ops/assets/...`) — never `assets/` at repo root.
5. If a future `scripts/*.py` emits JSON matching the schema, its docstring MUST point at the owner skill path (script ≠ schema owner).

**Anti-pattern (banned):** creating `assets/` at repo root because a script is shared — shared **code** lives in `scripts/`; shared **contracts** still belong to an owning skill.

---

## Generator-Critic-Loop (GCL) — Adversarial Quality Gate

> Full specification: [docs/gcl-spec.md](docs/gcl-spec.md) — purpose, roles, rubric, loop flow,
> termination, trace schema, prompt templates, and changelog.

### Summary

| Role | Job | Forbidden |
|------|-----|-----------|
| **Generator (G)** | Execute the cloud operation | modifying the rubric; self-scoring |
| **Critic (C)** | Independently audit G's output | calling `ve` / SDK / mutating anything |
| **Orchestrator (O)** | Loop control, termination | executing or scoring on its own |

**Hard constraint:** G and C MUST live in **isolated prompt contexts**. Shared context is banned.

### Rubric (5 dimensions)

| Dimension | Scale | Default threshold |
|-----------|-------|-------------------|
| **Correctness** | 0 / 0.5 / 1 | ≥ 0.5 (1.0 for `delete` / `stop` / IAM / KMS / DDL) |
| **Safety** | 0 / 1 | **= 1** (0 → ABORT immediately) |
| **Idempotency** | 0 / 0.5 / 1 | ≥ 0.5 |
| **Traceability** | 0 / 0.5 / 1 | ≥ 0.5 |
| **Spec Compliance** | 0 / 0.5 / 1 | ≥ 0.5 |

### Termination

| Condition | Behavior |
|-----------|----------|
| **PASS** | All dimensions meet threshold → return G's result |
| **MAX_ITER** | Reached max_iter → return best-so-far + unresolved items |
| **SAFETY_FAIL** | Safety=0 → **ABORT**; never partial return |

### Per-Skill Defaults

| Skill | GCL | max_iter | Notes |
|-------|------|----------|-------|
| `ve-ecs-ops` | **required** | 2 | instance Delete/Stop |
| `ve-redis-ops` | **required** | 2 | FlushAll / instance delete |
| `ve-rds-mysql-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-rds-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-rds-pg-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-polar-mysql-ops` | **required** | 2 | DDL / DELETE / TRUNCATE |
| `ve-mongodb-ops` | **required** | 2 | dropDatabase / delete |
| `ve-elasticsearch-ops` | **required** | 2 | delete index / cluster |
| `ve-tos-ops` | **required** | 2 | bucket delete + object lifecycle |
| `ve-iam-ops` | **required** | 2 | detach policy / delete role / rotate keys |
| `ve-kms-ops` | **required** | 2 | schedule key deletion (irreversible) |
| `ve-eip-ops` | **required** | 2 | release EIP |
| `ve-security-group-ops` | **required** | 2 | revoke-rule can lock out production |
| `ve-vpc-ops` | recommended | 3 | VPC / subnet delete |
| `ve-nat-ops` | recommended | 3 | SNAT / DNAT rule delete |
| `ve-vpn-ops` | recommended | 3 | tunnel / customer-gateway delete |
| `ve-clb-ops` | recommended | 3 | listener / backend delete |
| `ve-alb-ops` | recommended | 3 | listener / server group delete |
| `ve-vke-ops` | recommended | 3 | node / cluster delete |
| `ve-nas-ops` | recommended | 3 | filesystem / mount delete |
| `ve-cms-ops` | recommended | 3 | alarm rule delete |
| `ve-fg-ops` | recommended | 3 | function delete |
| `ve-ark-ops` | recommended | 3 | instance / template delete |
| `ve-cdn-ops` | optional | 5 | domain config / refresh |
| `ve-dns-ops` | optional | 5 | record delete |
| `ve-kafka-ops` | optional | 5 | topic delete |
| `ve-sls-ops` | optional | 5 | read-mostly |
| `ve-billing-ops` | optional | 5 | read-only |
| `ve-skill-generator` | optional | 3 | meta operation |

Each skill may override `max_iter` in its own `SKILL.md` (`## Quality Gate (GCL)`).

### Anti-Patterns (banned)

| Anti-pattern | Reason |
|---|---|
| Shared context G+C | Defeats independence |
| Subjective scoring | Critic must use rubric, not "vibes" |
| Unbounded loop | Always hard-cap iterations |
| Critic sees user request | Encourages rubber-stamping |
| Silently downgrade on Safety fail | Must ABORT visibly |
| Trace not persisted | No post-mortem |
| Critic mutates resources | Read-only by definition |
| Real `VOLCENGINE_SECRET_KEY` in trace | Credential leakage → use `<masked>` |
| GCL bypass for "obviously safe" ops | Even reads go through GCL |

### Cross-Skill Delegation

When GCL identifies cross-product gaps, the Orchestrator MUST delegate, not absorb:

| Critic finding | Delegate to |
|---|---|
| IAM policy gap | `ve-iam-ops` |
| KMS key / secret needed | `ve-kms-ops` |
| EIP / VPC concern in non-network skill | `ve-eip-ops` / `ve-vpc-ops` |
| Alarm rule change after destructive op | `ve-cms-ops` |
| Billing quota exceeded | `ve-billing-ops` |

The Critic itself MUST NOT call any skill — it only emits suggestions.

### GCL Rollout Complete — All 29 Skills Equipped

All 13 `required`-tier, 10 `recommended`-tier, and 6 `optional`-tier skills have GCL rubric + prompt templates.
See [docs/gcl-spec.md](docs/gcl-spec.md) §11 for the full rollout changelog.

## Evaluation

`assets/eval_queries.json` per skill holds intent-classification test cases (`should_trigger: true/false`). These are consumed by external evaluation harnesses, not by an in-repo test runner. When adding capability, add eval cases in the same change.

### Build-time regression

Automated validation scripts are available in `scripts/` for pre-commit checks (see **Files that DO NOT exist** below). After changing any skill, run `python3 scripts/validate_local.py` if available, then manually verify against the P0/P1 checklist in `ve-skill-generator/SKILL.md` and the 2-round self-review above.

### Runtime GCL

`scripts/gcl_runner.py` implements the GCL Orchestrator (Phase 2). Use it for automated GCL loops:
- External Critic via `--critic-json` or stdin (production mode)
- `--structural-critic-only` for CI/local structural smoke tests only (NOT for production mutations)

GCL runs MUST use externally supplied isolated Critic scores in production. Manual execution following `docs/gcl-spec.md` is also acceptable.

---

## Document Integrity & Link Validation (C12 / F9 / F10 — 强制)

> 每次文档变更后，自动对**所有受影响的文件**执行以下 3 层完整性检查。

### Layer 1 — 链接完整性（验证 F9）

对本次变更涉及的所有 Markdown 文件，扫描全部 `[...](...)` 引用链接并按目标类型处理：

| 链接类型 | 验证方法 | 失败时不通过则 |
|----------|---------|--------------|
| 相对路径 `.md` | `test -f <target>` | 修复路径或创建缺失文件 |
| 相对路径 `.yaml` / `.json` | `test -f <target>` | 修复路径或创建缺失文件 |
| `#section-anchor` | 在目标文件中 `grep -i 'section-name'` | 检查 anchor 名称拼写 |
| `http(s)://` | 确认 URL 可访问或指向官方文档 | 标记为 BLOCKED 待验证 |
| `{{env.*}}` / `{{user.*}}` / `{{output.*}}` | 仅检查语法，不检查值 | — |

**检查范围**: 本次变更的文件 + 所有引用了这些文件的其他文件。

```bash
# macOS 兼容的链接完整性扫描模板
echo "=== Link Integrity Check ==="
grep -oE '\[[^]]+\]\([^)]+\)' "$CHANGED_FILE" 2>/dev/null | \
  sed 's/\[[^]]*\]//' | tr -d '()' | while IFS= read -r target; do
  case "$target" in
    http*) echo "⏭️ EXTERNAL: $target" ;;
    #*) echo "⏭️ ANCHOR: $target" ;;
    *)  dir="$(dirname "$CHANGED_FILE")"
        if [ -f "$dir/$target" ] || [ -f "$target" ]; then
          echo "✅ $target"
        else
          echo "❌ MISSING: $target (referenced from $CHANGED_FILE)"
        fi ;;
  esac
done
```

### Layer 2 — 交叉引用一致性（验证 F9 / TE-7）

当 `AGENTS.md` 和 `docs/*.md` 同时变更时，检查：

| 检查项 | 方法 |
|--------|------|
| §引用编号一致 | `docs/gcl-spec.md §11` → 确认 `## 11.` 标题存在 |
| 文件引用路径一致 | AGENTS.md 中 `docs/token-efficiency.md` → 确认文件存在 |
| 双向引用对称 | A → B 时，B 中也应有 ← A 的反向引用或文档层级说明 |

```bash
# 交叉引用对称性检查（仅从 Markdown 链接 `[...](...)` 中提取）
echo "=== Cross-Reference Symmetry ==="
grep -oE '\[[^]]+\]\(docs/[^)]+\)' AGENTS.md 2>/dev/null | \
  sed 's/.*(docs/ds/' | tr -d ')' | while IFS= read -r href; do
  if grep -q 'AGENTS.md' "$href" 2>/dev/null; then
    echo "✅ $href has backlink to AGENTS.md"
  else
    echo "⚠️  $href missing backlink to AGENTS.md"
  fi
done
```

### Layer 3 — 内容去重（验证 C12 / TE-6）

在变更的文件中检测与同 skill 其他文件的重复段落：

1. 提取变更文件中的所有 **Markdown 表格**（至少 3 行）
2. 检查同 skill 目录下其他文件是否包含相同表格
3. 提取变更文件中的所有 **代码块**（至少 5 行）
4. 检查是否有完全相同的代码块出现在另一文件中

```bash
# 内容去重扫描模板
echo "=== Deduplication Check ==="
CHANGED="$(basename "$CHANGED_FILE")"
SKILL_DIR="$(dirname "$CHANGED_FILE")"
for peer_file in "$SKILL_DIR"/*.md "$SKILL_DIR"/references/*.md 2>/dev/null; do
  [ "$peer_file" = "$CHANGED_FILE" ] && continue
  [ ! -f "$peer_file" ] && continue
  # 检查是否有完全相同的代码块（≥5行）
  awk '/^```/{p=!p;if(p){b=$0;c=1}else{b=b"\n"$0;if(c>=5)print b;b=""}}p&&c++' "$CHANGED_FILE" | \
  while IFS= read -r block; do
    if grep -Fq "$block" "$peer_file" 2>/dev/null; then
      echo "❌ DUPLICATE block in $peer_file"
    fi
  done
done
```

### 发现任何问题 → 修复并重新验证 → 确认全部通过后方可继续。

---

## Files that DO NOT exist

- **`scripts/` directory** at repo root contains validation scripts (`validate_local.py`, `check_gcl_conformance.py`, `check_markdown_links.py`, etc.). These are used for pre-commit validation. There is no `.github/workflows/` directory — CI has not been set up yet.
- **No `package.json`, `Makefile`, `CI configs`, `build scripts`, `typechecker`, or non-stdlib test runner** — except:
- **No `CLAUDE.md` at repo root** — this file (`AGENTS.md`) is the agent guidance entry point, imported via `@CLAUDE.md`.
- **No `opencode.json`, `.cursorrules`, or similar IDE configs.**
- **No `audit-results/` directory** — GCL traces are stored in tool-local state only.
- `.omc/`, `.omo/` are gitignored cache data — not source.

---

## Runtime Quality Gates: GCL & Reflexion

Detailed runtime-quality specifications are intentionally externalized to reduce always-loaded context size:

| Spec | Read before modifying |
|---|---|
| `docs/gcl-spec.md` | any `## Quality Gate (GCL)` section, `references/rubric.md`, `references/prompt-templates.md`, or GCL-related runner code |
| `docs/reflexion-memory.md` | `docs/failure-patterns.md`, trace `failure_pattern` extraction, Reflexion retrieval/persistence logic, or failure-memory governance |
| `docs/failure-patterns.md` | only when retrieving or updating reusable failure patterns; keep it bounded and deduplicated |

### GCL hard constraints

- Production GCL requires isolated Generator and Critic contexts; shared-context G+C is banned.
- Critic is read-only: it MUST NOT call `ve`, use SDK clients, mutate resources, or self-score Generator output.
- Critic MUST NOT see the raw user request; it may use sanitized `{{output.operation_intent}}`, Generator output, trace, and rubric.
- Orchestrator owns `operation_intent` generation before Critic scoring; it MUST omit raw user wording, credentials, and unmasked sensitive identifiers.
- `Safety = 0` / `SAFETY_FAIL` MUST abort immediately; never return partial or best-effort output.
- Every GCL loop MUST be bounded by `max_iterations`; unbounded retry loops are banned.
- Every GCL run MUST persist a masked trace (tool-local; `.omc/` is gitignored — see **Files that DO NOT exist** below).
- Production GCL MUST use externally supplied isolated Critic scores; `--structural-critic-only` is allowed only for CI/local structural smoke tests and MUST NOT be used for production execution, human acceptance, or quality pass decisions.
- GCL prompt templates MUST use `{{env.*}}` / `{{user.*}}` / `{{output.*}}`; bare `{...}` placeholders are banned.
- GCL `required` / `recommended` skills MUST keep `## Quality Gate (GCL)` in `SKILL.md`, plus `references/rubric.md` and `references/prompt-templates.md`.

### Reflexion hard constraints

- Reflexion retrieval is an optional hint, not a mandatory gate.
- `docs/failure-patterns.md` MUST stay ≤ 200 lines; prune low-frequency entries when needed.
- Deduplicate patterns by `skill` + `command` + `error`; increment `count` on matches.
- Patterns MUST come from GCL trace `failure_pattern` fields or self-review findings, not ad-hoc subjective notes.
- Promote high-frequency patterns to anti-pattern docs and remove duplicates from memory.

### Relationship to build-time self-review

Build-time 2-round self-review and runtime GCL are independent gates. A clean self-review does not exempt runtime scoring; a passing GCL rubric does not exempt sloppy skill updates.

---

## Key References

| Document | Description |
|----------|-------------|
| `docs/gcl-spec.md` | **Runtime GCL spec** — purpose, roles, loop flow, trace, prompt templates, per-skill defaults, changelog |
| `docs/token-efficiency.md` | Detailed TE rules with code examples (TE-1 through TE-9) |
| `docs/reflexion-memory.md` | **Reflexion rules** — lightweight cross-session failure-pattern memory governance |
| `docs/failure-patterns.md` | **Reflexion memory store** — bounded structured failure patterns for cross-session learning |
| `ve-skill-generator/SKILL.md` | Meta-skill generator — full workflow, P0/P1 checklist, Token Efficiency rules |
| `ve-skill-generator/references/ve-skill-template.md` | Canonical SKILL.md template with GCL block |
| `ve-skill-generator/references/governance-and-adversarial-review.md` | Governance & adversarial review — R1-R4 pre-merge security/resilience/UX scenarios |
| `ve-skill-generator/references/cli-behavior.md` | Verified `ve` CLI conventions |
| `ve-skill-generator/references/execution-environment.md` | CLI + Go SDK setup details |
| `ve-skill-generator/references/user-experience-spec.md` | UX requirements every generated skill must follow |
| Each skill's `references/rubric.md` | The rubric instance |
| Each skill's `references/prompt-templates.md` | G/C/O prompt skeletons |