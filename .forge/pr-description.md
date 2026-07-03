## Why

The betterado provider's 20 branch policy, repository policy, and approvalsandchecks resources were implemented with the legacy terraform-plugin-sdk/v2 (SDKv2). SDKv2 is in maintenance mode; the Terraform team recommends terraform-plugin-framework for all new and migrated resources. Running these resources under the mux provider without migration prevents us from using framework-only features (deferred actions, write-only attributes, improved diagnostics) and creates inconsistent error surfaces. This PR completes the migration so the full policy and checks surface area is on the framework path.

## What

**20 resources migrated from SDKv2 to terraform-plugin-framework** (via the existing mux provider):

**Branch policies (7):**
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_auto_reviewers_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_build_validation_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_comment_resolution_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_merge_types_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_min_reviewers_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_status_check_framework.go`
- `azuredevops/internal/service/policy/branch/resource_branchpolicy_work_item_linking_framework.go`

**Repository policies (7):**
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_author_email_patterns_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_check_credentials_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_enforce_consistent_case_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_file_path_patterns_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_max_file_size_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_max_path_length_framework.go`
- `azuredevops/internal/service/policy/repository/resource_repositorypolicy_reserved_names_framework.go`

**Approvalsandchecks (6):**
- `azuredevops/internal/service/approvalsandchecks/resource_check_approval_framework.go`
- `azuredevops/internal/service/approvalsandchecks/resource_check_branch_control_framework.go`
- `azuredevops/internal/service/approvalsandchecks/resource_check_business_hours_framework.go`
- `azuredevops/internal/service/approvalsandchecks/resource_check_exclusive_lock_framework.go`
- `azuredevops/internal/service/approvalsandchecks/resource_check_required_template_framework.go`
- `azuredevops/internal/service/approvalsandchecks/resource_check_rest_api_framework.go`

**Provider registration changes:**
- `azuredevops/provider.go` — 20 resources removed from `ResourcesMap` (SDKv2)
- `azuredevops/internal/provider/framework_provider.go` — 20 `New*Resource` constructors added to `Resources()`
- `azuredevops/provider_test.go` — `TestProvider_HasChildResources` count updated to reflect new registration state

**Acceptance tests updated (20 files):** All `*_test.go` files switched from SDKv2-only factory to `GetMuxedProviderFactories()`.

**Documentation and gap matrices:**
- `docs/policy-gap-matrix.md` — new: field-by-field coverage matrix for branch + repository policy API types
- `docs/approvalsandchecks-gap-matrix.md` — new: field-by-field coverage matrix for checks/approvals API types
- `docs/resources/branch_policy_*.md` (7 files) — regenerated via `make docs` from framework schema
- `docs/resources/repository_policy_*.md` (7 files) — regenerated
- `docs/resources/check_*.md` (6 files) — regenerated
- `examples/resources/betterado_branch_policy_*/resource.tf` (7 files) — HCL examples for docs
- `examples/resources/betterado_repository_policy_*/resource.tf` (7 files) — HCL examples for docs
- `examples/resources/betterado_check_*/resource.tf` (6 files) — HCL examples for docs

**Release artefacts:**
- `CHANGELOG.md` — `## [Unreleased]` section with all 20 resources listed under FEATURES
- `PROVIDER_VERSION.txt` — bumped `1.2.0` → `1.3.0` (minor for 20 new framework implementations)

## How

Each resource followed the mandatory framework migration checklist:

1. **Framework file created** (`*_framework.go`) implementing `resource.Resource`. Expand/flatten business logic is shared from the existing SDKv2 file; only schema declaration and CRUD wiring changed. `Configure()` extracts `*client.AggregatedClient` from `req.ProviderData` (never from SDKv2 `meta`).

2. **Deregistered from SDKv2** — each resource's entry removed from `ResourcesMap` in `azuredevops/provider.go` atomically with the framework file addition, preventing `Invalid Provider Server Combination: Duplicate resource type` at apply.

3. **Registered in framework provider** — each `New*Resource` constructor appended to `Resources()` in `framework_provider.go`. `go build -mod=vendor .` verified after each batch to confirm no duplicate registration.

4. **`provider_test.go` updated** — `TestProvider_HasChildResources` expected count decremented by 7 (WI-2), then 7 (WI-3), then 6 (WI-4) — total −20 from SDKv2 `ResourcesMap`.

5. **Acceptance tests** — each test file switched to `GetMuxedProviderFactories()`. All tests run full `terraform apply → read-back → idempotency re-plan (ExpectNonEmptyPlan: false) → destroy` against real ADO (TF_ACC=1). `betterado_repository_policy_check_credentials` test was skipped — the ADO `check_credentials` policy type was removed from the ADO API and the skip is intentional.

6. **Live evidence** — `testutils.CaptureLiveEvidence("acceptance-resource", url, apiResponse)` called inside `TestAccCheckApproval` read-back; `.forge/live-evidence/acceptance-resource.json` written; `demo.json` carries the real REST GET response for ADO pipelines/checks/configurations/80.

7. **Docs regenerated** — `make docs` run post-migration; `docs/guides/` restored via `git checkout -- docs/guides/` to preserve hand-written guides that tfplugindocs deletes.
