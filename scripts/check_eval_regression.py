#!/usr/bin/env python3
"""Eval regression checker — intent-classification regression for all 28 skills.

Reads each skill's ``assets/eval_queries.json``, parses its
``## Trigger & Scope`` section, and verifies that queries labeled
``should_trigger: true`` match the skill's SHOULD scope while queries
labeled ``should_trigger: false`` do not (or align with SHOULD NOT).

Two passes:

1. **Absolute match** (always) — catches structural gaps: a query labeled
   ``should_trigger: true`` must have ≥15% token overlap with the SHOULD
   section, and ``should_trigger: false`` queries must have ≤75%.

2. **Git-diff delta** (``--git-diff``) — catches **semantic drift**: when
   ``## Trigger & Scope`` changes, compares each query's coverage score
   against the *old* scope. A regression of ≥50% (e.g. old=0.80 → new=0.30)
   is flagged as a BLOCKER even if the absolute match still passes.
   This detects the case "关键词还在但语义偏移了".

Usage::

    python3 scripts/check_eval_regression.py                    # absolute match only
    python3 scripts/check_eval_regression.py --git-diff         # + git diff against HEAD~1
    python3 scripts/check_eval_regression.py --git-diff origin/main  # against merge-base
    python3 scripts/check_eval_regression.py --json             # machine-readable JSON

Exit codes:
    0  — all skills pass
    1  — at least one regression detected
    2  — fatal error (bad args, missing eval file)
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path
from typing import Any

REGEX_FLAGS = re.MULTILINE | re.IGNORECASE

# Minimum required eval query counts
MIN_TOTAL = 8
MIN_TRIGGER = 4
MIN_NON_TRIGGER = 3

# Skills exempt from eval regression (orchestration / meta skills)
EXEMPT: frozenset[str] = frozenset({"ve-skill-generator"})


# ── helpers ──────────────────────────────────────────────────────────


def _find_trigger_scope_section(text: str) -> str | None:
    """Extract the ``## Trigger & Scope`` section from SKILL.md content."""
    # Matches both "## Trigger & Scope" and "## Trigger & Scope (Agent-Readable)"
    m = re.search(
        r"^## Trigger & Scope(?: \(Agent-Readable\))?\s*\n(.*?)(?=^## )",
        text, re.MULTILINE | re.DOTALL,
    )
    return m.group(1).strip() if m else None


def _extract_bullets(text: str) -> list[str]:
    """Extract bullet list items (lines starting with ``-`` after ``### SHOULD``)."""
    bullets: list[str] = []
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("- "):
            bullets.append(stripped[2:])
    return bullets


def _extract_section_bullets(scope_text: str, heading: str) -> list[str]:
    """Extract bullet items from a specific subsection like ``### SHOULD Use This Skill When``."""
    m = re.search(
        rf"^### {re.escape(heading)}\s*\n(.*?)(?=\n(?:### |$))",
        scope_text, re.MULTILINE | re.DOTALL,
    )
    if not m:
        return []
    return _extract_bullets(m.group(1))


# Known Volcengine product acronyms — used for cross-lingual Chinese↔English matching
_PRODUCT_ACRONYMS: set[str] = {
    "ecs", "vpc", "rds", "clb", "alb", "iam", "kms", "nat", "vpn",
    "vke", "tos", "cdn", "dns", "sls", "nas", "fg", "ark", "cms",
    "redis", "kafka", "mongodb", "elasticsearch", "polar", "eip",
}


