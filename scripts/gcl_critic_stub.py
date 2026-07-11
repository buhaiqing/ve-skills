#!/usr/bin/env python3
"""Isolated GCL Critic — reference implementation (AGENTS.md §9).

Runs as a SEPARATE process from the Generator, satisfying the hard constraint
that Generator and Critic live in isolated contexts. ``gcl_runner.py
--critic-command`` pipes a sanitized critic-input JSON to this process's stdin
(NEVER the raw user request) and expects a critic JSON on stdout.

This stub is rule-based (structural audit). A production deployment should
replace it with an LLM-backed Critic that reads ``references/rubric.md`` (the
``rubric_path`` field in the input) for product-specific scoring.

Usage (wired by gcl_runner.py):
  echo "$critic_input_json" | python3 scripts/gcl_critic_stub.py
"""
from __future__ import annotations

import json
import re
import sys
from typing import Any

RUBRIC_THRESHOLDS = {
    "correctness": 0.5,
    "safety": 1.0,
    "idempotency": 0.5,
    "traceability": 0.5,
    "spec_compliance": 0.5,
}

_SECRET_RE = re.compile(r"SecretKey\s*=\s*\S|VOLCENGINE_SECRET_KEY\s*=\s*\S|AKLT[A-Za-z0-9]{20,}", re.I)


def main() -> int:
    raw = sys.stdin.read()
    try:
        data = json.loads(raw)
    except json.JSONDecodeError as e:
        print(f"ERROR: invalid critic-input JSON: {e}", file=sys.stderr)
        return 2

    # Isolation guard: the runner must never send the raw request.
    if "request" in data:
        print("ERROR: Critic received raw `request` — violates AGENTS.md §9 isolation", file=sys.stderr)
        return 2

    gen = data.get("generator_output", {})
    excerpt = gen.get("result_excerpt", "") or ""
    cmd = gen.get("command", "") or ""
    exit_code = gen.get("exit_code")

    scores: dict[str, float] = {}
    suggestions: list[str] = []

    scores["correctness"] = 1.0 if exit_code == 0 else 0.0
    if exit_code != 0:
        suggestions.append(f"Generator exit_code={exit_code}; fix command or credentials")

    leak = _SECRET_RE.search(excerpt + cmd)
    scores["safety"] = 0.0 if leak else 1.0
    if leak:
        suggestions.append("Credential leak in generator output — mask and re-run")

    scores["idempotency"] = 0.5
    scores["traceability"] = 1.0 if (cmd and excerpt) else 0.5
    if not excerpt:
        suggestions.append("Empty generator output — capture stdout/stderr in trace")

    scores["spec_compliance"] = 1.0 if exit_code == 0 else 0.0

    blocking = scores["safety"] == 0.0 or scores["correctness"] == 0.0
    out = {
        "scores": {k: scores[k] for k in RUBRIC_THRESHOLDS},
        "suggestions": suggestions[:3],
        "blocking": blocking,
    }
    print(json.dumps(out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
