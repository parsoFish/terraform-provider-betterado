# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (completed all ACs)

1. Read WI-1 spec: needed `resource_release_folder.go`, provider registration, expand/flatten, round-trip test.
2. Examined `resource_release_definition.go` for idiom reference — no build tag on that file, but WI requires one for the folder resource.
3. Checked SDK `release.Folder` struct: fields are `Path *string`, `Description *string`, `CreatedBy`, `CreatedOn`, `LastChangedBy`, `LastChangedDate`. No project_id field — project_id is a Terraform-side attribute only, passed as a query param to the API.
4. Confirmed `converter.String()` is available in the converter package.
5. First attempt at CRUD stubs used `interface{}` types — go vet caught signature mismatch. Fixed to use `context.Context` + `diag.Diagnostics`.
6. Wrote `TestReleaseFolder_ExpandFlatten_Roundtrip` in `resource_release_folder_test.go`.
7. Quality gate `go test -tags all -count=1 -run TestReleaseFolder_ExpandFlatten` passes: `ok ... 0.004s`.
8. `go build -tags all -mod=vendor .` and `go vet -tags all -mod=vendor ./azuredevops/internal/service/release/` both clean.

## What worked

- Using `diag.Diagnostics` return type and `context.Context` first arg for CRUD stubs (matches `schema.CreateContextFunc` etc.).
- Using `r.TestResourceData()` (schema's own helper) in the test to get a correctly typed `*schema.ResourceData`.
- Calling `go vet -tags all` (not just `go vet`) to catch build-tag-guarded code.
- The `releaseapi.Folder` struct has no project_id; `project_id` is maintained solely in TF state and passed as an API argument by the CRUD functions (WI-2's job).

## What didn't work

- CRUD stubs with `ctx interface{}` — `go vet` rejects them as `schema.CreateContextFunc` requires `context.Context`.

## Open questions

_(none blocking — WI-1 complete)_

## Notes for reflection

- The `releaseapi.Folder` SDK struct does not carry `project_id`; the provider pattern stores it in schema state and passes it separately to every API call. Consistent with how `resource_release_definition.go` does it.
- The WI spec says CRUD stubs returning `nil` are acceptable; WI-2 fills in real CRUD logic.