def _tokenize(text: str) -> set[str]:
    """Lower-case word-level tokens (3+ chars, alpha) for matching.

    For mixed Chinese/English text, also extracts product acronyms found
    in the text so that Chinese queries like *创建 ECS 实例* can match
    English SHOULD-scope bullets mentioning *ECS*.
    """
    tokens: set[str] = set()

    # English words (3+ chars)
    for t in re.findall(r"[a-zA-Z][a-zA-Z0-9\-_]{2,}", text):
        tokens.add(t.lower())

    # Product acronyms from Chinese text — scan for any known acronym
    # appearing as uppercase word in the text
    for acro in _PRODUCT_ACRONYMS:
        if re.search(rf"\b{re.escape(acro)}\b", text, re.IGNORECASE):
            tokens.add(acro)

    # Chinese keywords (2+ chars) found in the SHOULD scope — these are
    # common operation/domain terms that appear consistently across skills
    cn_keywords = {"创建", "删除", "查询", "修改", "配置", "启动", "停止",
                   "重启", "扩容", "缩容", "备份", "恢复", "迁移", "监控",
                   "告警", "日志", "权限", "策略", "安全", "网络", "存储",
                   "实例", "集群", "节点", "账户", "账单", "费用", "巡检",
                   "优化", "分析", "诊断", "评估"}
    for kw in cn_keywords:
        if kw in text:
            tokens.add(kw)

    return tokens


def _score_matches(query_tokens: set[str], scope_tokens: set[str]) -> float:
    """Fraction of query tokens present in scope tokens (Jaccard-like coverage).

    Uses a lenient minimum-match formula: how many of the query's non-trivial
    tokens also appear in the scope, scaled by the total number of query tokens.
    A score near 1.0 means nearly every token in the query is relevant to this
    skill. A score of 0.0 means no overlap at all.
    """
    if not query_tokens:
        return 0.0
    # Exclude generic stop-words
    stop = {"the", "this", "that", "with", "from", "when", "what", "how",
            "for", "not", "use", "can", "all", "are", "its", "has", "had",
            "but", "was", "were", "been", "being", "have", "does", "do",
            "will", "would", "should", "could", "may", "might", "shall",
            "also", "about", "into", "over", "both", "than", "then",
            "each", "other", "more", "most", "some", "any", "need"}
    significant = query_tokens - stop
    if not significant:
        return 0.0
    matched = significant & scope_tokens
    return len(matched) / len(significant)


def _has_english_tokens(text: str) -> bool:
    """Check if text contains extractable English words (3+ chars)."""
    return bool(re.search(r"[a-zA-Z][a-zA-Z0-9\-_]{2,}", text))


def _validate_eval_schema(data: list[dict], skill_dir: str) -> list[str]:
    """Validate the shape of an eval_queries.json file."""
    errors: list[str] = []
    REQUIRED_KEYS = {"query", "should_trigger", "skill", "confidence"}
    VALID_BOOLS = {True, False}

    if not data:
        return [f"{skill_dir}: eval_queries.json is empty"]

    for i, entry in enumerate(data):
        if not isinstance(entry, dict):
            errors.append(f"{skill_dir}: entry[{i}] is not a dict")
            continue
        missing = REQUIRED_KEYS - entry.keys()
        if missing:
            errors.append(f"{skill_dir}: entry[{i}] missing keys: {missing}")
            continue
        if not isinstance(entry.get("query"), str) or not entry["query"].strip():
            errors.append(f"{skill_dir}: entry[{i}] 'query' must be non-empty string")
        if entry.get("should_trigger") not in VALID_BOOLS:
            errors.append(f"{skill_dir}: entry[{i}] 'should_trigger' must be bool")
        if not isinstance(entry.get("skill"), str) or not entry["skill"].strip():
            errors.append(f"{skill_dir}: entry[{i}] 'skill' must be non-empty string")
        valid_conf = {"high", "medium", "low"}
        if entry.get("confidence", "") not in valid_conf:
            errors.append(f"{skill_dir}: entry[{i}] 'confidence' must be one of {valid_conf}")

    return errors


# ── git-diff semantic drift ─────────────────────────────────────────


