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

## Iteration 3 (completed)

**Root cause fixed:** Gate failure was `.server_url: was cty.StringVal(""), but now cty.StringVal("https://management.azure.com/")`.

**Root cause:** `server_url` schema had `Default: seAzureRMDefaultString("")`. This pre-planned the attribute as `""`. After Create, `flattenFromServiceEndpoint` set `state.ServerURL` from `ep.Url` which is `"https://management.azure.com/"` (computed from `environment = "AzureCloud"`). Framework then detected: plan = `""`, returned = `"https://management.azure.com/"` → "provider produced inconsistent result after apply".

**Fix:** Removed `Default: seAzureRMDefaultString("")` from `server_url` schema. Added `seAzureRMUseStateForUnknown()` plan modifier. Now for Optional+Computed without a default: plan is `null` on Create (user didn't set it in config) → framework allows provider to fill in the computed value → no inconsistency.

**Idempotency:** On re-plan, Optional+Computed preserves prior state value (`"https://management.azure.com/"`) in plan → plan matches state → no diff.

**Build:** `go build -tags all ./...` passes. Quality gate passes.

## Iteration 4 (completed)

**Root cause fixed:** Gate failure was `.service_principal_id: was cty.StringVal("9be56814..."), but now cty.StringVal("8124569d...")` at Step 3/3 (the second apply/update step).

**Root cause:** `UseStateForUnknown` on `service_principal_id` froze the OLD SPN-ID in the plan during Update. After Update, `flattenFromServiceEndpoint` read the API response which returned the NEWLY-configured SPN-ID. Framework detected plan ≠ returned → "provider produced inconsistent result after apply".

**Fix:** Implemented `resource.ResourceWithModifyPlan` on `ServiceEndpointAzureRMResource`. In `ModifyPlan`:
1. If `credentials[0].serviceprincipalid` is known, propagate it directly to `service_principal_id` in the plan.
2. Otherwise (Automatic/MSI/WIF or unknown credentials), fall back to prior state value.

Removed `seAzureRMUseStateForUnknown()` from `service_principal_id` schema attribute (ModifyPlan supersedes it).

Also applied `gofumpt -w` to all 5 framework files to resolve golangci-lint gofumpt issues (0 issues after).

**CI gate:** `make test` (0 FAIL) + `golangci-lint run ./azuredevops/...` (0 issues) + `make terrafmt-check` (clean).

**Key lesson:** When a computed-only attribute mirrors a user-configurable input (credentials.serviceprincipalid → service_principal_id), UseStateForUnknown is wrong on it — state can be stale after update. Use ModifyPlan to derive the plan value from the user's config.

## Remaining work (iteration 5+)

- **AC3**: Live gate must re-run `TestAccServiceEndpointAzureRm_CreateAndUpdate` with the iteration 4 fix. The test flow is:
  1. Step 1: Create with configFirst (serviceprincipalidFirst) → ModifyPlan sets plan.ServicePrincipalID = serviceprincipalidFirst → API returns same → match ✓
  2. Step 2 (idempotency PlanOnly): credentials unchanged → ModifyPlan sets same ID → no diff ✓
  3. Step 3: Update with configSecond (serviceprincipalidSecond) → ModifyPlan sets plan.ServicePrincipalID = serviceprincipalidSecond → API returns same → match ✓

## Open questions

- If there are further gate failures, suspect the `credentials` block in flattenFromServiceEndpoint — the API may return a different casing or format for SPN IDs.

## Notes for reflection

- The "inline plan modifier per file" pattern is now established for this repo.
- ModifyPlan is the right tool when a computed attribute must be derived from user config values at plan time (not just preserved from state).
- gofumpt stricter than gofmt: always run gofumpt after editing framework files or will fail golangci-lint.
