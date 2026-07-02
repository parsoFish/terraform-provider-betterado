# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this iteration)

**Goal:** Fix gate failure — `TestBuildDefinitionFramework_Schema` found no tests because the file didn't exist.

**Actions taken:**
1. Created `azuredevops/internal/service/build/resource_build_definition_framework.go`:
   - Full schema with all AC1 attributes: name, project_id, revision, path, agent_pool_name, repository, variable, ci_trigger, pull_request_trigger, agent_specification, job_authorization_scope, queue_status, skip_first_run
   - CRUD methods + ImportState (stubs with real API calls)
   - Inline plan-modifiers (bdFw-prefixed) for String RequiresReplace, UseStateForUnknown, and Int64 UseStateForUnknown
   - Inline defaults (bdFwStaticString, bdFwStaticBool) — the vendored subpackages `booldefault` and `stringdefault` are NOT vendored; use the inline pattern from resource_build_folder_framework.go
   - Uses `defaults.Bool`/`defaults.String`/`defaults.BoolRequest`/`defaults.StringResponse` from `resource/schema/defaults` (IS vendored)
2. Created `azuredevops/internal/service/build/resource_build_definition_framework_test.go`:
   - Build tag: `//go:build all || resource_build_definition_framework`
   - `TestBuildDefinitionFramework_Schema` passes gate
3. Registered `build.NewBuildDefinitionResource` in `framework_provider.go`

**Gate result:** `go test -tags all -run TestBuildDefinitionFramework_Schema ./azuredevops/internal/service/build/...` → PASS

## What worked

- Inline defaults pattern from `resource_build_folder_framework.go` (same build package): use `bdFw*` prefix to avoid type name collisions with `bf*` modifiers in that file
- The `defaults` package IS vendored at `vendor/github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults/`
- The sub-packages `booldefault` and `stringdefault` are NOT vendored — must inline
- `schema.SetNestedAttribute` works fine for the `variable` block
- `schema.ListNestedAttribute` works for repository, ci_trigger, pull_request_trigger

## What didn't work

- First attempt imported `booldefault` and `stringdefault` — compiler error "cannot find module" (not vendored)
- Test file had reference to `resource.NestedBlockObject` which doesn't exist in this version of the framework

## Open questions

- Does the readIntoModel need to also populate `repository`, `variable`, `ci_trigger`, `pull_request_trigger` for the idempotency re-plan (AC2) to pass? Almost certainly yes — for AC2 we'll need to flattenBuildDefinitionFw helpers.
- The `betterado_build_definition` entry in `provider.go` ResourcesMap — needs to be removed/commented to avoid "Invalid Provider Server Combination" when the mux is active. This must be done before the acceptance test.

## Notes for reflection

- Pattern for inline defaults in go-framework resources in this codebase: implement the defaults.String/Bool interfaces inline with unique prefixes to avoid naming collisions within the same package.
- The acceptancetests file is a required output (`creates:` path). Next iteration should create it as a compiling stub.
