# AIOps — Skill Generator Meta-Operation

> AIOps deep content extracted per TE-7. This is a meta-skill; AIOps here means skill generation quality.

## Cross-Skill Diagnosis Decision Tree

```
[Skill Generation Quality Issue]
    │
    ├── Is it structural?
    │   ├── Missing frontmatter → Add required fields
    │   │   └── See ve-skill-template.md §1
    │   ├── Missing references/ → Create standard 6 files
    │   │   └── See ve-skill-generator/references/ve-skill-template.md
    │   └── Broken links → Run link integrity scan
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
    │   ├── TE-1 violation → Replace hardcoded tables with ve queries
    │   │   └── See docs/token-efficiency.md §TE-1
    │   ├── TE-6 violation → Dedup SKILL.md vs references/
    │   │   └── SKILL.md is authoritative
    │   └── Over-length content → Compress per TE-9
    │       └── Move detailed content to references/
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