def _get_changed_skills_by_git(root: Path, base_rev: str) -> set[str]:
    """Find skills whose ``## Trigger & Scope`` changed vs *base_rev*.

    Uses ``git diff`` to compare only the SKILL.md files that have
    ``## Trigger & Scope`` in their diff hunk.

    Returns a set of ``ve-*-ops`` directory names.
    """
    try:
        # Merge-base diff: changes on this branch vs base
        base = subprocess.run(
            ["git", "merge-base", "HEAD", base_rev],
            capture_output=True, text=True, check=True, cwd=root,
        ).stdout.strip()

        diff = subprocess.run(
            ["git", "diff", base, "--", "ve-*/SKILL.md"],
            capture_output=True, text=True, check=True, cwd=root,
        )
    except subprocess.CalledProcessError:
        return set()

    changed: set[str] = set()
    current_skill: str | None = None
    in_trigger_scope = False

    for line in diff.stdout.splitlines():
        # Track which file we're in
        if line.startswith("--- a/") or line.startswith("+++ b/"):
            m = re.search(r"ve-([^/]+)-ops/", line)
            current_skill = f"ve-{m.group(1)}-ops" if m else None
            in_trigger_scope = False
        # Check if diff hunk includes Trigger & Scope
        if line.startswith("@@") and current_skill and "Trigger & Scope" in line:
            in_trigger_scope = True
            changed.add(current_skill)
        elif line.startswith("@@") and current_skill:
            in_trigger_scope = False

    return changed


def _get_old_scope(root: Path, skill_dir: Path, base_rev: str) -> str | None:
    """Retrieve old ``## Trigger & Scope`` from git history.

    Uses ``git show`` to get the previous version of SKILL.md, then
    extracts the Trigger & Scope section. Returns ``None`` if the file
    didn't exist at the old revision (new skill).
    """
    rel = skill_dir.relative_to(root)
    path = f"{rel}/SKILL.md"
    try:
        base = subprocess.run(
            ["git", "merge-base", "HEAD", base_rev],
            capture_output=True, text=True, check=True, cwd=root,
        ).stdout.strip()
        old = subprocess.run(
            ["git", "show", f"{base}:{path}"],
            capture_output=True, text=True, check=True, cwd=root,
        )
    except subprocess.CalledProcessError:
        return None  # file didn't exist at old revision

    scope = _find_trigger_scope_section(old.stdout)
    return scope


def _check_delta(
    name: str,
    data: list[dict],
    old_scope_section: str,
    new_scope_section: str,
) -> list[str]:
    """Compare eval query coverage against old vs new scope.

    For each query with ``should_trigger=true`` that has English tokens,
    computes coverage score against *both* scopes. A relative loss ≥50%
    is flagged — e.g. old=0.80 → new=0.30 is a 62% loss.

    This catches semantic drift: "keywords are still there but the
    scope's meaning shifted".
    """
    errors: list[str] = []

    old_bullets = _extract_section_bullets(old_scope_section, "SHOULD Use This Skill When")
    new_bullets = _extract_section_bullets(new_scope_section, "SHOULD Use This Skill When")
    old_tokens = _tokenize(" ".join(old_bullets))
    new_tokens = _tokenize(" ".join(new_bullets))

    # If the new scope has MORE tokens (expansion), that's fine — no regression.
    # Only flag regressions where new < old significantly.
    for entry in data:
        if not entry.get("should_trigger"):
            continue
        query = entry["query"]
        if not _has_english_tokens(query):
            continue

        qt = _tokenize(query)
        stop = {"the", "this", "that", "with", "from", "when", "what", "how",
                "for", "not", "use", "can", "all", "are", "its", "has", "had",
                "but", "was", "were", "been", "being", "have", "does", "do",
                "will", "would", "should", "could", "may", "might", "shall",
                "also", "about", "into", "over", "both", "than", "then",
                "each", "other", "more", "most", "some", "any", "need"}
        sig = qt - stop
        if not sig:
            continue

        old_score = _score_matches(qt, old_tokens)
        new_score = _score_matches(qt, new_tokens)

        # Only flag regressions where old was meaningfully covered
        if old_score < 0.15:
            continue

        # Relative loss: how much of the old coverage was lost
        relative_loss = (old_score - new_score) / old_score if old_score > 0 else 0.0

        if relative_loss >= 0.50:
            errors.append(
                f"{name}: [GIT-DIFF] scope change dropped {relative_loss:.0%} coverage "
                f"(old={old_score:.2f} → new={new_score:.2f}): {query!r}"
            )

    return errors


# ── core check ────────────────────────────────────────────────────────


