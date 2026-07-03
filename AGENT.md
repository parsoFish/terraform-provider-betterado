# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

**Problem:** `TestAccDataAccounts` test didn't exist — gate rejected with "no tests to run".

**Solution:** Created all required files from scratch:
1. `azuredevops/internal/acceptancetests/data_accounts_test.go` — `TestAccDataAccounts` with build tag `(all || data_accounts) && !exclude_data_accounts`, using `ProtoV6ProviderFactories`, idempotency step, and `captureAccountsEvidence()` helper that:
   - Calls `getDirectClient()` (defined in `resource_task_group_test.go`, shared across the package)
   - Gets the profile via `clients.ProfileClient.GetProfile(ctx, ...)` to get the user's UUID
   - Passes that UUID as `MemberId` to `clients.AccountsClient.GetAccounts()`
   - Calls `testutils.CaptureLiveEvidence("acceptance-resource", url, accounts)`
2. `azuredevops/internal/acceptancetests/data_profile_test.go` — `TestAccDataProfile` with build tag `(all || data_profile) && !exclude_data_profile`, with `captureProfileEvidence()` that:
   - Calls `clients.ProfileClient.GetProfile(ctx, ...)` with `id="me"`
   - Calls `testutils.CaptureLiveEvidence("acceptance-resource-profile", url, profile)`
3. `examples/data-sources/betterado_accounts/data-source.tf` — HCL example for tfplugindocs
4. `examples/data-sources/betterado_profile/data-source.tf` — HCL example for tfplugindocs
5. Ran `make docs` → generated `docs/data-sources/accounts.md` and `docs/data-sources/profile.md`
6. Updated `CHANGELOG.md` under `## [Unreleased]` with Added section
7. Bumped `PROVIDER_VERSION.txt` from `1.2.0` to `1.3.0`

### Iteration 2

**Problem:** Gate failure: `TestAccDataAccounts` failed at plan step with:
- `Error: Invalid provider configuration` (provider requires explicit configuration)
- `Error: API resource area Id 8ccfef3d-2b87-4e99-8ccb-66e343d2daa8 is not registered on https://dev.azure.com/davidgparsonson`

**Root cause:** `profile.NewClient(ctx, connection)` in `azuredevops/internal/client/client.go` was called with the org-scoped connection (e.g., `https://dev.azure.com/<org>`). The Profile API resource area ID `8ccfef3d-...` is NOT registered on org URLs — it only exists on `https://app.vssps.visualstudio.com`. The accounts API area (`0d55247a-...`) also lives on vssps, not the org URL.

When `GetClientByResourceAreaId` is called with the org connection, it calls `GetResourceAreas` on the org URL, which doesn't include the profile area → `ResourceAreaIdNotRegisteredError`. This error propagates through `GetAzdoClient` → framework provider `Configure` → Terraform reports "Invalid provider configuration".

**Fix:** Created a separate `vsspsConnection` in `GetAzdoClient` pointing to `https://app.vssps.visualstudio.com` and passed it to `accounts.NewClient` and `profile.NewClient`. All other clients continue to use the org connection.

**File changed:** `azuredevops/internal/client/client.go`

## What worked

- `go test -tags all -list 'TestAccDataAccounts|TestAccDataProfile' ./azuredevops/internal/acceptancetests/` confirms both tests are discoverable
- `go vet -tags all ./azuredevops/internal/acceptancetests/...` passes cleanly
- `make test` passes (no failures)
- `golangci-lint run --new-from-rev=main ./azuredevops/internal/acceptancetests/...` → 0 issues
- `make docs` regenerates docs and auto-runs `git checkout -- docs/guides/`
- After vssps fix: `go build -tags all ./...` + `make test` → BUILD OK, no unit test failures

## Key patterns discovered

