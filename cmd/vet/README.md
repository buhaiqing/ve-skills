# vet — ve-skills unified CLI

Single statically-linked Go binary that consolidates the repository's Python
verification/CI scripts (`scripts/*.py`) so engineers and AI agents can run the
checks without a Python interpreter.

## Install / Update (one-shot, always latest)

```bash
curl -fsSL https://raw.githubusercontent.com/buhaiqing/ve-skills/main/cmd/vet/install.sh | bash
```

Re-running the script updates to the newest release. The script auto-detects
your OS/arch and pulls the matching artifact from the latest GitHub Release.

## Usage

```
vet version                                        # print version
vet check <frontmatter|aiops|assessment|gcl|links|eval> [--root DIR] [--json]
vet validate [--root DIR] [--list]
vet gcl <run|gate|trace> [flags]
```

| Subcommand | Replaces |
|------------|----------|
| `vet check frontmatter` | `scripts/validate_skills_frontmatter.py` |
| `vet check aiops` | `scripts/check_aiops_coverage.py` |
| `vet check assessment` | `scripts/validate_product_assessment.py` |
| `vet check gcl` | `scripts/check_gcl_conformance.py` |
| `vet check links` | `scripts/check_markdown_links.py` |
| `vet check eval` | `scripts/check_eval_regression.py` |
| `vet validate` | `scripts/validate_local.py` |
| `vet gcl run` | `scripts/gcl_runner.py` |
| `vet gcl gate` | `scripts/gcl_ci_gate.py` |
| `vet gcl trace` | `scripts/gcl_trace_aggregate.py` |

## Build from source

```bash
cd cmd/vet
go build -o vet .
```

## Release

Tagged `vet/vX.Y.Z` triggers GitHub Actions (`.github/workflows/release.yml`),
which uses GoReleaser to cross-compile for linux/darwin/windows on amd64/arm64
and publishes the assets to the GitHub Release.
