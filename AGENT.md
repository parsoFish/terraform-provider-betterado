# Unifier Agent Memory — INIT-2026-06-05-complete-release-definition

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Send-back round 2 (iteration 1 of the send-back)

**Feedback addressed:** The original demo (from the initial unifier pass, `da0f0cc5`) covered only WI-1 through WI-5. The send-back feedback (`INIT-2026-06-05-complete-release-definition.pr-feedback.md`) required the demo and PR description to be updated to cover WI-6, WI-7, and WI-8, which had been committed on the branch by the per-WI dev loops.

**What was done:**
- Confirmed all per-WI commits are on branch: WI-6 (`6a0e5d26`), WI-7 (`651c3b50` + `71db03b6`), WI-8 (`08a20287`, `b73c2d36`, `71db03b6`).
- Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → exit 0, all 3 packages ok.
- Ran `go test -tags all -count=1 -v -run TestReleaseDefinition_GatesTasks ./azuredevops/internal/service/release/` → PASS (WI-6 AC3).
- Ran `go test -tags all -count=1 -v -run TestReleaseDefinition_RoundTrip ./azuredevops/internal/service/release/` → PASS (WI-7 AC5, 4 subtests).
- Updated `demo/INIT-2026-06-05-complete-release-definition/demo.json` to include:
  - WI-6 checkpoint (gate tasks, TestReleaseDefinition_GatesTasks_ExpandFlatten)
  - WI-7 checkpoint (3 flatten bugs fixed, TestReleaseDefinition_RoundTrip 4 subtests, perpetual diff eliminated)
  - WI-8 checkpoint (TestAccReleaseDefinition_complete live ADO acceptance test, gates#>0, queue set, idempotency)
  - Updated summary (8 WIs), testEvidence (added WI-6/7/8 tests), impact bullets, filesChanged, usage_example
- Ran `forge demo render INIT-2026-06-05-complete-release-definition --dir .../demo/INIT-2026-06-05-complete-release-definition` → wrote DEMO.md + DEMO.html.
- Wrote `.forge/pr-description.md` with substantive Why/What/How covering all 8 WIs (no ## Demo section).
- Updated fix_plan.md: all 8 ACs ticked.

**forge demo render invocation note:** `forge demo render <id>` expects the demo.json at `<cwd>/demo/<id>/demo.json`. When running from outside the worktree root, use `--dir <absolute-path-to-demo-dir>` (the directory that *contains* demo.json, not the parent).

### Initial unifier pass (commit da0f0cc5)

- Covered WI-1 through WI-5 only (those were the only committed WIs at the time).
- Gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → exit 0.
- Demo and PR description written for the WI-1–5 feature surface.
- Per-WI dev loops for WI-6, WI-7, WI-8 then ran and committed their work on this branch.

## Notes for reflection

_(observations the reflector should capture into the brain)_

- The ado-demo SKILL.md contract (hardened 2026-06-05) is the definitive spec for what a live betterado demo requires: exhaustive HCL (every option non-default), apply → API GET → re-plan (No changes) → destroy. The unifier must update the demo after send-back WIs land — it cannot assume the initial demo covers newly committed WIs.
- `forge demo render` must be invoked from the worktree root OR with `--dir <absolute path to the dir containing demo.json>`. The bare `forge demo render <id>` resolves `demo/<id>/demo.json` relative to the process cwd.
- The three round-trip bugs (multipliers, spurious parallel_execution block, schedule_trigger branch_filter location) are a class of flatten error that offline gomock tests routinely miss because they don't exercise the ADO SDK wire format. The idempotency re-plan (`ExpectNonEmptyPlan: false`) in the acceptance test is the correct systematic catch.
