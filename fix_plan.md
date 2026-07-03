# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_workitemtracking_field is registered as a terraform-plugin-framework resource in framework_provider.go and removed from provider.go ResourcesMap WHEN terraform apply runs a config using betterado_workitemtracking_field THEN the resource is created, the provider read-back returns all fields including computed attributes (can_sort_by, is_queryable, is_identity, is_picklist, supported_operations), idempotency re-plan produces no diff, and destroy cleans up
- [x] AC2: GIVEN the SDKv2 resource_field.go and its unit tests are deleted WHEN go build -mod=vendor . is run THEN the provider compiles with no duplicate-type errors and no orphaned dead files remain
- [x] AC3: GIVEN the acceptance test TestAccWorkItemTrackingField_Basic runs with TF_ACC=1 WHEN the muxed provider is used (GetMuxedProviderFactories) THEN the test passes against real ADO and CaptureLiveEvidence is called with label acceptance-resource-workitemtracking-field

## Completion notes (iteration 0 → final)

All ACs were completed in the initial iteration (commit 5cd75dd9):

1. **AC1**: `resource_field_framework.go` created with full framework implementation; registered in `framework_provider.go` as `workitemtracking.NewFieldResource`; removed from `provider.go` ResourcesMap; `provider_test.go` updated.

2. **AC2**: `resource_field.go` (SDKv2) deleted; `go build -mod=vendor .` passes with no errors; `make test` all green.

3. **AC3**: All test functions use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`; `TestAccWorkItemTrackingField_Basic` calls `captureFieldEvidence` which invokes `CaptureLiveEvidence("acceptance-resource-workitemtracking-field", url, field)`.

Standing ACs satisfied:
- `make test` passes, `golangci-lint run --new-from-rev=main ./azuredevops/...` = 0 issues
- `docs/resources/workitemtracking_field.md` regenerated (shows all schema attributes)
- CHANGELOG.md updated under `## [Unreleased]` with bullet for this WI
- `examples/resources/betterado_workitemtracking_field/resource.tf` exists
