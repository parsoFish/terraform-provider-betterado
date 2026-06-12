---
title: Combined write+run WI for acceptance tests when live creds available
description: When secrets.env is present, a single WI that both authors a new acceptance test and runs it live completes in 1 iteration — no split write/run decomposition needed. The split pattern is only required when creds may be absent.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-12T12:47:08Z
updated_at: 2026-06-12T12:47:08Z
related_themes:
  - 2026-06-11-acceptance-test-wi-split-write-then-run
  - 2026-06-11-live-acc-tests-as-wi-gate
---

# Combined write+run WI for acceptance tests when live creds available

## Observation

All 3 WIs in INIT-2026-06-08-release-acceptance-test-fixes combined test authoring and live execution:

- WI-1: wrote `TestAccReleaseDefinition_updateAddEnvironment` + ran it live → 1 iteration
- WI-2: wrote `TestAccReleaseDefinition_import` + ran it live → 1 iteration
- WI-3: wrote `TestAccReleaseDefinition_completeWithNewFields` + ran it live → 1 iteration

Each gate was `go test -tags all -run <NewTestName> ./azuredevops/internal/acceptancetests/`. The test did not exist before the WI, so the gate correctly expected-failed at iteration 0. The dev authored the test and the live ADO run passed first try.

## When this works

- `secrets.env` is present and valid (`TF_ACC=1`, `AZDO_ORG_SERVICE_URL`, `AZDO_PERSONAL_ACCESS_TOKEN` all set)
- The gate targets a NEW test name (expected-fail before work; gate-tightening satisfied)
- The implementation being tested is already correct (the schema changes landed in prior initiatives)

## When to use the split pattern instead

Use the split (`WI-A: go build`, `WI-B: go test live`) from `2026-06-11-acceptance-test-wi-split-write-then-run` when:
- Creds may not be present at run time (gate would error at iteration 0 before test is written)
- The test is complex enough that the author WI and the live run WI benefit from separate focus

## Context

The prior cycle (INIT-2026-06-08-release-definition-approval-options-gates-comple) hit a gate-error-before-ralph-runs problem because credentials were absent. This cycle ran with full creds and the combined pattern worked cleanly.

## Sources

- `_logs/2026-06-12T12-19-27_INIT-2026-06-08-release-acceptance-test-fixes/events.jsonl` — ralph.end WI-1 at 12:32:32, WI-2 at 12:35:47, WI-3 at 12:41:29 (all `status: complete, iterations: 1`)
- `brain/cycles/_raw/2026-06-12T12-19-27_INIT-2026-06-08-release-acceptance-test-fixes.md`
