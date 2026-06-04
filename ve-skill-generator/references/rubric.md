---
name: ve-skill-generator-rubric
description: GCL rubric for ve-skill-generator (meta-skill). Optional tier, max_iter=3.
license: MIT
metadata:
  author: volcengine
  version: "1.0.0"
  last_updated: "2026-06-04"
  parent_skill: ve-skill-generator
  gcl_role: critic_input
  rubric_dimensions: 5
  default_max_iter: 3
---
# GCL Rubric — ve-skill-generator (Meta-Skill)
## 0. Operation Tiers
| Tier | Operations | max_iter | Safety |
|---|---|---|---|
| **Destructive** | — (generation is additive) | 3 | 1.0 |
| **State-changing** | GenerateSkill, UpdateSkill | 3 | 1.0 |
| **Mutating** | — (all generation is state-changing) | 3 | ≥0.5 |
| **Read-only** | AnalyzeOpenAPI, ReviewSkill | 3 | ≥0 |
Safety: Generated skill MUST NOT contain real credentials in examples — only `<masked>` and {{env.*}} placeholders. Generated rubric MUST include 5 dimensions. Generated prompt-templates MUST include Critic that hides user request. Generated skill MUST reference references/rubric.md and references/prompt-templates.md. VOLCENGINE_SECRET_KEY never in generated output.
## Changelog
| 1.0.0 | 2026-06-04 | Initial GCL rubric for ve-skill-generator |