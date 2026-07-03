# Agent Memory — WI-9

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed all ACs)

1. Read WI-9.md, oriented on existing framework migrations (github, azurerm, azurecr, dockerregistry, bitbucket all done as prior WIs).
2. Read SDKv2 source files:
   - `data_serviceendpoint_npm.go`: exposes `url` as computed field (from `ep.Url`); base schema: project_id, service_endpoint_id, service_endpoint_name, authorization (TypeMap), description.
   - `data_serviceendpoint_sonarcloud.go`: only base schema (no extra fields).
3. Created framework implementations following the github/azurecr pattern (minimal: project_id, service_endpoint_id, service_endpoint_name; npm adds `url` from `ep.Url`).
4. Registered both in `framework_provider.go` DataSources().
5. Deregistered both from `provider.go` DataSourcesMap (replaced with comments).
6. Removed both from `provider_test.go` expectedDataSources (replaced with comments).
7. Updated `data_serviceendpoint_npm_test.go` and `data_serviceendpoint_sonarcloud_test.go` to use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` instead of `Providers: testutils.GetProviders()`.
8. Verified: `go build -mod=vendor .` clean, `TestProvider_HasChildDataSources` passes, `make test` all green, `golangci-lint --new-from-rev=main` 0 issues.

## What worked

- The existing framework data source files (github, azurecr, azurerm, bitbucket, dockerregistry) are the perfect template — just copy + adapt names and extra fields.
- npm's only extra field is `url` (from `ep.Url`). The framework model just adds `URL types.String \`tfsdk:"url"\`` and sets it from `ep.Url`.
- sonarcloud has no extra fields beyond base schema (project_id, service_endpoint_id, service_endpoint_name).
- `testutils.GetMuxedProviderFactories()` (with 'd', not `GetMuxProviderFactories()`) is the correct function for data source acceptance tests.

## What didn't work

_(nothing to record — all succeeded in iteration 0)_

## Open questions

_(none)_

## Notes for reflection

- WI-9 pattern is identical to prior WIs (5, 6, 7, 8) — all data source migrations follow the same 7-step recipe. The pattern is fully established.
- The npm data source's SDKv2 schema also has `authorization` (TypeMap) and `description`, but framework migrations in this initiative consistently omit those from the data model (keeping minimal: project_id, service_endpoint_id, service_endpoint_name + type-specific extras). This is consistent with github, bitbucket, dockerregistry framework data sources.
