# Unifier Agent Memory — INIT-2026-07-01-migrate-framework-git

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

### Iteration 1 (UWI-3 — terminal re-prep)

**State at start:** Branch clean at c2b0aaec (prior unifier commit). Quality gate green: ok release 0.008s | ok taskagent 0.006s | ok taskagent/validate 0.005s. UWI-2 (code-fix) addressed review concerns: go.mod tidy for tflog direct dep, initialization block validators restored, gap matrix updated, fixture lookup hardened.

**Actions taken:**

1. Read AGENT.md, fix_plan.md, UWI-1/2/3 specs. Noted that UWI-3 is the terminal re-prep packaging step.
2. Ran quality gate: go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... → GREEN.
3. Verified actual diffStat from git diff --stat main...HEAD = "36 files changed, 4022 insertions(+), 624 deletions(-)".
4. Updated demo.json: corrected diffStat; added initialization-validators checkpoint reflecting UWI-2 fix; updated essence to mention UWI-2 fixes; updated AC4 evidence to mention go.mod fix; updated quality-gate afterNote with fresh run output.
5. Re-rendered DEMO.md (derived from updated demo.json).
6. Updated .forge/pr-description.md ## Why/What/How to include UWI-2 context (go.mod tidy, initialization validators, fixture hardening).
7. Committed as feat(INIT-2026-07-01-migrate-framework-git): unify and demo.
8. Pushed to origin.

**AC status (UWI-3):**
- All ACs 1-13, 15: met.
- AC14 (liveEvidence.url): partial — credentials-environment constraint as documented.


_(updated by each iteration — most recent at the top)_

### Iteration 1 (UWI-1)

**State at start:** Prior unifier run had already committed `demo.json`, `DEMO.md`, and `.forge/pr-description.md` (commit `e6824376`). Two additional fix commits were on top (`9aa2dbed`, `7d1281c2`) — fixing initialization block validators and updating the gap matrix after review. Branch was clean and up to date with origin.

**Actions taken:**

1. Read `AGENT.md`, `fix_plan.md`, all 6 WI specs to understand scope.
2. Ran quality gate verbatim: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → **GREEN** (ok release 0.007s | ok taskagent 0.005s | ok taskagent/validate 0.004s).
3. Detected stale `diffStat` in `demo.json` ("33 files changed" but actual branch diff = "36 files changed, 4039 insertions(+), 624 deletions(-)"). Updated.
4. Updated `afterNote` for quality-gate checkpoint with actual captured output (exact stdout from quality gate run).
5. Updated AC15 evidence string with concrete test result from actual gate run.
6. Re-rendered `DEMO.md` via direct Node invocation of `cli/demo-model.ts::renderDemoBundle` (see note below about forge CLI workaround).
7. Committed as `feat(INIT-2026-07-01-migrate-framework-git): unify and demo`.
8. Pushed — `origin/forge/INIT-2026-07-01-migrate-framework-git` == local HEAD (c2b0aaec).

**AC status:**
- AC1–AC13, AC15: **met** (concrete evidence in acEvaluations).
- AC14 (liveEvidence.url on acceptance-resource checkpoint): **partial** — `CaptureLiveEvidence("acceptance-resource", ...)` call confirmed in test code; `liveEvidence.url` empty because TF_ACC not available in unifier env.

## Notes for reflection

_(observations the reflector should capture into the brain)_

- `forge demo render` CLI fails from worktrees due to eager top-level import chain in `cli.ts` → `architect-runner.ts` → `pm-invocation.ts` → `deriveAgentSpec('skills/project-manager/SKILL.md')` (missing from worktrees). Workaround: invoke `renderDemoBundle` from `cli/demo-model.ts` directly via Node. Forge team should consider lazy-loading the PM invocation module or skipping the load for `demo render`.
- Unifier "resume" scenario: if prior unifier commits exist, only update what is stale (diffStat, afterNotes, evidence strings); do not re-author from scratch — it wastes iterations.
