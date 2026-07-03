# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What was done (iteration 0 — completed all ACs)

All three ACs were completed in a single iteration:

### AC1 — Framework resource implementation

- **Created** `azuredevops/internal/service/workitemtracking/resource_field_framework.go`:
  - Full terraform-plugin-framework implementation of `betterado_workitemtracking_field`
  - All schema parity: `name`, `reference_name`, `type`, `description`, `usage`, `read_only`, `can_sort_by`, `is_queryable`, `is_identity`, `is_picklist`, `is_picklist_suggested`, `picklist_id`, `is_locked`, `url`, `restore` (WriteOnly), `supported_operations` (ListNestedAttribute, Computed)
  - Custom inline validators: `notWhitespaceValidator`, `lengthBetweenValidator`, `doesNotMatchValidator`, `oneOfValidator`, `isUUIDValidator`
  - `RequiresReplace` plan modifiers for all ForceNew fields
  - `UseStateForUnknown`-style plan modifiers for all computed attributes (prevents spurious diffs)
  - Identity field type override: API returns `string`, resource returns `identity` (preserved from SDKv2)
  - `WriteOnly: true` for `restore` field (framework v1.8+)
  - `supported_operations` as `ListNestedAttribute` with computed `name`/`reference_name` nested attrs
  - `ImportState` via reference name
- **Registered** `workitemtracking.NewFieldResource` in `framework_provider.go` `Resources()`
- **Removed** `"betterado_workitemtracking_field"` from `provider.go` `ResourcesMap`; added comment
- **Updated** `provider_test.go` `expectedResources` to remove entry; added comment

### AC2 — SDKv2 deletion

- **Deleted** `azuredevops/internal/service/workitemtracking/resource_field.go`
- No `resource_field_test.go` existed (confirmed with Glob before deleting)
- `go build -mod=vendor .` passes with zero errors
- `make test` all green

### AC3 — Acceptance test update

- **Updated** `resource_workitemtracking_field_test.go`:
  - All 17 test functions use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
  - `checkFieldDestroyed` builds direct ADO client (no `Meta()` with ProtoV6)
  - `captureFieldEvidence` helper calls `CaptureLiveEvidence("acceptance-resource-workitemtracking-field", url, field)`
  - `TestAccWorkItemTrackingField_Basic` includes `captureFieldEvidence(tfNode)` in its `Check` step

### Standing ACs verified

- `make test` passes
- `golangci-lint run --new-from-rev=main ./azuredevops/...` = 0 issues
- `make terrafmt-check` passes
- `docs/resources/workitemtracking_field.md` regenerated — all schema attributes documented
- `examples/resources/betterado_workitemtracking_field/resource.tf` exists
- `CHANGELOG.md` has draft bullet under `## [Unreleased]`

## Key patterns learned

1. **`checkFieldDestroyed` must not use `Meta()`** — ProtoV6ProviderFactories don't expose it; must build direct client from env vars.
2. **`WriteOnly` fields** must be in the tfsdk model but cannot be read back from state or API; just read from plan on Create.
3. **Custom validators inline** (not from `stringvalidator` package) work fine for simple cases that don't need framework's built-in validators.
4. **`UseStateForUnknown`** pattern for computed attributes: skip if `PlanValue.IsUnknown() == false`; skip if `StateValue.IsNull()` (first apply).
5. **`supported_operations`**: Use `ListNestedAttribute` with `Computed: true`; build using `types.ObjectValue()` + `types.ListValue()`.

## What didn't work

_(no dead ends — all ACs completed in one iteration)_

## Open questions

_(none)_
