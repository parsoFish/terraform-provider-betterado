# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (this iteration)

**Root cause of gate failure (AC1 — second attempt):** `TestAccMuxSdkv2Passthrough` still creates a `betterado_project` resource in its HCL config, even though it had been switched to `ProtoV6ProviderFactories`. The org is at the 1000-project cap: ADO rejects any `QueueCreateProject` call with "organization already has 1000 projects". The project-create happened inside the Terraform apply step (resource `betterado_project.smoke`), not in SharedReleaseFixture.

**Fix applied (AC1):** Rewrote `TestAccMuxSdkv2Passthrough` to call `SharedReleaseFixture(t)` (exactly as `TestAccReleaseFolderFramework` does) and removed the `betterado_project` resource block from `hclMuxSmokeFolder`. The HCL now only creates `betterado_release_folder.smoke` using the pre-existing shared project ID.

Commit: `f3925364`

**Verified:** `go build -tags all ./...` and `go vet -tags all ./azuredevops/internal/acceptancetests/...` both clean.

### Iteration 0

**Root cause of gate failure (AC1):** `TestAccMuxSdkv2Passthrough` used `Providers: testutils.GetProviders()` (SDKv2-only map). Since `betterado_release_folder` was migrated to the framework provider in prior iterations, the SDKv2 provider no longer serves that resource type — causing "The provider hashicorp/betterado does not support resource type betterado_release_folder".

**Fix applied (AC1):** Changed the test to use:
- `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` (instead of `Providers: testutils.GetProviders()`)
- `getDirectClient()` in both `checkMuxSmokeFolderDestroyed` and `captureMuxPassthroughEvidence` (instead of `testutils.GetProvider().Meta()` which is nil with ProtoV6ProviderFactories)
- Removed unused `"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"` import

**Fix applied (AC3):** Renamed example files for tfplugindocs convention:
- `examples/resources/betterado_release_folder/main.tf` → `resource.tf`
- `examples/data-sources/betterado_release_folder/main.tf` → `data-source.tf`

**Fix applied (AC4):**
- Added 5 data source entries to CHANGELOG.md `## [Unreleased]` section
- Bumped `PROVIDER_VERSION.txt` from `1.0.5` → `1.1.0`

**Verified:** `go build -tags all ./...` and `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` pass cleanly.

## What worked

- Using `SharedReleaseFixture(t)` is the ONLY correct approach for any test that needs a project — it reuses `betterado-standing-demo` and never creates a new project
- `ProtoV6ProviderFactories` + `GetMuxProviderFactories()` is the correct approach for any test that exercises a framework resource
- `getDirectClient()` (defined in `resource_task_group_test.go`) is the pattern for CheckDestroy when using ProtoV6ProviderFactories — it builds a client directly from env vars rather than relying on the SDKv2 singleton Meta
- tfplugindocs expects `resource.tf` for resources and `data-source.tf` for data sources (not `main.tf`)

## What didn't work

- `Providers: testutils.GetProviders()` — SDKv2-only; cannot serve framework resources
- `testutils.GetProvider().Meta()` — nil when using ProtoV6ProviderFactories
- Having `ProtoV6ProviderFactories` but STILL creating a `betterado_project` resource in HCL → project-create fails at 1000-project cap

## Pre-existing issues (not introduced by this WI)

- `go vet` has 2 pre-existing failures in unrelated packages (`data_users_test.go:236` and `resource_serviceendpoint_aws_test.go:64`). These are present on the branch before our changes; the gate only covers the release/taskagent packages.

## Remaining work

- AC1: ✅ Fix committed (f3925364) — needs live gate run to confirm
- AC2: ✅ docs/ files exist; verified by live gate
- AC3: ✅ resource.tf + data-source.tf in place (17bddc0f)
- AC4: ✅ CHANGELOG + PROVIDER_VERSION.txt done (fbb7d7da)

## Notes for reflection

- Whenever a resource is migrated from SDKv2 to framework, every acceptance test that uses that resource must be updated to use `ProtoV6ProviderFactories` instead of `Providers` — AND must not create new projects (use SharedReleaseFixture).
- The 1000-project cap pattern is documented in brain theme `2026-06-20-ado-org-project-limit-blocks-test-creates`: every test that creates resources in ADO must use the persistent shared project.
