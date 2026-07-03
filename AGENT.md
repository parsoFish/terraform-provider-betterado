# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

- Read all existing SDKv2 source files for azurecr and dockerregistry, plus reference framework migrations (azurerm, generic).
- Created 4 new framework files:
  - `resource_serviceendpoint_dockerregistry_framework.go` — full CRUD, UsernamePassword auth scheme, computed auth map
  - `data_serviceendpoint_dockerregistry_framework.go` — lookup by ID or name
  - `resource_serviceendpoint_azurecr_framework.go` — full CRUD + ModifyPlan to preserve computed SPN/WIF fields for idempotency
  - `data_serviceendpoint_azurecr_framework.go` — lookup by ID or name
- Deregistered from `provider.go` ResourcesMap and DataSourcesMap (comment conventions per project pattern).
- Registered in `framework_provider.go` Resources() and DataSources().
- Updated `provider_test.go` expected resource + data source counts.
- Rewrote `resource_serviceendpoint_dockerregistry_test.go`:
  - Added build tag `(all || resource_serviceendpoint_dockerregistry) && !exclude_resource_serviceendpoint_dockerregistry`
  - Added `TestAccServiceEndpointDockerRegistry_basic` with `GetMuxedProviderFactories()`, idempotency step, CaptureLiveEvidence
  - Updated `TestAccServiceEndpointDockerRegistry_CreateAndUpdate` to `GetMuxedProviderFactories()`
- Rewrote `resource_serviceendpoint_azurecr_test.go`:
  - Added build tag
  - Switched all tests to `GetMuxedProviderFactories()`
- Build: `go build ./...` passes, `golangci-lint` 0 issues, `go test ./azuredevops/ ./azuredevops/internal/provider/...` pass.

## What worked

- Pattern: Copy the structure from `resource_serviceendpoint_azurerm_framework.go` and `resource_serviceendpoint_generic_framework.go` for plan modifiers, defaults, and the basic CRUD pattern.
- Pattern: Use `seAzureCRBuildEndpoint` returning `(*serviceendpoint.ServiceEndpoint, diag.Diagnostics)` instead of `resource.ModifyPlanResponse` - the `diag` package must be imported separately from `resource`.
- Pattern: `checkServiceEndpointDockerRegistryDestroyed` must use `getDirectClient()` (not `testutils.GetProvider().Meta()`) because `ProtoV6ProviderFactories` doesn't wire the SDKv2 singleton Meta.
- Pattern: `captureServiceEndpointDockerRegistryEvidence` label must be `"acceptance-resource-dockerregistry"` (matches forge gate checkpoint).
- Pattern: `testutils.GetMuxedProviderFactories()` (NOT `GetMuxProviderFactories()`; the task group data source uses `GetMuxedProviderFactories()`).
- `gofumpt` is required in addition to `gofmt` — run `gofumpt -w <files>` before committing new framework files.

## What didn't work

- `resource.ModifyPlanResponse` is NOT the right type for `diag.Diagnostics` return from a helper function. Use `diag.Diagnostics` from the `"github.com/hashicorp/terraform-plugin-framework/diag"` package.
- Using `testutils.GetProviders()` in tests for framework-migrated resources causes the resource type not to be found in the test provider.

## Open questions

- The azurecr resource's `seAzureCRBuildEndpoint` function may need refinement if the API returns scope in a different format for different regions. Monitor live test failures on this.
- The dockerregistry Read doesn't restore password from state since API never returns it — this is the same behavior as the SDKv2 version; the test should not assert on password equality in read-back.

## Notes for reflection

- The project uses `gofumpt` (stricter than `gofmt`) — golangci-lint enforces this. Always run `gofumpt -w` on new framework files.
- `EndpointAuthenticationScheme` constants (ServicePrincipal, WorkloadIdentityFederation) are defined in the same package — can be referenced directly.
