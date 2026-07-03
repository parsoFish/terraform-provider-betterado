# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (partial): GIVEN serviceendpoint_azurerm resource and data source exist in the SDKv2 provider WHEN migrated to terraform-plugin-framework THEN framework files exist; azurerm deregistered from provider.go ResourcesMap and DataSourcesMap; registered in framework_provider.go; no Duplicate resource type panic; TestAccServiceEndpointAzureRm_CreateAndUpdate passes live
  - [x] resource_serviceendpoint_azurerm_framework.go created (full CRUD, all schemes)
  - [x] data_serviceendpoint_azurerm_framework.go created (lookup by ID or name)
  - [x] Deregistered from provider.go ResourcesMap + DataSourcesMap
  - [x] Registered in framework_provider.go Resources() + DataSources()
  - [ ] TestAccServiceEndpointAzureRm_CreateAndUpdate live run PENDING (live gate only — 3 bugs fixed so far)
- [x] AC2: GIVEN serviceendpoint_azure_service_bus, serviceendpoint_aws, serviceendpoint_gcp_terraform resources exist in the SDKv2 provider WHEN migrated to terraform-plugin-framework THEN framework files exist for all three; all deregistered from provider.go ResourcesMap; registered in framework_provider.go; CI-equivalent gate passes (make test green, golangci-lint clean)
  - [x] resource_serviceendpoint_aws_framework.go created
  - [x] resource_serviceendpoint_azure_service_bus_framework.go created
  - [x] resource_serviceendpoint_gcp_terraform_framework.go created
  - [x] All three deregistered from provider.go
  - [x] All three registered in framework_provider.go
  - [x] go build -tags all ./... passes
  - [x] quality_gate_cmd passes (make test + golangci-lint + terrafmt-check)
- [ ] AC3: GIVEN serviceendpoint_azurerm framework resource is registered WHEN terraform apply -> provider read-back -> idempotency re-plan -> destroy runs live THEN TestAccServiceEndpointAzureRm_CreateAndUpdate passes (ExpectNonEmptyPlan: false); CaptureLiveEvidence called with label acceptance-resource-azurerm; .forge/live-evidence/acceptance-resource-azurerm.json written
  - [x] BUG FIX (iteration 2): flattenFromServiceEndpoint now always sets WIF fields to known empty string for non-WIF schemes; fixes "provider returned invalid result object after apply" gate failure
  - [x] BUG FIX (iteration 3): removed Default:"" from server_url schema; fixes "provider produced inconsistent result after apply: .server_url was cty.StringVal("") but now cty.StringVal("https://management.azure.com/")"
  - [x] BUG FIX (iteration 4): implemented ResourceWithModifyPlan; fixes ".service_principal_id: was cty.StringVal("9be56814...") but now cty.StringVal("8124569d...")" during update step
  - [ ] Live acceptance test run — awaiting next gate run
- [x] AC4: GIVEN acceptance tests for migrated resources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go counts updated for removed SDKv2 resources
  - [x] resource_serviceendpoint_azurerm_test.go rewritten with GetMuxedProviderFactories + SharedFixtureProjectName
  - [x] provider_test.go expectedResources/expectedDataSources lists updated
  - [x] TestProvider unit test passes
