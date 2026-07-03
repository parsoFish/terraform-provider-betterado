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

## Open questions

_(none)_

## Notes for reflection

- All ACs complete as of iteration 2. The main blocker was the vssps connection issue for profile/accounts clients.
