# Add betterado_release_definition_permissions resource

> _Derived from `demo.json` (ADR 021). Essence:_ Introduces betterado_release_definition_permissions — a new Terraform resource that assigns, reads, and removes permissions on Azure DevOps release definitions via the ReleaseManagement2 security namespace (namespace ID: c788c23e-1b46-4162-8f5e-d7585343b5de). A live ADO spike confirmed the definition-level token format is {projectId}/{releaseDefinitionId} (simpler than the WI spec's initial assumption). Before this change, operators had no Terraform-managed way to control who can view, edit, approve, or delete specific release definitions. After this change, permissions on any release definition are declared as HCL desired state, applied idempotently, and cleaned up on destroy.

## Summary

- New resource betterado_release_definition_permissions manages ACLs on ADO release definitions via Terraform
- Token format confirmed via live ADO probe: {projectId}/{releaseDefinitionId} (simpler than initial hypothesis, no namespace prefix)
- Pure-function token derivation — no live API call required to build the ACL token
- Full acceptance test (set + idempotency check) verified against live Azure DevOps (TF_ACC=1)
- Resource registered in provider.go with paired provider_test.go update and HCL usage example

## Test Evidence

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... — passes green on HEAD

- **Before:** No release definition permissions resource existed. Packages compiled but had no permissions coverage.
- **After:** All three packages (release, taskagent, taskagent/validate) pass. The permissions package separately passes its unit suite including the new TestReleaseDefinitionPermissions_TokenFormatSpike.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release | ok (no permissions tests) | ok  (0.023s) | — | match |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent | ok (unchanged) | ok  (0.009s) | — | match |
| github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate | ok (unchanged) | ok  (0.004s) | — | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### Live ADO probe confirmed ReleaseManagement2 token format before building

- **Before:** Token format for ReleaseManagement2 namespace (c788c23e-1b46-4162-8f5e-d7585343b5de) was unconfirmed. WI spec hypothesised 'ReleaseManagement2/Project/{projectId}/{definitionId}' — a prefix pattern from the ReleaseManagement (project-level) namespace.
- **After:** Live probe against org davidgparsonson: created release definition ID=1 in project 21cff396-…, POSTed to _apis/accesscontrolentries/{namespaceId} with token='{projectId}/{releaseDefinitionId}' — API returned HTTP 200 with valid ACE; subsequent GET confirmed token in ACL. Confirmed format: '{projectId}/{releaseDefinitionId}' (no namespace prefix). Simpler than hypothesised and identical to the Build namespace pattern.

## API / Behaviour Diff

### betterado_release_definition_permissions (added)

**After:**
```
resource: project_id (Required, UUID), release_definition_id (Optional, int — omit for project-scope token), principal (Required, string), permissions (Required, map of string)
```

## Acceptance criteria

- ReleaseManagement2 token format confirmed via live ADO probe: token = {projectId}/{releaseDefinitionId}
- TestReleaseDefinitionPermissions_TokenFormatSpike asserts constant non-empty, is a format string, and renders correct token for fake IDs
- ResourceReleaseDefinitionPermissions() has CRUD functions, project_id and release_definition_id fields in schema, and timeout block
- provider.go ResourcesMap contains betterado_release_definition_permissions
- TestProvider_HasChildResources passes with betterado_release_definition_permissions in expected list
- examples/resources/betterado_release_definition_permissions/main.tf exists with valid HCL
- TestAccReleaseDefinitionPermissions_SetPermissions acceptance test exercises apply → idempotency check lifecycle
- Quality gate go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... passes green

## Files Changed

- `azuredevops/internal/service/permissions/resource_release_definition_permissions.go` — New resource: confirmed token format comment (live probe), releaseDefinitionTokenFormat constant, createReleaseDefinitionToken pure function, ResourceReleaseDefinitionPermissions() CRUD via securityhelper
- `azuredevops/internal/service/permissions/resource_release_definition_permissions_test.go` — Unit test TestReleaseDefinitionPermissions_TokenFormatSpike: asserts constant non-empty, is format string, renders {projectId}/{releaseDefinitionId} for fake IDs
- `azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go` — Live acceptance test TestAccReleaseDefinitionPermissions_SetPermissions: apply → idempotency re-plan against real ADO (TF_ACC=1)
- `azuredevops/provider.go` — Registers betterado_release_definition_permissions in ResourcesMap (alphabetical order)
- `azuredevops/provider_test.go` — Adds betterado_release_definition_permissions to TestProvider_HasChildResources expected list (paired with provider.go)
- `examples/resources/betterado_release_definition_permissions/main.tf` — HCL example: project_id, release_definition_id, principal, and permissions attributes demonstrated

```
azuredevops/internal/acceptancetests/resource_release_definition_permissions_test.go        | 177 +++++++++++++++++++++
azuredevops/internal/service/permissions/resource_release_definition_permissions.go         | 152 ++++++++++++++++++
azuredevops/internal/service/permissions/resource_release_definition_permissions_test.go    |  57 +++++++
azuredevops/provider.go                                                                     |   1 +
azuredevops/provider_test.go                                                                |   1 +
examples/resources/betterado_release_definition_permissions/main.tf                         |  22 +++
8 files changed, 466 insertions(+)
```

## Usage

```
```hcl
resource "betterado_release_definition_permissions" "example" {
  project_id            = data.betterado_project.example.id
  release_definition_id = betterado_release_definition.example.id
  principal             = data.betterado_group.readers.descriptor

  permissions = {
    ViewReleases      = "Allow"
    EditReleaseStage  = "NotSet"
    DeleteReleases    = "Deny"
    CreateReleases    = "Deny"
  }
}
```
```

## Impact

- Operators can now declare release definition permission ACLs as Terraform desired state — no more manual portal clicks or raw API calls.
- Permissions are applied idempotently: re-running terraform apply on an unchanged config produces an empty plan.
- Clean destroy removes the permission entry from the release definition's ACL in ADO.
- Uses the ReleaseManagement2 security namespace (definition-level scope, c788c23e-1b46-4162-8f5e-d7585343b5de) — more granular than the project-level ReleaseManagement namespace.
- Token derivation is a pure function of projectId + releaseDefinitionId (format: {projectId}/{releaseDefinitionId}) — no live API call required to build the ACL token, keeping unit tests fast and offline.
- Spike-driven development: token format confirmed against live ADO before implementation began, avoiding rework.
