---
title: Split acceptance test WIs — WI-A writes, WI-B runs live
description: For acceptance tests requiring live ADO credentials, split into two WIs: WI-A authors the test function (gate = go build), WI-B runs it live (gate = go test -tags all -run ...). Prevents WI-B gate from firing before any code exists.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-11T12:30:00Z
updated_at: 2026-06-11T12:30:00Z
related_themes:
  - 2026-06-11-live-acc-tests-as-wi-gate
  - 2026-06-11-live-acc-wi-gate-errors-before-ralph-runs
---

# Split acceptance test WIs — WI-A writes, WI-B runs live

## Pattern

When a new acceptance test must be both authored AND run live against ADO, decompose into two WIs:

**WI-A: Write the test function**
- Gate: `go build ./azuredevops/internal/acceptancetests/...`
- Agent authors `TestAcc<Resource>_<Case>` + HCL fixture
- No credentials needed
- Gate passes when the code compiles

**WI-B: Run the test live**
- Gate: `source secrets.env && go test -tags all -count=1 -run TestAcc<Resource>_<Case> ./azuredevops/internal/acceptancetests/`
- No new code authoring needed
- Requires live env with `TF_ACC=1` + `secrets.env`
- Gate passes when the live ADO round-trip succeeds

## Why

If a single WI carries both responsibilities and the gate is the live command, `gate.errored` fires at iteration 0 when creds are absent — ralph runs 0 iterations and the test is never written. The unifier must rescue (as in this cycle).

The split pattern avoids this: WI-A always passes in any env; WI-B is advisory (can be deferred or re-run when creds are available).

## Evidence from this cycle

This exact failure occurred with WI-2 (`TestAccReleaseDefinition_approvalsAndGates`) in INIT-2026-06-08-release-definition-approval-options-gates-comple. The unifier authored the test in UWI-1, but UWI-2 then wedged trying to verify credentials (froze for 33 hours). The split would have avoided both the gate-error AND the unifier wedge.

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple/events.jsonl` — `EV_mq55dnrt_p26ki973` (gate.errored, WI-2)
- `brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple.md`
