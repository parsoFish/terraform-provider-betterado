# AGENT.md — INIT-2026-06-05-complete-release-definition

## Iteration 1 (unifier) — completed

### What was done

1. **Read state**: AGENT.md and fix_plan.md were absent (first unifier iteration). Read all 5 WI specs and the initiative manifest to understand scope.

2. **Quality gate**: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — **green** (all packages ok, 20 release tests + taskagent suite).

3. **Demo authored**: `demo/INIT-2026-06-05-complete-release-definition/demo.json` written with:
   - 5 harness checkpoints (one per WI), each with beforeNote/afterNote describing the behavioural delta
   - Accurate metrics scraped from the live gate run (11→20 release tests, +82%)
   - Full `testEvidence[]` table (21 rows)
   - `usage_example` (HCL showing all 5 new schema features)
   - `impact[]` (5 bullets)
   - `acceptanceCriteria[]`, `summary`, `filesChanged`
   - Fixed `deltaPct` type bug (must be `number | null`, not string) before render succeeded

4. **Demo rendered**: `forge demo render ... --dir <absolute-path>` (forge's `process.chdir(FORGE_ROOT)` means the `--dir` flag must be an absolute path).
   - `DEMO.md` + `DEMO.html` written and committed.

5. **PR description**: `.forge/pr-description.md` written (gitignored — read by orchestrator for `gh pr create --body-file`). Substantive Why/What/How sections. No `## Demo` section.

6. **Committed**: `feat(INIT-2026-06-05-complete-release-definition): unify and demo` (3 files: demo.json, DEMO.md, DEMO.html).

7. **Pushed**: `origin/forge/INIT-2026-06-05-complete-release-definition` == local HEAD.

### Key findings

- `forge demo render` requires `--dir <absolute-path>` when called from a worktree, because `forge` changes its cwd to `FORGE_ROOT` (`/home/parso/forge`) at startup.
- `deltaPct` in `HarnessMetricRow` must be `number | null` — string values cause a `toFixed is not a function` error in the renderer.
- All 5 WIs were complete and committed before the unifier ran; no code changes needed.

### Gate status

- `initiative_gate`: ✅ green
- `demo_runs_clean`: ✅ (harness shape — demo.command not applicable; `forge demo render` exited 0)
- `pr_self_contained`: ✅ demo.json validates; .forge/pr-description.md has substantive Why/What/How; no ## Demo section
- `branches_in_sync`: ✅ origin == local HEAD; main == merge-base
