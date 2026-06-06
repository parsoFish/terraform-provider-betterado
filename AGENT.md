# Unifier Agent Memory — INIT-2026-06-05-release-folder

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (2026-06-06)

**Goal**: Initial prep — confirm WI-1 committed, run quality gate, author demo + PR description, commit + push.

**What I did**:
1. Read AGENT.md, fix_plan.md, PROMPT.md, and initiative manifest.
2. Verified `git log --oneline main...HEAD` shows 2 commits: the safety-net WIP commit and the actual WI-1 feature commit (`adfa7576 feat: implement betterado_release_folder resource with CRUD and 5 unit tests`).
3. Confirmed `git diff --stat main...HEAD`: 6 files changed, 405 insertions, 65 deletions — resource impl, test, provider registration, HCL example, AGENT.md, fix_plan.md.
4. Wrote skeleton `demo/INIT-2026-06-05-release-folder/demo.json` and `.forge/pr-description.md` immediately (committed as `wip: unifier skeleton`).
5. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — **GREEN** (3 packages ok).
6. Ran `go test -tags all -count=1 -v -run TestReleaseFolder ./azuredevops/internal/service/release/...` — all 5 tests PASS: TestReleaseFolderExpand, TestReleaseFolderCreate_Error, TestReleaseFolderRead_404ClearsID, TestReleaseFolderUpdate_Args, TestReleaseFolderDelete_Error.
7. Updated demo.json with real test names + accurate gate output.
8. Fixed `apiDiff[].before: null` → `"resource did not exist"` (forge demo-html.ts calls `.trim()` on before/after strings, null caused TypeError).
9. Ran `forge demo render INIT-2026-06-05-release-folder --dir /home/parso/forge/_worktrees/INIT-2026-06-05-release-folder/demo/INIT-2026-06-05-release-folder` — **SUCCESS** (DEMO.md + DEMO.html written). Note: forge chdir's to FORGE_ROOT at startup, so absolute path is required for `--dir`.
10. Committed `feat(INIT-2026-06-05-release-folder): unify and demo` and pushed.
11. Ticked AC1 in fix_plan.md as verified.

**Status**: All 4 orchestrator gates should pass:
- `initiative_gate`: ✅ quality gate green
- `demo_runs_clean`: ✅ `forge demo render` exits 0 (shape is "harness", no demo.command required)
- `pr_self_contained`: ✅ demo.json exists + validates, pr-description.md has Why/What/How, no `## Demo` section
- `branches_in_sync`: ✅ pushed to origin

**Key learning**: `forge demo render <id>` requires `--dir <absolute-path-to-demo-dir>` because forge chdir's to FORGE_ROOT. Also, `apiDiff[].before` must be a string (not null) — the HTML renderer calls `.trim()` on it.

## Notes for reflection

- The `forge demo render` CLI requires an absolute path when using `--dir`; relative paths resolve against FORGE_ROOT (not the worktree). Consider documenting this in SKILL.md or adding a guard in demo-model.ts.
- Null values in `apiDiff[].before`/`after` cause a TypeError in demo-html.ts (`.trim()` on null). The TypeScript type allows `undefined` (skips the field) but the renderer doesn't guard against `null`. A type guard or optional-chaining fix in demo-html.ts would prevent this footgun.
