# Agent Memory — UWI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (2026-07-04)

**AC1 — Validator parity (DONE):**
- `data_security_namespace_token_framework.go`: Added `stringvalidator.RegexMatches(UUID regex)` on `namespace_id`, `stringvalidator.LengthAtLeast(1)` on `namespace_name`
- `data_securityrole_definitions_framework.go`: Added `stringvalidator.LengthAtLeast(1)` on `scope`
- Import path: `github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator` + `github.com/hashicorp/terraform-plugin-framework/schema/validator` + `regexp`

**AC2 — SDKv2 file deletion (DONE):**
- Deleted 19 SDKv2 files total (14 permissions + 4 security + 2 securityroles)
- NOTE: WI says "13 permissions" but there were actually 14 (resource_release_definition_permissions.go was also migrated to framework)
- Created 3 new helper files:
  - `azuredevops/internal/service/security/namespace_token_helpers.go`: `TokenTemplate`, `namespaceTokenTemplates`, `createClassificationNodeToken`, `getQueryIDsFromPath`, `getQueryName`
  - `azuredevops/internal/service/security/security_helpers.go`: `resolveIdentityDescriptor`
  - `azuredevops/internal/service/permissions/permissions_helpers.go`: `transformPath`, `getQueryIDsFromPath` (uses go-linq)

**AC3 — Demo artifact fixes (DONE):**
- demo.json acEvaluations entries 4 and 9: replaced `c0ac3757`/`2026-07-02T12:44:24Z` with `6ddb680c`/`2026-07-03T02:14:27Z`
- DEMO.md rows 4 and 9: same fix

**AC5 — Vendor/push (DONE):**
- Vendor was already clean; `go mod vendor` produced no changes
- All changes pushed to origin as commit `7394d78c`

**AC4 — Live evidence (BLOCKED):**
- Requires TF_ACC=1 with real AZDO_ORG_SERVICE_URL + AZDO_PERSONAL_ACCESS_TOKEN
- Cannot run live tests without credentials in this environment

## What worked

- `stringvalidator.RegexMatches(UUID regex)` is the correct framework replacement for `validation.IsUUID`
- `stringvalidator.LengthAtLeast(1)` is the correct framework replacement for `validation.StringIsNotEmpty`
- Extract-then-delete approach for SDKv2 files with shared helpers works cleanly
- The permissions package `getQueryIDsFromPath` uses `go-linq` (github.com/ahmetb/go-linq) for filtering — must keep that import

## What didn't work

_(nothing failed in this iteration)_

## Open questions

- AC4: Requires live AZDO credentials. If the gate runner has TF_ACC=1 and AZDO credentials configured, AC4 may be verifiable by the orchestrator externally.

## Notes for reflection

- WI AC2 count "13 permissions resource_*.go" was off by 1 — there were actually 14 permissions SDKv2 files (release_definition_permissions was also migrated to framework but not counted in the WI). All 14 were deleted.
- The `data_security_namespace_token.go` → `namespace_token_helpers.go` rename was detected by git as a rename (72% similarity).
