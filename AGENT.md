# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 — COMPLETE

Wrote `resource_task_group_framework.go` + `resource_task_group_framework_test.go`.
Both ACs pass: `go test -tags all -run TestTaskGroupFramework_Schema` → OK, `go build -mod=vendor .` → OK.
Committed as `feat(taskagent): add terraform-plugin-framework TaskGroupResource`.

## What worked

- **Inline static default helpers**: The `stringdefault`, `booldefault`, `int64default` sub-packages
  are NOT in the vendor. Implement static defaults inline by satisfying the `defaults.String`,
  `defaults.Bool`, `defaults.Int64` interfaces directly in the resource file.
- **Inline plan modifiers**: `stringplanmodifier` package is NOT in the vendor. Implement
  `requiresReplaceModifier` and `useStateForUnknownModifier` satisfying `planmodifier.String`
  directly in the resource file.
- **`resource.ResourceWithMetadata` does NOT exist** in this framework version. `Metadata()`
  is part of the base `resource.Resource` interface. Test just calls `r.Metadata()` directly.
- **`Default` field requires `Computed: true`**: The framework ValidateImplementation check
  rejects `Default` on non-Computed attributes. Use `Optional: true, Computed: true` for all
  fields that have a default.
- **`types.ListValueFrom` + typed slice**: Works well for nested objects using
  `types.ObjectType{AttrTypes: ...}` as element type, with Go struct slice.
- **`attr.Type` map** from `types.StringType`, `types.BoolType`, `types.Int64Type`,
  `types.MapType{ElemType: types.StringType}`, `types.ListType{ElemType: types.StringType}`.

## What didn't work

- Importing `stringplanmodifier`, `booldefault`, `stringdefault` — these packages do not exist
  in the vendor. Don't attempt to use them.
- Using `resource.ResourceWithMetadata` — doesn't exist in this framework version.

## Open questions

## Notes for reflection

- The framework vendor is missing the commonly-documented typed sub-packages (stringdefault, etc.).
  Future WIs should implement them inline or add the packages to vendor.