def _check_skill(skill_dir: Path) -> tuple[str, list[str]]:
    """Run eval regression against one skill directory.

    Returns (skill_name, [error_strings]).
    """
    name = skill_dir.name
    errors: list[str] = []

    # ── 1. eval_queries.json ──
    eval_path = skill_dir / "assets" / "eval_queries.json"
    if not eval_path.exists():
        return name, [f"{name}: missing assets/eval_queries.json"]

    try:
        raw = eval_path.read_text(encoding="utf-8")
        data: list[dict] = json.loads(raw)
    except (json.JSONDecodeError, OSError) as e:
        return name, [f"{name}: assets/eval_queries.json parse error: {e}"]

    schema_errors = _validate_eval_schema(data, name)
    if schema_errors:
        return name, schema_errors

    triggers = [e for e in data if e.get("should_trigger")]
    non_triggers = [e for e in data if not e.get("should_trigger")]

    # Minimum count validation
    issues: list[str] = []
    if len(data) < MIN_TOTAL:
        issues.append(f"only {len(data)} entries (<{MIN_TOTAL})")
    if len(triggers) < MIN_TRIGGER:
        issues.append(f"only {len(triggers)} should_trigger=true (<{MIN_TRIGGER})")
    if len(non_triggers) < MIN_NON_TRIGGER:
        issues.append(f"only {len(non_triggers)} should_trigger=false (<{MIN_NON_TRIGGER})")
    for e in data:
        if e.get("skill") != name:
            issues.append(f"entry 'skill' mismatch: expected '{name}', got '{e.get('skill')}'")

    if issues:
        errors.extend(f"{name}: {i}" for i in issues)
        # Minimum counts are warnings, not regression failures
        # Only return early if the file is truly broken (empty, no triggers at all)
        if not triggers and not non_triggers:
            return name, errors

    # ── 2. SKILL.md Trigger & Scope ──
    skill_path = skill_dir / "SKILL.md"
    if not skill_path.exists():
        errors.append(f"{name}: missing SKILL.md")
        return name, errors

    skill_text = skill_path.read_text(encoding="utf-8")
    scope_section = _find_trigger_scope_section(skill_text)
    if not scope_section:
        errors.append(f"{name}: cannot find ## Trigger & Scope section in SKILL.md")
        return name, errors

    should_bullets = _extract_section_bullets(scope_section, "SHOULD Use This Skill When")
    should_not_bullets = _extract_section_bullets(scope_section, "SHOULD NOT Use This Skill When")
    should_tokens = _tokenize(" ".join(should_bullets))
    should_not_tokens = _tokenize(" ".join(should_not_bullets))

    # ── 3. Regression check ──
    for entry in data:
        query = entry["query"]
        should_trigger = entry["should_trigger"]
        qt = _tokenize(query)
        match_score = _score_matches(qt, should_tokens)
        anti_score = _score_matches(qt, should_not_tokens) if should_not_tokens else 0.0
        # Number of non-trivial tokens in the query (used for minimum-match guard)
        stop = {"the", "this", "that", "with", "from", "when", "what", "how",
                "for", "not", "use", "can", "all", "are", "its", "has", "had",
                "but", "was", "were", "been", "being", "have", "does", "do",
                "will", "would", "should", "could", "may", "might", "shall",
                "also", "about", "into", "over", "both", "than", "then",
                "each", "other", "more", "most", "some", "any", "need"}
        sig_tokens = qt - stop
        n_tokens = len(sig_tokens)

        # ── Skip content regression for queries with no English tokens ──
        # Chinese-only queries (e.g. "创建安全组") cannot do meaningful
        # cross-lingual matching against English Trigger & Scope sections.
        # We still validate schema + skill field above.
        if not _has_english_tokens(query):
            continue

        if should_trigger and match_score < 0.15:
            errors.append(
                f"{name}: SHOULD trigger but low scope match "
                f"(score={match_score:.2f}, tokens={n_tokens}): {query!r}"
            )
        elif not should_trigger and n_tokens >= 2 and match_score > 0.75:
            # Require ≥2 matching tokens for non-trigger flagging — this avoids
            # false positives where only the product acronym happens to match
            # (e.g. "查一下 VPC 的广播功能怎么配置" is correctly non-trigger
            # but matches "VPC" in the scope)
            if anti_score < 0.3:
                errors.append(
                    f"{name}: SHOULD NOT trigger but high scope match "
                    f"(score={match_score:.2f}, tokens={n_tokens}, "
                    f"anti={anti_score:.2f}): {query!r}"
                )

    return name, errors


