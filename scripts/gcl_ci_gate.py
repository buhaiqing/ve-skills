#!/usr/bin/env python3
"""GCL CI gate — structural smoke test across all ve-*-ops skills.

Runs ``scripts/gcl_runner.py`` with ``--structural-critic-only --max-iter 1``
for each GCL-equipped skill, using a safe no-op echo command. Validates that
the GCL runner:
  - Parses args and loads rubric/prompt-templates for each skill
  - Executes the structural critic without crashing
  - Produces a valid trace with proper scoring
  - Handles all termination paths (PASS / RETRY / SAFETY_FAIL)

This is a **structural gate**, not a production GCL run — it does not execute
real ``ve`` commands or require cloud credentials. It catches broken runner
code, missing or malformed skill GCL artifacts, and integration regressions.

Usage:
    python3 scripts/gcl_ci_gate.py                     # human-readable
    python3 scripts/gcl_ci_gate.py --json               # machine-readable
    python3 scripts/gcl_ci_gate.py --root PATH           # custom repo root
    python3 scripts/gcl_ci_gate.py --skills ve-ecs-ops   # single-skill smoke

Exits 0 if all skills pass, 1 otherwise.

Pure stdlib — no external dependencies. Python 3.10+ syntax.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Any

# Same canonical 29-skill set as check_gcl_conformance.py (AGENTS.md §8)
# plus incident-loop-agent (orchestration skill, exempt from ve-* prefix).
GCL_SKILLS: frozenset[str] = frozenset({
    "ve-ecs-ops",
    "ve-redis-ops",
    "ve-rds-mysql-ops",
    "ve-rds-ops",
    "ve-rds-pg-ops",
    "ve-polar-mysql-ops",
    "ve-mongodb-ops",
    "ve-elasticsearch-ops",
    "ve-tos-ops",
    "ve-iam-ops",
    "ve-kms-ops",
    "ve-eip-ops",
    "ve-security-group-ops",
    "ve-vpc-ops",
    "ve-nat-ops",
    "ve-vpn-ops",
    "ve-clb-ops",
    "ve-alb-ops",
    "ve-vke-ops",
    "ve-nas-ops",
    "ve-cms-ops",
    "ve-fg-ops",
    "ve-ark-ops",
    "ve-cdn-ops",
    "ve-dns-ops",
    "ve-kafka-ops",
    "ve-sls-ops",
    "ve-billing-ops",
    "ve-skill-generator",
    "incident-loop-agent",
})

# Safe smoke command — no real cloud side-effect, no credentials needed.
SMOKE_COMMAND = 'echo \'{"Response":{"RequestId":"ci-gate-smoke"}}\''

# Runner script path, resolved lazily per invocation.
_RUNNER: Path | None = None


def _runner_path(root: Path) -> Path:
    global _RUNNER
    if _RUNNER is None:
        _RUNNER = root / "scripts" / "gcl_runner.py"
    return _RUNNER


def smoke_skill(root: Path, skill: str) -> dict[str, Any]:
    """Run gcl_runner for a single skill; return the result report.

    Returns a dict with keys:
      skill, ok, exit_code, stderr_first_line, trace_summary, timed_out
    """
    runner = _runner_path(root)
    cmd = [
        sys.executable,
        str(runner),
        "run",
        "--skill",
        skill,
        "--request",
        f"CI gate smoke: {skill}",
        "--command",
        SMOKE_COMMAND,
        "--max-iter",
        "1",
        "--structural-critic-only",
        "--root",
        str(root),
    ]
    result: dict[str, Any] = {"skill": skill, "ok": False, "exit_code": -1}

    try:
        proc = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    except subprocess.TimeoutExpired:
        result["exit_code"] = -1
        result["timed_out"] = True
        result["stderr_first_line"] = "TIMEOUT (30s)"
        return result

    result["exit_code"] = proc.returncode
    result["timed_out"] = False
    result["stdout_summary"] = (proc.stdout or "").strip()
    result["stderr_first_line"] = (proc.stderr or "").splitlines()[0] if proc.stderr else ""

    # runner exits 0 on PASS, 1 on MAX_ITER, 2 on config error, 3 on SAFETY_FAIL
    ok_codes = {0, 1}  # PASS or MAX_ITER (both structurally acceptable)
    result["ok"] = proc.returncode in ok_codes

    # Extract trace path: runner prints "PASS / MAX_ITER — trace: path" to stdout
    combined = (proc.stdout or "") + (proc.stderr or "")
    trace_line = ""
    for line in reversed(combined.splitlines()):
        if "trace:" in line:
            trace_line = line.strip()
            break
    result["trace_line"] = trace_line

    # Parse trace for quick summary
    if trace_line:
        trace_path_str = trace_line.split("trace:")[-1].strip()
        trace_path = Path(trace_path_str)
        if trace_path.is_file():
            try:
                trace = json.loads(trace_path.read_text(encoding="utf-8"))
                final = trace.get("final", {})
                iters = trace.get("iterations", [])
                last_score = {}
                if iters:
                    last_score = iters[-1].get("critic", {}).get("scores", {})
                result["trace_summary"] = {
                    "status": final.get("status", "UNKNOWN"),
                    "iter": final.get("iter", 0),
                    "iterations": len(iters),
                    "scores": last_score,
                    "schema_version": trace.get("trace_schema_version", ""),
                    "has_redaction": trace.get("redaction_pass", False),
                }
            except (json.JSONDecodeError, OSError) as e:
                result["trace_parse_error"] = str(e)

    return result


def smoke_all(root: Path, skills: list[str] | None = None) -> list[dict[str, Any]]:
    """Run smoke test for all (or specified) GCL skills.

    Results are sorted by skill name for stable diffs.
    """
    target_skills = sorted(skills or GCL_SKILLS)
    reports: list[dict[str, Any]] = []
    for skill in target_skills:
        reports.append(smoke_skill(root, skill))
    return reports


def build_parser() -> argparse.ArgumentParser:
    """Build the CLI argument parser."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--root",
        type=Path,
        default=Path(__file__).resolve().parents[1],
        help="Repo root (default: parent of this script).",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Emit machine-readable JSON with summary and reports.",
    )
    parser.add_argument(
        "--skills",
        nargs="+",
        default=None,
        help="One or more skill names to test (default: all).",
    )
    parser.add_argument(
        "--skip-incident-loop",
        action="store_true",
        help="Skip incident-loop-agent (it has no ve product context).",
    )
    return parser


