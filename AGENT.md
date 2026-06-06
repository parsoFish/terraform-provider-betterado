# Unifier Agent Memory — INIT-2026-06-05-release-folder

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (unifier)

- Read AGENT.md, fix_plan.md, WI-1.md, initiative manifest.
- Confirmed per-WI dev commits are on branch (`adfa7576 feat: implement betterado_release_folder resource` + follow-up build-tag fix).
- **Ran quality gate** (`go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`): **GREEN** — all 3 packages pass (release 0.019s, taskagent 0.008s, taskagent/validate 0.004s).
- Updated `demo.json` `diffStat` field to match current `git diff --stat main...HEAD` (9 files, 823 insertions, 56 deletions — now includes demo artefacts).
- Ticked AC1 in `fix_plan.md` as proven green.
- Re-ran `forge demo render INIT-2026-06-05-release-folder --dir ...` — wrote updated `DEMO.md` + `DEMO.html`.
- Committed and pushed: `feat(INIT-2026-06-05-release-folder): unify and demo`.

### Notes for reflection

- `forge demo render` requires `--dir <absolute-path>` when the forge CLI is invoked from a different CWD than the worktree. Document this in `forge demo` skill.
- Build tag on the production resource file (`resource_release_folder.go`) must NOT carry a test-only build tag — this was fixed in commit `d2f84860`.

## Notes for reflection

_(observations the reflector should capture into the brain)_
