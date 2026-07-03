# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_workitem is registered as a terraform-plugin-framework resource in framework_provider.go and removed from provider.go ResourcesMap WHEN terraform apply runs a config using betterado_workitem THEN the resource is created, the provider read-back returns all fields, idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy cleans up
- [x] AC2: GIVEN the SDKv2 resource_workitem.go and resource_workitem_test.go are deleted and the old SDKv2 registration removed WHEN go build -mod=vendor . is run THEN the provider compiles with no duplicate-type errors and no orphaned dead files remain
- [ ] AC3: GIVEN the acceptance test TestAccWorkItem_basic runs with TF_ACC=1 WHEN the muxed provider is used (GetMuxedProviderFactories) THEN the test passes against real ADO and CaptureLiveEvidence is called with label acceptance-resource-workitem
  - AC3 requires live ADO run; code is fully wired, waiting for forge gate
