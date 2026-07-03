# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)
- Created `azuredevops/internal/service/profile/` package with `data_profile.go`
- The data source stores `profileapi.Client` (not `*client.AggregatedClient`) to avoid needing a connection reference
- Added `ProfileClient profile.Client` field to `AggregatedClient` in `client.go` and wired it in `GetAzdoClient()`
- The profile SDK client uses `connection.GetClientByResourceAreaId()` internally (resource area ID: `8ccfef3d-2b87-4e99-8ccb-66e343d2daa8`)
- `CoreAttributes` in the `Profile` struct is `*map[string]CoreProfileAttribute` — keys are "DisplayName", "Email", "PublicAlias", "Avatar"
- Test file has NO `//go:build` tag (per WI constraints)
- All 8 subtests of `TestDataProfileSchema` pass offline
- Registered `profile.NewProfileDataSource` in `framework_provider.go` DataSources(); NOT in `provider.go`

## What worked

- Pattern: store the typed SDK client (`profileapi.Client`) directly in the data source struct, wired from `c.ProfileClient` in `Configure()`; this keeps the data source testable without needing a full AggregatedClient
- Adding `ProfileClient` to `AggregatedClient` follows the existing pattern (all other SDK clients are there)

## What didn't work

- Initially tried to call `d.client.GetConnection()` — `AggregatedClient` has no `GetConnection()` method; solution was to add `ProfileClient` to the aggregated client struct

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
