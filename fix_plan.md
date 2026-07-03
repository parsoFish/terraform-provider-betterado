# Fix Plan

> Checklist for WI-8. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN serviceendpoint_artifactory, serviceendpoint_dynamics_lifecycle_services exist in the SDKv2 provider WHEN both migrated to terraform-plugin-framework THEN framework resource files exist for both; both deregistered from provider.go ResourcesMap; registered in framework_provider.go Resources(); no Duplicate resource type panic; CI-equivalent gate passes
- [x] AC2: GIVEN acceptance tests for migrated enterprise endpoint resources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go counts updated for removed SDKv2 resources
- [x] AC3: GIVEN framework resource files compile correctly WHEN go build -mod=vendor . is run THEN provider binary builds without errors; TypeName for each uses req.ProviderTypeName + suffix pattern
