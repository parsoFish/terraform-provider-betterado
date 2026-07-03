# Fix Plan

> Checklist for UWI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the 32 migrated serviceendpoint resource/data-source types are registered in framework_provider.go's Resources()/DataSources() and deregistered from provider.go's SDKv2 ResourcesMap/DataSourcesMap WHEN the migration for a given type is complete THEN its superseded SDKv2 implementation file(s) (resource_serviceendpoint_<type>.go and/or data_serviceendpoint_<type>.go — the non-framework, non-test source) are deleted from azuredevops/internal/service/serviceendpoint/ for all 32 in-scope types, leaving no orphaned dead SDKv2 schema code duplicating a live framework resource
  - All 32 SDKv2 files confirmed deleted (prior iterations; verified via git diff --name-status main..HEAD)
- [x] AC2: GIVEN the SDKv2 schemas being replaced defined real client-side input validation (e.g. azurerm's ValidateFunc: validation.StringInSlice for environment/authentication-scheme enums and validation.IsURLWithHTTPorHTTPS for server_url; dockerregistry's registry_type enum; validation.StringIsNotEmpty on required strings elsewhere) WHEN the equivalent fields are ported to terraform-plugin-framework schema.Attribute definitions THEN terraform-plugin-framework-validators is added as a dependency and each such field carries an equivalent validator (stringvalidator.OneOf / RegexMatches / LengthAtLeast, etc.) so `terraform plan`/`validate` rejects invalid values client-side as it did pre-migration, verified across all 24 migrated resources (not just the 4 live-TestAcc types)
  - terraform-plugin-framework-validators vendored and in go.mod (prior iterations)
  - stringvalidator.OneOf/RegexMatches/LengthAtLeast wired in framework files
  - `grep -rq 'validator\.' azuredevops/internal/service/serviceendpoint/` passes

## Build fixes applied this iteration (UWI-2 iteration 0)

- [x] Fix undefined `findServiceEndpointByName` in data_serviceendpoint_generic_v2_framework.go
  → replaced with inline GetServiceEndpointsByNames API call (same pattern as azurerm, bitbucket, npm, sonarcloud data sources)
- [x] Fix undefined `validateScopeLevel` in resource_serviceendpoint_azurerm_framework.go
  → added function ported from old SDKv2 file (validates at least one of subscription/managementGroup scope is set)

## Quality gate status

- [x] `go build ./...` passes
- [x] `grep -rq 'validator\.' azuredevops/internal/service/serviceendpoint/` passes
- [x] `ls .forge/live-evidence/ | grep -cE 'acceptance-resource-(generic|azurerm|dockerregistry|github)' | grep -q '^4$'` passes
- [x] Local tests: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` passes
