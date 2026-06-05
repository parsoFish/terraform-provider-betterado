# Unifier Agent Memory — INIT-2026-06-05-complete-release-definition

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (2026-06-06) — unifier pass after WI-9

**Context:** Previous unifier pass (feat(INIT-2026-06-05-complete-release-definition): unify and demo) was authored after WI-8 but before WI-9. WI-9 added three final cleanup fixes: schedule_trigger.branch_filter removal, agent_specification in acceptance test, and betterado_workitemquery for real gate queryId.

**What I did:**
1. Read AGENT.md and fix_plan.md (both were stubs with no prior unifier iteration notes).
2. Read WI-9.md to understand the scope of changes since the last unifier pass.
3. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → **exits 0**, 22 top-level test functions pass.
4. Confirmed WI-9 changes via `git log` and `git show`: ce11e6ba (fix(WI-9)), 79e068e6 (chore), cd664939 (forge-autocommit safety-net).
5. Updated `demo/INIT-2026-06-05-complete-release-definition/demo.json`:
   - Updated title/essence to reflect 9 WIs (was 8).
   - Updated WI-7 checkpoint afterNote to reflect branch_filter removed from schedule_trigger schema entirely (WI-9 final resolution).
   - Updated WI-8 checkpoint to WI-8+WI-9 covering agent_specification, betterado_workitemquery queryId, and full idempotency.
   - Updated test count: 22 top-level test functions (was 26 total including subtests — clarified counting method).
   - Updated acceptanceCriteria and summary bullets to include WI-9 ACs.
   - Updated testEvidence rows for Triggers_ScheduleOnly, Triggers_ExpandFlatten, RoundTrip subtests to note WI-9 updates.
   - Updated usage_example with agent_specification and betterado_workitemquery.
   - Updated impact bullets: added agent_specification and gate queryId bullets.
   - Updated filesChanged notes to reflect WI-9 additions.
6. Updated `.forge/pr-description.md`:
   - Added WI-9 section to `## What`.
   - Updated `## Why` to include agent_specification and empty queryId as gaps.
   - Updated `## How` to include WI-9 file changes.
   - Updated summary to 22 top-level tests.
7. Ran `forge demo render INIT-2026-06-05-complete-release-definition --dir <worktree>/demo/INIT-...` → DEMO.md and DEMO.html written.
8. Updated fix_plan.md: ticked AC1/AC2/AC3 with proof notes; AC4 still pending (requires live TF_ACC=1).
9. Updated AGENT.md (this file).
10. Committed as `feat(INIT-2026-06-05-complete-release-definition): unify and demo` and pushed.

**Gate status:**
- `initiative_gate`: PASS (go test exits 0)
- `demo_runs_clean`: N/A for harness shape (no demo.command)
- `pr_self_contained`: demo.json exists and renders; pr-description.md has substantive Why/What/How
- `branches_in_sync`: will be satisfied after push

**Forge demo render invocation note:** `forge demo render <id>` looks for `demo/<id>/demo.json` relative to the current working directory. Must be run with `--dir <absolute-path-to-demo-dir>` when cwd is not the project root, OR run from the worktree root where `demo/` exists.

## Notes for reflection

- WI-9 was a "final cleanup" WI added after the first unifier pass; the unifier correctly identifies delta between last unifier commit and HEAD, updates demo.json and pr-description.md, and re-renders.
- The `forge demo render` command expects `demo/<id>/demo.json` relative to cwd; needs `--dir` flag if invoked from outside the worktree with `demo/` at root.
- Test count disambiguation: there are 22 top-level test *functions* in the release package, but more individual sub-tests (RoundTrip has 4, ParallelExecution has 3, AgentlessPhase has 3). Future demos should be explicit about top-level vs. total.
