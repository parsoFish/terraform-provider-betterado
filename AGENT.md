# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-04) — ALL ACs COMPLETE
- Created `azuredevops/internal/service/extensionmanagement/` (new package).
- Implemented `resource_extension_install_framework.go` with:
  - `ExtensionInstallResource` struct + `NewExtensionInstallResource()` constructor.
  - Schema: required `publisher_id` + `extension_id` (both RequiresReplace), computed+optional `version`, optional+computed `disabled`.
  - Custom `stringNotEmptyValidator` (inline, no external dep) — `terraform-plugin-framework-validators` is NOT vendored; inline impl satisfies AC.
  - Custom `requiresReplaceString` and `useStateForUnknownString` plan modifiers (pattern from `resource_task_group_framework.go`).
  - CRUD: Create=InstallExtensionByName, Read=GetInstalledExtensionByName, Update=UpdateInstalledExtension, Delete=UninstallExtensionByName.
  - Read: 404 → `resp.State.RemoveResource(ctx)` + return, no error.
  - `expandExtensionInstall` + `flattenExtensionInstall` exported helpers for testability.
- Implemented `resource_extension_install_framework_test.go` (build tag `all || resource_extension_install`): 4 tests all pass.
- Quality gate `go test -tags all -run TestExtensionInstallResource ./azuredevops/internal/service/extensionmanagement/...` → PASS.
- No changes to `framework_provider.go`, `provider.go`, or any existing file.

## What worked

- `terraform-plugin-framework-validators` is NOT vendored in this project — implement validator interface inline.
- Follow `taskagent/resource_task_group_framework.go` pattern for inline plan modifiers + state model struct with `tfsdk` tags.
- `utils.ResponseWasNotFound` for 404 detection; disabled-flag parsing mirrors SDKv2 `resource_extension.go` (string split on ",").

## What didn't work

_(none — first iteration succeeded for all ACs)_

## Open questions

_(none outstanding)_

## Notes for reflection

- The `terraform-plugin-framework-validators` external package is referenced in the WI spec but is not vendored. Projects should either vendor it or document the inline-validator pattern as the preferred approach for new framework resources.
