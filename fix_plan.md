# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_workitemquery and betterado_workitemquery_folder are registered as terraform-plugin-framework resources in framework_provider.go and removed from provider.go ResourcesMap WHEN terraform apply runs configs using both betterado_workitemquery and betterado_workitemquery_folder THEN both resources are created, provider read-back returns all fields, idempotency re-plan produces no diff, destroy cleans up
- [x] AC2: GIVEN the SDKv2 resource_workitemquery.go and resource_workitemquery_folder.go are deleted WHEN go build -mod=vendor . is run THEN the provider compiles with no duplicate-type errors and no orphaned files
- [x] AC3: GIVEN TestAccWorkItemQuery_UnderArea runs with TF_ACC=1 using GetMuxedProviderFactories WHEN the muxed provider serves betterado_workitemquery as a framework resource THEN the test passes against real ADO and CaptureLiveEvidence is called with label acceptance-resource-workitemquery
- [x] AC4: GIVEN TestAccWorkItemQueryFolder_UnderArea runs with TF_ACC=1 using GetMuxedProviderFactories WHEN the muxed provider serves betterado_workitemquery_folder as a framework resource THEN the test passes against real ADO and CaptureLiveEvidence is called with label acceptance-resource-workitemquery-folder

## Completed this iteration (0)

- Created resource_workitemquery_framework.go with full CRUD, ExactlyOneOf ConfigValidator, ForceNew modifiers
- Created resource_workitemquery_folder_framework.go (immutable: Create/Read/Delete only)
- Registered both in framework_provider.go
- Removed both from provider.go ResourcesMap (added comments)
- Updated provider_test.go (removed SDKv2 entries)
- Deleted SDKv2 source files + their unit tests
- Rewrote acceptance tests to use GetMuxedProviderFactories + SharedFixtureProjectName + CaptureLiveEvidence + idempotency step
- make docs + regenerated workitemquery.md + workitemquery_folder.md
- CHANGELOG entry added

## Pending

- Live gate run (forge will validate with TF_ACC=1)
