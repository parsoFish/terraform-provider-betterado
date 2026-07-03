# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

All 4 ACs implemented in a single iteration:

1. **AC1** — Created `data_feature_flag_framework.go` implementing `datasource.DataSource` for `betterado_feature_flag`. Added `featuremanagement.NewFeatureFlagDataSource` to `framework_provider.go` DataSources(). `NewFeatureFlagResource` was already registered in Resources() from WI-4. Confirmed `azuredevops/provider.go` (SDKv2) has zero entries for betterado_feature_flag.

2. **AC2** — Created `examples/resources/betterado_feature_flag/resource.tf` and `examples/data-sources/betterado_feature_flag/data-source.tf`. Ran `make docs` which generated `docs/resources/feature_flag.md` and `docs/data-sources/feature_flag.md` (tfplugindocs strips provider prefix from filenames). The Makefile already includes `git checkout -- docs/guides/` to restore guides. All attributes documented.

3. **AC3** — `TestProvider_HasChildResources` and `TestProvider_HasChildDataSources` test the SDKv2 ResourcesMap/DataSourcesMap. Framework resources (betterado_feature_flag, release resources, task_group) are NOT in those maps, so no count change was needed. Tests pass as-is.

4. **AC4** — Added `### Added` section under `## [Unreleased]` in CHANGELOG.md. Bumped PROVIDER_VERSION.txt from `1.2.0` to `1.2.1`.

Also fixed pre-existing gofumpt issue in `resource_feature_flag_framework.go` (line 276 extra blank line) flagged by golangci-lint --new-from-rev=main.

## What worked

- `make docs` automatically runs `git checkout -- docs/guides/` at the end (Makefile has that step)
- tfplugindocs generates docs as `docs/resources/feature_flag.md` (strips provider prefix)
- Framework resources don't appear in SDKv2 ResourcesMap, so provider_test.go counts didn't need updating
- `gofumpt -w <file>` fixes gofumpt formatting issues

## What didn't work

_(none — first iteration completed everything)_

## Open questions

_(none)_

## Notes for reflection

_(none)_
