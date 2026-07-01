# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN a `betterado_release_folder` block in HCL, with the mux provider active WHEN `terraform apply` runs THEN the folder is created via the Release Management API at vsrm.dev.azure.com; the provider reads it back; a re-plan produces `ExpectNonEmptyPlan: false`; `terraform destroy` removes it cleanly
- [ ] AC2: GIVEN the `TestAccReleaseFolderFramework` acceptance test is run with TF_ACC=1 WHEN the live apply → read-back → idempotency → destroy cycle completes THEN the test passes; `CaptureLiveEvidence("acceptance-resource", <vsrm-GET-url>, apiResponse)` is called during the live read-back so `.forge/live-evidence/acceptance-resource.json` is written
- [x] AC3 (partial): `make test`, `golangci-lint run ./azuredevops/...`, `make terrafmt-check` all pass with zero new lint findings

## Sub-tasks

- [x] Fix "Duplicate resource type: betterado_release_folder" mux error
  - Removed `betterado_release_folder` registration from `azuredevops/provider.go` (SDKv2)
  - Updated `azuredevops/provider_test.go` `TestProvider_HasChildResources` to not expect it in SDKv2 map
  - Resource now exclusively lives in `framework_provider.go` → `release.NewReleaseFolderResource`

- [x] Fix nil-Meta panic in checkReleaseFolderFrameworkDestroyed / captureReleaseFolderFrameworkEvidence
  - Both helpers called testutils.GetProvider().Meta().(*client.AggregatedClient)
  - ProtoV6ProviderFactories never configures the SDKv2 singleton → Meta() == nil → panic
  - Fixed: use getDirectClient() (same package, resource_task_group_test.go) which builds client from env vars
  - Dropped unused `client` import from the file

- [ ] AC1 + AC2 live gate: TestAccReleaseFolderFramework must pass with TF_ACC=1
  - The framework resource implementation is committed (resource_release_folder_framework.go)
  - The acceptance test now uses SharedReleaseFixture (no project created)
  - The mux duplicate is fixed
  - The nil-Meta panic is fixed
  - Next gate run will verify the full live CRUD cycle
