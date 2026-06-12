---
title: golangci-lint only in ci_gate_cmd — lint errors surface too late
description: The dev-loop quality gate runs only `go test`; golangci-lint runs only in ci_gate_cmd. Agent-introduced lint errors (e.g. unchecked errcheck) abort the cycle at CI gate, forcing a full retry.
category: antipattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## What happened

First run of INIT-2026-06-08-release-data-sources-completion: unifier wrote `doc_audit_test.go` with `filepath.Rel` return value discarded:

```go
rel, _ := filepath.Rel(repoRoot, docPath)
```

The dev-loop quality gate (`go test -tags all -run ... ./azuredevops/internal/service/release/`) passed. But `ci_gate_cmd` is `make test && golangci-lint run ./... && make terrafmt-check`. `golangci-lint` flagged `errcheck` on line 56. Cycle aborted as terminal (`failure_mode: terminal`). Full retry required (lost ~$0.80 unifier cost).

## Root cause

Dev-loop gate is scoped to the package under test; `golangci-lint` is only in the full CI gate run after dev-close. Linting runs too late.

## Mitigations

1. **PM acceptance criterion**: add "passes `golangci-lint run ./azuredevops/internal/service/release/...`" as an AC in any WI that creates new `.go` files in this project.
2. **AGENT.md note**: document that `errcheck` is enforced; always handle error return from `filepath.Rel`, `os.*`, `json.*`.
3. Optionally add a fast `golangci-lint run --fast <changed-packages>` to the dev-loop gate command.

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion/events.jsonl` — EV_mq55twvx_q1m71sp0 (ci-gate error), EV_mq55twvx_ci3zuddp (ci-gate-failed message)
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion.md`
