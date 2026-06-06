---
title: betterado_release_definition_permissions security token format (confirmed)
description: The ReleaseManagement2 security-namespace token for a release definition is `{projectId}/{releaseDefinitionId}` — NO `ReleaseManagement2/Project/...` namespace prefix, identical in shape to the Build namespace. The WI-1 spike DISPROVED the manifest's original `ReleaseManagement2/Project/{projectId}/{definitionId}` hypothesis. release_definition_id is Optional (project-scope token = projectId alone; definition-scope = projectId/releaseDefinitionId).
category: reference
project: terraform-provider-betterado
created_at: '2026-06-06T00:00:00Z'
updated_at: '2026-06-06T00:00:00Z'
---

# Release-definition permissions: confirmed ADO security token format

## Confirmed format

For `betterado_release_definition_permissions` (security namespace **ReleaseManagement2**):

- **Definition-scope token:** `{projectId}/{releaseDefinitionId}`
- **Project-scope token:** `{projectId}` (alone)
- **No** `ReleaseManagement2/Project/...` namespace prefix — the token shape is
  identical to the Build namespace.

`release_definition_id` is therefore **Optional** in the resource schema: omit it
for a project-scoped permission, supply it for a definition-scoped one. (This
diverged from the initiative spec's `Required`, and the Optional behaviour is the
intended/correct one — it mirrors ADO's dual-scope permission model.)

## Why this is recorded

The INIT-2026-06-05-release-definition-permissions manifest carried a WRONG
hypothesis — `ReleaseManagement2/Project/{projectId}/{definitionId}` — which the
WI-1 token spike disproved live. An inline code comment in
`resource_release_definition_permissions.go` isn't discoverable during planning,
so a future permissions initiative could re-encode the same wrong guess. This theme
is the durable record so the planner uses the confirmed format.

## Open follow-up

The cycle delivered `TestAccReleaseDefinitionPermissions_SetPermissions` but NOT
`_UpdatePermissions` (the spec required both). The update path has its own code;
add the update-path acceptance test in a follow-up.
