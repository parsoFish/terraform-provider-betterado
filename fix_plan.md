# Fix Plan

> Checklist for WI-7. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the framework migration of graph + identity resources/data sources is complete WHEN make docs runs THEN docs/resources/group.md, docs/data-sources/group.md, docs/data-sources/groups.md, docs/data-sources/descriptor.md, docs/data-sources/storage_key.md, docs/data-sources/group_membership.md, docs/data-sources/user.md, docs/data-sources/users.md, docs/data-sources/service_principal.md, docs/data-sources/identity_group.md, docs/data-sources/identity_groups.md, docs/data-sources/identity_user.md are current; docs/guides/ is restored via git checkout -- docs/guides/
- [x] AC2: GIVEN the provider version has been bumped in PROVIDER_VERSION.txt WHEN cat PROVIDER_VERSION.txt THEN the version is higher than the version present at the start of this initiative
- [x] AC3: GIVEN CHANGELOG.md is read WHEN the ## Unreleased section is viewed THEN it lists all migrated graph and identity resources/data sources under a 'Migration' or 'Changed' heading with a one-line description per resource
- [x] AC4: GIVEN examples/ directory WHEN examples/data-sources/betterado_group/data-source.tf and similar exist THEN each migrated resource/data source has an example HCL file that the docs embed

## All ACs complete

- Prior commit 7afbf8a0 completed ACs 2-4 (examples, version bump 1.2.0→1.2.1, changelog, gate test).
- This iteration (commit 565618ed) ran `make docs` to regenerate the 13 docs pages from framework schemas and restored docs/guides/ via git checkout.
- All required files are present in git diff main...HEAD.
- make test: all pass. golangci-lint: 0 issues. make terrafmt-check: pass. go build: clean.
