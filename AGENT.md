# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration)

**Problem (from last-gate-failure.md):** "Invalid Provider Server Combination: Duplicate resource type: betterado_release_folder" — mux fails to load provider schema because `betterado_release_folder` is registered in BOTH the SDKv2 provider (`azuredevops/provider.go`) and the framework provider (`azuredevops/internal/provider/framework_provider.go`).

**Fix applied:**
1. Removed `"betterado_release_folder": release.ResourceReleaseFolder()` from `azuredevops/provider.go` line 91 (SDKv2 resources map). Left a comment explaining why.
2. Updated `azuredevops/provider_test.go` `TestProvider_HasChildResources` to not expect `betterado_release_folder` in the SDKv2 resource map (it's now framework-only).
3. Ran `make fmt` to fix gofmt formatting issues (comment caused whitespace change).
4. All offline gates now pass: `make test` ✓, `golangci-lint run ./azuredevops/...` (0 issues) ✓, `make terrafmt-check` ✓.

**Commit:** `fix: remove betterado_release_folder from SDKv2 provider to eliminate mux duplicate`

## What worked

- `make fmt` is required after any edit to provider.go — gofmt enforced by `make test` (fmtcheck step).
- The `release` import in `provider.go` stays — it's still needed for data sources (`DataReleaseFolder()`, etc.).
- The data source `betterado_release_folder` (line 186 of provider.go) stays in SDKv2 — only the **resource** was migrated.
- `TestProvider_HasChildResources` in `azuredevops/provider_test.go` checks the SDKv2 resource map exhaustively — must be updated when a resource is migrated out.

## What didn't work

_(nothing to record yet — issue was clear from the gate failure message)_

## Open questions

- Will the live acceptance test (`TestAccReleaseFolderFramework` with TF_ACC=1) pass with the current framework resource implementation? The mux conflict is fixed, but the resource CRUD logic hasn't been exercised live yet. The next gate run will reveal any API/logic issues.

## Notes for reflection

- Pattern: when migrating a resource from SDKv2 to framework via mux, always remove it from both `provider.go` resource map AND `provider_test.go` `TestProvider_HasChildResources`. Otherwise the duplicate causes "Invalid Provider Server Combination" at plan time.
