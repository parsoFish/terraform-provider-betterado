# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the pure-framework provider binary is built WHEN the live acceptance test TestAccProviderMuxFree runs with TF_ACC=1 THEN the test passes (betterado_project resource is plannable and readable against real ADO), CaptureLiveEvidence is called with real REST GET response, and the .forge/live-evidence/acceptance-provider-mux-free.json file is written
  - Added TestAccProviderMuxFree to resource_project_test.go using GetProviderFactories() + data.betterado_project
  - `go test -tags all -list TestAccProviderMuxFree ./azuredevops/internal/acceptancetests/` confirms test is registered (not "no tests to run")
- [x] AC2: GIVEN the mux scaffold has been removed WHEN CHANGELOG.md is inspected THEN an entry exists under ## [Unreleased] describing the breaking removal of the mux scaffold and documenting that terraform >= 1.x is now required
  - Added ### BREAKING CHANGES and ### INTERNAL sections under ## [Unreleased] in CHANGELOG.md
- [x] AC3: GIVEN PROVIDER_VERSION.txt contains 1.22.0 WHEN the version bump is applied THEN PROVIDER_VERSION.txt contains a major-bumped version (2.0.0)
  - PROVIDER_VERSION.txt changed from 1.22.0 → 2.0.0

## Status: All ACs implemented, committed. Awaiting live gate (TF_ACC=1).
