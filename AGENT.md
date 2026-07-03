# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (completed)

- Created 5 new framework files:
  - `resource_serviceendpoint_azurerm_framework.go` — full CRUD for azurerm (ServicePrincipal/MSI/WIF)
  - `data_serviceendpoint_azurerm_framework.go` — data source lookup by ID or name
  - `resource_serviceendpoint_aws_framework.go` — UsernamePassword scheme
  - `resource_serviceendpoint_azure_service_bus_framework.go` — None scheme
  - `resource_serviceendpoint_gcp_terraform_framework.go` — JWT scheme
- Modified `framework_provider.go` to register all 4 resources + 1 data source
- Modified `provider.go` to deregister (comment out) all 4 resources + 1 data source
- Modified `provider_test.go` to remove deregistered resources from expected lists
- Rewrote `resource_serviceendpoint_azurerm_test.go` to use `GetMuxedProviderFactories` + `SharedFixtureProjectName`
- Build: `go build -tags all ./...` passes clean
- Gate: `go test release/... taskagent/...` passes

## What worked

- **Inline plan-modifier and default pattern**: The vendored `terraform-plugin-framework` does NOT include sub-packages like `stringplanmodifier`, `stringdefault`, `booldefault`. Must define inline struct types implementing `planmodifier.String`, `defaults.String`, `defaults.Bool` in each file. Follow pattern from `resource_serviceendpoint_generic_framework.go`.
- **Name-based service endpoint lookup**: Use `GetServiceEndpointsByNames` with `GetServiceEndpointsByNamesArgs{EndpointNames: &[]string{name}}` — NOT `GetServiceEndpoints`/`GetServiceEndpointsArgs.EndpointNames` (that field doesn't exist).
- **Sensitive fields on Read**: For secrets (passwords, keys, connection strings) that the API never returns, preserve the value from the current state by reading `req.State` into the model before setting computed values.
- **`getDirectClient()` pattern for CheckDestroy**: When using `ProtoV6ProviderFactories`, the SDKv2 provider singleton's Meta is not wired. Use `getDirectClient()` (builds `*client.AggregatedClient` from env vars) in `CheckDestroy` funcs.
- **`SharedFixtureProjectName`**: Org is at the 1000-project cap. Use `data "betterado_project"` lookup of `"betterado-standing-demo"` — never create new projects.
- **Build tag for acceptance tests**: Use `//go:build (all || resource_serviceendpoint_azurerm) && !exclude_resource_serviceendpoint_azurerm` pattern (matching existing tests in the repo).

## What didn't work

- **`stringplanmodifier`/`stringdefault`/`booldefault` sub-packages**: Not in vendor directory; importing them causes `cannot find module providing package` errors. Must use inline types.
- **`GetServiceEndpointsArgs.EndpointNames`**: Field does not exist on that struct. Wrong method. Use `GetServiceEndpointsByNamesArgs`.

## Pre-existing test failures (NOT introduced by WI-3)

The `serviceendpoint` unit test package has build failures because many `expandServiceEndpoint*` functions return 1 value but old test files expect 2 values. These existed before iteration 1. The `graph` package also has a pre-existing test failure. These are out of scope.

## Iteration 2 (completed)

**Root cause fixed:** Gate failure was `"provider still indicated an unknown value for workload_identity_federation_issuer/subject after apply"`. Root cause: `flattenFromServiceEndpoint` only set WIF fields when `scheme == WorkloadIdentityFederation`. For `ServicePrincipal` scheme (what the test uses), these computed-only fields were never set → remained unknown in the state → Terraform error.

**Fix applied:** In `flattenFromServiceEndpoint`, initialise all three computed-only fields (`WorkloadIdentityFederationIssuer`, `WorkloadIdentityFederationSubject`, `ServicePrincipalID`) to empty string `""` before reading API params, if they are currently unknown. This ensures they are always a known value after apply, regardless of auth scheme. For WIF scheme, the actual API values then overwrite the empty string.

**Build:** `go build -tags all ./...` passes. Quality gate (`go test release/... taskagent/...`) passes.

**Idempotency reasoning:** After re-plan, plan for these fields triggers `UseStateForUnknown`: plan is unknown → state is `""` (not null) → plan becomes `""` → matches state → no diff. ✓

## Remaining work (iteration 3+)

- **AC3**: Live gate must re-run `TestAccServiceEndpointAzureRm_CreateAndUpdate` with the fix applied. The test should now pass — WIF fields will have known empty string values for `ServicePrincipal` scheme.

## Open questions

- Do the azurerm test env vars (`AZDO_AZR_SPN_ID`, `AZDO_AZR_SPN_KEY`) exist in the live gate environment? The HCL template uses them as Go env vars passed in from test setup.

## Notes for reflection

- The "inline plan modifier per file" pattern is now established for this repo. A shared helper package could reduce duplication across framework files — future refactor candidate.
