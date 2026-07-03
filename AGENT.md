# Agent Memory — UWI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 6 (gate fix: testplans inconsistent result)

**Problem:** `TestAccProjectFeatures_roundtrip` failing with:
```
--- FAIL: TestAccProjectFeatures_roundtrip (5.73s)
    Step 2/2 error: Provider produced inconsistent result after apply
    .features["testplans"]: was cty.StringVal("enabled"), but now cty.StringVal("disabled").
```
Gate command: `go test -tags all -run TestAccProjectFeatures ./azuredevops/internal/acceptancetests/`

**Root cause:** ADO org lacks TestPlans license. `SetFeatureStateForScope` for `ms.vss-test-web.test`
returns HTTP 200 but feature stays `disabled`. Framework's `readFeaturesIntoModel` reads back `disabled`,
writes to state, causing plan=enabled vs state=disabled → Terraform "inconsistent result after apply".

**Fix applied (committed b154c6de):**
1. Test now uses `artifacts`+`boards` (both license-free) instead of `testplans`+`artifacts`.
2. `applyFeatureStates` checks `result.State` from `SetFeatureStateForScope` — explicit error if returned
   state doesn't match requested state (prevents silent mismatch turning into opaque Terraform panic).

## What worked

- Using `artifacts` and `boards` for feature toggle tests (both license-free, reliably togglable on any ADO project type).
- Checking `SetFeatureStateForScope` return value to detect silent API license failures.
- `smokeResolveProject` correctly resolves existing projects without creation (avoids 1000-project cap).

## What didn't work

- Testing `testplans` feature: fails silently when org lacks TestPlans license — API returns 200 but ignores the change.
- The original test used testplans (mirroring SDKv2 tests) but the license restriction wasn't surfaced in prior test infrastructure.

## Open questions

- Does `boards` have any restrictions on public projects? Should be universally available, but worth monitoring.

## Notes for reflection

- ADO feature management API returns HTTP 200 even when a license restriction prevents a state change — always verify by checking the returned `ContributedFeatureState`, don't assume success.
- `SetFeatureStateForScope` response body contains the actual resulting state — use it.
- `testplans` requires Basic+TestPlans or Visual Studio subscription; `artifacts`, `boards`, `repositories`, `pipelines` do not.
