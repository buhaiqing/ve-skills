#!/usr/bin/env python3
# DEPRECATED: superseded by cmd/vet — run `vet gcl run`
"""GCL Orchestrator (Phase 2) — Generator execution loop with external Critic injection.

Implements the **Orchestrator (O)** role from AGENTS.md GCL spec. Generator runs
``ve``/shell commands; Critic scores MUST come from an **isolated** context — via
``--critic-json``, stdin, or a separate ``--critic-command`` process (this script
never self-scores as Critic in production mode).

Loop closure (fixes the earlier open-loop defect):
  * Prior Critic suggestions are injected into the Generator via the
    ``GCL_CRITIC_FEEDBACK`` env var on every iteration (AGENTS.md §4 [3]).
  * Known failure patterns are injected via ``GCL_KNOWN_FAILURE_PATTERNS``
    (Reflexion pre-flight, AGENTS.md §12).
  * Non-null failure patterns are written back to ``docs/failure-patterns.md``
    at the end of a MAX_ITER / SAFETY_FAIL run (Reflexion write-back).

Usage:
  python3 scripts/gcl_runner.py run \\
    --skill ve-ecs-ops \\
    --request "List ECS instances read-only" \\
    --command 've ecs DescribeInstances --Region cn-beijing' \\
    [--max-iter 2] \\
    [--critic-json path/to/critic.json] \\
    [--critic-command 'python3 scripts/gcl_critic_stub.py']

  # Rule-based structural audit only (CI / dry-run; NOT a substitute for isolated Critic):
  python3 scripts/gcl_runner.py run ... --structural-critic-only

Trace output: ``audit-results/gcl-trace-YYYYMMDD-HHMMSS.json``
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

# Allow importing sibling scripts (gcl_trace_aggregate) for Reflexion write-back.
sys.path.insert(0, str(Path(__file__).resolve().parent))

# Per AGENTS.md §8 defaults (override via --max-iter)
SKILL_MAX_ITER: dict[str, int] = {
    "ve-ecs-ops": 2,
    "ve-redis-ops": 2,
    "ve-rds-mysql-ops": 2,
    "ve-rds-ops": 2,
    "ve-rds-pg-ops": 2,
    "ve-polar-mysql-ops": 2,
    "ve-mongodb-ops": 2,
    "ve-elasticsearch-ops": 2,
    "ve-tos-ops": 2,
    "ve-iam-ops": 2,
    "ve-kms-ops": 2,
    "ve-eip-ops": 2,
    "ve-security-group-ops": 2,
    "ve-vpc-ops": 3,
    "ve-nat-ops": 3,
    "ve-vpn-ops": 3,
    "ve-clb-ops": 3,
    "ve-alb-ops": 3,
    "ve-vke-ops": 3,
    "ve-nas-ops": 3,
    "ve-cms-ops": 3,
    "ve-fg-ops": 3,
    "ve-ark-ops": 3,
    "ve-cdn-ops": 5,
    "ve-dns-ops": 5,
    "ve-kafka-ops": 5,
    "ve-sls-ops": 5,
    "ve-billing-ops": 5,
    "ve-skill-generator": 3,
}

RUBRIC_THRESHOLDS: dict[str, float] = {
    "correctness": 0.5,
    "safety": 1.0,
    "idempotency": 0.5,
    "traceability": 0.5,
    "spec_compliance": 0.5,
}

SECRET_PATTERNS = [
    re.compile(r"SecretKey\s*=\s*[^<\s][^\s\"']+", re.I),
    re.compile(r"VOLCENGINE_SECRET_KEY\s*=\s*[^<\s][^\s\"']+", re.I),
    re.compile(r"AKLT(?![<\s])[A-Za-z0-9]{20,}"),
]


def mask_secrets(text: str) -> str:
    out = text
    out = re.sub(r"(SecretKey\s*=\s*)([^\s\"']+)", r"\1<masked>", out, flags=re.I)
    out = re.sub(r"(VOLCENGINE_SECRET_KEY\s*=\s*)([^\s\"']+)", r"\1<masked>", out, flags=re.I)
    out = re.sub(r"(AKLT)([A-Za-z0-9]{20,})", r"\1<masked>", out)
    return out


def has_credential_leak(text: str) -> bool:
    return any(p.search(text) for p in SECRET_PATTERNS)


def run_command(
    command: str,
    timeout: int = 120,
    extra_env: dict[str, str] | None = None,
) -> dict[str, Any]:
    """Execute generator command; capture exit code and masked output.

    ``extra_env`` carries loop-state into the Generator without rewriting the
    command string: ``GCL_CRITIC_FEEDBACK`` (prior Critic suggestions) and
    ``GCL_KNOWN_FAILURE_PATTERNS`` (Reflexion hints). This is how the
    Orchestrator "injects suggestions into G" (AGENTS.md §4 [3]) — the Generator
    reads these from its environment and adapts, closing the G↔C loop.
    """
    try:
        env = dict(os.environ)
        if extra_env:
            env.update({k: v for k, v in extra_env.items() if v})
        proc = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=timeout,
            env=env,
        )
        combined = (proc.stdout or "") + (proc.stderr or "")
        masked = mask_secrets(combined)
        excerpt = masked[:2000] + ("..." if len(masked) > 2000 else "")
        return {
            "command": mask_secrets(command),
            "exit_code": proc.returncode,
            "result_excerpt": excerpt,
            "stdout_len": len(proc.stdout or ""),
            "stderr_len": len(proc.stderr or ""),
        }
    except subprocess.TimeoutExpired:
        return {
            "command": mask_secrets(command),
            "exit_code": -1,
            "result_excerpt": f"TIMEOUT after {timeout}s",
            "stdout_len": 0,
            "stderr_len": 0,
        }


def structural_critic(generator: dict[str, Any]) -> dict[str, Any]:
    """Rule-based structural audit — for CI/dry-run only, not production Critic."""
    scores: dict[str, float] = {}
    suggestions: list[str] = []

    exit_code = generator.get("exit_code", -1)
    excerpt = generator.get("result_excerpt", "")
    cmd = generator.get("command", "")

    scores["correctness"] = 1.0 if exit_code == 0 else 0.0
    if exit_code != 0:
        suggestions.append(f"Generator exit_code={exit_code}; fix command or credentials")

    leak = has_credential_leak(excerpt) or has_credential_leak(cmd)
    scores["safety"] = 0.0 if leak else 1.0
    if leak:
        suggestions.append("Credential leak in trace — mask SecretKey and re-run")

    scores["idempotency"] = 0.5
    scores["traceability"] = 1.0 if cmd and excerpt else 0.5
    if not excerpt:
        suggestions.append("Empty generator output — capture stdout/stderr in trace")

    scores["spec_compliance"] = 1.0 if exit_code == 0 else 0.0
    if exit_code == 0 and "ve" not in cmd and "python" not in cmd.lower():
        scores["spec_compliance"] = 0.5  # structural smoke: command succeeded

    blocking = scores["safety"] == 0.0 or scores["correctness"] == 0.0
    return {
        "scores": scores,
        "suggestions": suggestions[:3],
        "blocking": blocking,
        "_mode": "structural-only",
    }


def load_critic(path: Path | None, stdin: bool) -> dict[str, Any] | None:
    if path:
        return json.loads(path.read_text(encoding="utf-8"))
    if stdin and not sys.stdin.isatty():
        return json.loads(sys.stdin.read())
    return None


def run_isolated_critic(
    skill: str,
    operation_intent: dict[str, Any],
    generator: dict[str, Any],
    iterations: list[dict[str, Any]],
    critic_cmd: str,
    root: Path,
    timeout: int = 120,
) -> dict[str, Any] | None:
    """Run an ISOLATED Critic in a separate process (AGENTS.md §9).

    The Critic receives ONLY sanitized inputs — ``operation_intent``,
    ``generator_output``, and the prior ``trace`` — never the raw user
    ``request``. This satisfies the hard constraint that the Critic MUST NOT
    see the raw request (prevents answer-aligned rubber-stamping) and that G
    and C live in isolated contexts. The external command must print a valid
    critic JSON to stdout.
    """
    rubric_path = root / skill / "references" / "rubric.md"
    critic_input = {
        "skill": skill,
        "operation_intent": operation_intent,
        "generator_output": {
            "command": generator.get("command", ""),
            "exit_code": generator.get("exit_code"),
            "result_excerpt": generator.get("result_excerpt", ""),
        },
        "trace": {"iterations": iterations},
        "rubric_path": (
            str(rubric_path.relative_to(root)) if rubric_path.exists() else "AGENTS.md"
        ),
        # Raw `request` is intentionally excluded from the Critic input.
    }
    try:
        proc = subprocess.run(
            critic_cmd,
            shell=True,
            input=json.dumps(critic_input, ensure_ascii=False),
            capture_output=True,
            text=True,
            timeout=timeout,
            env=dict(os.environ),
        )
    except subprocess.TimeoutExpired:
        print("ERROR: Critic command timed out", file=sys.stderr)
        return None
    if proc.returncode != 0:
        print(
            f"ERROR: Critic command failed ({proc.returncode}): {proc.stderr[:500]}",
            file=sys.stderr,
        )
        return None
    try:
        critic = json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        print(f"ERROR: Critic output is not valid JSON: {e}", file=sys.stderr)
        return None
    errs = validate_critic_payload(critic)
    if errs:
        print("ERROR: Invalid critic JSON from subprocess:", "; ".join(errs), file=sys.stderr)
        return None
    return critic


def validate_critic_payload(critic: dict[str, Any]) -> list[str]:
    errs: list[str] = []
    scores = critic.get("scores")
    if not isinstance(scores, dict):
        return ["critic.scores must be object"]
    for dim in RUBRIC_THRESHOLDS:
        if dim not in scores:
            errs.append(f"critic.scores missing '{dim}'")
        elif scores[dim] not in (0, 0.5, 1, 0.0, 1.0):
            errs.append(f"critic.scores.{dim} must be 0, 0.5, or 1")
    if "suggestions" not in critic:
        errs.append("critic.suggestions required")
    if "blocking" not in critic:
        errs.append("critic.blocking required")
    return errs


def decide(scores: dict[str, float]) -> str:
    if scores.get("safety", 1) == 0:
        return "SAFETY_FAIL"
    for dim, threshold in RUBRIC_THRESHOLDS.items():
        if scores.get(dim, 0) < threshold:
            return "RETRY"
    return "PASS"


# Reflexion: failure-pattern extraction (AGENTS.md §14.6).
# Maps Generator output + Critic suggestions to a structured failure_pattern
# block that callers (or Reflexion pre-flight) can persist to
# docs/failure-patterns.md. Categories match the schema in that file:
#   cli_parameter | skill_generation | cross_skill | runtime | token_efficiency
_FAILURE_SIGNATURES: list[tuple[str, re.Pattern[str]]] = [
    ("cli_parameter", re.compile(r"InvalidParameter|MissingParameter|AuthFailure|UnauthorizedOperation", re.I)),
    ("runtime", re.compile(r"TIMEOUT|RequestLimitExceeded|InternalError|ConnectionError", re.I)),
    ("cross_skill", re.compile(r"delegate-to|not found in target skill|cross-skill", re.I)),
    ("token_efficiency", re.compile(r"token budget|exceeds.*token|too long|truncated", re.I)),
    ("skill_generation", re.compile(r"frontmatter missing|missing rubric|broken link", re.I)),
]


def extract_failure_pattern(
    skill: str,
    command: str,
    generator: dict[str, Any],
    critic: dict[str, Any],
) -> dict[str, Any] | None:
    """Heuristic failure-pattern extraction. Returns None if no pattern matched.

    The schema mirrors ``docs/failure-patterns.md`` so that traces can feed
    Reflexion memory directly. Count starts at 1; downstream tooling is
    expected to dedup-and-increment before persisting.
    """
    corpus_parts = [
        command or "",
        generator.get("result_excerpt", "") or "",
        *(critic.get("suggestions") or []),
    ]
    corpus = "\n".join(corpus_parts)
    for category, pattern in _FAILURE_SIGNATURES:
        match = pattern.search(corpus)
        if not match:
            continue
        fix = (critic.get("suggestions") or ["Investigate failure pattern and add fix"])[0]
        return {
            "category": category,
            "skill": skill,
            "command": command[:200] if command else None,
            "error": match.group(0),
            "fix": fix[:200],
            "count": 1,
            "reusable": category in {"cli_parameter", "runtime"},
        }
    return None


def load_known_failure_patterns(root: Path, skill: str, limit: int = 10) -> str:
    """Reflexion pre-flight (AGENTS.md §12 / reflexion-memory.md §5).

    Load known failure patterns for this skill from ``docs/failure-patterns.md``
    and return them as a hint string the Orchestrator injects into the Generator
    environment (``GCL_KNOWN_FAILURE_PATTERNS``). This closes the Reflexion read
    side; matches are filtered by skill name. Returns "" when absent.
    """
    fp = root / "docs" / "failure-patterns.md"
    if not fp.exists():
        return ""
    out: list[str] = []
    for line in fp.read_text(encoding="utf-8").splitlines():
        s = line.strip()
        if not s.startswith("|") or skill not in s:
            continue
        if "---" in s or ("skill" in s.lower() and "command" in s.lower()):
            continue  # separator or header row
        out.append(s)
        if len(out) >= limit:
            break
    return "\n".join(out)


def _writeback_failure_pattern(root: Path, skill: str, failure_pattern: Any) -> None:
    """Reflexion write-back: persist a non-null failure_pattern to docs/failure-patterns.md.

    Uses scripts/gcl_trace_aggregate.update_failure_patterns_file so the schema
    stays consistent. Never raises — Reflexion must not break the main loop.
    """
    if not failure_pattern:
        return
    try:
        import gcl_trace_aggregate as gcl_agg

        pattern = failure_pattern["error"] if isinstance(failure_pattern, dict) else str(failure_pattern)
        category = (failure_pattern.get("category") if isinstance(failure_pattern, dict) else None) or "runtime"
        source = (failure_pattern.get("command") if isinstance(failure_pattern, dict) else None) or "gcl-runner"
        summary = {
            "failure_patterns": [
                {"skill": skill, "pattern": pattern, "category": category, "source": source}
            ]
        }
        gcl_agg.update_failure_patterns_file(root, summary)
    except Exception as e:  # Reflexion must never break the main loop
        print(f"WARN: Reflexion write-back skipped: {e}", file=sys.stderr)


def detect_credential_fields(text: str) -> list[str]:
    """Detect which credential fields are present in raw text (before masking).

    Returns field names that would be redacted by mask_secrets().
    """
    fields: list[str] = []
    if re.search(r"SecretKey\s*=", text, re.I):
        fields.append("SecretKey")
    if re.search(r"VOLCENGINE_SECRET_KEY\s*=", text, re.I):
        fields.append("VOLCENGINE_SECRET_KEY")
    if re.search(r"AKLT[A-Za-z0-9]{20,}", text):
        fields.append("AKLT_token")
    return fields


DESTRUCTIVE_VERBS = re.compile(
    r"\b(Delete|Remove|Terminate|Destroy|Stop|Shutdown|PowerOff|Release|Revoke|Disable|Deactivate|"
    r"Flush|Purge|Drop|Truncate|Detach|Disassociate|Revoke\w*Access|Revoke\w*Permission)\b",
    re.I,
)
MUTATING_VERBS = re.compile(
    r"\b(Create|Add|Allocate|Attach|Assign|Authorize|Enable|Activate|"
    r"Modify|Update|Set|Change|Resize|Rebuild|Reboot|Restart)\b",
    re.I,
)
NEGATION_PATTERN = re.compile(
    r"\b(Enable|Activate|Allow|Grant|Create)\w*(Protection|Policy|Rule|Firewall)\b",
    re.I,
)


def derive_operation_intent(skill: str, command: str) -> dict[str, Any]:
    if not command:
        return {
            "operation": "unknown",
            "resource_scope": [],
            "expected_state": "unknown",
            "safety_class": "read_only",
        }

    resource = skill.replace("ve-", "").replace("-ops", "")
    cmd_stripped = re.sub(r"#.*$", "", command, flags=re.M)

    if NEGATION_PATTERN.search(cmd_stripped):
        return {
            "operation": f"enable_{resource}",
            "resource_scope": [],
            "expected_state": "ACTIVE",
            "safety_class": "mutating",
        }

    if DESTRUCTIVE_VERBS.search(cmd_stripped):
        return {
            "operation": f"destructive_{resource}",
            "resource_scope": [],
            "expected_state": "DELETED",
            "safety_class": "destructive",
        }

    if MUTATING_VERBS.search(cmd_stripped):
        return {
            "operation": f"modify_{resource}",
            "resource_scope": [],
            "expected_state": "MODIFIED",
            "safety_class": "mutating",
        }

    return {
        "operation": "describe",
        "resource_scope": [],
        "expected_state": "UNCHANGED",
        "safety_class": "read_only",
    }


def persist_trace(root: Path, trace: dict[str, Any]) -> Path:
    out_dir = root / "audit-results"
    out_dir.mkdir(parents=True, exist_ok=True)
    ts = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    path = out_dir / f"gcl-trace-{ts}.json"
    path.write_text(json.dumps(trace, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return path


def cmd_run(args: argparse.Namespace) -> int:
    root = args.root
    max_iter = args.max_iter or SKILL_MAX_ITER.get(args.skill, 3)
    command = args.command

    operation_intent = derive_operation_intent(args.skill, command)
    masked_fields = detect_credential_fields(command)
    known_patterns = load_known_failure_patterns(root, args.skill)

    trace: dict[str, Any] = {
        "trace_schema_version": "v1",
        "skill": args.skill,
        "request": args.request,
        "rubric_version": "v1",
        "operation_intent": operation_intent,
        "masked_fields": masked_fields,
        "redaction_pass": True,
        "iterations": [],
    }

    critic_feedback = ""

    for iteration in range(1, max_iter + 1):
        generator = run_command(
            command,
            timeout=args.timeout,
            extra_env={
                "GCL_CRITIC_FEEDBACK": critic_feedback,
                "GCL_KNOWN_FAILURE_PATTERNS": known_patterns,
            },
        )
        generator["args"] = {"iter": iteration, "critic_feedback": critic_feedback or None}

        if args.structural_critic_only:
            critic = structural_critic(generator)
        else:
            if has_credential_leak(generator.get("result_excerpt", "")):
                trace["iterations"].append(
                    {
                        "iter": iteration,
                        "generator": generator,
                        "critic": {
                            "scores": {
                                "correctness": 0,
                                "safety": 0,
                                "idempotency": 0.5,
                                "traceability": 0.5,
                                "spec_compliance": 0.5,
                            },
                            "suggestions": ["Credential leak in generator output — mask secrets and re-run"],
                            "blocking": True,
                        },
                        "decision": "SAFETY_FAIL",
                    }
                )
                trace["final"] = {
                    "status": "SAFETY_FAIL",
                    "iter": iteration,
                    "output": None,
                    "failure_pattern": extract_failure_pattern(
                        args.skill, command, generator, trace["iterations"][-1]["critic"]
                    ),
                }
                path = persist_trace(root, trace)
                _writeback_failure_pattern(root, args.skill, trace["final"].get("failure_pattern"))
                print(f"SAFETY_FAIL — credential leak detected — trace: {path}", file=sys.stderr)
                return 3
            critic = load_critic(args.critic_json, args.critic_stdin)
            if critic is None and args.critic_command:
                critic = run_isolated_critic(
                    args.skill, operation_intent, generator,
                    trace["iterations"], args.critic_command, root,
                )
            if critic is None:
                print(
                    "ERROR: No Critic payload. Pass --critic-json, pipe JSON to stdin, "
                    "--critic-command <cmd>, or use --structural-critic-only for rule-based audit.",
                    file=sys.stderr,
                )
                return 2
            errs = validate_critic_payload(critic)
            if errs:
                print("ERROR: Invalid critic JSON:", "; ".join(errs), file=sys.stderr)
                return 2

        decision = decide(critic["scores"])
        trace["iterations"].append(
            {
                "iter": iteration,
                "generator": generator,
                "critic": {
                    "scores": critic["scores"],
                    "suggestions": critic.get("suggestions", []),
                    "blocking": critic.get("blocking", False),
                },
                "decision": decision,
            }
        )

        if decision == "SAFETY_FAIL":
            trace["final"] = {
                "status": "SAFETY_FAIL",
                "iter": iteration,
                "output": None,
                "failure_pattern": extract_failure_pattern(
                    args.skill, command, generator, critic
                ),
            }
            path = persist_trace(root, trace)
            _writeback_failure_pattern(root, args.skill, trace["final"].get("failure_pattern"))
            print(f"SAFETY_FAIL — trace: {path}", file=sys.stderr)
            return 3

        if decision == "PASS":
            trace["final"] = {
                "status": "PASS",
                "iter": iteration,
                "output": generator.get("result_excerpt", ""),
            }
            path = persist_trace(root, trace)
            print(f"PASS (iter {iteration}) — trace: {path}")
            return 0

        critic_feedback = "; ".join(critic.get("suggestions", [])[:3])

    trace["final"] = {
        "status": "MAX_ITER",
        "iter": max_iter,
        "output": trace["iterations"][-1]["generator"].get("result_excerpt", "") if trace["iterations"] else None,
        "unresolved": [
            d for d, t in RUBRIC_THRESHOLDS.items()
            if trace["iterations"][-1]["critic"]["scores"].get(d, 0) < t
        ],
        "failure_pattern": extract_failure_pattern(
            args.skill, command, trace["iterations"][-1]["generator"], trace["iterations"][-1]["critic"]
        ),
    }
    path = persist_trace(root, trace)
    _writeback_failure_pattern(root, args.skill, trace["final"].get("failure_pattern"))
    print(f"MAX_ITER — trace: {path}", file=sys.stderr)
    return 1


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description=__doc__)
    sub = p.add_subparsers(dest="cmd", required=True)

    run = sub.add_parser("run", help="Execute GCL loop")
    run.add_argument(
        "--root",
        type=Path,
        default=Path(__file__).resolve().parents[1],
        help="Repository root (default: parent of scripts/)",
    )
    run.add_argument("--skill", required=True, help="Skill id, e.g. ve-ecs-ops")
    run.add_argument("--request", required=True, help="Sanitized user request (stored in trace)")
    run.add_argument("--command", required=True, help="Shell command for Generator")
    run.add_argument("--max-iter", type=int, default=None)
    run.add_argument("--timeout", type=int, default=120)
    run.add_argument("--critic-json", type=Path, default=None, help="External Critic JSON file")
    run.add_argument("--critic-stdin", action="store_true", help="Read Critic JSON from stdin")
    run.add_argument(
        "--structural-critic-only",
        action="store_true",
        help="Use rule-based structural critic (CI/dry-run; not for production mutations)",
    )
    run.add_argument(
        "--critic-command",
        type=str,
        default=None,
        help="Isolated Critic command (separate process). Receives a sanitized critic-input "
        "JSON on stdin (operation_intent + generator_output + trace, NEVER the raw request) "
        "and must print critic JSON to stdout. Implements AGENTS.md §9 G/C isolation when "
        "--critic-json / --critic-stdin are absent.",
    )
    run.set_defaults(func=cmd_run)
    return p


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
