# Fix Plan

> Checklist for WI-11. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN all taskagent resources and data sources have been migrated to framework WHEN make docs is run THEN docs/ directory is regenerated (run git checkout -- docs/guides/ to preserve hand-written guides); docs/taskagent-gap-matrix.md is up to date
- [x] AC2: GIVEN the migration is complete WHEN CHANGELOG.md is inspected THEN an '## Unreleased' entry exists documenting migration of all taskagent resources/data-sources to terraform-plugin-framework
- [x] AC3: GIVEN the provider ships a user-visible change (all taskagent types now framework) WHEN PROVIDER_VERSION.txt is inspected THEN version is bumped by one minor semver increment from the pre-initiative value
- [x] AC4: GIVEN provider_test.go counts are all correct after migration WHEN TestProvider_HasChildResources and TestProvider_HasChildDataSources run THEN both tests pass with the updated counts reflecting all taskagent types removed from SDKv2

## Notes

- All ACs completed in iteration 0 (first pass)
- `make test` passes (no FAIL)
- `TestProvider_HasChildResources` and `TestProvider_HasChildDataSources` both pass
- PROVIDER_VERSION.txt bumped: 1.2.0 → 1.3.0
- CHANGELOG.md: added `### Changed` summary section under `## [Unreleased]`
- docs/taskagent-gap-matrix.md: updated section headings from azuredevops_* to betterado_*; added COMPLETE migration status banner
- `make docs` ran; Makefile automatically runs `git checkout -- docs/guides/` at end
