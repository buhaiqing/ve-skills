@CLAUDE.md

# Agent-Specific Rules

These supplement `CLAUDE.md` (imported above). They encode lessons that have already cost agent sessions in this repo.

## Repository Type — Read This First

This repo contains **only Markdown skill specifications**. There is:

- **No build system, no tests, no lint, no package manager.** Do not run `go build`, `go test`, `npm`, `pip`, `make`, etc. — they will fail or do nothing.
- **No runtime code.** The `ve` CLI and Go SDK referenced in skills are *consumed by* generated skills at execution time, not built here.
- **No CI workflows yet.** Verification is human review against the P0/P1 checklist in `ve-skill-generator/SKILL.md`.

Verification for skill edits = re-reading the file plus walking the P0/P1 checklist. Do not invent build/test commands.

## MANDATORY: Two-Round Self-Review After Every Skill Update

After ANY edit to a `ve-*-ops/SKILL.md`, its `references/*.md`, `assets/*`, or to `ve-skill-generator/**`, you MUST run **two rounds of self-review** before declaring done. Fix every issue surfaced — do not just list them.

### Round 1 — Structural & Specification Compliance

Verify the file against `ve-skill-generator/SKILL.md` and `ve-skill-generator/references/ve-skill-template.md`:

- [ ] Frontmatter matches agentskills.io OpenSpec (name, description, license, compatibility, metadata)
- [ ] `cli_applicability` is set; if `dual-path`, BOTH `ve` CLI step and SDK step exist for every operation
- [ ] Five Core Standards section present and accurate (boundaries, I/O, steps, failures, single responsibility)
- [ ] Placeholders use the right channel: `{{env.*}}` for secrets/region, `{{user.*}}` for interactive, `{{output.*}}` for captured
- [ ] No credential value is logged, printed, or echoed anywhere in examples — only `<masked>` or existence checks
- [ ] Every destructive operation (delete, stop, release, restore) has an explicit safety gate / confirmation
- [ ] Error taxonomy has ≥ 10 product-specific codes with HALT vs retry classification
- [ ] All 6 standard reference files exist when applicable: `core-concepts`, `api-sdk-usage`, `cli-usage`, `troubleshooting`, `monitoring`, `integration`
- [ ] Cross-product operations declare delegation targets (e.g. ECS skill delegates IAM work to IAM skill)
- [ ] `last_updated` and `version` bumped if behavior changed

### Round 2 — Accuracy & Anti-Pattern Sweep

Verify against reality, not memory:

- [ ] Every API field name and CLI flag is grounded in official OpenAPI or verified via `ve <service> --help` — no guessed names
- [ ] Command prefix is `ve` (not `volcengine-cli`, which was renamed at v1.0.20)
- [ ] CLI command shape matches verified pattern: `ve <service> <Action> --<Param> value` (PascalCase actions and params)
- [ ] JSON bodies use `--body '{"Key":"Value"}'` form, not invented flags
- [ ] No generic prose, no speculative claims, no copy-pasted boilerplate from other products that doesn't apply
- [ ] Links to sibling references use correct relative paths
- [ ] Examples are minimal and runnable, not pseudocode
- [ ] Markdown renders cleanly (tables aligned, code fences closed, headings hierarchical)

**Fix every issue found in both rounds before responding "done".** If a finding requires information you don't have (e.g. an unverified API field), mark it `[blocked: needs OpenAPI verification]` rather than guessing.

## Skill Authoring Guardrails

- **Never create a new `ve-*-ops/` directory by hand.** Use the `ve-skill-generator` meta-skill flow so layout, frontmatter, and checklists stay consistent.
- **Never invent CLI flags or API parameters.** If a field isn't in the official OpenAPI doc or `ve <service> <action> --help`, do not write it. Mark it as needing verification instead.
- **Single-product rule.** One `ve-*-ops` skill = one product = one primary resource model. Cross-product work is delegated, not absorbed.
- **Prefer editing existing skills over creating new files.** Do not add tutorial-style `.md` files at the repo root.

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
```

`.omc/` and `.omo/` are tool-local state and are gitignored — do not commit anything inside them.
