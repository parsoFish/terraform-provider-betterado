# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (completed all ACs — commit 50d648df)

**What was done:**

1. **AC1 — make docs**: Created 20 `examples/resources/<resource_name>/resource.tf` example files (7 branch policy, 7 repository policy, 6 check resources). `make docs` ran cleanly via `tfplugindocs generate --provider-name betterado`. All 5 `docs/guides/` hand-written files restored automatically by the GNUmakefile `docs:` target.

2. **AC2 — CHANGELOG**: Extended existing `## [Unreleased]` entry (already had branch policy bullet) with two more bullets listing all repository policy and checks resources by name. Also mentioned gap matrix docs.

3. **AC3 — Version bump**: `PROVIDER_VERSION.txt` bumped from `1.2.0` (main) to `1.3.0` (minor bump for 20 new framework resource implementations).

4. **AC4 — Demo evidence**: Live evidence already existed in `.forge/live-evidence/acceptance-resource.json` (check_approval, id=80, capturedAt: 2026-07-03T01:25:01Z, URL: `https://dev.azure.com/davidgparsonson/.../_apis/pipelines/checks/configurations/80?api-version=7.1`). Created `forge/history/INIT-2026-07-01-migrate-framework-policy-branch/demo/demo.json` with `acceptance-resource` checkpoint carrying that `liveEvidence.url`.

**Quality gate**: `go test -tags all -run TestProvider_HasChildResources -count=1 ./azuredevops/` → `ok 0.006s` ✓

## What worked

- `make docs` auto-restores `docs/guides/` via `git checkout -- docs/guides/` — already in the GNUmakefile `docs:` target.
- `tfplugindocs` requires exactly `examples/resources/<exact-type-name>/resource.tf`. Resource type names must match `resp.TypeName` in `Metadata()`. Repository policy names use singular `_pattern` not `_patterns`.
- Live evidence was already in `.forge/live-evidence/acceptance-resource.json` from WI-3/4 acceptance tests. No re-running needed.
- Demo.json format follows `forge/history/INIT-2026-07-01-migrate-framework-release-folder-permissions/demo/demo.json`.

## What didn't work

_(none — all ACs completed in iteration 1)_

## Open questions

_(none)_

## Notes for reflection

- `make docs` GNUmakefile target handles `git checkout -- docs/guides/` internally — good pattern to document.
- Repository policy resource type names: singular `_pattern` (not `_patterns`) per `resp.TypeName` in framework files.
