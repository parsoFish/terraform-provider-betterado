# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

- Wrote `getReleaseDefinitionPermissionsResource` helper using `schema.TestResourceDataRaw` with the raw map, NOT `d.Set` after construction. This is critical for TypeInt zero-value detection (see below).
- Wrote `TestReleaseDefinitionPermissions_ProjectScopedToken` (AC1): passes a raw map with only `project_id`, calls `createReleaseDefinitionToken`, asserts token == projectID with no slash.
- Wrote `TestReleaseDefinitionPermissions_TokenEdgeCases` (AC2): table-driven with definitionID=0 and definitionID=99999. Initially failed because production code used `GetOk` which returns false for zero.
- Fixed production `createReleaseDefinitionToken` to use `GetOkExists` (with `//nolint:staticcheck // SA1019`) instead of `GetOk` for `release_definition_id`. This matches the existing pattern in `resource_group.go`.

## What worked

- `schema.TestResourceDataRaw(t, schema, raw)` where `raw` is passed with `"release_definition_id": 0` correctly sets the field — the diff is computed from the raw config, so `GetOkExists` returns `true`.
- `GetOkExists` with `//nolint:staticcheck` suppression is the established project pattern for distinguishing unset TypeInt from zero-value TypeInt.
- `definitionID < 0` as the sentinel for "omit this field" in the helper.

## What didn't work

- `d.Set("release_definition_id", 0)` after `schema.TestResourceDataRaw(t, schema, nil)` does NOT make `GetOk` or `GetOkExists` return true for that field — the underlying diff doesn't track it as "changed from config". Must pass it in the raw map to `TestResourceDataRaw`.
- `GetOk` for TypeInt zero value: `GetOk` explicitly checks `!reflect.DeepEqual(value, zero)` and returns false for 0. Cannot use `GetOk` to detect "release_definition_id set to 0".

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

- The `GetOk` vs `GetOkExists` distinction for Optional TypeInt fields is a common footgun in Terraform SDK v2. Worth capturing as a project pattern/antipattern in brain/cycles/themes/.
- Both ACs are satisfied; all tests pass; committed in ec272374.
