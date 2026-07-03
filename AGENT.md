# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-04) — ALL ACs COMPLETED

**Gate failure before this iteration:** Directory `azuredevops/internal/service/accounts/` did not exist → `go test` exited non-zero with "directory not found".

**Actions taken:**
1. Created `azuredevops/internal/service/accounts/data_accounts.go` — `betterado_accounts` framework data source (terraform-plugin-framework pattern). Implements `Metadata`, `Schema`, `Configure`, `Read`. Uses `accounts.Client.GetAccounts()`. Optional `member_id` attribute filters by UUID. 404 → returns empty list.
2. Added `AccountsClient accounts.Client` to `AggregatedClient` struct in `azuredevops/internal/client/client.go`, initialized via `accounts.NewClient(ctx, connection)` in `GetAzdoClient()`.
3. Created `azuredevops/internal/service/accounts/data_accounts_test.go` with `TestDataAccountsSchema` (4 sub-tests: non-nil, TypeName, schema has accounts+member_id+id, struct compiles).
4. Registered `accounts.NewAccountsDataSource` in `framework_provider.go` `DataSources()`. NOT added to `provider.go`.
5. Created `docs/accounts-profile-gap-matrix.md` covering all fields from `Account` and `Profile` structs, with implemented/gap/out-of-scope status.

**Quality gate result:** `TestDataAccountsSchema` PASS (4/4 sub-tests). Committed as `feat(accounts): add betterado_accounts framework data source + gap matrix`.

## What worked

- `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts/` package exists with `accounts.Client` interface and `accounts.NewClient()`. Uses `GetClientByResourceAreaId` internally.
- Pattern for framework data source: look at `azuredevops/internal/service/release/datasource_release_folder_framework.go`.
- `TestDataAccountsSchema` test pattern: call `Metadata()` + `Schema()` directly using `datasource.MetadataRequest/Response` and `datasource.SchemaRequest/Response` — no mocking needed.
- Build tag `//go:build (all || data_accounts) && !exclude_data_accounts` needed in test file.

## What didn't work

- Pre-existing test failures in `serviceendpoint` (build mismatch on 2-var returns) and `identity` (error message format mismatch) — these are NOT related to this WI; confirmed by stashing my changes and seeing the same failures on the branch tip.

## Open questions

_(nothing blocking)_

## Notes for reflection

- `betterado_accounts` is a framework-only data source. AC2 confirmed it is registered ONLY in `framework_provider.go` `DataSources()`, not in `provider.go` SDKv2 DataSourcesMap.
- Profile API (`betterado_profile`) is deferred to a follow-on WI per gap matrix conclusions.
- All 3 ACs completed in one iteration.
