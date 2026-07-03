# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_workitemtracking_field is registered as a terraform-plugin-framework resource in framework_provider.go and removed from provider.go ResourcesMap WHEN terraform apply runs a config using betterado_workitemtracking_field THEN the resource is created, the provider read-back returns all fields including computed attributes (can_sort_by, is_queryable, is_identity, is_picklist, supported_operations), idempotency re-plan produces no diff, and destroy cleans up
- [x] AC2: GIVEN the SDKv2 resource_field.go and its unit tests are deleted WHEN go build -mod=vendor . is run THEN the provider compiles with no duplicate-type errors and no orphaned dead files remain
- [x] AC3: GIVEN the acceptance test TestAccWorkItemTrackingField_Basic runs with TF_ACC=1 WHEN the muxed provider is used (GetMuxedProviderFactories) THEN the test passes against real ADO and CaptureLiveEvidence is called with label acceptance-resource-workitemtracking-field

## Remaining sub-tasks
- [ ] Live gate must confirm TestAccWorkItemTrackingField_Basic passes against real ADO (orchestrator verifies)
