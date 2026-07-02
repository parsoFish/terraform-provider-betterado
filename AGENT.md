# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration — WI-3 completes)

- CHANGELOG.md already had `## [Unreleased]` with `### FEATURES` section documenting both dashboard and extension migrations (written by prior WI-1/WI-2 iterations).
- The WI spec required `### Changed` and `### Added` subsections in addition to the detailed FEATURES entries.
- Added `### Changed` entries for both migrations and `### Added` entries for the two gap-matrix docs files under `## [Unreleased]`, preserving the existing `### FEATURES` block.
- Bumped `PROVIDER_VERSION.txt` from `1.2.0` → `1.2.1` (patch increment).
- `make test` passed (gofmt clean, all non-acceptance tests passed).
- Committed as `docs(changelog): add WI-3 Unreleased entries for dashboard/extension migration; bump patch to 1.2.1`.

## What worked

- The CHANGELOG.md `## [Unreleased]` section already existed with full detail from WI-1/WI-2 work; only needed to prepend the canonical `### Changed` / `### Added` entries required by the WI spec.
- `make test` passes trivially for docs-only changes (no Go code changed).

## What didn't work / Gate rejections

### Iteration 1 (post-commit gate failure)
- Gate ran `go test -count=1 ./azuredevops/internal/service/dashboard/...` → **REJECTED** with "no test files" — the forge gate-tightening rule treats "no tests found" as a gate failure, even though the WI spec said it expected existing unit tests.
- Root cause: dashboard package had no `_test.go` files at all; gate requires actual tests to run.

### Iteration 2 (fix)
- Wrote `resource_dashboard_framework_test.go` with 16 unit tests covering:
  - All three plan modifiers (dashboardUseStateForUnknownString, dashboardRequiresReplaceString, dashboardUseStateForUnknownInt64Struct) — Description, MarkdownDescription, and logic paths
  - DashboardResource.Metadata → type name
  - DashboardResource.Schema → required/optional/computed attribute presence
  - NewDashboardResource interface compliance (ResourceWithConfigure, ResourceWithImportState)
- **No build tags** — the gate runs without `-tags all`, so tests must be unconditional.
- Type assertions via `r.(resource.ResourceWithConfigure)` — cannot use `var _ resource.ResourceWithConfigure = r` when r is typed as `resource.Resource`.
- `go test -count=1 ./azuredevops/internal/service/dashboard/...` → 16 tests PASS, `ok ... 0.003s`.
- `make test` (full module) → 0 FAIL lines.

## Open questions

_(none)_

## Notes for reflection

- The WI spec said "The gate runs existing unit tests from WI-1" but no unit tests existed — gate-tightening rejected this. Wrote new tests to satisfy the gate.
- Key: gate runs WITHOUT `-tags all`, so tests need no build tags to run under the live gate command.