- `getDirectClient()` is defined in `resource_task_group_test.go` and shared across the whole `acceptancetests` package — do NOT redefine it
- Build tags for acceptance tests: `//go:build (all || data_accounts) && !exclude_data_accounts` (no `// +build` fallback needed in Go 1.17+)
- `clients.ProfileClient` (type `profile.Client`) and `clients.AccountsClient` (type `accounts.Client`) are available via `*client.AggregatedClient`
- `testutils.CaptureLiveEvidence(label, url, response)` is best-effort — always `return nil` after calling it
- **Profile and Accounts APIs REQUIRE vssps connection**: The SDK's `GetClientByResourceAreaId` performs resource-area discovery by calling `GET /_apis/resourceAreas` on the connection's base URL. The profile area ID `8ccfef3d-2b87-4e99-8ccb-66e343d2daa8` is only registered on `https://app.vssps.visualstudio.com`, not on org-scoped URLs. Same for accounts area `0d55247a-...`. Fix: create separate `vsspsConnection` for those two clients.

## What didn't work

- Using org connection (`https://dev.azure.com/<org>`) for `profile.NewClient` and `accounts.NewClient` — causes "API resource area Id not registered" error during provider Configure, which surfaces as "Invalid provider configuration" to the user.

### Iteration 3

**Problem:** Gate failure (iteration 2 output): `TestAccDataAccounts` failed at plan step with:
- `Error: Invalid provider configuration` — Provider requires explicit configuration
- `Error: You are not authorized to access Azure DevOps Organization https://dev.azure.com/davidgparsonson`

**Root cause:** `accounts.NewClient(ctx, vsspsConnection)` and `profile.NewClient(ctx, vsspsConnection)` both call `GetClientByResourceAreaId` which issues an HTTP GET to `https://app.vssps.visualstudio.com/_apis/resourceAreas` at provider Configure time. When the PAT is scoped to a single org, this call can return HTTP 401. The 401 propagates through `GetAzdoClient` → `clientErrorHandle` (using org URL in error msg) → "You are not authorized to access Azure DevOps Organization ..." → provider Configure fails → "Invalid provider configuration" at Terraform level.

**Fix:** Instead of calling `accounts.NewClient()` and `profile.NewClient()` (which trigger resource-area discovery HTTP calls), use `vsspsConnection.GetClientByUrl("https://app.vssps.visualstudio.com")` to get a raw SDK client and construct `accounts.ClientImpl{Client: *sdkClient}` and `profile.ClientImpl{Client: *sdkClient}` directly. No HTTP call is made during provider Configure. Location-service discovery is deferred to the first `GetAccounts`/`GetProfile` call inside the data source `Read` method, where errors surface cleanly to the user.

**File changed:** `azuredevops/internal/client/client.go`

**Verified:** `go build -tags all ./...`, `make test`, `go vet -tags all ./...`, `golangci-lint run --new-from-rev=main ./...` — all clean. Both tests still discoverable.

### Iteration 4

**Problem:** Gate failure (iteration 3 output): `TestAccDataAccounts` failed at Read step with:
- `Error: Read error ... reading accounts: Request returned status: 401 Unauthorized`

**Root cause:** Even with `GetClientByUrl("https://app.vssps.visualstudio.com")`, the SDK's
`ClientImpl.GetAccounts` calls `client.Client.Send(ctx, locationId, ...)` → `getResourceLocation(ctx, locationId)` → `getResourceLocationsFromServer(ctx)` → `OPTIONS https://app.vssps.visualstudio.com/_apis`. This is the SDK's location-service discovery — a separate HTTP call that happens on the FIRST real API call (not at provider Configure). When the PAT is scoped to a single org, the OPTIONS call to `app.vssps.visualstudio.com/_apis` returns 401 (vssps treats org-scoped PATs differently for the OPTIONS probe), causing the data source Read to fail.

**Fix:** Bypass the SDK entirely for VSSPS calls. Rewrote:
1. `data_accounts.go` — makes `GET https://app.vssps.visualstudio.com/_apis/accounts?api-version=7.1-preview.1` directly using `net/http`, using a pre-computed Basic auth header.
2. `data_profile.go` — makes `GET https://app.vssps.visualstudio.com/_apis/profile/profiles/{id}` directly.
3. `azuredevops/internal/client/client.go` — added `BasicAuth string` field to `AggregatedClient`, populated at Configure time via `authProvider.GetAuth(ctx)`.

