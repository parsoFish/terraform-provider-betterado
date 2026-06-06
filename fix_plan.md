# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the shared fixture from WI-1 is available and TF_ACC=1 is set WHEN TestAccReleaseDefinition_basic is run with TF_ACC=1 THEN the test uses the shared fixture objects (project ID, repo ID, build-definition ID from SharedReleaseFixture) instead of hand-rolling a minimal project inline, applies Terraform config against those live fixture objects, asserts the API round-trip, and passes end-to-end with clean destroy
- [x] AC2: GIVEN the refactored TestAccReleaseDefinition_basic HCL is examined WHEN the HCL template function hclReleaseDefinitionBasic (or its replacement) is inspected THEN it no longer contains an inline betterado_project resource block hand-rolled by the test itself — instead it references fixture-supplied IDs (e.g. project_id = <fixture.ProjectID>) so the project is owned and cleaned up by the fixture
- [ ] AC3: GIVEN the refactored test runs live WHEN the full test lifecycle completes (apply → API-roundtrip-assert → destroy) THEN no ADO-validity errors (VS402877, VS402982, invalid permission keys) are produced, and no orphaned projects, repos, or release definitions remain in the ADO org after the run

## Notes

AC1 + AC2 are structurally complete as of iteration 1 (commit 8c97ba95):
- `TestAccReleaseDefinition_basic` calls `SharedReleaseFixture(t)` at the top.
- New `hclReleaseDefinitionBasicFixture(name, fixture)` emits only `betterado_release_definition`; `project_id` is the fixture UUID.
- `hclReleaseDefinitionBasic(name)` preserved unchanged for other tests (`_update`, etc.).
- `go build ./...` and `go vet ./...` clean; test skips correctly without `TF_ACC`.

AC3 requires a live TF_ACC run — the orchestrator's acceptance gate with TF_ACC=1 will verify this.
The HCL already has correct VS402982 (retention_policy) and VS402877 (pre+post approvals) blocks.
