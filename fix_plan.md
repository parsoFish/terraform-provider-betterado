# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_dashboard and betterado_extension have both been migrated to framework (WI-1 and WI-2 complete) WHEN WI-3 implementation runs THEN CHANGELOG.md has a new entry under '## Unreleased' that documents the framework migration of betterado_dashboard and betterado_extension
- [x] AC2: GIVEN PROVIDER_VERSION.txt exists with the current semver WHEN WI-3 implementation runs THEN PROVIDER_VERSION.txt has its patch version incremented by 1 (e.g. 0.x.y → 0.x.(y+1)) to reflect two user-visible resource migrations
- [x] AC3: GIVEN changed files WHEN CI-equivalent gate runs (make test) THEN make test exits 0 (gofmt + whole-module go test without TF_ACC)
- [x] GATE: go test -count=1 ./azuredevops/internal/service/dashboard/... now finds 16 tests (resource_dashboard_framework_test.go) and exits 0
