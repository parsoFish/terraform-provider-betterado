---
title: Use path as Terraform resource ID for ADO folder resources
description: ADO folder APIs address folders by (project_id, path) — the numeric Folder.Id is not stable or returned consistently. Use path as the Terraform resource ID and set path ForceNew to keep updates simple.
category: decision
project: terraform-provider-betterado
created_at: 2026-06-06T02:36:29Z
updated_at: 2026-06-06T02:36:29Z
related_themes:
  - 2026-05-31-forge-onboarding-findings
---

# Use path as Terraform resource ID for ADO folder resources

## Decision

For `betterado_release_folder` (and any future ADO folder-like resources), use `path`
as the Terraform resource ID — not `Folder.Id`.

## Why

The ADO `/release/folders` API does not expose a stable, globally-unique numeric ID
that survives folder operations. Folders are addressed by `(projectId, path)` in all
CRUD operations:

- `CreateFolder` POST → returns a `Folder` struct; the `Id` field is not guaranteed
  stable across rename/recreate cycles.
- `GetFolders` → takes a `path` filter to locate a specific folder.
- `UpdateFolder` → takes `path` to identify which folder to update.
- `DeleteFolder` → takes `path`.

Using `path` as the resource ID means:
- Read uses `path` from `d.Id()` to call `GetFolders` with filter.
- A 404 / empty result → `d.SetId("")` → Terraform marks as deleted.
- No dependency on a numeric ID that might change.

## ForceNew on path

Set `path: ForceNew: true` in the schema. This means:

- Folder path changes trigger destroy + recreate (the safe, simple path).
- In-place updates only change `description`.

If path-rename support is needed in future (call `UpdateFolder` with a new path),
that can be added as an explicit in-place update — but the complexity cost is not
worth it for initial implementation.

## Applied in

`azuredevops/internal/service/release/resource_release_folder.go`:
- Schema: `path` is Required, ForceNew.
- `resourceReleaseFolderCreate` sets `d.SetId(path)`.
- `resourceReleaseFolderRead` uses `d.Id()` as the path filter to `GetFolders`.

## Sources

- `_logs/2026-06-06T02-00-02_INIT-2026-06-05-release-folder/work-items-snapshot/WI-1.md`
  (Implementation guide section — Create and Read description)
- `brain/cycles/_raw/2026-06-06T02-00-02_INIT-2026-06-05-release-folder.md`
