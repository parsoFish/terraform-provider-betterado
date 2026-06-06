# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (WI-4)

- Created `azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go`
- Modelled on `resource_build_definition_permissions_test.go` (same package, testutils pattern)
- `TestAccReleaseDefinitionPermissions_SetPermissions`: apply step + idempotency PlanOnly step; 4 permissions on Readers group
- `TestAccReleaseDefinitionPermissions_UpdatePermissions`: two apply steps (initial + changed values), each followed by idempotency PlanOnly step
- `hclReleaseDefinitionPermissions` helper: project → betterado_release_definition → Readers group data source → permissions resource
- Release definition HCL cloned from `hclReleaseDefinitionBasic` (requires retention_policy + pre_deploy_approval for ADO REST 7.2)
- `release_definition_id = betterado_release_definition.release.id` — the release definition resource's Terraform `id` IS the numeric definition ID as a string; SDK v2 coerces to TypeInt

## What worked

- `gofmt -w` cleaned up whitespace in permissions map
- `go test -list TestAccReleaseDefinitionPermissions ./azuredevops/internal/acceptancetests/` lists both test functions
- `go build ./azuredevops/internal/acceptancetests/` compiles cleanly
- `make fmtcheck` and `./scripts/terrafmt.sh` both pass

## What didn't work

- `betterado_release_definition.release.definition_id` — does NOT exist on the release definition resource; the only attribute carrying the integer ID is `.id` (the standard Terraform resource ID)

## Open questions

- Will the live ADO gate confirm that `ViewReleases`, `EditReleaseStage`, `DeleteReleases`, `CreateReleases` are valid permission action names for the ReleaseManagement2 namespace? The live gate will confirm (or reject with the correct names).

## Notes for reflection

- The release definition resource does NOT expose a separate `definition_id` computed attribute — the Terraform `.id` carries the integer definition ID as a string. Future agents must NOT look for `definition_id` on `betterado_release_definition`.
- Pre-existing build failures in `azuredevops/internal/service/graph` and `azuredevops/internal/service/serviceendpoint` packages are NOT introduced by this initiative (verified on stash).
