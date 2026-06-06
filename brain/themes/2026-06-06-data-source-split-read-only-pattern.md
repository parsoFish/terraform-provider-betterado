---
title: Read-only data source pattern — single-lookup vs list data sources in this provider
description: betterado data sources follow upstream's pattern of split single-lookup (data.betterado_release_definition) and list (data.betterado_release_definitions) data sources; each maps to a single SDK call and registers in provider.go's data source map.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T05:10:00Z
updated_at: 2026-06-06T05:10:00Z
related_themes:
  - 2026-05-18-stack-and-test-layout
---

# Read-only data source pattern — single-lookup vs list

## Pattern

This provider implements data sources as two separate resources per entity:

1. **Single-lookup** (`data.betterado_release_definition`) — resolves one entity by ID or name. Uses `GetReleaseDefinition` SDK call. Schema: `project_id` + `release_definition_id` (int, Optional/Computed) + `name` (string, Optional/Computed).
2. **List** (`data.betterado_release_definitions`) — returns a slice of entities with optional filters. Uses `GetReleaseDefinitions` SDK call. Schema: `project_id` + optional `path`/`name` filters + `release_definitions` list attribute.

Both are registered in `azuredevops/provider.go`'s `DataSourcesMap` alongside upstream data sources (e.g., `azuredevops_build_definition`).

## File layout

```
azuredevops/internal/service/release/
  data_release_definition.go       # single-lookup impl
  data_release_definition_test.go  # unit tests (gomock)
  data_release_definitions.go      # list impl
  data_release_definitions_test.go # unit tests (gomock)
```

Acceptance tests live in:
```
azuredevops/internal/acceptancetests/
  data_release_definition_test.go  # TestAccDataReleaseDefinition_ById, _ByName, TestAccDataReleaseDefinitions_List
```

## Quality gate split

- Dev-loop offline gate: `go test -run TestDataReleaseDefinition|TestDataReleaseDefinitions ./azuredevops/internal/service/release/`
- Pre-merge live gate: `TF_ACC=1 go test -run TestAccDataReleaseDefinition|TestAccDataReleaseDefinitions ./azuredevops/internal/acceptancetests/`

## Why two separate resources vs one with a `list` mode

Matches upstream `azuredevops` provider convention (e.g., `azuredevops_build_definition` vs `azuredevops_build_definitions`). Keeps Terraform plan output clean — one resource, one purpose.

## Sources

- `_logs/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources/work-items-snapshot/WI-1.md` through `WI-5.md`
- `brain/cycles/_raw/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources.md`
