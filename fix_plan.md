# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a clean worktree with the data source implementations already present WHEN go test -tags all -count=1 ./azuredevops/internal/service/release/... is run THEN all tests pass including TestDataReleaseDefinitionRevision_Read_ReturnsJSON, TestDataReleaseDefinitionRevision_Read_Error, TestDataReleaseDefinitionHistory_Read_Populates, TestDataReleaseDefinitionHistory_Read_Empty, TestDataReleaseDefinitionHistory_Read_Error, TestDataSourceDocPagesExist
- [x] AC2: GIVEN examples/data-sources/betterado_release_definition_revision/ does not yet exist WHEN the WI is completed THEN examples/data-sources/betterado_release_definition_revision/main.tf exists and contains a valid HCL example
- [x] AC3: GIVEN examples/data-sources/betterado_release_definition_history/ does not yet exist WHEN the WI is completed THEN examples/data-sources/betterado_release_definition_history/main.tf exists and contains a valid HCL example
- [x] AC4: GIVEN all implementation files are present WHEN make test is run (gofmt check + go test ./... without TF_ACC) THEN the command exits 0 — gofmt clean and all unit tests pass
- [x] AC5: GIVEN all implementation files are present WHEN golangci-lint run ./... is run THEN exits 0 with no lint errors
