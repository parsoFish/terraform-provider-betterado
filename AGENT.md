# Agent Memory — UWI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration — completed all ACs)

**AC1 — Validators in resource_workitem_framework.go:**
- Added imports: `github.com/google/uuid`, `github.com/hashicorp/terraform-plugin-framework/path`, `github.com/hashicorp/terraform-plugin-framework/schema/validator`
- Added compile-time check: `_ resource.ResourceWithConfigValidators = (*WorkItemResource)(nil)`
- Added validator types inline (same pattern as resource_workitemquery_framework.go and resource_field_framework.go):
  - `wiIsUUIDValidator{}` — ValidateString using uuid.Parse
  - `wiNotWhitespaceValidator{}` — ValidateString using strings.TrimSpace
  - `wiTagsSizeFloorValidator{}` — ValidateSet checking len(Elements()) >= 1
  - `wiParentIDAtLeastValidator{}` — ValidateInt64 checking ValueInt64() >= 1
  - `wiConflictingFieldsValidator{}` — ValidateResource (resource.ConfigValidator) for custom_fields vs additional_fields_json
- Added `ConfigValidators()` method returning wiConflictingFieldsValidator{}
- Wired validators into Schema() attributes

**AC2 — Delete classification.go / classification_test.go:**
- Confirmed no external callers (grep found functions only self-referenced within the package)
- Deleted both files; `utils/` directory is now empty

**AC3 — Convert acceptance tests to SharedFixtureProjectName:**
- Removed all `betterado_project` resource creation from TestAccWorkItem_* functions
- Replaced with `data "betterado_project" "test" { name = SharedFixtureProjectName }` pattern
- Created new shared helper functions: workItemTagUpdateShared, workItemParentShared, workItemParentDeleteShared, workItemParentUpdateShared, workItemDescriptionShared, workItemDescriptionNoneShared, workItemAdditionalFieldsShared
- All 8 test functions now use shared fixture project

**AC4 — workitemquery_folder CaptureLiveEvidence:**
- Already wired in TestAccWorkItemQueryFolder_UnderArea via captureQueryFolderEvidence(tfNode) call
- No changes needed

**AC5 — Version/CHANGELOG:**
- Bumped PROVIDER_VERSION.txt from 1.2.1 → 1.9.1 (main was 1.9.0)
- Updated CHANGELOG.md [Unreleased] section with validator parity, classification helper deletion, acceptance test conversion

## What worked

- **Inline validator pattern**: No external validator library needed; all other framework files in the package use inline validator structs. Copied the exact pattern from resource_workitemquery_framework.go.
- **Set validator**: Use `ValidateSet()` method signature (not `ValidateString`) with `validator.SetRequest`/`validator.SetResponse` — found in schema/validator/set.go.
- **Int64 validator**: Use `ValidateInt64()` method with `validator.Int64Request`/`validator.Int64Response` — found in schema/validator/int64.go.
- **Resource-level ConflictsWith**: Implement as `resource.ConfigValidator` via `ValidateResource()` method, then return from `ConfigValidators()` method.

## What didn't work

_(none — all approaches succeeded on first try)_

## Open questions

_(none)_

## Notes for reflection

- The `wiTagsSizeFloorValidator` adds a size-floor of 1, but the ADO tags test removes all tags by passing an empty set `[]`. This might cause a validator failure if someone tries to set `tags = []`. The validator only fires when the attribute is non-null (i.e. the user specifies `tags`); if they set `tags = []` the validator will reject it with "must contain at least 1 element". This mirrors the SDKv2 MinItems:1 behavior but may be surprising. The workItemTagUpdateShared step tests removing tags by going back to `workItemBasicShared` (which has no `tags` attribute at all, not `tags = []`).
