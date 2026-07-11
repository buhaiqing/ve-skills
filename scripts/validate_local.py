#!/usr/bin/env python3
"""Run the local validation suite that mirrors the CI quality gates."""

from __future__ import annotations

import argparse
import shlex
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class Step:
    name: str
    argv: tuple[str, ...]


def _check_file_integrity(root: Path) -> int:
    errors: list[str] = []
    for f in sorted(root.glob("ve-*/SKILL.md")):
        data = f.read_bytes()
        if b"\0" in data:
            errors.append(f"{f.relative_to(root)}: contains null bytes")
    if errors:
        for e in errors:
            print(f"  FAIL: {e}")
        return 1
    print("  OK: all SKILL.md files are clean UTF-8 text")
    return 0


def _check_required_sections(root: Path) -> int:
    HARD: set[str] = {"## Trigger & Scope", "## Quality Gate (GCL)"}
    SOFT: set[str] = {"### What This Skill Does", "## Operational Best Practices", "### Next Steps"}
    errors: list[str] = []
    warnings: list[str] = []

    for f in sorted(root.glob("ve-*/SKILL.md")):
        skill = f.parent.name
        # ve-skill-generator is a meta-skill, not a product-ops skill
        if skill == "ve-skill-generator":
            continue
        text = f.read_text(encoding="utf-8")

        # Some skills use "(Agent-Readable)" suffix on Trigger & Scope
        has_ts = "## Trigger & Scope" in text or "## Trigger & Scope (Agent-Readable)" in text
        has_shall = "### SHOULD Use This Skill When" in text
        has_shall_not = "### SHOULD NOT Use This Skill When" in text
        has_gcl = "## Quality Gate (GCL)" in text
        has_what = "### What This Skill Does" in text
        has_ops = "## Operational Best Practices" in text
        has_next = "### Next Steps" in text

        if not has_ts:
            errors.append(f"{skill}: missing ## Trigger & Scope")
        elif not has_shall or not has_shall_not:
            errors.append(f"{skill}: ## Trigger & Scope lacks SHOULD/SHOULD NOT subsections")
        if not has_gcl:
            errors.append(f"{skill}: missing ## Quality Gate (GCL)")
        if not has_what:
            errors.append(f"{skill}: missing ### What This Skill Does (IMPORTANT — MUST exist)")
        if not has_ops:
            errors.append(f"{skill}: missing ## Operational Best Practices (IMPORTANT — MUST exist)")
        if not has_next:
            warnings.append(f"{skill}: missing ### Next Steps")

    for e in errors:
        print(f"  FAIL: {e}")
    for w in warnings:
        print(f"  WARN: {w}")
    if errors:
        return 1
    print("  OK: all harness-critical sections present")
    return 0


def _check_error_taxonomy(root: Path) -> int:
    import re

    warnings: list[str] = []
    for f in sorted(root.glob("ve-*/SKILL.md")):
        skill = f.parent.name
        if skill == "ve-skill-generator":
            continue
        text = f.read_text(encoding="utf-8")

        if "## Error Taxonomy" not in text:
            warnings.append(f"{skill}: missing ## Error Taxonomy")
            continue

        severity_classifications = re.findall(r"^\|\s*`[^`]+`\s*\|\s*[^|]+?\|\s*[^|]*?\*\*(HALT|RETRY)\*\*", text, re.MULTILINE)

        if len(severity_classifications) < 10:
            warnings.append(f"{skill}: ## Error Taxonomy has only {len(severity_classifications)} codes, need ≥10")
        elif "HALT" not in severity_classifications:
            warnings.append(f"{skill}: ## Error Taxonomy missing HALT classification")
        elif "RETRY" not in severity_classifications:
            warnings.append(f"{skill}: ## Error Taxonomy missing RETRY classification")

    for w in warnings:
        print(f"  WARN: {w}")
    if warnings:
        print(f"  → {len(warnings)} error taxonomy issue(s) found (advisory)")
        return 0
    print("  OK: all skills have ## Error Taxonomy with ≥10 codes including HALT/RETRY")
    return 0


