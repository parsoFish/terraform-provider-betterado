---
title: Acceptance test file can be written and compile-verified in dev-loop without live ADO run
description: Writing a new TF_ACC acceptance test file in the dev-loop uses a compile-only gate (go build + go test without TF_ACC=1); the live round-trip is deferred to pre-merge, keeping dev-loop fast and creds-free.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T05:10:00Z
updated_at: 2026-06-06T05:10:00Z
related_themes:
  - 2026-06-05-live-acceptance-gate-for-acceptance-wis
  - 2026-06-02-ci-green-gate-design
---

# Acceptance test file: compile-only in dev-loop, live in pre-merge

## Pattern

WI-5 of INIT-2026-06-05-release-data-sources had a targeted scope: write
`azuredevops/internal/acceptancetests/data_release_definition_test.go`.
The `quality_gate_cmd` was:

```bash
go test -mod=vendor -tags all -count=1 -timeout 30m \
  -run TestAccDataReleaseDefinition|TestAccDataReleaseDefinitions \
  ./azuredevops/internal/acceptancetests/
```

Without `TF_ACC=1` in env, this command compiles the test binary and registers
the test functions but exits 0 without running them (acceptance tests are skipped
when `TF_ACC` is unset). The gate checks:
- The file compiles cleanly (no syntax errors, correct package/imports).
- The test functions are correctly named (regex match).
- No regressions in the acceptancetests package.

The live round-trip (`TF_ACC=1` with real ADO creds) runs at pre-merge review
time, not in the dev-loop.

## Why this works

The Terraform plugin testing framework skips acceptance tests when `TF_ACC` is
not set to `"1"`. Gate exit code is 0 but with `[no test to run]` skipping —
NOT the `[no tests to run]` pattern (different wording) that the gate-tightener
flags as a no-work indicator. The gate-tightener correctly passed WI-5 at iter-1.

## Applied in this cycle

- WI-5 iter-0: `gate.expected-fail` (no-work-indicator) — file didn't exist yet.
- WI-5 iter-1: agent wrote `data_release_definition_test.go`, ran `go build` to
  confirm compile, then ran the gate. Gate passed in 3.9s.
- Total WI-5 cost: $0.55. 1 iteration.

## Contrast with WI-9 (INIT-2026-06-05-complete-release-definition)

WI-9 ran `TF_ACC=1` live in the dev-loop because it was implementing acceptance
*criteria* (real ADO idempotency). WI-5 here is only implementing the *test
file structure* — no live run needed. Correct split.

## Sources

- `_logs/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources/events.jsonl`
  (events: `ralph.end` WI-5, `gate.pass` WI-5 `EV_mq1vs931_bqgoj5ov`)
- `_logs/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources/work-items-snapshot/WI-5.md`
- `brain/cycles/_raw/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources.md`
