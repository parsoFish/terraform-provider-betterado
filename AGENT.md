# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-03)

**Gate failure at start:** `TestAccWorkItemQuery_UnderArea` failed with "organization already has 1000 projects". Root cause: test was using `testutils.GenerateResourceName()` to create new projects with `betterado_project.project`, hitting the 1000-project ADO org cap.

**Fix strategy:**
1. Created framework resources (resource_workitemquery_framework.go, resource_workitemquery_folder_framework.go)
2. Removed SDKv2 implementations (deleted source + test files)
3. Registered framework resources in framework_provider.go, deregistered from provider.go
4. Rewrote acceptance tests to use:
   - `GetMuxedProviderFactories()` instead of `GetProviders()`/`Providers:`
   - `SharedFixtureProjectName` ("betterado-standing-demo") via `data "betterado_project" "test"` lookup
   - `CaptureLiveEvidence("acceptance-resource-workitemquery", url, q)` + matching folder label
   - Idempotency step `PlanOnly: true, ExpectNonEmptyPlan: false`
   - `CheckDestroy` functions that build their own ADO client from env vars (ProtoV6ProviderFactories doesn't expose Meta)

**Build:** `go build -mod=vendor .` ✅ Clean
**Offline tests:** `go test -tags all ./azuredevops/...` — all packages green except pre-existing failures in `serviceendpoint` (assignment mismatch in tests, not my code) and `graph` (same pattern). Neither is related to WI-4.
**provider_test.go:** `TestProvider_HasChildResources` ✅ passes

## What worked

- Inline plan modifiers (wiqStringRequiresReplace, wiqStringUseStateForUnknown, wiqBoolUseStateForUnknown) instead of sub-packages like stringplanmodifier (not vendored)
- Inline validators (wiqIsUUIDValidator, wiqNotEmptyValidator, wiqAreaValidator, wiqWiqlLengthValidator) instead of terraform-plugin-framework-validators (not vendored)
- Custom ConfigValidator (wiqExactlyOneOfValidator, wiqfExactlyOneOfValidator) implementing resource.ConfigValidator interface
- SharedFixtureProjectName pattern avoids project creation → avoids 1000-project cap

## What didn't work

- First attempt had bad imports (boolplanmodifier, stringplanmodifier, frameworkvalidator packages not in vendor)
- resourcevalidator.Default was a nil type — removed that spurious import

## Open questions

_(none blocking)_

## Notes for reflection

- The 1000-project ADO org cap requires ALL acceptance tests that used `betterado_project` resource to use SharedFixtureProjectName pattern instead
- The wiqfExactlyOneOfValidator (folder variant) could share logic with wiqExactlyOneOfValidator but it's simpler to keep separate since they're in the same package
