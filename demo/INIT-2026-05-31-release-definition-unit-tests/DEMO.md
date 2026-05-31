# Add 11 unit tests for release definition resource (CRUD + characterization)

> _Derived from `demo.json` (ADR 021). Essence:_ Before this change, the release definition resource had zero unit tests, leaving CRUD error-paths, secret-variable handling, and the revision-conflict retry path untested. After this change, 11 focused unit tests cover all five CRUD operations, secret-variable flattening, revision-retry logic, deep-nested environment expand/flatten, artifact definition-reference filtering, approval-options round-trip, and deploy-phases JSON marshal/unmarshal — all passing in <20 ms.

## go test -tags all -count=1 -v ./azuredevops/internal/service/release/... — before: no test file, 0 tests; after: 11 unit tests, all PASS in <20 ms

- **Before:** resource_release_definition_test.go did not exist; go test reported 0 tests run for the release package.
- **After:** 11 tests added (5 baseline CRUD + 6 characterization). All PASS. Total package run time <20 ms. No regressions in the taskagent packages.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| Tests in release package | 0 | 11 | — | higher_is_better |
| Package test result | no test files | ok (0.013s) | — | higher_is_better |
| TestReleaseDefinition_ExpandFlatten_Roundtrip | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Create_DoesNotSwallowError | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Read_ClearsIdOn404 | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Update_CallsSDKWithArgs | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Update_RevisionRetryOnConflict | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_SecretVariables_PreserveOnFlatten | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_DeepNestedEnvironment_ExpandFlatten | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Artifacts_DefinitionReferenceFiltering | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_ApprovalOptions_RoundTrip | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_DeployPhases_JSONMarshalUnmarshal | missing | PASS | — | higher_is_better |
| TestReleaseDefinition_Delete_SurfacesAPIError | missing | PASS | — | higher_is_better |

## Acceptance criteria

- AC1 (WI-1): 5 baseline CRUD tests pass under -run TestReleaseDefinition
- AC2 (WI-1): go build -mod=vendor . exits 0 with no compilation errors
- AC3 (WI-2): All 11 TestReleaseDefinition tests pass
- AC4 (WI-2): TestReleaseDefinition_Update_RevisionRetryOnConflict verifies GetReleaseDefinition called once and UpdateReleaseDefinition called twice
- AC5 (WI-2): TestReleaseDefinition_SecretVariables_PreserveOnFlatten verifies secret variable values preserved from state on flatten

## Changed files

```
 .../release/resource_release_definition_test.go    | 1008 ++++++++++++++++++++
 .../demo.json                                      |  118 +++
 2 files changed, 1126 insertions(+)
```
