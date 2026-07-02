# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-02) — COMPLETE

- Read WI-1 spec: documentation-only WI, no code changes. Quality gate is `TestProvider_HasChildResources`.
- Explored source files:
  - `azuredevops/internal/service/policy/branch/`: 7 branch policy resources + common.go
  - `azuredevops/internal/service/policy/repository/`: 7 repo policy resources + common.go
  - `azuredevops/internal/service/approvalsandchecks/`: 6 check resources + common_check.go
- Created `docs/policy-gap-matrix.md`: covers all 9 branch policy types + 7 repo policy types with per-field Coverage columns.
- Created `docs/approvalsandchecks-gap-matrix.md`: covers all 8 check types with per-field Coverage columns.
- Ran `go test -tags all -run TestProvider_HasChildResources ./azuredevops/` — PASS (0.005s).
- Committed as `docs: add policy and approvalsandchecks gap matrices` (bc358c12).

## What worked

- Read source files directly for ground-truth field mapping (common.go for policy type UUIDs, resource files for settings schemas).
- Followed `docs/release-definition-gap-matrix.md` as the format template exactly.
- Quality gate is a structural provider test; docs-only changes pass trivially.

## What didn't work

_(none — completed in iteration 0)_

## Open questions

_(none)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
