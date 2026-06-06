---
title: ADO REST silently discards unknown fields — live round-trip required to detect stale schema
description: ADO REST 7.x silently accepts and discards unknown fields on PUT, so unit mocks pass while the live provider remains non-idempotent; only TF_ACC plan with ExpectNonEmptyPlan:false reliably detects the drift.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-06T00:00:00Z
updated_at: 2026-06-06T00:00:00Z
related_themes:
  - 2026-05-31-release-definition-unit-test-substrate
  - 2026-06-02-ci-green-gate-design
---

# ADO REST silently discards unknown fields — live round-trip required to detect stale schema

## Antipattern

ADO REST APIs silently accept fields that have no effect (e.g. `BranchFilters` on a time-based schedule trigger, or a comma-joined string vs. a list for `parallel_execution.multipliers`). Gomock unit tests pass because the mock returns whatever the fixture says — no ADO round-trip occurs. The stale schema field lives undetected until `terraform plan` (with a real ADO apply) reports a perpetual diff.

## Observed cases in this project

| Field | Why ADO ignores it | Symptom |
|---|---|---|
| `schedule_trigger.branch_filter` | Classic schedule triggers are time-based; no branch filter concept in ADO | Perpetual diff on every `terraform plan` after apply |
| `parallel_execution.multipliers` | ADO stores comma-joined `*string`; expand was sending `[]interface{}` | Perpetual diff on every `terraform plan` |
| `parallel_execution` block for non-parallel phases | Flatten emitted a spurious empty block for `parallelExecutionType: "none"` | Perpetual diff |

All three were discovered only when `TestAccReleaseDefinition_complete` ran with `ExpectNonEmptyPlan: false` (WI-7, WI-9).

## Rule

**For every new schema field:**
1. Confirm the ADO REST docs show the field is used by the relevant API endpoint.
2. Add a `TestReleaseDefinition_RoundTrip` subtest that goes through expand → JSON-marshal → JSON-unmarshal → flatten and asserts equality.
3. For acceptance tests, always include a `PlanOnly: true` step with `ExpectNonEmptyPlan: false` to gate against silent discard.

Do not assume a field is live just because gomock unit tests pass.

## Sources

- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/events.jsonl` (events: gate.pass WI-9 live idempotency)
- `_logs/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition/pr-description.md` (WI-7 and WI-9 sections)
- `brain/cycles/_raw/2026-06-05T15-06-02_INIT-2026-06-05-complete-release-definition.md`
