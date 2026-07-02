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

## Remaining work

All ACs are complete. Build, lint, and offline tests pass. Live acceptance tests should pass now that `resolveOrCreateFixtureProject` no longer attempts to create a project on an org at capacity.
