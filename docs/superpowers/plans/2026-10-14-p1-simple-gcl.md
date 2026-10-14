# P1 Simple-GCL Plan

## Milestones

| M | Task | DoD | Verify |
|---|---|---|---|
| M1 | gcl-spec.md degraded-mode section | section present w/ disclosure checklist | grep 'degraded' gcl-spec |
| M2 | duprefs detector package + `vet check duprefs` | package builds, command runs | go build/test |
| M3 | One SKILL.md compression demo (`ve-ark-ops`) | SKILL.md <800 lines, refs created, content traceable | vet validate passes |
| M4 | end-to-end run | all checks pass | make check |

## Verification Command Matrix
- cd cmd/vet && go build ./... && go vet ./...
- go test ./internal/check/duprefs/...
- ./cmd/vet/vet-bin validate --root .
- wc -l ve-ark-ops/SKILL.md
