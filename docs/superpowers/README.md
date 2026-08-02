# Superpowers — Spec & Plan Archive

> **Entry point for AI agents**: Read [ARCHITECTURE.md](./ARCHITECTURE.md) first. It contains the compressed knowledge artifact: current architecture, contracts, state, and pending work — all in one file.

## Structure

| Path | Status | Content |
|------|--------|---------|
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | **Authoritative** | Compressed knowledge: architecture, contracts, ADRs, current state, pending work |
| `l2-to-l3/` | Historical (DONE) | L2→L3 evolution specs/plans (T01-T08). All implemented. |
| `l3-to-l4/` | Historical (DONE) | L3→L4 evolution specs/plans (T09-T12). All implemented. |
| `specs/` | Historical (DONE) | Phase 1/2/3, P0, Wave A/B designs. All implemented. |
| `plans/` | Historical (DONE) | Implementation plans for above. All executed. |
| `rubrics/` | Reference | GCL quality gate rubrics. |
| `plans/golang-migration/` | Historical (DONE) | Python→Go migration. Complete. |

## When to Read Historical Specs

Only read historical specs/plans when:
- You need implementation details for a specific feature not covered in ARCHITECTURE.md
- You're debugging and need to understand why a contract was designed a certain way
- You're working on Wave C (pending) and need the original design intent

For all other purposes, ARCHITECTURE.md is sufficient.

## Wave C (Pending Work)

See ARCHITECTURE.md §5 "Pending (Wave C)" for the list of unimplemented work. No specs/plans exist yet for Wave C — they should be written when work begins.
