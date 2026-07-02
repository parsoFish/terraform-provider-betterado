# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (completed — all 3 ACs done)

- Read all source files in `azuredevops/internal/service/security/`, `azuredevops/internal/service/securityroles/`, and `azuredevops/internal/service/permissions/`
- Read `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/security/models.go` for ADO SDK struct reference
- Read `azuredevops/utils/sdk/securityroles/models.go` + `client.go` for the hand-rolled securityroles SDK
- Read `azuredevops/internal/service/permissions/utils/` for common schema (`baseSchema.go`) and namespace UUIDs (`namespaces.go`)
- Modelled all 3 gap-matrix docs after `docs/release-definition-gap-matrix.md` (the reference format)
- Created `docs/security-gap-matrix.md`, `docs/securityroles-gap-matrix.md`, `docs/permissions-gap-matrix.md`
- Quality gate: `go test -tags all -run TestProvider_HasChildResources ./azuredevops/` → PASS (0.005s)
- Committed: fb9c2884

## What worked

- Reading source files directly was fast and sufficient — all token formats and schema fields are coded inline in each resource file
- The `baseSchema.go` CreatePermissionResourceSchema helper explained the shared `principal`/`replace`/`permissions` fields across all 13 permissions resources
- `namespaces.go` had all namespace UUID constants in one place

## What didn't work

_(nothing failed)_

## Open questions

- `deny_permissions` in `betterado_securityrole_definitions` is declared `Optional` instead of `Computed` — minor schema bug noted in the matrix, deferred

## Notes for reflection

- The permissions package uses a uniform helper pattern (securityhelper.{New,Set,Get}PrincipalPermissions) — all 13 resources share the same 3-field base schema; the only variation is the token construction function and namespace ID
- The securityroles API is a hand-rolled preview client (not in the official Go SDK)
- `InheritPermissions` on ACL tokens is the one genuinely useful writable gap across all permissions resources — worth a dedicated WI in the initiative
