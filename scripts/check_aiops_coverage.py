#!/usr/bin/env python3
# DEPRECATED: superseded by cmd/vet — run `vet check aiops`
"""Check AIOps coverage across all 29 skills.

Validates:
1. references/advanced/aiops.md exists for all required+recommended skills (23)
2. references/advanced/finops.md exists for all required+recommended skills (23)
3. assets/eval_queries.json exists for all 29 skills

Usage:
    python3 scripts/check_aiops_coverage.py
    python3 scripts/check_aiops_coverage.py --json
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


# Required+recommended skills (23) need aiops.md + finops.md
# Optional skills (6) only need eval_queries.json: cdn, dns, kafka, sls, billing, skill-generator
REQUIRED_RECOMMENDED = frozenset({
    # required (13)
    "ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops",
    "ve-rds-pg-ops", "ve-polar-mysql-ops", "ve-mongodb-ops",
    "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops", "ve-kms-ops",
    "ve-eip-ops", "ve-security-group-ops",
    # recommended (10)
    "ve-vpc-ops", "ve-nat-ops", "ve-vpn-ops", "ve-clb-ops", "ve-alb-ops",
    "ve-vke-ops", "ve-nas-ops", "ve-cms-ops", "ve-fg-ops", "ve-ark-ops",
})

# finops.md required for all required+recommended (23) skills

ALL_SKILLS = frozenset({
    "ve-ecs-ops", "ve-redis-ops", "ve-rds-mysql-ops", "ve-rds-ops",
    "ve-rds-pg-ops", "ve-polar-mysql-ops", "ve-mongodb-ops",
    "ve-elasticsearch-ops", "ve-tos-ops", "ve-iam-ops", "ve-kms-ops",
    "ve-eip-ops", "ve-security-group-ops", "ve-vpc-ops", "ve-nat-ops",
    "ve-vpn-ops", "ve-clb-ops", "ve-alb-ops", "ve-vke-ops", "ve-nas-ops",
    "ve-cms-ops", "ve-fg-ops", "ve-ark-ops", "ve-cdn-ops", "ve-dns-ops",
    "ve-kafka-ops", "ve-sls-ops", "ve-billing-ops", "ve-skill-generator",
})


def check_skill(root: Path, skill: str) -> dict:
    base = root / skill / "references" / "advanced"
    eval_path = root / skill / "assets" / "eval_queries.json"

    # Validate eval_queries.json exists AND is parseable as JSON
    eval_ok = False
    eval_trigger = 0
    eval_non_trigger = 0
    if eval_path.is_file():
        try:
            d = json.load(open(eval_path))
            arr = d if isinstance(d, list) else d.get("queries", [])
            eval_trigger = sum(1 for q in arr if q.get("should_trigger"))
            eval_non_trigger = sum(1 for q in arr if not q.get("should_trigger"))
            eval_ok = True
        except (json.JSONDecodeError, ValueError):
            eval_ok = False

    return {
        "skill": skill,
        "aiops_md": (base / "aiops.md").is_file(),
        "finops_md": (base / "finops.md").is_file(),
        "eval_queries": eval_ok,
        "eval_trigger": eval_trigger,
        "eval_non_trigger": eval_non_trigger,
        "has_advanced": base.is_dir(),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args()

    results = []
    for skill in sorted(ALL_SKILLS):
        r = check_skill(args.root, skill)
        results.append(r)

    # Coverage stats
    advanced_skills = [r for r in results if r["has_advanced"]]
    aiops_md_skills = [r for r in results if r["aiops_md"]]
    finops_md_skills = [r for r in results if r["finops_md"]]
    eval_skills = [r for r in results if r["eval_queries"]]

    rr_skills = [r for r in results if r["skill"] in REQUIRED_RECOMMENDED]
    rr_aiops = [r for r in rr_skills if r["aiops_md"]]
    rr_finops = [r for r in rr_skills if r["finops_md"]]

    # eval content quality: trigger >= 5, non_trigger >= 2
    eval_quality_ok = [r for r in results if r["eval_queries"]
                       and r["eval_trigger"] >= 5 and r["eval_non_trigger"] >= 2]
    eval_quality_bad = [r for r in results if r["eval_queries"]
                        and (r["eval_trigger"] < 5 or r["eval_non_trigger"] < 2)]
    eval_parse_fail = [r["skill"] for r in results
                       if not r["eval_queries"] and (root / r["skill"] / "assets" / "eval_queries.json").is_file()]

    report = {
        "total_skills": len(ALL_SKILLS),
        "required_recommended_count": len(REQUIRED_RECOMMENDED),
        "advanced_dir_count": len(advanced_skills),
        "aiops_md_count": len(aiops_md_skills),
        "finops_md_count": len(finops_md_skills),
        "eval_parseable_count": len(eval_skills),
        "eval_quality_ok_count": len(eval_quality_ok),
        "rr_aiops_coverage": f"{len(rr_aiops)}/{len(REQUIRED_RECOMMENDED)}",
        "rr_finops_coverage": f"{len(rr_finops)}/{len(REQUIRED_RECOMMENDED)}",
        "eval_coverage": f"{len(eval_skills)}/{len(ALL_SKILLS)}",
        "skills_missing_aiops": [r["skill"] for r in results if not r["aiops_md"]],
        "skills_missing_finops": [r["skill"] for r in results if not r["finops_md"]],
        "skills_missing_eval": [r["skill"] for r in results if not r["eval_queries"]],
        "eval_quality_bad": eval_quality_bad,
        "eval_parse_fail": eval_parse_fail,
        "details": results,
    }

    finops_missing_rr = [s for s in REQUIRED_RECOMMENDED if s not in [r["skill"] for r in finops_md_skills]]

    if args.json:
        print(json.dumps(report, indent=2))
        return 0

    print(f"AIOps Coverage Report")
    print(f"{'='*50}")
    print(f"Total skills: {len(ALL_SKILLS)}")
    print(f"Required+Recommended: {len(REQUIRED_RECOMMENDED)}")
    print(f"")
    print(f"advanced/aiops.md:  {len(aiops_md_skills)}/{len(ALL_SKILLS)} ({len(rr_aiops)}/{len(REQUIRED_RECOMMENDED)} for R+R tier)")
    print(f"advanced/finops.md: {len(finops_md_skills)}/{len(ALL_SKILLS)} ({len(rr_finops)}/{len(REQUIRED_RECOMMENDED)} for R+R tier)")
    print(f"eval_queries.json:  {len(eval_skills)}/{len(ALL_SKILLS)} (trigger≥5, non_trigger≥2: {len(eval_quality_ok)}/{len(eval_skills)})")

    if report["skills_missing_aiops"]:
        print(f"\nMissing advanced/aiops.md:")
        for s in report["skills_missing_aiops"]:
            print(f"  - {s}")
    if finops_missing_rr:
        print(f"\nMissing advanced/finops.md (R+R tier only):")
        for s in sorted(finops_missing_rr):
            print(f"  - {s}")
    elif report["skills_missing_finops"]:
        print(f"\nMissing advanced/finops.md (optional tier - ok):")
        for s in report["skills_missing_finops"]:
            print(f"  - {s}")
    if report["skills_missing_eval"]:
        print(f"\nMissing eval_queries.json:")
        for s in report["skills_missing_eval"]:
            print(f"  - {s}")
    if report["eval_parse_fail"]:
        print(f"\nUnparseable eval_queries.json (JSON syntax error):")
        for s in report["eval_parse_fail"]:
            print(f"  - {s}")
    if report["eval_quality_bad"]:
        print(f"\nLow-quality eval_queries.json (need trigger≥5, non_trigger≥2):")
        for r in report["eval_quality_bad"]:
            print(f"  - {r['skill']}: trigger={r['eval_trigger']}, non_trigger={r['eval_non_trigger']}")

    # Pass if all required+recommended have aiops+finops and all have eval (parseable)
    ok = (len(rr_aiops) == len(REQUIRED_RECOMMENDED) and
          len(rr_finops) == len(REQUIRED_RECOMMENDED) and
          len(eval_skills) == len(ALL_SKILLS) and
          len(eval_parse_fail) == 0 and
          len(eval_quality_bad) == 0)
    print(f"\n{'✅ PASS' if ok else '❌ FAIL'}: AIOps coverage check")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
