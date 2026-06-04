---
name: ve-skill-generator-prompt-templates
description: GCL prompt templates for ve-skill-generator meta-skill.
license: MIT
metadata: {author: volcengine, version: "1.0.0", last_updated: "2026-06-04", parent_skill: ve-skill-generator, default_max_iter: 3}
---
# GCL Prompt Templates — ve-skill-generator
## 1.G / 2.C / 3.O (standard)
Generator creates/updates skill files. Critic verifies generated skill against GCL checklist. Orchestrator: loop max 3 iterations.
## 4. Safety (meta-skill specific)
### Verification checklist (Critic scores each):
- Generated SKILL.md has ## Quality Gate (GCL) section
- references/rubric.md has 5 dimensions with product-specific rules
- references/prompt-templates.md has G/C/O prompts + safety prompts for destructive/state-changing ops
- No real credentials in examples — only {{env.*}} and <masked>
- Dual-path documented when cli_applicability=dual-path
- Error taxonomy ≥ 10 codes
## 5. Changelog
| 1.0.0 | 2026-06-04 | Initial |