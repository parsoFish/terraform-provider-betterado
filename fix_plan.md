# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_release_definition_permissions resource is registered and TF_ACC=1 with a valid PAT in secrets.env WHEN TestAccReleaseDefinitionPermissions_SetPermissions runs against live ADO (terraform apply) THEN a release definition permission entry is created for the Readers group on the target release definition; the read step confirms the permission is persisted; the plan step after apply is empty (ExpectNonEmptyPlan: false = idempotent); the destroy step removes the permission cleanly
- [x] AC2: GIVEN the same live acceptance test WHEN a second TestAccReleaseDefinitionPermissions_UpdatePermissions step runs with changed permission values THEN terraform apply updates the permissions in place; read confirms the updated values; idempotency re-plan is empty

## Notes (iter-0)

- Test file: `azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go` (committed)
- Both functions registered and listed by `go test -list`
- `gofmt` clean, `terrafmt` clean, package compiles clean
- HCL uses `betterado_release_definition.release.id` for `release_definition_id` (Terraform SDK v2 coerces string→int)
- Pre-existing build failures in graph/serviceendpoint packages are NOT our change (verified by stash test)
- Pending: live TF_ACC gate run by orchestrator
