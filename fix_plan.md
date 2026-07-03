# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN all four feed resources/data-sources have been migrated to framework (WI-2 through WI-5 complete) WHEN make docs runs (tfplugindocs) THEN docs/resources/feed.md, docs/resources/feed_permission.md, docs/resources/feed_retention_policy.md and docs/data-sources/feed.md are up-to-date; docs/guides/ is restored via git checkout -- docs/guides/
- [x] AC2: GIVEN the provider version and changelog WHEN the migration is complete and all four resources are verified live THEN PROVIDER_VERSION.txt is bumped (semver patch), CHANGELOG.md has a new ## Unreleased entry documenting the framework migration of betterado_feed, betterado_feed_permission, betterado_feed_retention_policy, and data.betterado_feed
- [x] AC3: GIVEN examples/resources/ and examples/data-sources/ directories WHEN docs generation runs THEN examples/resources/betterado_feed/resource.tf, examples/resources/betterado_feed_permission/resource.tf, examples/resources/betterado_feed_retention_policy/resource.tf, and examples/data-sources/betterado_feed/data-source.tf exist and contain valid non-trivial HCL
- [x] AC4: GIVEN CI-equivalent gate WHEN make test && golangci-lint run ./azuredevops/... && make terrafmt-check runs THEN all checks pass; the migrated code is golangci-clean for changed files; docs HCL is terrafmt-clean
