# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (initial implementation)
- Created `resource_feed_permission_framework.go` — full framework resource implementing Create/Read/Update/Delete/ImportState
- Registered `feed.NewFeedPermissionResource` in `framework_provider.go`
- Deregistered `betterado_feed_permission` from SDKv2 `provider.go` ResourcesMap
- Added `TestAccFeedPermissionFramework_basic` to `resource_feed_permission_test.go`

### Iteration 1 (display_name fix)
- Fixed display_name "provider still indicated an unknown value after apply" by ensuring display_name is always set to a known value post-create (fallback to empty string if ADO doesn't return it)

### Iteration 2 (current — verification)
- Confirmed all offline gates pass: `make test` green, golangci-lint 0 issues, terrafmt clean
- Confirmed `.forge/live-evidence/acceptance-resource.json` was written by live test run with real ADO URL
- Both ACs confirmed complete; no code changes needed this iteration

## What worked

- Framework resource pattern mirrors `resource_feed_framework.go` (WI-2) — same Configure/Model/Schema structure
- `useStateForUnknown()` / `requiresReplace()` plan modifiers from `framework_defaults.go`
- `getFeedDirectClient()` and `nilIfEmptyStr()` helper functions are defined in `resource_feed_framework_test.go` and accessible to `resource_feed_permission_test.go` within the same package
- `SharedFixtureProjectName` from `shared_fixtures.go` — avoids creating a new project (org is at 1000-project cap)
- Polling pattern via `retry.StateChangeConf` for ADO propagation delay
- `CaptureLiveEvidence("acceptance-resource", permURL, permissions)` — best-effort, returns error caller ignores
- Build tag: `//go:build (all || resource_feed_permission) && !exclude_feed_permission` — no new tag needed, reuses existing

## What didn't work

- Initially `display_name` caused "provider still indicated an unknown value after apply" — fixed by explicitly setting to empty string when ADO doesn't return it

## Open questions

_(none)_

## Notes for reflection

- The `GetMuxedProviderFactories()` function is in `testutils/mux_provider.go` (with "ed" suffix), distinct from `GetMuxProviderFactories()` in `testutils/commons.go` (without "ed"); these are two separate helpers
- Live evidence shows the test created a permission with role=reader matching a real ADO group descriptor
