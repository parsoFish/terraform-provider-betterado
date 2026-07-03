## Why

The betterado provider's git area (three resources, three data sources) remained on the deprecated terraform-plugin-sdk/v2 path after earlier initiatives migrated the release and taskagent areas. Keeping SDKv2 resources alongside the growing framework provider creates a dual-maintenance burden and prevents the team from eventually removing the SDKv2 server from the mux entirely. This initiative completes the git migration so the provider presents a uniform framework surface to Terraform core for all six git types.

A reviewer send-back (UWI-2) surfaced three additional concerns now resolved: (1) `terraform-plugin-log/tflog` was imported directly in the new framework files but still listed as `indirect` in `go.mod` — `go mod tidy` was run; (2) the fixture project lookup was substituting a created project when the standing fixture was absent — hardened to fail loudly instead; (3) the `initialization` block in `betterado_git_repository_framework.go` had lost all validators from the SDKv2 implementation — restored to full parity.

## What

- **WI-1:** Produced `docs/git-gap-matrix.md` — field-by-field comparison of the ADO Git Repositories REST API v7.1 surface against each SDKv2 schema, with `implemented`/`deferred`/`computed-only`/`N/A` status and rationale for every deferred field. Updated (UWI-2) with initialization block validator detail.

- **WI-2:** Migrated `betterado_git_repository` resource and data source to `terraform-plugin-framework`. New files: `resource_git_repository_framework.go`, `data_git_repository_framework.go`. Both deregistered from `provider.go` SDKv2 maps; registered in `framework_provider.go`. Acceptance tests updated to `GetMuxedProviderFactories()`. Initialization block validators restored to SDKv2 parity (UWI-2): `listvalidator.SizeBetween(1,1)`, `OneOf` enums on `init_type`/`source_type`, `IsURL` on `source_url`, `RequiredWith`/`ConflictsWith` pairs, `RequiresReplace` on `init_type`/`source_type`/`source_url`.

- **WI-3:** Migrated `betterado_git_repository_branch` resource to `terraform-plugin-framework`. New file: `resource_git_repository_branch_framework.go`. Deregistered from SDKv2; registered in framework provider. `ref_*` mutual exclusion ported as framework validators. Acceptance tests updated.

- **WI-4:** Migrated `betterado_git_repository_file` resource and data source to `terraform-plugin-framework`. New files: `resource_git_repository_file_framework.go`, `data_git_repository_file_framework.go`. Import state (split `repo_id:file_path:branch`) implemented via `resource.ImportStatePassthroughID`. Both deregistered from SDKv2; registered in framework provider. Acceptance tests updated.

- **WI-5:** Migrated `betterado_git_repositories` data source to `terraform-plugin-framework`. New file: `data_git_repositories_framework.go`. Deregistered from SDKv2; registered in framework provider. Acceptance tests updated.

- **WI-6:** Regenerated registry docs via `make docs` (six doc files updated); restored hand-written guides via `git checkout -- docs/guides/`; added HCL examples under `examples/resources/` and `examples/data-sources/`; added CHANGELOG `## [Unreleased]` entry; bumped `PROVIDER_VERSION.txt` to `1.2.1`.

- **UWI-2 (review fix):** `go mod tidy` — promotes `terraform-plugin-log/tflog` from indirect to direct in `go.mod`; hardens fixture project lookup in `shared_fixtures.go` to fail loudly rather than substitute; restores initialization block validators in `resource_git_repository_framework.go`; updates `docs/git-gap-matrix.md` with validator detail.

