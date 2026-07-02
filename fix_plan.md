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

## Remaining work

All ACs are complete and all code quality gates (build, lint, offline tests) pass. The live acceptance tests should pass once the gate runs against the shared fixture project.
