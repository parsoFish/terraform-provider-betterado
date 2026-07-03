# Agent Memory — WI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (current)

- Checked `.forge/last-gate-failure.md` — not present (no prior live gate run).
- Reviewed git log: no prior WI-6-specific iteration commits on this branch.
- **All implementation was already present** when iteration 0 ran:
  - `resource_environment_resource_kubernetes_framework.go` — fully implemented (`EnvironmentResourceKubernetes` struct, CRUD, schema with all attrs, `requiresReplaceSet` planmodifier).
  - SDKv2 `resource_environment_resource_kubernetes.go` — DELETED (confirmed absent).
  - SDKv2 unit test `resource_environment_resource_kubernetes_test.go` in taskagent/ — DELETED.
  - `framework_provider.go` already includes `taskagent.NewEnvironmentResourceKubernetesResource` at line 209.
  - `provider.go` has comment noting resource is in framework, not in SDKv2 ResourcesMap.
  - `azuredevops/internal/acceptancetests/resource_environment_resource_kubernetes_test.go` — complete with `TestAccEnvironmentResourceKubernetes_createUpdate`, `captureEnvironmentKubernetesEvidence` calling `CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, k8sResource)`.
  - `CHANGELOG.md` — has entry under `## [Unreleased]` FEATURES.
  - `docs/resources/environment_resource_kubernetes.md` — generated and present.
  - `examples/resources/betterado_environment_resource_kubernetes/resource.tf` — present.
- Added missing comment to `provider_test.go` noting `betterado_environment_resource_kubernetes` is now a framework resource, committed.
- All offline gates pass: `go build ./azuredevops/...`, `go test -v ./...`, `make terrafmt-check`, `golangci-lint run --new-from-rev=main ./azuredevops/... → 0 issues`.

## What worked

- All implementation was essentially pre-built by the branching strategy (prior WI work carried over).
- Framework file uses shared helper functions from other files in the taskagent package: `requiresReplace()`, `requiresReplaceInt64()`, `useStateForUnknown()`, `defaultString()`, `requiresReplaceSet()` (defined inline in the framework file itself).
- `expandStringSetAttr` / `flattenStringSetAttr` helpers are defined inline in the framework file.
- Acceptance test uses `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` and `SharedFixtureProjectName` to avoid ADO org project limit.
- `captureEnvironmentKubernetesEvidence` calls `CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, k8sResource)` for AC3.
- The second acceptance test step (name change) tests ForceNew behavior since all attrs are RequiresReplace.

## What didn't work

_(none — clean first pass)_

## Open questions

- Live gate required: `TF_ACC=1 go test -tags all -run TestAccEnvironmentResourceKubernetes ./azuredevops/internal/acceptancetests/` must run with real ADO credentials. The offline run shows `ok ... 0.007s` (tests skipped — expected).
- The `serviceendpoint` package has a pre-existing build failure (return value count mismatch in test files) that exists on main too — not our issue.

## Notes for reflection

- The kubernetes environment resource migration followed the exact same pattern as `betterado_environment`: framework file with RequiresReplace on all attrs (no in-place update), `readIntoModel` helper, Set attr with custom planmodifier.
- `requiresReplaceSet()` is defined directly in the kubernetes framework file (not in a shared helper file) — consistent with how project-specific plan modifiers are handled per resource.
