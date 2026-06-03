# Unifier Agent Memory — INIT-2026-06-04-release-folder

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

### Iteration 1 (unifier initial prep)

- Read AGENT.md, fix_plan.md, PROMPT.md, initiative manifest.
- Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → **GREEN** (release ok, taskagent ok, taskagent/validate ok).
- Verified WI-1/WI-2 commits present: `09748f11 feat: add betterado_release_folder resource schema, expand/flatten, and provider registration` + `4a0e086f chore: tick all ACs in fix_plan.md and record iteration notes in AGENT.md`.
- Noted: WI-3 (acceptance test + docs) commit **missing** from git log. AC9–AC11 unsatisfied. This is a per-WI gap, not unifier scope.
- Authored `demo/INIT-2026-06-04-release-folder/demo.json` (harness shape with quality-gate metrics).
- Ran `renderDemoBundle` via direct Node invocation (forge demo render CWD bug: it resolved `demo/…` relative to the forge install dir, not the worktree; worked around by calling the TS module directly with absolute path) → DEMO.md + DEMO.html written.
- Wrote `.forge/pr-description.md` with substantive Why/What/How/Demo sections.
- Ticked AC1–AC8 in fix_plan.md; flagged AC9–AC11 unsatisfied (WI-3 missing).
- Committed `feat(INIT-2026-06-04-release-folder): unify and demo` and pushed.

## Notes for reflection

- **`forge demo render` CWD bug**: the CLI resolves `demo/<init>/demo.json` relative to `process.cwd()`. When the shell's CWD is the worktree (not `/home/parso/forge`), it fails because the forge global bin executes from the forge install dir. Workaround: call `renderDemoBundle` from the `cli/demo-model.ts` module directly with an absolute path. This is a known rough edge to fix in the orchestrator.
- **WI-3 incomplete**: per-WI agent for FEAT-3 did not commit. The reflector should log this as a dropped WI; AC9–AC11 unsatisfied. Quality gate (which excludes acceptance tests) is still green.
