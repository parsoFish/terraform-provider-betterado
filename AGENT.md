# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-xx)

**Root cause of gate failure:** The gate ran `go test -tags all -run TestAccProviderMuxFree ./azuredevops/internal/acceptancetests/` and got `[no tests to run]` because `TestAccProviderMuxFree` didn't exist in the codebase yet.

**What was done:**
1. Added `TestAccProviderMuxFree` function to `azuredevops/internal/acceptancetests/resource_project_test.go` BEFORE `TestAccProject_basic`.
   - Uses `testutils.GetProviderFactories()` (pure-framework, proto v6, no mux)
   - Data source: `data "betterado_project" "muxfree"` with `name = "betterado-standing-demo"` (existing standing fixture — never creates a project)
   - Asserts `name` attribute matches
   - Calls `testutils.CaptureLiveEvidence("acceptance-provider-mux-free", ...)` to write `.forge/live-evidence/acceptance-provider-mux-free.json`
   - Has an early `t.Skip` guard when `TF_ACC` is unset (but the gate runs with `TF_ACC=1`)
2. Added `### BREAKING CHANGES` + `### INTERNAL` sections under `## [Unreleased]` in `CHANGELOG.md` documenting removal of mux scaffold + Terraform >= 1.x requirement
3. Bumped `PROVIDER_VERSION.txt` from `1.22.0` → `2.0.0`

**Verified:**
- `go build -tags all ./...` → clean
- `go test -tags all -list TestAccProviderMuxFree ./azuredevops/internal/acceptancetests/` → lists `TestAccProviderMuxFree` (no longer "no tests to run")
- Committed as `5e7ef0dd`

## What worked

- `testutils.GetProviderFactories()` is defined in `mux_provider.go` and returns a pure-framework `ProtoV6ProviderFactories` map
- `SharedFixtureProjectName = "betterado-standing-demo"` (defined in `shared_fixtures.go` in the same `acceptancetests` package)
- `testutils.CaptureLiveEvidence(label, url, payload)` signature: takes label string, url string, and any JSON-serializable interface{}
- The framework provider has `core.NewProjectDataSource` registered, which serves `data.betterado_project`
- The test file already imported `os` and `fmt` — no new imports needed

## What didn't work

- N/A (this was a green-field addition; no dead ends)

## Open questions

_(none blocking)_

## Notes for reflection

- The gate failure was "no tests to run" — the test simply didn't exist. Pattern: always ensure the quality_gate_cmd target function exists before iteration 0.
- `SharedFixtureProjectName` is in the same package (`acceptancetests`), directly accessible — no testutils import needed for the constant.
