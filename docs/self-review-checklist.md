# Two-Round Self-Review Checklist

> 配套 `AGENTS.md` §MANDATORY: Two-Round Self-Review。AGENTS.md 只保留强制声明，完整清单在此维护。
> 任何 `ve-*-ops/SKILL.md`、`references/*.md`、`assets/*` 或 `ve-skill-generator/**` 的编辑，都必须在"完成"前跑完两轮自审并修复全部问题。

---

## Round 1 — Structural & Specification Compliance

Verify against `ve-skill-generator/SKILL.md` and `ve-skill-generator/references/ve-skill-template.md`:

| # | Check | Detail |
|---|-------|--------|
| C1 | Frontmatter | name, description, license, compatibility, metadata complete |
| C2 | cli_applicability | set; if `dual-path`, both CLI and SDK steps exist |
| C3 | Five Core Standards | boundaries, I/O, steps, failures, single responsibility present & accurate |
| C4 | Placeholder channels | `{{env.*}}` for secrets/region, `{{user.*}}` for interactive, `{{output.*}}` for captured |
| C5 | Credential safety | No credential value logged/printed/echoed — only `<masked>` or existence checks |
| C6 | **Token Efficiency** | See TE rules; verify TE-1~TE-7 |
| C7 | Safety gates | Every destructive operation (delete, stop, release, restore) has explicit confirmation |
| C8 | Error taxonomy | ≥ 10 product-specific codes with HALT vs retry classification |
| C9 | Reference files | All 6 standard files exist when applicable |
| C10 | Cross-product delegation | ECS skill delegates IAM work to IAM skill, etc. |
| C11 | Version bump | `last_updated` and `version` bumped if behavior changed |
| C12 | **Content dedup** | No duplicate operation flows across SKILL.md ↔ references/; TE-6 verified |
| C13 | **AIOps coverage** | `references/advanced/aiops.md` and `references/advanced/finops.md` exist for all required+recommended skills (23 total) |
| C15 | **`### What This Skill Does`** | MUST exist with a clear 2-3 sentence description of purpose and boundary |
| C16 | **`## Operational Best Practices`** | MUST exist with actionable operational guidance |
| C17 | **SecurityOps coverage** | `references/advanced/securityops.md` exists for security-critical skills (8); recommended for all others |

---

## Round 2 — Accuracy & Anti-Pattern Sweep

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
| F9 | **Link integrity** | Cross-references between docs/ and AGENTS.md verified; all internal `[...](...)` resolve |
| F10 | **Dedup integrity** | No repeated content across multiple files; run dedup check after every change |
| F11 | **Eval data coverage** | `assets/eval_queries.json` exists for all 29 skills with ≥ 5 trigger + ≥ 2 non-trigger cases |
| F12 | **FinOps coverage** | `references/advanced/finops.md` exists for required+recommended skills |
| F13 | **SecurityOps coverage** | `references/advanced/securityops.md` exists for security-critical skills (8); recommended for all others |

**Fix every issue found in both rounds before responding "done".** If a finding requires information you don't have, mark it `[blocked: needs OpenAPI verification]`.

---

## Validation Command Matrix

| Change scope | Required command |
|---|---|
| Full local validation before PR / handoff | Walk P0/P1 checklist in `ve-skill-generator/SKILL.md` (no automated runner in this repo) |
| Any `SKILL.md` frontmatter or metadata change | Re-check P0/P1 checklist items C1–C12 manually |
| Any GCL rubric, prompt template, or `## Quality Gate (GCL)` section change | Verify against `docs/gcl-spec.md` §Rubric + §Prompt Templates |
| Any Markdown spec, README, or path-reference change | Run the link-integrity scan in `docs/document-integrity.md` |
| Any `docs/*.md` change | Verify cross-reference symmetry |
| Any skill harness review | Walk `docs/skill-harness-review-checklist.md` D0-D4 |
| Any Go SDK example code change | `python3 -m py_compile <file>` if Python; Go examples are illustrative — no build step |

> **Note:** This repo has a Go validation CLI, `vet` (built from `cmd/vet/`), for pre-commit checks. After changes, run `vet validate --root .` (or `vet check frontmatter|links|gcl|eval|aiops|assessment`), then supplement with manual checks against the P0/P1 checklist.