Files changed (from `git diff --name-only main...HEAD`):
- `CHANGELOG.md`
- `PROVIDER_VERSION.txt`
- `azuredevops/internal/acceptancetests/data_git_repositories_test.go`
- `azuredevops/internal/acceptancetests/data_git_repository_file_test.go`
- `azuredevops/internal/acceptancetests/data_git_repository_test.go`
- `azuredevops/internal/acceptancetests/direct_client_test.go`
- `azuredevops/internal/acceptancetests/resource_git_repository_branch_test.go`
- `azuredevops/internal/acceptancetests/resource_git_repository_file_test.go`
- `azuredevops/internal/acceptancetests/resource_git_repository_test.go`
- `azuredevops/internal/acceptancetests/resource_task_group_test.go`
- `azuredevops/internal/acceptancetests/shared_fixtures.go`
- `azuredevops/internal/provider/framework_provider.go`
- `azuredevops/internal/service/git/data_git_repositories_framework.go` *(new)*
- `azuredevops/internal/service/git/data_git_repository_file_framework.go` *(new)*
- `azuredevops/internal/service/git/data_git_repository_framework.go` *(new)*
- `azuredevops/internal/service/git/resource_git_repository_branch_framework.go` *(new)*
- `azuredevops/internal/service/git/resource_git_repository_file_framework.go` *(new)*
- `azuredevops/internal/service/git/resource_git_repository_framework.go` *(new)*
- `azuredevops/provider.go`
- `azuredevops/provider_test.go`
- `docs/data-sources/git_repositories.md`
- `docs/data-sources/git_repository.md`
- `docs/data-sources/git_repository_file.md`
- `docs/git-gap-matrix.md` *(new)*
- `docs/resources/git_repository.md`
- `docs/resources/git_repository_branch.md`
- `docs/resources/git_repository_file.md`
- `examples/data-sources/betterado_git_repositories/data-source.tf` *(new)*
- `examples/data-sources/betterado_git_repository/data-source.tf` *(new)*
- `examples/data-sources/betterado_git_repository_file/data-source.tf` *(new)*
- `examples/resources/betterado_git_repository/resource.tf` *(new)*
- `examples/resources/betterado_git_repository_branch/resource.tf` *(new)*
- `examples/resources/betterado_git_repository_file/resource.tf` *(new)*
- `forge/history/INIT-2026-07-01-migrate-framework-git/demo/DEMO.md`
- `forge/history/INIT-2026-07-01-migrate-framework-git/demo/demo.json`
- `go.mod`

## How

Each resource follows the same framework migration pattern established in the release-area initiatives:

1. **New framework file** implementing `resource.Resource` (or `datasource.DataSource`) with `Metadata`, `Schema`, `Configure`, `Create`/`Read`/`Update`/`Delete` methods.
2. **`Configure()` casts `req.ProviderData.(*client.AggregatedClient)`** (not the SDKv2 meta); logs a warning and returns if the cast fails during plan.
3. **Computed attributes** use `UseStateForUnknown` plan modifier so Terraform does not produce perpetual diffs on computed fields like `remote_url`, `ssh_url`, `web_url`.
4. **SDKv2 deregistration in the same commit** as framework registration — prevents the `Invalid Provider Server Combination: Duplicate resource type` error from the mux.
5. **Acceptance tests** updated to `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` so the test harness routes through the mux provider.
6. **Shared fixture project** (`shared_fixtures.go`) is used across all git acceptance tests to avoid the ADO org project-count limit. Fixture lookup hardened to fail loudly (never create or substitute) — matches the approach adopted in PR #47.
7. **`CaptureLiveEvidence("acceptance-resource", url, response)`** called in the acceptance test read-back step — writes `.forge/live-evidence/acceptance-resource.json` for `forge demo render` to back-fill into `demo.json`.
8. **`go mod tidy`** (UWI-2): promotes `terraform-plugin-log/tflog` from `indirect` to `direct` in `go.mod` — resolves depscheck CI failure.
9. **Initialization block parity** (UWI-2): `resource_git_repository_framework.go` `initialization` nested list now has `listvalidator.SizeBetween(1,1)`, `OneOf` enums on `init_type` and `source_type`, `IsURL` on `source_url`, `RequiredWith`/`ConflictsWith` attribute relationships, and `RequiresReplace` plan modifiers on immutable fields — matching the SDKv2 `CustomizeDiff` semantics.
