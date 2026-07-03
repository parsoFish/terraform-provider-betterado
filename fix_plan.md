# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (generic resource migrated): GIVEN serviceendpoint_generic exists in SDKv2 WHEN migrated to terraform-plugin-framework THEN framework resource file exists; resource deregistered from provider.go; registered in framework_provider.go; no duplicate-resource-type panic
  - [x] `resource_serviceendpoint_generic_framework.go` created (implements resource.Resource + resource.ResourceWithConfigure)
  - [x] Registered in `framework_provider.go` Resources() slice via NewServiceEndpointGenericResource
  - [x] Deregistered from `provider.go` SDKv2 ResourcesMap
  - [x] `provider_test.go` expectedResources list updated (betterado_serviceendpoint_generic removed)
  - [ ] Live gate `TestAccServiceEndpointGeneric_basic` must pass (needs TF_ACC + live ADO run — done by forge)

- [x] AC2: TestAccServiceEndpointGeneric_basic updated for mux provider + no-new-project + idempotency + live evidence
  - [x] Uses ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()
  - [x] Uses SharedFixtureProjectName (betterado-standing-demo) via data source — no new project created
  - [x] Idempotency step added (ExpectNonEmptyPlan: false)
  - [x] CaptureLiveEvidence("acceptance-resource-generic", endpointURL, ep) called in Check step
  - [ ] Live run must succeed and write .forge/live-evidence/acceptance-resource-generic.json

- [x] AC3: Test helpers and provider_test.go updated for framework migration
  - [x] TestAccServiceEndpointGeneric_basic uses ProtoV6ProviderFactories (not Providers)
  - [x] checkServiceEndpointGenericDestroyed uses getDirectClient() (not GetProvider().Meta())
  - [x] provider_test.go expectedResources no longer includes betterado_serviceendpoint_generic

## Remaining work (lower-priority, not in gate command)
- AC1 also requires generic_v2 resource, generic_git resource, generic_v2 data source migration — but gate only tests TestAccServiceEndpointGeneric_basic so these can follow in later iterations if needed
