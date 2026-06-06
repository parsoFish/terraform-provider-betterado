# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

**Iteration 1 (complete — all 3 ACs satisfied, quality gate PASS):**
- Created `azuredevops/internal/acceptancetests/shared_fixtures.go` with `SharedReleaseFixture(t)` helper
- Created `azuredevops/internal/acceptancetests/shared_fixtures_test.go` with `TestSharedReleaseFixture` smoke test
- Quality gate `go test -tags all -count=1 -run TestSharedReleaseFixture ./azuredevops/internal/acceptancetests/` → PASS (live ADO)
- Offline skip verified: `TF_ACC="" go test ... -run TestSharedReleaseFixture` → SKIP

## What worked

- **Direct client construction**: Built `*client.AggregatedClient` directly from env vars using `azuredevops.NewAuthProviderPAT(pat)` + `client.GetAzdoClient(authProvider, orgURL)` rather than `testutils.GetProvider().Meta()`. The latter is nil in plain `TestXxx` functions not wrapped in `resource.Test`. Direct construction is cleaner.
- **`_test.go` for the smoke test**: Putting `TestSharedReleaseFixture` in `shared_fixtures_test.go` (not `shared_fixtures.go`) is required for Go to recognize it as a test function. Non-`_test.go` files don't expose `Test*` funcs to the test runner.
- **`canonicalApproval()` automated-approver pattern**: UUID `00000000-0000-0000-0000-000000000000` + `IsAutomated: true` is the established pattern in this codebase for ADO-valid automated approvals.
- **`VariableGroupParameters.Variables` type**: Must be `*map[string]interface{}`, not `*map[string]taskagent.VariableValue`. The ADO SDK uses interface{} for flexible JSON types.
- **gofmt required**: `gofmt -w` was needed; the aligned struct literal spacing in `canonicalStage` was flagged.

## What didn't work

- `testutils.GetProvider().Meta().(*client.AggregatedClient)` panics in standalone `TestXxx` functions because `Meta()` is only set after `resource.Test` configures the provider. Don't use this pattern outside of `resource.Test` callbacks.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
