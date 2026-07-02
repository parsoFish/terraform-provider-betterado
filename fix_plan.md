# Fix Plan — WI-2: Migrate 7 branch policy resources to framework

## Acceptance Criteria

- [x] AC1: All 7 *_framework.go files created; TestAccBranchPolicy* acceptance tests use GetMuxedProviderFactories()
- [x] AC2: All 7 branch policy resources removed from provider.go ResourcesMap and added to framework_provider.go Resources()
- [x] AC3: TestProvider_HasChildResources passes with updated resource count

## Status

**Iteration 1 (complete):** Created all 7 framework files, updated provider registration, updated tests to use GetMuxedProviderFactories().

**Iteration 2 (complete):** Fixed `settings` and `scope` schema definition from `ListNestedAttribute` (requires `= [{}]` HCL) to `ListNestedBlock` (requires `{}` block HCL). This matches the block syntax used in all acceptance test HCL.

## Remaining work

The code is complete. The gate runs live acceptance tests with TF_ACC. Those should pass once the schema block fix is in place.
