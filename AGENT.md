# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (WI-2)

On entry, all three ACs were already implemented from prior iterations on this branch:

**AC1 — Schema unit test:**
- `azuredevops/internal/service/build/resource_build_folder_framework_test.go` — `TestBuildFolderFramework_Schema` passes
- Schema declares `project_id`, `path`, `description` — verified by calling `r.Schema()` directly

**AC2 — Framework provider registration + compile:**
- `azuredevops/internal/service/build/resource_build_folder_framework.go` — full framework resource (361 lines)
- `azuredevops/internal/provider/framework_provider.go` line 210: `build.NewBuildFolderResource`
- `azuredevops/provider.go` — SDKv2 `betterado_build_folder` entry removed/commented (line 58-61 shows comment explaining move)
- `go build -mod=vendor .` — compiles cleanly

**AC3 — Acceptance test:**
- `azuredevops/internal/acceptancetests/resource_build_folder_framework_test.go` — `TestAccBuildFolder_Framework_basic`
- Uses `SharedReleaseFixture(t)` for pre-existing project (no new ADO project created — org at 1000-project cap)
- `CaptureLiveEvidence("acceptance-resource", apiURL, folder)` called with REST GET URL
- `ExpectNonEmptyPlan: false` set on idempotency step
- `checkBuildFolderFrameworkDestroyed` verifies cleanup

**CI gates verified:**
- `make test` — PASS (no failures)
- `golangci-lint run --new-from-rev=main ./azuredevops/...` — 0 issues
- `make terrafmt-check` — PASS
- `go test -tags all -run TestBuildFolderFramework_Schema ./azuredevops/internal/service/build/...` — PASS

## What worked

- Framework resource pattern matches other betterado framework resources in the codebase
- `SharedReleaseFixture(t)` for project fixture (avoids ADO project creation limit)
- Removing SDKv2 `betterado_build_folder` from `provider.go`'s `ResourcesMap` and adding it only to `framework_provider.go` prevents "Invalid Provider Server Combination" at mux time
- `CaptureLiveEvidence` is best-effort (failure doesn't fail the test) — correct pattern

## What didn't work

_(none encountered in this iteration)_

## Open questions

_(none)_

## Notes for reflection

- WI-2 was already complete when iteration 0 ran — status field in WI-2.md was already `complete`
- All ACs verified: AC1 schema test passes, AC2 builds, AC3 acceptance test written with all required elements
- The quality gate cmd `go test -tags all -run TestBuildFolderFramework_Schema ./azuredevops/internal/service/build/...` runs in 0.004s and PASSES