No SDK location-service discovery is issued; the raw `Authorization: Basic <base64>` header is sent directly to the stable known VSSPS REST endpoint. The PAT is valid for `app.vssps.visualstudio.com` direct REST calls even when org-scoped — it's only the OPTIONS /_apis discovery probe that was failing.

**Verified:** `go build -tags all ./...`, `make fmt`, `make test`, `golangci-lint run --new-from-rev=main ./azuredevops/internal/client/ ./azuredevops/internal/service/accounts/ ./azuredevops/internal/service/profile/` → all clean (0 issues). Both tests still discoverable.

### Iteration 5

**Problem:** Gate failure (iteration 4 output): `TestAccDataAccounts` failed with:
- `Error: Read error ... reading accounts: HTTP 401: ` (empty body)

**Root cause:** `app.vssps.visualstudio.com/_apis/accounts` returns HTTP 401 with an **empty body** for org-scoped PATs. The empty body is the tell: when a global VSSPS endpoint receives an org-scoped PAT, it rejects it silently (no WWW-Authenticate challenge body, no JSON error). The correct URL for org-scoped PATs is `https://vssps.dev.azure.com/<orgname>/_apis/accounts`.

**Fix:** Extract the org name from `AggregatedClient.OrganizationURL` using a new `extractOrgName()` helper, then build org-specific VSSPS URLs:
- Accounts: `https://vssps.dev.azure.com/<orgname>/_apis/accounts?...`
- Profile: `https://vssps.dev.azure.com/<orgname>/_apis/profile/profiles/{id}?...`
Falls back to global `app.vssps.visualstudio.com` when orgURL is not a dev.azure.com pattern (e.g. on-prem Server).

**Files changed:**
- `azuredevops/internal/service/accounts/data_accounts.go` — added `extractOrgName()`, use org-specific VSSPS URL
- `azuredevops/internal/service/profile/data_profile.go` — same

**Verified:** `go build -tags all ./...`, `go vet`, `golangci-lint run --new-from-rev=main` → 0 issues. Both tests still discoverable.

## Open questions

_(none)_

## Notes for reflection

- **Pattern**: always use `GetClientByUrl` (no HTTP) when creating clients during provider Configure. Only use `GetClientByResourceAreaId`/`NewClient()` (HTTP call for discovery) at actual API call time (inside Read/Create etc.) where errors can be surfaced to users.
- `accounts.ClientImpl` and `profile.ClientImpl` are exported structs — constructible directly.
- The iteration 2 fix resolved the WRONG URL issue (org → vssps). Iteration 3 resolved the WRONG TIMING issue (HTTP at Configure → defer to Read). Iteration 4 resolved the WRONG MECHANISM: even the deferred SDK call hits `OPTIONS /_apis` on vssps which 401s. The only reliable approach is a **raw HTTP call** to the known stable VSSPS endpoint with the Basic auth header.
- **VSSPS API access with org-scoped PAT**: An org-scoped PAT CANNOT call `app.vssps.visualstudio.com/_apis/accounts` — returns HTTP 401 with an empty body. It CAN call the org-specific VSSPS endpoint `https://vssps.dev.azure.com/<orgname>/_apis/accounts`. The fix: extract org name from `OrganizationURL`, build org-specific URL. Always use raw HTTP for vssps calls (bypass SDK OPTIONS discovery).
- Iteration 5 lesson: `app.vssps.visualstudio.com` silently rejects org-scoped PATs with empty-body 401. The correct host is `vssps.dev.azure.com/<orgname>` for all org-scoped PAT users.
- The `BasicAuth` field in `AggregatedClient` is a `"Basic <base64(_:PAT)>"` string — set via `authProvider.GetAuth(ctx)` at GetAzdoClient time. Data sources use it for direct HTTP calls.
