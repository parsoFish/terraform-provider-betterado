# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 2

**Root cause of gate failure:**
`TestAccEnvironmentKubernetes_createUpdate` used `Providers: testutils.GetProviders()` (SDKv2-only provider map). Since `betterado_environment` is now a framework resource, the SDKv2 provider no longer knows about it. The Kubernetes test's HCL references `resource "betterado_environment" "test"`, so plan fails with "does not support resource type betterado_environment."

**Fix applied:**
1. `resource_environment_resource_kubernetes_test.go`: Changed `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
2. Same file: replaced `testutils.GetProvider().Meta().(*client.AggregatedClient)` with `testutils.GetDirectClient()`
3. Added `GetDirectClient()` to `testutils/commons.go` (no build tag) so it compiles without `-tags all`

**Offline gates:** `make test` pass, golangci-lint 0 issues, terrafmt-check pass, `go build -tags all ./...` clean.

## What worked

- Moving `GetDirectClient` to `testutils/commons.go` (no build tag) so all test files can use it
- `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` is required for any test whose HCL references ANY framework resource (even as a dependency — not just the resource under test)
- Pattern: `Providers: testutils.GetProviders()` is incompatible with any HCL that touches framework resources

## What didn't work

_(none so far)_

## Open questions

- AC1 and AC3 need live TF_ACC gate to confirm (offline gates pass).

## Notes for reflection

- Any test that uses `Providers: GetProviders()` (SDKv2-only) but references a framework resource in HCL will fail with "does not support resource type". The fix is always `ProtoV6ProviderFactories: GetMuxedProviderFactories()`.
- `GetDirectClient()` should be in a non-tagged shared location (testutils/commons.go) so all test files can use it regardless of build tags.
