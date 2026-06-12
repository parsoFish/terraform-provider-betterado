---
title: Live-acceptance WI gate errors before ralph runs if env creds absent
description: When a WI's quality_gate_cmd is a live acceptance test and TF_ACC/creds are not in the cycle env, gate.errored fires at iteration 0 — ralph runs 0 iterations and the WI fails before any code is written; unifier must rescue.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-11T12:30:00Z
updated_at: 2026-06-11T12:30:00Z
related_themes:
  - 2026-06-11-live-acc-tests-as-wi-gate
  - 2026-06-08-live-env-fast-fail-guard-confirmed
---

# Live-acceptance WI gate errors before ralph runs if env creds absent

## Observation

WI-2 in INIT-2026-06-08-release-definition-approval-options-gates-comple had gate:

```
go test -tags all -run TestAccReleaseDefinition_approvalsAndGates ./azuredevops/internal/acceptancetests/
```

Gate exit code: `-5` (`live-env-missing`). Reason: `TF_ACC`, `AZDO_ORG_SERVICE_URL`, `AZDO_PERSONAL_ACCESS_TOKEN` not in cycle env.

Ralph ran **0 iterations** — it never even read the WI. Status: `failed`, `stop_reason: gate-errored`.

Result: the acceptance test function (`TestAccReleaseDefinition_approvalsAndGates`) did not exist on the branch. The unifier (UWI-1) had to detect the absence and write 124 lines of test code + HCL fixture.

## Root cause

In the prior data-sources cycle that succeeded (INIT-2026-06-08-release-data-sources-completion), WI-2's live test ran because WI-1 had already shipped the implementation and `secrets.env` was loaded in the ralph context. Here, WI-2 was the *sole* WI responsible for authoring the acceptance test — so ralph's very first act (gate pre-check) killed the WI before any work began.

## Pattern

A WI whose gate requires live ADO credentials MUST either:
1. Be launched in a cycle env with `secrets.env` sourced (i.e. `forge serve` with creds exported), OR
2. Have its gate split: WI-A writes the test (gate = `go build`); WI-B runs it live (gate = `go test -tags all -run ...`).

If neither condition holds, ralph produces zero work and the unifier must rescue.

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple/events.jsonl` — `EV_mq55dnrt_p26ki973` (gate.errored, live-env-missing), `EV_mq55dnru_h5ydppfq` (ralph.end status:failed 0 iterations)
- `brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple.md`
