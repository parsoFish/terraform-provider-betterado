# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a `betterado_release_folder` block in HCL, with the mux provider active WHEN `terraform apply` runs THEN the folder is created via the Release Management API at vsrm.dev.azure.com; the provider reads it back; a re-plan produces `ExpectNonEmptyPlan: false`; `terraform destroy` removes it cleanly
- [x] AC2: GIVEN the `TestAccReleaseFolderFramework` acceptance test is run with TF_ACC=1 WHEN the live apply → read-back → idempotency → destroy cycle completes THEN the test passes; `CaptureLiveEvidence("acceptance-resource", <vsrm-GET-url>, apiResponse)` is called during the live read-back so `.forge/live-evidence/acceptance-resource.json` is written
- [x] AC3: GIVEN the CI-equivalent gate runs (no TF_ACC) WHEN `make test && golangci-lint run ./azuredevops/... && make terrafmt-check` executes THEN all checks pass with zero new lint findings on changed files

## Status: COMPLETE

All ACs verified:
- resource_release_folder_framework.go: full CRUD + ImportState
- resource_release_folder_framework_test.go: TestAccReleaseFolderFramework with SharedReleaseFixture + CaptureLiveEvidence
- framework_provider.go: release.NewReleaseFolderResource registered
- .forge/live-evidence/acceptance-resource.json: written by live acceptance test (vsrm.dev.azure.com API response)
- CI gate: go test EXIT:0, golangci-lint 0 issues, terrafmt-check EXIT:0
