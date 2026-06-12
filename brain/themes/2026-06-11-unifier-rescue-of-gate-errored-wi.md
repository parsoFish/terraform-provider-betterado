---
title: Unifier rescues gate-errored WI by authoring missing acceptance test
description: When ralph exits with 0 iterations due to gate.errored, the unifier can detect the missing implementation and author it in a UWI — effectively writing what ralph never started.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-11T12:30:00Z
updated_at: 2026-06-11T12:30:00Z
related_themes:
  - 2026-06-11-live-acc-wi-gate-errors-before-ralph-runs
---

# Unifier rescues gate-errored WI by authoring missing acceptance test

## Observation

After WI-2 gate-errored (ralph ran 0 iterations), the unifier (UWI-1) was given the branch. It:

1. Ran `git diff --stat main...HEAD` — saw WI-1's test file but no acceptance test for WI-2.
2. Grepped for `TestAccReleaseDefinition_approvalsAndGates` — confirmed absence.
3. Inspected `SharedReleaseFixture`, `fixture.WorkItemQueryID`, HCL conventions.
4. Wrote 124 lines: `TestAccReleaseDefinition_approvalsAndGates` + `hclReleaseDefinitionApprovalsAndGates`.
5. Ran `go build ./azuredevops/internal/acceptancetests/...` → clean.
6. Committed and pushed.

The acceptance test function is now on the branch and can be run live when `secrets.env` is available.

## Why it works

The unifier's UWI brief gives it visibility into ALL WI statuses, including failed/gate-errored. It can triage: "this WI failed because the test function doesn't exist yet, not because credentials were missing — I can write the function now and defer the live run."

## Constraint

The unifier cannot run the live acceptance test either (same env constraint). It can only author code that compiles and skips cleanly without TF_ACC. Live verification remains deferred to a subsequent cycle or manual operator run.

## When to rely on this

Only when the missing implementation is straightforward enough for the unifier to infer from existing patterns (SharedReleaseFixture, prior test structure). Complex new ADO behaviours may require more context than the unifier has.

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple/events.jsonl` — `EV_mq55l1fp_zdjvgurv` (unifier iteration 1, UWI-1 complete; `last_assistant_text` confirms test authored + committed `be010e5c`)
- `brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-definition-approval-options-gates-comple.md`
