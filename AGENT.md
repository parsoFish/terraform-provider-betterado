# Agent Memory — WI-7

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

**Root cause of gate failure**: All three TestAccDeploymentGroup tests failed because the ADO test org has hit its 1000-project limit. The old tests created new projects per test run.

**Fix applied**: Rewrote acceptance tests to use the standing fixture project (`SharedFixtureProjectName = "betterado-standing-demo"`) via `ProtoV6ProviderFactories` (mux provider), following the same pattern as `TestAccEnvironmentResourceKubernetes_createUpdate`. The HCL now uses `data "betterado_project" "fixture"` instead of `resource "betterado_project"`.

**All changes committed in `0ee972ad`:**
1. Created `resource_deployment_group_framework.go` — full CRUD + ImportState
2. Deleted `resource_deployment_group.go` (SDKv2)
3. Removed `betterado_deployment_group` from `provider.go` ResourcesMap
4. Added `taskagent.NewDeploymentGroupResource` to `framework_provider.go` Resources()
5. Removed `betterado_deployment_group` from `provider_test.go` expectedResources list
6. Rewrote acceptance test to use fixture project + `ProtoV6ProviderFactories` + `ExpectNonEmptyPlan: false`
7. Added `captureDeploymentGroupEvidence()` with `CaptureLiveEvidence("acceptance-resource-deployment-group", ...)`
8. Created `examples/resources/betterado_deployment_group/resource.tf`
9. Ran `make docs` → updated `docs/resources/deployment_group.md` (removes SDKv2 timeouts block)
10. Added CHANGELOG entry under Unreleased FEATURES

## What worked

- Standing fixture project pattern (same as WI-5/WI-6) avoids 1000-project limit
- `ProtoV6ProviderFactories` + `GetMuxedProviderFactories()` required for framework resources
- `GetDirectClient()` required in `CheckDestroy` when using `ProtoV6ProviderFactories`
- `pool_id` handled as `Int64Attribute` with `requiresReplaceInt64()` + Computed=true (ADO always assigns a pool)
- `else if` instead of `else { if }` to satisfy gocritic linter
- `make fmt` + `golangci-lint run --new-from-rev=main` + `make terrafmt-check` all pass green

## What didn't work

_(nothing — iteration 0 was clean)_

## Open questions

_(none)_

## Notes for reflection

- Deployment group `pool_id` is always assigned by ADO (even when not specified) — `Computed: true` is required to avoid plan drift
- The `withPoolId` test pattern works with the fixture project by creating two deployment groups in the same project (pool_source + test), rather than needing two separate projects
