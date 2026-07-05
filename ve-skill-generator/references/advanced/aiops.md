# AIOps — Skill Generator Meta-Operation

> AIOps deep content extracted per TE-7. This is a meta-skill; AIOps here means skill generation quality.

## Cross-Skill Diagnosis Decision Tree

```
[Skill Generation Quality Issue]
    │
    ├── Is it structural?
    │   ├── Missing frontmatter (missing name/description/compatibility in > 3 skills) → Add required fields
    │   │   └── See ve-skill-template.md §1
    │   ├── Missing references/ (< 6 standard files per skill) → Create standard 6 files
    │   │   ├── aiops.md missing for ≥ 3 skills → Run check_aiops_coverage.py
    │   │   └── See ve-skill-generator/references/ve-skill-template.md
    │   └── Broken links (link integrity check fails > 5 errors) → Run link integrity scan
    │       ├── Cross-skill references pointing to nonexistent files
    │       └── Fix relative paths
    │
    ├── Is it GCL-related?
    │   ├── Missing rubric.md → Generate from template
    │   │   └── See ve-skill-generator/references/rubric.md
    │   ├── Missing prompt-templates.md → Generate from template
    │   │   └── See ve-skill-generator/references/prompt-templates.md
    │   └── GCL not in SKILL.md → Add ## Quality Gate (GCL) section
    │       └── See docs/gcl-spec.md §3
    │
    ├── Is it Token Efficiency-related?
    │   ├── TE-1 violation (hardcoded versions in > 5 references files) → Replace hardcoded tables with ve queries
    │   │   └── See docs/token-efficiency.md §TE-1
    │   ├── TE-6 violation (duplicate blocks across SKILL.md and references/ in ≥ 3 skills) → Dedup
    │   │   └── SKILL.md is authoritative — remove from references/
    │   ├── Over-length content (single SKILL.md > 2000 lines) → Compress per TE-9
    │   │   └── Move detailed content to references/te-7/ or references/advanced/
    │   └── eval_queries.json missing for > 2 skills → Generate with ≥ 5 trigger + ≥ 2 non-trigger cases
    │
    ├── Is it evaluation quality related?
    │   ├── Skill coverage < 80% (21 of 29 skills not fully compliant) → Prioritize required tier first
    │   │   └── Run validate_local.py on affected skills
    │   ├── > 3 P1 violations across skills → Batch-fix systematic issues
    │   │   └── Track in progress.md per skill batch
    │   └── GCL conformance score < 70% → Review rubric and prompt templates
    │       └── Run check_gcl_conformance.py
    │
    └── Unknown → Run P0/P1 checklist
```

## Alarm Storm Handling

**Detection Criteria:**
- > 3 skills failing P0 checklist simultaneously
- > 5 structural violations across skills
- Pattern of same violation type across multiple skills

**Suppression Workflow:**
1. Run check_gcl_conformance.py to identify all violations
2. Run check_aiops_coverage.py for coverage gaps
3. Prioritize by skill tier (required > recommended > optional)
4. Batch-fix common violations across skills
5. Re-validate all skills with validate_local.py

## Proactive Inspection Checklist

```markdown
## Skill Generator Proactive Inspection — [Date]

### Coverage
- [ ] All 29 skills have advanced/aiops.md
- [ ] All 29 skills have eval_queries.json
- [ ] All required+recommended skills have advanced/finops.md

### Quality Gates
- [ ] All skills pass check_gcl_conformance.py
- [ ] All skills pass check_aiops_coverage.py
- [ ] All skills pass validate_local.py
- [ ] No broken links across skill ecosystem

### GCL Coverage
- [ ] All 29 skills have rubric.md
- [ ] All 29 skills have prompt-templates.md
- [ ] All 29 skills have ## Quality Gate (GCL) section

### Token Efficiency
- [ ] No hardcoded API versions in references/
- [ ] No duplicate content across SKILL.md and references/
- [ ] Error tables compact (<=3 cols)
```