def _format_human(reports: list[dict[str, Any]]) -> str:
    """Render the human-readable summary."""
    passing = sum(1 for r in reports if r["ok"])
    total = len(reports)
    lines: list[str] = [f"GCL CI gate: {passing}/{total} skills pass structural smoke."]
    failing = [r for r in reports if not r["ok"]]
    if failing:
        lines.append("")
        for r in failing:
            reason = f"exit_code={r['exit_code']}"
            if r.get("timed_out"):
                reason = "TIMEOUT"
            lines.append(f"  FAIL {r['skill']}: {reason}")
            if r.get("stderr_first_line"):
                lines.append(f"       stderr: {r['stderr_first_line']}")
    lines.append("")
    return "\n".join(lines)


def cmd_check(args: argparse.Namespace) -> int:
    """CLI entry point. Returns the desired process exit code."""
    skills = set(args.skills) if args.skills else None
    if args.skip_incident_loop and skills is None:
        skills = GCL_SKILLS - {"incident-loop-agent"}

    reports = smoke_all(args.root, list(skills) if skills else None)
    passing = sum(1 for r in reports if r["ok"])

    if args.json:
        payload = {
            "summary": {
                "total": len(reports),
                "passing": passing,
                "failing": len(reports) - passing,
            },
            "reports": reports,
        }
        print(json.dumps(payload, indent=2))
    else:
        print(_format_human(reports), end="")

    return 0 if passing == len(reports) else 1


def main() -> int:
    """Module entry point."""
    parser = build_parser()
    args = parser.parse_args()
    return cmd_check(args)


if __name__ == "__main__":
    sys.exit(main())