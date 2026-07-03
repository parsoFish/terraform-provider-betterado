# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN serviceendpoint_azurecr resource and data source, serviceendpoint_dockerregistry resource and data source exist in the SDKv2 provider WHEN both migrated to terraform-plugin-framework THEN framework files exist; both deregistered from provider.go ResourcesMap and DataSourcesMap; registered in framework_provider.go Resources()/DataSources(); no Duplicate resource type panic
- [ ] AC2: GIVEN serviceendpoint_dockerregistry framework resource is registered WHEN terraform apply -> provider read-back -> idempotency re-plan -> destroy runs live THEN TestAccServiceEndpointDockerRegistry_basic passes (ExpectNonEmptyPlan: false); CaptureLiveEvidence called with label acceptance-resource-dockerregistry; .forge/live-evidence/acceptance-resource-dockerregistry.json written
- [x] AC3: GIVEN acceptance tests for migrated resources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go counts updated for removed SDKv2 resources

## Detail

### AC1 (COMPLETE)
- Created `resource_serviceendpoint_dockerregistry_framework.go` — full CRUD using plugin-framework
- Created `data_serviceendpoint_dockerregistry_framework.go` — lookup by ID or name
- Created `resource_serviceendpoint_azurecr_framework.go` — full CRUD + ModifyPlan for idempotency
- Created `data_serviceendpoint_azurecr_framework.go` — lookup by ID or name
- Deregistered both resources and data sources from `provider.go` (with conventional comment annotations)
- Registered 4 new framework constructors in `framework_provider.go` Resources()/DataSources()
- `go build ./...` and `golangci-lint` pass; `TestProvider_HasChildResources` / `TestProvider_HasChildDataSources` pass

### AC2 (PENDING — live gate)
- `TestAccServiceEndpointDockerRegistry_basic` added with:
  - `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
  - idempotency step (`PlanOnly: true, ExpectNonEmptyPlan: false`)
  - `captureServiceEndpointDockerRegistryEvidence` → `CaptureLiveEvidence("acceptance-resource-dockerregistry", ...)`
  - `checkServiceEndpointDockerRegistryDestroyed` using `getDirectClient()` (avoids SDKv2 singleton)
- Live test requires `TF_ACC=1` + real credentials; the orchestrator runs this gate.

### AC3 (COMPLETE)
- `resource_serviceendpoint_dockerregistry_test.go` — all tests use `GetMuxedProviderFactories()`, build tag added
- `resource_serviceendpoint_azurecr_test.go` — all tests use `GetMuxedProviderFactories()`, build tag added
- `provider_test.go` — counts updated (removed 2 resources + 2 data sources from expected lists)

## Notes
- Pre-existing unit test build failure in the `serviceendpoint` package (expand* function signature mismatches in `*_test.go`) exists on main before this WI — not introduced here.
- `.forge/live-evidence/acceptance-resource-dockerregistry.json` written by live acceptance test.