def _check_te1_hardcodes(root: Path) -> int:
    import re

    # Scans for user-choice version params in example bodies (not API spec versions)
    PATTERNS: list[tuple[str, str]] = [
        ("EngineVersion", r'"EngineVersion":\s*"\d+\.\d+"'),
        ("MongoVersion", r'"MongoVersion":\s*"\d+\.\d+"'),
        ("--MongoVersion", r'--MongoVersion\s+\d+\.\d+'),
        ("--Version", r'--Version\s+"\d+\.\d+"'),
        ("--TargetVersion", r'--TargetVersion\s+"\d+\.\d+"'),
    ]
    warnings: list[str] = []

    for glob_pat in ("ve-*/references/cli-usage.md", "ve-*/SKILL.md"):
        for f in sorted(root.glob(glob_pat)):
            text = f.read_text(encoding="utf-8")
            rel = f.relative_to(root)
            for label, pattern in PATTERNS:
                for m in re.finditer(pattern, text):
                    warnings.append(f"{rel}: TE-1 hardcoded {label} → {m.group()}")

    for w in warnings:
        print(f"  WARN: {w}")
    if not warnings:
        print("  OK: no hardcoded version numbers detected")
    else:
        print(f"  → {len(warnings)} TE-1 candidate(s) found (advisory)")
    return 0


def build_steps(python: str = sys.executable) -> list[Step]:
    return [
        Step("File integrity (null byte check)", (python, "-c", _inline_script(_check_file_integrity))),
        Step("Frontmatter validation", (python, "scripts/validate_skills_frontmatter.py")),
        Step("Required sections presence", (python, "-c", _inline_script(_check_required_sections))),
        Step("Error Taxonomy (≥10 codes, HALT/RETRY)", (python, "-c", _inline_script(_check_error_taxonomy))),
        Step("TE-1 hardcoded version scan", (python, "-c", _inline_script(_check_te1_hardcodes))),
        Step("Markdown local links", (python, "scripts/check_markdown_links.py")),
        Step(
            "GCL runner smoke test",
            (
                python,
                "scripts/gcl_runner.py",
                "run",
                "--skill",
                "ve-skill-generator",
                "--request",
                "CI smoke test",
                "--command",
                'echo {"Response":{"RequestId":"ci-smoke"}}',
                "--max-iter",
                "1",
                "--structural-critic-only",
            ),
        ),
        Step("GCL trace aggregate", (python, "scripts/gcl_trace_aggregate.py", "--since-hours", "168")),
        Step(
            "Script unit tests",
            (python, "-m", "unittest", "discover", "-s", "scripts", "-p", "*_test.py", "-v"),
        ),
        Step("GCL Tier-A conformance", (python, "scripts/check_gcl_conformance.py")),
        Step("Eval regression", (python, "scripts/check_eval_regression.py")),
    ]


def _inline_script(fn):
    import inspect, textwrap
    source = inspect.getsource(fn)
    lines = source.splitlines()
    body = textwrap.indent(textwrap.dedent("\n".join(lines[1:])), "    ")
    return (
        "import sys\n"
        "from pathlib import Path\n"
        "def main():\n"
        "    root = Path.cwd()\n"
        f"{body}\n"
        'sys.exit(main())\n'
    )


def run_step(root: Path, step: Step) -> int:
    print(f"\n==> {step.name}")
    print("$ " + shlex.join(step.argv))
    proc = subprocess.run(step.argv, cwd=root)
    return proc.returncode


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--list", action="store_true", help="Print commands without running them")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)
    root = args.root.resolve()
    steps = build_steps()

    if args.list:
        for step in steps:
            print(f"{step.name}: {shlex.join(step.argv)}")
        return 0

    for step in steps:
        rc = run_step(root, step)
        if rc != 0:
            print(f"\nFAILED: {step.name} exited with {rc}", file=sys.stderr)
            return rc

    print("\nOK: local validation suite passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
