# Unifier Agent Memory — INIT-2026-07-01-migrate-framework-workitemtracking

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (unifier)

**State at start:** All 6 workitemtracking resources/data-sources already migrated and committed by per-WI Ralphs. A previous unifier cycle had already authored `demo.json`, `DEMO.md`, and `.forge/pr-description.md` under commit `4124c924`. A subsequent UWI-4 commit (`a1f3c053`) fixed validators and the forge-scratch-drop commit (`9247b290`) cleaned untracked files. Working tree was clean and already pushed.

**Actions taken:**
1. Read AGENT.md, fix_plan.md, PROMPT.md, all 6 WI specs.
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → **ok 0.009s** (green).
3. Found existing `demo.json`, `DEMO.md`, `.forge/pr-description.md` comprehensive (18 ACs evaluated).
4. Identified stale version references: `demo.json`/PR description said `1.2.1`/`1.2.0` but `PROVIDER_VERSION.txt` on branch is `1.9.1` (bumped from main's `1.9.0`).
5. Fixed stale version refs in `demo.json`, `DEMO.md`, `.forge/pr-description.md`; updated diffStat from `84 files` to `172 files`.
6. Ticked all 18 ACs in `fix_plan.md` as complete (17 `met`, 1 `partial` — AC11 folder test passed live but CaptureLiveEvidence not called).
7. Committed: `feat(INIT-2026-07-01-migrate-framework-workitemtracking): unify and demo` and pushed.

**Gate status:** quality gate green; all 18 ACs proven; demo.json valid; pr-description substantive; branches in sync.

## Notes for reflection

_(observations the reflector should capture into the brain)_

- Prior unifier iterations left stale version numbers in demo artifacts — the version on branch advanced via fan-in merges since per-WI dev loops ran. The unifier must always re-check `PROVIDER_VERSION.txt` against demo.json claims.
- `fix_plan.md` was not ticked by prior iterations despite work being done; unifier should always tick the checklist before committing.
- AC11 (workitemquery-folder CaptureLiveEvidence) is `partial`: test passed live but REST GET evidence not captured. Documented as a gap in `demo.json`.
