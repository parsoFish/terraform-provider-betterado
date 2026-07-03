# Agent Memory — UWI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (UWI-2 first run)

**Problem:** `go build ./...` failed with two undefined function errors:
1. `findServiceEndpointByName` in `data_serviceendpoint_generic_v2_framework.go:149`
2. `validateScopeLevel` in `resource_serviceendpoint_azurerm_framework.go:438`

**Fixes applied:**
1. Replaced `findServiceEndpointByName(ctx, d.client, endpointName, projectID)` with direct `GetServiceEndpointsByNames` API call — same pattern used by `data_serviceendpoint_azurerm_framework.go`, `data_serviceendpoint_bitbucket_framework.go`, etc.
2. Added `validateScopeLevel(scopeMap map[string][]string) error` function to `resource_serviceendpoint_azurerm_framework.go` — ported directly from the deleted SDKv2 file (was at line 513 in main branch). Validates at least one of subscription/managementGroup scope is set.

**Status:** All 3 quality gate conditions pass after these fixes.

## Prior iteration work (before UWI-2 iter 0)

- All 32 SDKv2 files deleted (confirmed via `git diff --name-status main..HEAD | grep "^D.*serviceendpoint"`)
- All 32 framework files created and registered in framework_provider.go
- `terraform-plugin-framework-validators` added as dependency (go.mod + vendor/)
- `stringvalidator.*` validators wired in azurerm and other framework files
- 4 live-evidence files present in `.forge/live-evidence/`
- Provider registrations: framework_provider.go has all 32, provider.go SDKv2 map has only comments

## What worked

- `GetServiceEndpointsByNames` API pattern for data source name-based lookup — consistent across all data source framework files
- Porting `validateScopeLevel` directly from old SDKv2 file to framework file works cleanly

## What didn't work

- `findServiceEndpointByName` was a hallucinated/placeholder helper never defined anywhere — don't use it; use `GetServiceEndpointsByNames` inline
- When SDKv2 files are deleted, any helper functions they contained that are called by framework files must be moved to the framework file (or commons.go); they don't auto-migrate

## Open questions

_(none blocking)_

## Notes for reflection

- When migrating from SDKv2 to framework, check if the SDKv2 file exports/defines helpers called by the new framework file; those must be explicitly ported
- The `findServiceEndpointByName` anti-pattern: new framework data sources should always use the inline `GetServiceEndpointsByNames` API call pattern
