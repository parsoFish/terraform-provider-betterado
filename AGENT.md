# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (this run)

- Wrote `resource_build_folder_framework.go` — full framework resource (Create/Read/Update/Delete/ImportState/Configure) reusing the existing `createBuildFolder()` helper from `resource_build_folder.go`.
- Used inline plan-modifiers (`bfUseStateForUnknown`, `bfRequiresReplace`) and inline default (`bfStaticString`) — matching the release package pattern exactly since `stringplanmodifier` and `stringdefault` sub-packages are NOT vendored.
- Wrote `resource_build_folder_framework_test.go` — `TestBuildFolderFramework_Schema` (the gate test). Uses `//go:build all || resource_build_folder_framework`.
- Registered `build.NewBuildFolderResource` in `framework_provider.go` Resources().
- Removed `betterado_build_folder` from SDKv2 `provider.go` ResourcesMap (commented with migration note matching release_folder pattern).
- Updated `provider_test.go` expectedResources list to comment out betterado_build_folder (now framework-served).
- Fixed gofmt issue in provider.go after editing.
- Wrote acceptance test `azuredevops/internal/acceptancetests/resource_build_folder_framework_test.go` — `TestAccBuildFolder_Framework_basic` using SharedReleaseFixture + ProtoV6ProviderFactories, idempotency step, CaptureLiveEvidence with label "acceptance-resource".

## What worked

- Directly reusing `createBuildFolder()` from existing `resource_build_folder.go` — avoids duplicating API code.
- The inline plan-modifier pattern from `release/framework_defaults.go` compiles cleanly; the `defaults` and `planmodifier` packages ARE vendored in the right path.
- `TestBuildFolderFramework_Schema` runs and passes (PASS in 0.003s).
- `go build -mod=vendor .` compiles cleanly.
- `make test` passes (gofmt + full unit suite).
- `golangci-lint run --new-from-rev=main ./azuredevops/...` → 0 issues.

## What didn't work

- First attempt imported `stringplanmodifier` — that package is NOT vendored; replaced with inline implementations.
- First attempt used a custom DefaultString interface with wrong method signature — replaced with correct `defaults.StringRequest / *defaults.StringResponse` from the vendored `defaults` package.

## Open questions

- AC3 (acceptance test) requires `TF_ACC=1` and live ADO environment — cannot run offline. The orchestrator will run this via the live gate.
- Build folder REST ID pattern: the SDKv2 implementation uses `project.Id.String()` as the Terraform state ID (not the path). The framework resource maintains this.

## Notes for reflection

- The `betterado_build_folder` provider_test.go count list must be updated when removing from SDKv2 — same pattern as release_folder.
- Build folder REST endpoint uses the standard ADO org URL (not vsrm host): `{orgURL}/{project}/_apis/build/folders{path}?api-version=7.1-preview.2`.
- `stringplanmodifier` and `stringdefault` sub-packages from terraform-plugin-framework are NOT vendored — always use inline implementations matching `release/framework_defaults.go`.
