# Fix Plan — WI-2: Migrate 7 branch policy resources to framework

## Acceptance Criteria

- [x] AC1: All 7 *_framework.go files created; TestAccBranchPolicy* acceptance tests use GetMuxedProviderFactories()
- [x] AC2: All 7 branch policy resources removed from provider.go ResourcesMap and added to framework_provider.go Resources()
- [x] AC3: TestProvider_HasChildResources passes with updated resource count

## Status

**Iteration 1 (complete):** Created all 7 framework files, updated provider registration, updated tests to use GetMuxedProviderFactories().

**Iteration 2 (complete):** Fixed `settings` and `scope` schema definition from `ListNestedAttribute` (requires `= [{}]` HCL) to `ListNestedBlock` (requires `{}` block HCL). This matches the block syntax used in all acceptance test HCL.

**Iteration 3 (complete):** Fixed root cause of gate failures — all 7 acceptance tests were creating new ADO projects, which fails at the 1000-project org limit. Switched all tests to use `data "betterado_project"` with `SharedFixtureProjectName`. Per-run git repos are still created uniquely and destroyed by Terraform. Also:
- Added `captureMinReviewersPolicyEvidence` live evidence to min_reviewer_test.go
- Fixed gofmt/gofumpt formatting in 3 framework files
- Added CHANGELOG.md [Unreleased] entry

**Iteration 4 (complete):** Fixed iteration 3's gate failure — all 7 acceptance tests used `data "betterado_project"` in HCL which fails with "Project does not exist" on the live org. Replaced with `SharedFixtureProjectID(t)` helper that resolves project UUID via CoreClient.GetProject SDK before HCL is generated.

**Iteration 5 (complete):** Fixed iteration 4's gate failure — `SharedFixtureProjectID(t)` called `resolveOrCreateFixtureProject` which fell through to `QueueCreateProject` when GetProject returned an error. The live org is at the 1000-project cap so `QueueCreateProject` always fails. Removed the create fallback; `resolveOrCreateFixtureProject` now does a GetProject-only lookup and `t.Fatal`s if the project is absent.

**Iteration 6 (complete):** Fixed iteration 5's gate failure — `resolveOrCreateFixtureProject` called `GetProject("betterado-standing-demo")` which returned TF200016 (project does not exist) on the live org. The org may have deleted or renamed this project. Applied same auto-discovery pattern as `smokeResolveProject` in state_upgrade_smoke_test.go: try AZDO_TEST_EXISTING_PROJECT env var first, then try the named project, then auto-discover first WellFormed project from GetProjects. Also ran gofmt on provider.go (alignment fix).

**Iteration 7 (complete):** Fixed two live gate failures:

1. `TestAccBranchPolicyStatusCheck_complete`: replaced `betterado_user_entitlement` resource
   (requires billing; always fails on live org with error 5015) with `data.betterado_group "Project
   Administrators"` — group's `origin_id` serves as `author_id` for status check policy.

2. `TestAccBranchPolicyMinReviewers_requiresImportError`: updated `ExpectError` regex from
   `` ` creating policy in Azure DevOps: The update is rejected by policy` `` (SDKv2 common.go
   error format with leading space) to `` `The update is rejected by policy` `` which matches
   the framework resource diagnostic format (summary + detail) and also the legacy format.

## Remaining work

All ACs should now be complete. All known live gate failures addressed.
