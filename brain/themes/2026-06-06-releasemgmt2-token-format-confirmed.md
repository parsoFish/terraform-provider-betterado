---
title: ReleaseManagement2 namespace token format confirmed — no prefix
description: Live ADO probe (2026-06-06) confirmed the ReleaseManagement2 token format is {projectId}/{releaseDefinitionId} — identical to the Build namespace, NOT the hypothesised 'ReleaseManagement2/Project/{projectId}/{definitionId}' prefix format.
category: decision
project: terraform-provider-betterado
created_at: 2026-06-06T06:04:52Z
updated_at: 2026-06-06T06:04:52Z
related_themes:
  - 2026-06-06-spike-wis-prevent-wrong-builds
---

# ReleaseManagement2 namespace token format confirmed — no prefix

## Decision

For the `ReleaseManagement2` security namespace (`c788c23e-1b46-4162-8f5e-d7585343b5de`), the correct Azure DevOps ACL token format is:

- **Project-level:** `{projectId}` (plain UUID, no prefix)
- **Definition-level:** `{projectId}/{releaseDefinitionId}` (UUID + `/` + int ID)

## What was expected vs actual

The initiative manifest and WI specs hypothesised the format might be `ReleaseManagement2/Project/{projectId}/{definitionId}`, analogous to how the namespace is named. This was **wrong**.

The actual format is structurally identical to the `ReleaseManagement` (Build) namespace — no string prefix, just `{projectId}` or `{projectId}/{defId}`.

## Evidence from the spike

WI-1 queried `_apis/accesscontrollists/c788c23e-1b46-4162-8f5e-d7585343b5de` against the live org. Created a release definition (ID=1) in project `21cff396-a36f-4d05-bccf-91e3a2a8b4bb`. POSTed an ACE with token `21cff396-a36f-4d05-bccf-91e3a2a8b4bb/1` → HTTP 200, valid ACE returned, confirmed in subsequent GET. Namespaced prefix token was not present.

## Implementation

`azuredevops/internal/service/permissions/resource_release_definition_permissions.go`:

```go
// Token format: {projectId}/{releaseDefinitionId}
// Confirmed live 2026-06-06.
func createReleaseDefinitionToken(d *schema.ResourceData, ...) (string, error) {
    // ... builds "{projectId}/{releaseDefinitionId}" directly
}
```

## Implication for future initiatives

Any initiative touching `ReleaseManagement2` permissions should use this token format directly — no probe needed. The confirmed format is simpler than the hypothesised one and requires no live API call to construct (pure function of projectId + defId).

## Sources

- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/work-items-snapshot/WI-1.md` (spike methodology)
- `brain/cycles/_raw/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions.md`