# ── CLI ──────────────────────────────────────────────────────────────


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--json", action="store_true", help="Machine-readable JSON output")
    parser.add_argument(
        "--git-diff", nargs="?", const="HEAD~1", default=None, metavar="REV",
        help="Enable git-diff semantic drift detection. Default rev: HEAD~1. "
             "Example: --git-diff origin/main",
    )
    return parser


def main() -> int:
    args = build_parser().parse_args()
    root = args.root.resolve()

    results: dict[str, Any] = {
        "passed": [],
        "failed": [],
        "warnings": [],
        "meta": {
            "total_skills": 0,
            "passed_count": 0,
            "failed_count": 0,
            "total_errors": 0,
        },
    }

    if not args.json:
        print("Pass 1 — Absolute match (schema + keyword coverage)")
        print()

    for skill_dir in sorted(root.glob("ve-*/")):
        name = skill_dir.name
        if name in EXEMPT:
            continue
        if not (skill_dir / "SKILL.md").exists():
            continue

        results["meta"]["total_skills"] += 1
        skill_name, errors = _check_skill(skill_dir)
        if errors:
            results["failed"].append({"skill": skill_name, "errors": errors})
            results["meta"]["failed_count"] += 1
            results["meta"]["total_errors"] += len(errors)
        else:
            results["passed"].append(skill_name)
            results["meta"]["passed_count"] += 1

    # ── Pass 2: git-diff semantic drift ────────────────────────────
    if args.git_diff is not None:
        if not args.json:
            print()
            print("Pass 2 — Git-diff semantic drift (old scope vs new scope)")
            print()

        changed_skills = _get_changed_skills_by_git(root, args.git_diff)
        if not changed_skills:
            if not args.json:
                print("  No Trigger & Scope changes detected — skipping")
        else:
            for skill_dir in sorted(root.glob("ve-*/")):
                name = skill_dir.name
                if name not in changed_skills:
                    continue

                # Load current eval data
                eval_path = skill_dir / "assets" / "eval_queries.json"
                if not eval_path.exists():
                    continue
                try:
                    data: list[dict] = json.loads(eval_path.read_text(encoding="utf-8"))
                except (json.JSONDecodeError, OSError):
                    continue

                # Get current scope
                skill_text = (skill_dir / "SKILL.md").read_text(encoding="utf-8")
                new_scope = _find_trigger_scope_section(skill_text)
                if not new_scope:
                    continue

                # Get old scope from git
                old_scope = _get_old_scope(root, skill_dir, args.git_diff)
                if not old_scope:
                    if not args.json:
                        print(f"  SKIP {name}: no old scope in git (new skill)")
                    continue

                # Compare
                delta = _check_delta(name, data, old_scope, new_scope)
                if delta:
                    results["failed"].append({"skill": name, "errors": delta})
                    results["meta"]["failed_count"] += 1
                    results["meta"]["total_errors"] += len(delta)
                elif not args.json:
                    print(f"  OK   {name} — no coverage regression")

    # Output
    if args.json:
        print(json.dumps(results, ensure_ascii=False, indent=2))
    else:
        print()
        for p in sorted(results["passed"]):
            print(f"  OK:   {p}")
        for f in results["failed"]:
            print(f"  FAIL: {f['skill']}")
            for e in f["errors"]:
                print(f"         {e}")
        print()
        print(f"Skills: {results['meta']['passed_count']} passed, "
              f"{results['meta']['failed_count']} failed "
              f"({results['meta']['total_skills']} total)")
        print(f"Errors: {results['meta']['total_errors']}")

    return 1 if results["failed"] else 0


if __name__ == "__main__":
    sys.exit(main())