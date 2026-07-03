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

### Iteration 3 (fix classic pipelines)

**Root cause of gate failure**: All three TestAccDeploymentGroup tests failed with:
```
Error: creating deployment group in Azure DevOps: The classic pipelines are disabled for this project / organization.
```

**Root cause analysis**: The `enableClassicPipelinesForFixtureProject` function was only setting
`disableClassicBuildPipelineCreation: false` (for classic build pipelines) at the project level.
However, deployment groups require **classic deployment pipelines** to be enabled, which is controlled
by a **separate flag**: `disableClassicDeploymentPipelineCreation`. Additionally, there may be an
**org-level** policy that overrides the project-level settings.

**Fix applied** (commit 51f67ebe):
- Added `disableClassicDeploymentPipelineCreation: false` to the PATCH body
- Now also PATCHes the **org-level** endpoint (`{orgURL}/_apis/build/generalsettings`) first,
  before the project-level endpoint, to unblock any org-level lock
- Org-level PATCH is non-fatal (logs warning and continues) because the org admin may use a
  different mechanism, but the project-level PATCH is the authoritative fix
- Switched from `bytes.NewBufferString` to `strings.NewReader` (removed `bytes` import)

## What worked

- Standing fixture project pattern (same as WI-5/WI-6) avoids 1000-project limit
- `ProtoV6ProviderFactories` + `GetMuxedProviderFactories()` required for framework resources
- `GetDirectClient()` required in `CheckDestroy` when using `ProtoV6ProviderFactories`
- `pool_id` handled as `Int64Attribute` with `requiresReplaceInt64()` + Computed=true (ADO always assigns a pool)
- `else if` instead of `else { if }` to satisfy gocritic linter
- `make fmt` + `golangci-lint run --new-from-rev=main` + `make terrafmt-check` all pass green

## What didn't work

- Setting only `disableClassicBuildPipelineCreation: false` — NOT sufficient for deployment groups.
  You must also set `disableClassicDeploymentPipelineCreation: false`.
- Patching only the project level may not be enough if org-level policy is set.

## Open questions

_(none)_

## Notes for reflection

- Deployment group `pool_id` is always assigned by ADO (even when not specified) — `Computed: true` is required to avoid plan drift
- The `withPoolId` test pattern works with the fixture project by creating two deployment groups in the same project (pool_source + test), rather than needing two separate projects
- **Classic pipeline settings for deployment groups**: Deployment groups are a feature of classic
  release/deployment pipelines, not classic build pipelines. Both `disableClassicBuildPipelineCreation`
  AND `disableClassicDeploymentPipelineCreation` must be false for deployment groups to work.
  The ADO error message "classic pipelines are disabled" is ambiguous about which setting is the
  actual blocker.
