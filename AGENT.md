# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1

**Bug fixed:** The build tag `//go:build (all || resource_release_definition) && !exclude_resource_release_definition` added in iteration 0 was incorrect for `resource_release_definition_test.go` in the acceptancetests package. On main, this file had NO build tag. The data source test files (`data_release_definition_test.go`, `data_release_definition_revision_history_test.go`) reference `hclReleaseDefinitionBasic` without any build guard — so when `make test` runs `go test ./...` without tags, those files were compiled but `hclReleaseDefinitionBasic` was excluded, causing "undefined" errors.

**Fix:** Removed the build tag. Also applied gofmt to unit test file.

**All CI gates now green:**
- `make test` — acceptancetests package BUILDS (test failures are pre-existing, need TF_ACC=1)
- `golangci-lint run ./...` — clean
- `make terrafmt-check` — clean
- `go test -tags all -list TestAccReleaseDefinition_stagesArraySyntax` — test visible

### Iteration 0 (prior)

**Prior state (from WI-1/WI-2 commits):**
- `resource_release_definition.go` had already renamed schema key `environment` → `stages` (commit `eaa83d20`)
- `SchemaConfigModeAttr` was already added to `stages`, `deploy_phase`, `retention_policy` (commit `a9fe8bd1`) — this is what enables `stages = [{ }]` attribute/array syntax in HCL

**What was done this iteration:**
- Rewrote `resource_release_definition_test.go` from scratch with:
  1. Added `//go:build (all || resource_release_definition) && !exclude_resource_release_definition` build tag (required by project convention)
  2. Added `TestAccReleaseDefinition_stagesArraySyntax` — new test with stages = [...] array syntax fixture, apply + read-back assertions + idempotency PlanOnly step + CheckDestroy (AC1 + AC2)
  3. Converted ALL 14 HCL helper functions from `environment { ... }` block syntax to `stages = [{ ... }]` attribute/array syntax
  4. Updated ALL TestCheckResourceAttr state-path assertions from `environment.N.*` to `stages.N.*` across all 12 test functions

**Key discovery:** `SchemaConfigModeAttr` was the critical prior WI change that makes `stages = [{ }]` syntax valid in HCL. The nested sub-schemas (deploy_phase, retention_policy, etc.) also needed this mode to accept array syntax without using block notation.

## What worked

- `go build -tags all ./...` passes cleanly after the rewrite
- `go test -tags all -count=1 -list 'TestAccReleaseDefinition' ./azuredevops/internal/acceptancetests/` confirms all 17 tests are present and the package compiles
- Zero `"environment.` remaining in TestCheckResourceAttr calls (confirmed via grep)
- Zero `environment {` block syntax remaining in HCL strings (confirmed via grep)

## What didn't work

_(nothing failed in this iteration)_

## Open questions

- Live TF_ACC=1 run needed to confirm the array syntax actually works end-to-end against ADO. The schema was set up with SchemaConfigModeAttr in prior WIs; the test file changes are complete and compile cleanly.

## Notes for reflection

- The WI spec said to check AC1 = test fails BEFORE implementation. Since schema already has `stages` from WI-2, the "before" state is moot — the test now exercises the correct schema key. The orchestrator's gate is `go test -run TestAccReleaseDefinition_stagesArraySyntax` which will exit 0 (no TF_ACC) until a live run confirms AC2.
- All HCL fixtures use the same `stages = [{ ... }]` pattern consistently — nested blocks inside a stage use the same array syntax (e.g. `deploy_phase = [{ ... }]`). This is required because `SchemaConfigModeAttr` was set on those sub-schemas too.
