---
title: Acceptance WIs must use live TF_ACC gate, not offline unit test, as quality_gate_cmd
description: Work items whose acceptance criteria are inherently live (idempotency, real ADO field persistence, gate task execution) must use the TF_ACC acceptance test as the quality_gate_cmd — offline unit tests cannot verify these properties.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T00:00:00Z
updated_at: 2026-06-06T00:00:00Z
related_themes:
  - 2026-06-02-ci-green-gate-design
  - 2026-06-05-ado-silent-field-discard-idempotency
---

# Acceptance WIs must use live TF_ACC gate as quality_gate_cmd

## Pattern

When a WI's acceptance criteria involve live ADO behaviour (idempotency, real field persistence, gate task execution, agent pool resolution), the `quality_gate_cmd` MUST be the live acceptance test:

```bash
go test -tags all -count=1 -run TestAccReleaseDefinition_complete -timeout 30m ./azuredevops/internal/acceptancetests/...
```

**Do not** use the offline unit gate (`./azuredevops/internal/service/release/...`) for an acceptance WI. It will pass-green while the live resource remains non-idempotent or incomplete.

## The two-gate model

This project uses a **two-gate model** declared in the initiative manifest:

1. **Offline unit gate** — `go test -tags all -count=1 -run <Prefix> ./azuredevops/internal/service/release/...`
   - Creds-free, fast (~1s), default dev-loop gate for schema WIs (expand/flatten logic, gomock unit tests).
2. **Live acceptance gate** — `TF_ACC=1 go test -run TestAccReleaseDefinition_complete -timeout 30m ./azuredevops/internal/acceptancetests/...`
   - Requires `secrets.env` (PAT). Slow (~28s). Mandatory for acceptance WIs and final-cleanup WIs like WI-9.
   - Always includes a `PlanOnly: true` step with `ExpectNonEmptyPlan: false` to prove idempotency.

## Observed in WI-9

WI-9 used the live gate. Ralph ran it 3× during iteration 1 (at 15:06, 15:11, 15:20). Each pass was ~28s. Total cycle cost $8.34 for a focused cleanup — reasonable for a live-gated WI.

The offline unit suite still ran at dev-loop baseline-green check and was verified passing by the unifier (`go test ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` exits 0, 22 tests).

## Sources

- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/events.jsonl` (gate.pass events: 15:06:35, 15:11:45, 15:20:38)
- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/work-items-snapshot/WI-9.md` (quality_gate_cmd)
- `brain/cycles/_raw/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition.md`
