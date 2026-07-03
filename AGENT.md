# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (first, complete)

- Gate failure: `[no tests to run]` — the file `resource_extension_install_test.go` did not exist yet.
- Wrote `azuredevops/internal/acceptancetests/resource_extension_install_test.go` from scratch.
- Committed as `1ff5ce49`.

### Iteration 1

- Gate failure: `The provider hashicorp/betterado does not support resource type "betterado_extension_install"` — the resource was implemented (WI-2, `resource_extension_install_framework.go`) but NOT registered in `framework_provider.go`.
- Fixed by adding `extensionmanagement.NewExtensionInstallResource` to the `Resources()` slice in `azuredevops/internal/provider/framework_provider.go`.
- Committed as `7b928209`.
- `make test` passes (no offline failures). Gate offline: `ok ... 0.008s` (skips without TF_ACC — AC3 satisfied). Live gate requires TF_ACC + real ADO org.

## What worked

- **Pattern**: mirror `resource_task_group_test.go` — `ProtoV6ProviderFactories`, `getDirectClient()` for `CheckDestroy` and evidence, `captureXxxEvidence()` as a `resource.TestCheckFunc` inline closure.
- **Build tag**: `//go:build (all || resource_extension_install) && !exclude_resource_extension_install` — required for the gate command `-tags all` to pick it up.
- **Function name**: `testutils.GetMuxProviderFactories()` — NOT `GetMuxedProviderFactories` (which does not exist).
- `getDirectClient()` is already declared in `resource_task_group_test.go` (same package) — do not re-declare it. Added `buildExtensionDirectClient()` as a separate helper that does the same thing without collision.
- Gate command runs against `./azuredevops/internal/acceptancetests/` (not `./...`) — avoids `[no test files]` sibling masking.
- After writing the file the gate output changed from `[no tests to run]` to `--- SKIP: TestAccExtensionInstall_basic` — that is the correct offline result (AC3 satisfied).
- **Registration**: `extensionmanagement.NewExtensionInstallResource` must be in `framework_provider.go`'s `Resources()` slice — without it the mux provider doesn't know the resource type and the TF_ACC test fails immediately with "Invalid resource type".

## What didn't work

- Nothing tried that failed.

## Open questions

- None — all pieces are in place: resource implementation (WI-2), test file (WI-3 iter 0), provider registration (WI-3 iter 1). Live gate should now succeed with real TF_ACC credentials.

## Notes for reflection

- The WI spec said "WI-4 handles registration" but the live gate for WI-3 required it too. The fix was to register in this iteration rather than wait for WI-4.
