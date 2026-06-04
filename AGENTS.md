@CLAUDE.md

# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

> **Content hierarchy**: This file is the entry point. Detailed specifications live in `docs/`:
> - `docs/gcl-spec.md` — full GCL specification (purpose, roles, loop flow, trace, prompt templates, changelog)
> - `docs/token-efficiency.md` — full TE rules with code examples (TE-1 through TE-7)
>
> Keep this file in sync with `docs/` files. When updating either, verify the other stays aligned.

---

## Repository Type — Read This First

This repo contains **only Markdown skill specifications**. There is:

- **No build system, no tests, no lint, no package manager.** Do not run `go build`, `go test`, `npm`, `pip`, `make`, etc. — they will fail or do nothing.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md`.

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

**不可压缩的内容**：Agent 可执行命令本身（参数、JSON paths）、错误恢复逻辑、安全门、Credential 规则、跨技能编排链。

**发现任一违规 → 立即修复 → 重新检查直到全部通过。**

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

## See also

- `docs/gcl-spec.md` — full GCL specification (purpose, roles, loop flow, trace, prompt templates, changelog)
- `docs/token-efficiency.md` — detailed TE rules with code examples (TE-1 through TE-7)
- `ve-skill-generator/SKILL.md` — the meta-skill that scaffolds new skills
- `ve-skill-generator/references/ve-skill-template.md` — canonical SKILL.md template with GCL block
- Each skill's `references/rubric.md` — the rubric instance
- Each skill's `references/prompt-templates.md` — G/C/O prompt skeletons