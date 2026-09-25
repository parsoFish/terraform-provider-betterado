# Profile corrections — brain/projects/terraform-provider-betterado/profile.md

> Corrections produced by INIT-2026-09-25-gap-registry-and-matrix-normalization (WI-5).
> Apply these patches to `brain/projects/terraform-provider-betterado/profile.md`
> after the initiative merges.

---

## Correction: API-coverage discipline stat line

**File:** brain/projects/terraform-provider-betterado/profile.md

The stat line under `## API-coverage discipline` cites a stale writable-gap count
from before the normalization run. The 31 normalized gap matrices show
release-definition now has 0 writable gaps open (all previously gap-open writable
fields were implemented in INIT-2026-06-17 / INIT-2026-06-19); the 25 remaining
deferred items are all read-only computed fields.

old:
```
The release_definition surface has a field-by-field gap matrix at
`docs/release-definition-gap-matrix.md` (93 mapped / 8 writable gaps open).
```

new:
```
The release_definition surface has a field-by-field gap matrix at
`docs/release-definition-gap-matrix.md` (132 fields mapped / 0 writable gaps open /
25 gap-deferred read-only computed fields). Gap-deferred items are all server-generated
read-only metadata (createdBy, modifiedBy, url, badgeUrl, etc.) with no declarative value.
```

---

## Correction: Net-new resources table

**File:** brain/projects/terraform-provider-betterado/profile.md

The net-new resources table predates the framework migration initiatives
(INIT-2026-06-19, INIT-2026-07-01). The file paths and notes need updating:
all net-new resources have been migrated from SDKv2 to terraform-plugin-framework
and the SDKv2 files deleted. The data source row for task_group was also missing.

old:
```
| `betterado_release_definition` (resource + data) | `release/resource_release_definition.go` (~1490 lines), `release/data_release_definition*.go` | vsrm |
| `betterado_release_folder` (resource + data) | `release/resource_release_folder.go`, `release/data_release_folder.go` | vsrm |
| `betterado_release_definition_permissions` | `release/` (permissions sub-package) | vsrm |
| `betterado_task_group` (resource + data) | `taskagent/resource_task_group.go`, `taskagent/data_task_group.go` | core |
| release definition history / revision / list data sources | `release/data_release_definition_{history,revision}.go`, `release/data_release_definitions.go` | vsrm |
```

new:
```
| `betterado_release_definition` (resource) | `release/resource_release_definition_framework.go` | vsrm |
| `betterado_release_definition` (data — by id/name) | `release/datasource_release_definition_framework.go` | vsrm |
| `betterado_release_definitions` (data — list) | `release/datasource_release_definitions_framework.go` | vsrm |
| `betterado_release_definition_history` (data) | `release/datasource_release_definition_history_framework.go` | vsrm |
| `betterado_release_definition_revision` (data) | `release/datasource_release_definition_revision_framework.go` | vsrm |
| `betterado_release_folder` (resource + data) | `release/resource_release_folder_framework.go`, `release/datasource_release_folder_framework.go` | vsrm |
| `betterado_release_definition_permissions` | `permissions/resource_release_definition_permissions_framework.go` | vsrm |
| `betterado_task_group` (resource) | `taskagent/resource_task_group_framework.go` | core |
| `betterado_task_group` (data) | `taskagent/data_task_group_framework.go` | core |
All files above are terraform-plugin-framework implementations (SDKv2 originals deleted).
```
