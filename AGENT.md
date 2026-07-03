# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed all ACs)

**All three ACs implemented and committed in a single iteration:**

1. **Created `resource_field_framework.go`** — full framework implementation with:
   - All schema attributes ported from SDKv2 `resource_field.go`
   - Custom inline validators (no `terraform-plugin-framework-validators` package — NOT vendored)
   - Inline plan modifiers (RequiresReplace for ForceNew fields, UseStateForUnknown for computed)
   - Identity field type override in Read (ADO API returns "string" for identity fields — must override to "identity")
   - `WriteOnly: true` on `restore` field (framework v1.19 supports this)
   - `supported_operations` as `schema.ListNestedAttribute` (Computed)
   - `ImportState` implemented (import by reference name)

2. **Deleted `resource_field.go`** (SDKv2 impl) — no unit test file existed

3. **Registered in `framework_provider.go`**: added `workitemtracking.NewFieldResource` to `Resources()`

4. **Updated `provider.go`**: removed `betterado_workitemtracking_field` from ResourcesMap, replaced with comment

5. **Updated `provider_test.go`**: removed `betterado_workitemtracking_field` from expectedResources, added comment

6. **Updated acceptance test** `resource_workitemtracking_field_test.go`:
   - All tests use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
   - `checkFieldDestroyed` uses direct ADO client (not `Meta()` — ProtoV6 doesn't expose it)
   - `TestAccWorkItemTrackingField_Basic` adds `captureFieldEvidence(tfNode)` check
   - CaptureLiveEvidence label: `"acceptance-resource-workitemtracking-field"`
   - No build tag added (original had none; `resource_workitemtrackingprocess_field_test.go` depends on `generateFieldName`, `fieldBasic`, `checkFieldDestroyed` from this file without a build tag — adding one breaks the no-tag build)

7. **Docs & examples**: created `examples/resources/betterado_workitemtracking_field/resource.tf`, regenerated `docs/resources/workitemtracking_field.md` via `make docs`

8. **CHANGELOG.md**: added entry under `## [Unreleased]`

## What worked

- **Inline validators** (not using `terraform-plugin-framework-validators` package — it is NOT in the vendor tree)
- **Inline plan modifiers** (same pattern as `resource_workitem_framework.go`)
- **`nolint:nilerr`** comment on best-effort returns to silence golangci-lint `nilerr` rule
- **gofumpt** required (stricter than gofmt); run `/home/parso/go/bin/gofumpt -w` on new files
- **Direct ADO client in CheckDestroy** (using `getFieldDirectClient()` pattern same as workitem test)

## What didn't work

- Adding a build tag to the test file — `resource_workitemtrackingprocess_field_test.go` (no build tag) calls `generateFieldName`, `fieldBasic`, `checkFieldDestroyed` from the field test file; adding a tag hides those symbols at compile time and causes a build failure

## Open questions

- Whether `description` should be `Optional+Computed` or just `Optional` with empty string default. Currently `Optional+Computed` with UseStateForUnknown to avoid spurious diffs.

## Notes for reflection

- `resource_workitemtrackingprocess_field_test.go` has no build tag and shares helpers from `resource_workitemtracking_field_test.go` — this cross-file helper sharing is a pattern to be careful about when adding build tags to test files.
- The `terraform-plugin-framework-validators` package is not vendored; always use inline validators for this project's framework resources.
